package admission

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/hawky-4s-/k8s-ndots-admission-controller/internal/config"
)

func TestMutator_Mutate(t *testing.T) {
	logger := slog.Default()
	cfg := &config.Config{NdotsValue: 2, AnnotationKey: "change-ndots", AnnotationMode: "opt-out"}

	ndotsTwo := "2"
	ndotsFive := "5"

	tests := []struct {
		name           string
		pod            *corev1.Pod
		ns             *corev1.Namespace
		wantPatchLen   int
		wantOp         string
		wantPath       string
		wantValCheck   func(*testing.T, interface{})
	}{
		{
			name:           "no dnsConfig -> add full config",
			pod:            &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"}, Spec: corev1.PodSpec{}},
			ns:             &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
			wantPatchLen:   1,
			wantOp:         "add",
			wantPath:       "/spec/dnsConfig",
			wantValCheck: func(t *testing.T, res interface{}) {
				val, ok := res.(map[string]interface{})
				require.True(t, ok)
				opts, ok := val["options"].([]map[string]interface{})
				require.True(t, ok)
				require.Len(t, opts, 1)
				assert.Equal(t, "ndots", opts[0]["name"])
				assert.Equal(t, "2", opts[0]["value"])
			},
		},
		{
			name: "Namespace annotation disables (Opt-out mode)",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"},
				Spec: corev1.PodSpec{DNSConfig: &corev1.PodDNSConfig{}},
			},
			ns: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "default",
					Annotations: map[string]string{"change-ndots": "false"},
				},
			},
			wantPatchLen: 0,
		},
		{
			name: "Pod annotation overrides NS annotation (Pod:True, NS:False) -> Mutate",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "pod1",
					Namespace:   "default",
					Annotations: map[string]string{"change-ndots": "true"},
				},
				Spec: corev1.PodSpec{DNSConfig: &corev1.PodDNSConfig{}},
			},
			ns: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "default",
					Annotations: map[string]string{"change-ndots": "false"},
				},
			},
			wantPatchLen: 1,
			wantOp:       "add",
			wantPath:     "/spec/dnsConfig/options",
		},
		{
			name: "Pod annotation overrides NS annotation (Pod:False, NS:True) -> Skip",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "pod1",
					Namespace:   "default",
					Annotations: map[string]string{"change-ndots": "false"},
				},
				Spec: corev1.PodSpec{DNSConfig: &corev1.PodDNSConfig{}},
			},
			ns: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "default",
					Annotations: map[string]string{"change-ndots": "true"},
				},
			},
			wantPatchLen: 0,
		},
		{
			name: "empty options -> add options array",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"},
				Spec: corev1.PodSpec{
					DNSConfig: &corev1.PodDNSConfig{},
				},
			},
			ns: &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
			wantPatchLen: 1,
			wantOp:       "add",
			wantPath:     "/spec/dnsConfig/options",
			wantValCheck: func(t *testing.T, res interface{}) {
				val, ok := res.([]map[string]interface{})
				require.True(t, ok)
				require.Len(t, val, 1)
				assert.Equal(t, "ndots", val[0]["name"])
				assert.Equal(t, "2", val[0]["value"])
			},
		},
		{
			name: "options without ndots -> append ndots",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"},
				Spec: corev1.PodSpec{
					DNSConfig: &corev1.PodDNSConfig{
						Options: []corev1.PodDNSConfigOption{
							{Name: "timeout", Value: nil},
						},
					},
				},
			},
			ns: &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
			wantPatchLen: 1,
			wantOp:       "add",
			wantPath:     "/spec/dnsConfig/options/-",
			wantValCheck: func(t *testing.T, res interface{}) {
				val, ok := res.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "ndots", val["name"])
				assert.Equal(t, "2", val["value"])
			},
		},
		{
			name: "ndots with wrong value -> replace",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"},
				Spec: corev1.PodSpec{
					DNSConfig: &corev1.PodDNSConfig{
						Options: []corev1.PodDNSConfigOption{
							{Name: "ndots", Value: &ndotsFive},
						},
					},
				},
			},
			ns: &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
			wantPatchLen: 1,
			wantOp:       "replace",
			wantPath:     "/spec/dnsConfig/options/0/value",
			wantValCheck: func(t *testing.T, res interface{}) {
				val, ok := res.(string)
				require.True(t, ok)
				assert.Equal(t, "2", val)
			},
		},
		{
			name: "ndots correct -> no patch",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"},
				Spec: corev1.PodSpec{
					DNSConfig: &corev1.PodDNSConfig{
						Options: []corev1.PodDNSConfigOption{
							{Name: "ndots", Value: &ndotsTwo},
						},
					},
				},
			},
			ns: &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
			wantPatchLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := fake.NewSimpleClientset()
			if tt.ns != nil {
				_, err := client.CoreV1().Namespaces().Create(context.Background(), tt.ns, metav1.CreateOptions{})
				require.NoError(t, err)
			}

			mutator := NewMutator(cfg, logger, client)

			patches, err := mutator.Mutate(tt.pod)
			require.NoError(t, err)
			require.Len(t, patches, tt.wantPatchLen)

			if tt.wantPatchLen > 0 {
				patch := patches[0]
				assert.Equal(t, tt.wantOp, patch.Op)
				assert.Equal(t, tt.wantPath, patch.Path)
				if tt.wantValCheck != nil {
					tt.wantValCheck(t, patch.Value)
				}
			}
		})
	}
}
