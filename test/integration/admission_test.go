//go:build integration

package integration

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"github.com/hawky-4s-/k8s-ndots-admission-controller/internal/admission"
	"github.com/hawky-4s-/k8s-ndots-admission-controller/internal/config"
	"github.com/hawky-4s-/k8s-ndots-admission-controller/internal/logging"
	"github.com/hawky-4s-/k8s-ndots-admission-controller/internal/metrics"
)

// TestIntegration_FullAdmissionFlow tests the complete admission flow
// from HTTP request to mutation response
func TestIntegration_FullAdmissionFlow(t *testing.T) {
	// Create a full stack with real components
	cfg := &config.Config{
		NdotsValue:     2,
		AnnotationKey:  "change-ndots",
		AnnotationMode: "opt-out",
	}

	logger := logging.NewLogger("info", "json", os.Stdout)
	reg := prometheus.NewRegistry()
	metricsRecorder := metrics.NewRecorder(reg)

	mutator := admission.NewMutator(cfg, logger)
	handler := admission.NewHandlerWithMetrics(mutator, logger, metricsRecorder)

	// Create test server
	mux := http.NewServeMux()
	mux.HandleFunc("/mutate", handler.HandleMutate)
	server := httptest.NewServer(mux)
	defer server.Close()

	tests := []struct {
		name        string
		pod         corev1.Pod
		namespace   string
		wantMutated bool
		wantNdots   string
	}{
		{
			name: "pod without dnsConfig gets ndots added",
			pod: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "default",
				},
				Spec: corev1.PodSpec{},
			},
			namespace:   "default",
			wantMutated: true,
			wantNdots:   "2",
		},
		{
			name: "pod with opt-out annotation is skipped",
			pod: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod-skip",
					Namespace: "default",
					Annotations: map[string]string{
						"change-ndots": "false",
					},
				},
				Spec: corev1.PodSpec{},
			},
			namespace:   "default",
			wantMutated: false,
		},
		{
			name: "pod with correct ndots is not mutated",
			pod: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod-correct",
					Namespace: "default",
				},
				Spec: corev1.PodSpec{
					DNSConfig: &corev1.PodDNSConfig{
						Options: []corev1.PodDNSConfigOption{
							{Name: "ndots", Value: stringPtr("2")},
						},
					},
				},
			},
			namespace:   "default",
			wantMutated: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create admission review
			podBytes, err := json.Marshal(tt.pod)
			require.NoError(t, err)

			review := admissionv1.AdmissionReview{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "admission.k8s.io/v1",
					Kind:       "AdmissionReview",
				},
				Request: &admissionv1.AdmissionRequest{
					UID:       "test-uid",
					Namespace: tt.namespace,
					Kind: metav1.GroupVersionKind{
						Group:   "",
						Version: "v1",
						Kind:    "Pod",
					},
					Object: runtime.RawExtension{
						Raw: podBytes,
					},
				},
			}

			reviewBytes, err := json.Marshal(review)
			require.NoError(t, err)

			// Send request
			resp, err := http.Post(server.URL+"/mutate", "application/json", bytes.NewReader(reviewBytes))
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			// Parse response
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			var responseReview admissionv1.AdmissionReview
			err = json.Unmarshal(body, &responseReview)
			require.NoError(t, err)

			assert.True(t, responseReview.Response.Allowed)

			if tt.wantMutated {
				assert.NotEmpty(t, responseReview.Response.Patch)
				assert.NotNil(t, responseReview.Response.PatchType)
			} else {
				assert.Empty(t, responseReview.Response.Patch)
			}
		})
	}
}

// TestIntegration_ConcurrentAdmissions tests handling of concurrent requests
func TestIntegration_ConcurrentAdmissions(t *testing.T) {
	cfg := &config.Config{
		NdotsValue:     1,
		AnnotationKey:  "change-ndots",
		AnnotationMode: "always",
	}

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	reg := prometheus.NewRegistry()
	metricsRecorder := metrics.NewRecorder(reg)

	mutator := admission.NewMutator(cfg, logger)
	handler := admission.NewHandlerWithMetrics(mutator, logger, metricsRecorder)

	mux := http.NewServeMux()
	mux.HandleFunc("/mutate", handler.HandleMutate)
	server := httptest.NewServer(mux)
	defer server.Close()

	// Send 10 concurrent requests
	numRequests := 10
	done := make(chan bool, numRequests)

	for i := 0; i < numRequests; i++ {
		go func(id int) {
			pod := corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "concurrent-pod",
					Namespace: "test-ns",
				},
			}
			podBytes, _ := json.Marshal(pod)
			review := admissionv1.AdmissionReview{
				Request: &admissionv1.AdmissionRequest{
					UID:       "test-uid",
					Namespace: "test-ns",
					Kind:      metav1.GroupVersionKind{Kind: "Pod"},
					Object:    runtime.RawExtension{Raw: podBytes},
				},
			}
			reviewBytes, _ := json.Marshal(review)

			resp, err := http.Post(server.URL+"/mutate", "application/json", bytes.NewReader(reviewBytes))
			if err != nil {
				t.Logf("Request %d failed: %v", id, err)
				done <- false
				return
			}
			defer resp.Body.Close()

			done <- resp.StatusCode == http.StatusOK
		}(i)
	}

	// Wait for all requests
	successCount := 0
	for i := 0; i < numRequests; i++ {
		if <-done {
			successCount++
		}
	}

	assert.Equal(t, numRequests, successCount, "All concurrent requests should succeed")
}

// TestIntegration_MetricsRecorded tests that metrics are properly recorded
func TestIntegration_MetricsRecorded(t *testing.T) {
	cfg := &config.Config{
		NdotsValue:     2,
		AnnotationKey:  "change-ndots",
		AnnotationMode: "opt-out",
	}

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	reg := prometheus.NewRegistry()
	metricsRecorder := metrics.NewRecorder(reg)

	mutator := admission.NewMutator(cfg, logger)
	handler := admission.NewHandlerWithMetrics(mutator, logger, metricsRecorder)

	mux := http.NewServeMux()
	mux.HandleFunc("/mutate", handler.HandleMutate)
	server := httptest.NewServer(mux)
	defer server.Close()

	// Make a request
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "test-pod", Namespace: "default"},
	}
	podBytes, _ := json.Marshal(pod)
	review := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Namespace: "default",
			Kind:      metav1.GroupVersionKind{Kind: "Pod"},
			Object:    runtime.RawExtension{Raw: podBytes},
		},
	}
	reviewBytes, _ := json.Marshal(review)

	_, err := http.Post(server.URL+"/mutate", "application/json", bytes.NewReader(reviewBytes))
	require.NoError(t, err)

	// Check metrics are recorded
	metricFamilies, err := reg.Gather()
	require.NoError(t, err)

	foundMutations := false
	foundDuration := false
	for _, mf := range metricFamilies {
		if mf.GetName() == "ndots_webhook_mutations_total" {
			foundMutations = true
		}
		if mf.GetName() == "ndots_webhook_request_duration_seconds" {
			foundDuration = true
		}
	}

	assert.True(t, foundMutations, "Mutations metric should be recorded")
	assert.True(t, foundDuration, "Duration metric should be recorded")
}

// TestIntegration_TLSServer tests the server with TLS (when certs are available)
func TestIntegration_TLSServer(t *testing.T) {
	// Skip if running in CI without certs
	if os.Getenv("CI") != "" {
		t.Skip("Skipping TLS test in CI environment")
	}

	cfg := &config.Config{
		NdotsValue:     2,
		AnnotationKey:  "change-ndots",
		AnnotationMode: "opt-out",
	}

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	mutator := admission.NewMutator(cfg, logger)
	handler := admission.NewHandler(mutator, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("/mutate", handler.HandleMutate)

	// Create TLS test server
	server := httptest.NewTLSServer(mux)
	defer server.Close()

	// Create client that accepts self-signed cert
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "test-pod", Namespace: "default"},
	}
	podBytes, _ := json.Marshal(pod)
	review := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Namespace: "default",
			Kind:      metav1.GroupVersionKind{Kind: "Pod"},
			Object:    runtime.RawExtension{Raw: podBytes},
		},
	}
	reviewBytes, _ := json.Marshal(review)

	resp, err := client.Post(server.URL+"/mutate", "application/json", bytes.NewReader(reviewBytes))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestIntegration_WorkloadTypes tests that pods created by different controllers are properly mutated
// This simulates how the Kubernetes API server sends AdmissionReview for pods with ownerReferences
func TestIntegration_WorkloadTypes(t *testing.T) {
	cfg := &config.Config{
		NdotsValue:     2,
		AnnotationKey:  "change-ndots",
		AnnotationMode: "opt-out",
	}

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	reg := prometheus.NewRegistry()
	metricsRecorder := metrics.NewRecorder(reg)

	mutator := admission.NewMutator(cfg, logger)
	handler := admission.NewHandlerWithMetrics(mutator, logger, metricsRecorder)

	mux := http.NewServeMux()
	mux.HandleFunc("/mutate", handler.HandleMutate)
	server := httptest.NewServer(mux)
	defer server.Close()

	tests := []struct {
		name            string
		workloadType    string
		ownerKind       string
		ownerAPIVer     string
		generatePodName string
		wantMutated     bool
	}{
		{
			name:            "pod created by Deployment (via ReplicaSet)",
			workloadType:    "Deployment",
			ownerKind:       "ReplicaSet",
			ownerAPIVer:     "apps/v1",
			generatePodName: "my-deployment-abc123-xyz",
			wantMutated:     true,
		},
		{
			name:            "pod created by StatefulSet",
			workloadType:    "StatefulSet",
			ownerKind:       "StatefulSet",
			ownerAPIVer:     "apps/v1",
			generatePodName: "my-statefulset-0",
			wantMutated:     true,
		},
		{
			name:            "pod created by DaemonSet",
			workloadType:    "DaemonSet",
			ownerKind:       "DaemonSet",
			ownerAPIVer:     "apps/v1",
			generatePodName: "my-daemonset-abc123",
			wantMutated:     true,
		},
		{
			name:            "pod created by Job",
			workloadType:    "Job",
			ownerKind:       "Job",
			ownerAPIVer:     "batch/v1",
			generatePodName: "my-job-xyz",
			wantMutated:     true,
		},
		{
			name:            "pod created by CronJob (via Job)",
			workloadType:    "CronJob",
			ownerKind:       "Job",
			ownerAPIVer:     "batch/v1",
			generatePodName: "my-cronjob-12345-xyz",
			wantMutated:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create pod with ownerReference simulating controller-created pod
			pod := corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      tt.generatePodName,
					Namespace: "default",
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: tt.ownerAPIVer,
							Kind:       tt.ownerKind,
							Name:       "my-" + tt.workloadType,
							UID:        "owner-uid-123",
							Controller: boolPtr(true),
						},
					},
					Labels: map[string]string{
						"app": "test-app",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "main",
							Image: "nginx:latest",
						},
					},
				},
			}

			podBytes, err := json.Marshal(pod)
			require.NoError(t, err)

			review := admissionv1.AdmissionReview{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "admission.k8s.io/v1",
					Kind:       "AdmissionReview",
				},
				Request: &admissionv1.AdmissionRequest{
					UID:       types.UID("test-uid-" + tt.workloadType),
					Namespace: "default",
					Kind: metav1.GroupVersionKind{
						Group:   "",
						Version: "v1",
						Kind:    "Pod",
					},
					Object: runtime.RawExtension{
						Raw: podBytes,
					},
				},
			}

			reviewBytes, err := json.Marshal(review)
			require.NoError(t, err)

			resp, err := http.Post(server.URL+"/mutate", "application/json", bytes.NewReader(reviewBytes))
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			var responseReview admissionv1.AdmissionReview
			err = json.Unmarshal(body, &responseReview)
			require.NoError(t, err)

			assert.True(t, responseReview.Response.Allowed, "Pod should be allowed")

			if tt.wantMutated {
				assert.NotEmpty(t, responseReview.Response.Patch, "Pod from %s should be mutated", tt.workloadType)

				// Verify patch contains dnsConfig
				var patchOps []map[string]interface{}
				err = json.Unmarshal(responseReview.Response.Patch, &patchOps)
				require.NoError(t, err)
				assert.NotEmpty(t, patchOps, "Patch should have operations")
			}
		})
	}
}

// TestIntegration_NamespaceExclusion tests that pods in excluded namespaces are not mutated
func TestIntegration_NamespaceExclusion(t *testing.T) {
	cfg := &config.Config{
		NdotsValue:       2,
		AnnotationKey:    "change-ndots",
		AnnotationMode:   "opt-out",
		NamespaceExclude: []string{"kube-system", "kube-public"},
	}

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	mutator := admission.NewMutator(cfg, logger)
	handler := admission.NewHandler(mutator, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("/mutate", handler.HandleMutate)
	server := httptest.NewServer(mux)
	defer server.Close()

	tests := []struct {
		name        string
		namespace   string
		wantMutated bool
	}{
		{
			name:        "default namespace - should mutate",
			namespace:   "default",
			wantMutated: true,
		},
		{
			name:        "kube-system - should NOT mutate",
			namespace:   "kube-system",
			wantMutated: false,
		},
		{
			name:        "kube-public - should NOT mutate",
			namespace:   "kube-public",
			wantMutated: false,
		},
		{
			name:        "custom namespace - should mutate",
			namespace:   "my-app",
			wantMutated: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pod := corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: tt.namespace,
				},
			}
			podBytes, _ := json.Marshal(pod)

			review := admissionv1.AdmissionReview{
				Request: &admissionv1.AdmissionRequest{
					UID:       "test-uid",
					Namespace: tt.namespace,
					Kind:      metav1.GroupVersionKind{Kind: "Pod"},
					Object:    runtime.RawExtension{Raw: podBytes},
				},
			}
			reviewBytes, _ := json.Marshal(review)

			resp, err := http.Post(server.URL+"/mutate", "application/json", bytes.NewReader(reviewBytes))
			require.NoError(t, err)
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			var responseReview admissionv1.AdmissionReview
			json.Unmarshal(body, &responseReview)

			assert.True(t, responseReview.Response.Allowed)

			if tt.wantMutated {
				assert.NotEmpty(t, responseReview.Response.Patch, "Namespace %s should be mutated", tt.namespace)
			} else {
				assert.Empty(t, responseReview.Response.Patch, "Namespace %s should NOT be mutated", tt.namespace)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
