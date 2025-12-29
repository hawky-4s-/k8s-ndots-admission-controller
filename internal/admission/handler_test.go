package admission

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// MockMutator is a mock implementation of PodMutator
type MockMutator struct {
	mock.Mock
}

func (m *MockMutator) Mutate(pod *corev1.Pod) ([]PatchOperation, error) {
	args := m.Called(pod)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]PatchOperation), args.Error(1)
}

// MockMetricsRecorder is a mock implementation of MetricsRecorder
type MockMetricsRecorder struct {
	mock.Mock
}

func (m *MockMetricsRecorder) RecordMutation(namespace, action string) {
	m.Called(namespace, action)
}

func (m *MockMetricsRecorder) RecordError(errorType string) {
	m.Called(errorType)
}

func (m *MockMetricsRecorder) ObserveRequestDuration(seconds float64) {
	m.Called(seconds)
}

func TestHandler_HandleMutate(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockMutator)
		wantStatusCode int
		wantAllowed    bool
		wantPatch      bool
	}{
		{
			name:           "invalid body returns 400",
			requestBody:    "invalid-json",
			setupMock:      func(m *MockMutator) {},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "valid pod request calls mutator",
			requestBody: admissionv1.AdmissionReview{
				Request: &admissionv1.AdmissionRequest{
					UID: "test-uid",
					Kind: metav1.GroupVersionKind{
						Group:   "",
						Version: "v1",
						Kind:    "Pod",
					},
					Object: runtime.RawExtension{
						Raw: []byte(`{"apiVersion":"v1","kind":"Pod","metadata":{"name":"test"}}`),
					},
				},
			},
			setupMock: func(m *MockMutator) {
				m.On("Mutate", mock.AnythingOfType("*v1.Pod")).Return(
					[]PatchOperation{{Op: "add", Path: "/foo", Value: "bar"}},
					nil,
				)
			},
			wantStatusCode: http.StatusOK,
			wantAllowed:    true,
			wantPatch:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMutator := new(MockMutator)
			tt.setupMock(mockMutator)

			h := NewHandler(mockMutator, slog.Default())

			var body []byte
			if s, ok := tt.requestBody.(string); ok {
				body = []byte(s)
			} else {
				body, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest("POST", "/mutate", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			h.HandleMutate(w, req)

			resp := w.Result()
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)

			if resp.StatusCode == http.StatusOK {
				var review admissionv1.AdmissionReview
				err := json.NewDecoder(w.Body).Decode(&review)
				assert.NoError(t, err)
				assert.NotNil(t, review.Response)
				assert.Equal(t, tt.wantAllowed, review.Response.Allowed)

				if tt.wantPatch {
					assert.NotEmpty(t, review.Response.Patch)
					assert.Equal(t, admissionv1.PatchTypeJSONPatch, *review.Response.PatchType)
				}
			}

			mockMutator.AssertExpectations(t)
		})
	}
}

func TestNewHandlerWithMetrics(t *testing.T) {
	mockMutator := new(MockMutator)
	mockMetrics := new(MockMetricsRecorder)
	logger := slog.Default()

	h := NewHandlerWithMetrics(mockMutator, logger, mockMetrics)

	require.NotNil(t, h)
	assert.Equal(t, mockMutator, h.mutator)
	assert.Equal(t, logger, h.logger)
	assert.Equal(t, mockMetrics, h.metrics)
}

func TestHandler_HandleMutateWithMetrics(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMutator   func(*MockMutator)
		setupMetrics   func(*MockMetricsRecorder)
		wantStatusCode int
	}{
		{
			name:        "successful mutation records mutated metric",
			requestBody: createValidAdmissionReview("test-pod", "default"),
			setupMutator: func(m *MockMutator) {
				m.On("Mutate", mock.AnythingOfType("*v1.Pod")).Return(
					[]PatchOperation{{Op: "add", Path: "/spec/dnsConfig", Value: map[string]interface{}{}}},
					nil,
				)
			},
			setupMetrics: func(m *MockMetricsRecorder) {
				m.On("ObserveRequestDuration", mock.AnythingOfType("float64")).Once()
				m.On("RecordMutation", "default", "mutated").Once()
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:        "skipped mutation records skipped metric",
			requestBody: createValidAdmissionReview("test-pod", "default"),
			setupMutator: func(m *MockMutator) {
				m.On("Mutate", mock.AnythingOfType("*v1.Pod")).Return(nil, nil)
			},
			setupMetrics: func(m *MockMetricsRecorder) {
				m.On("ObserveRequestDuration", mock.AnythingOfType("float64")).Once()
				m.On("RecordMutation", "default", "skipped").Once()
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:         "decode error records error metric",
			requestBody:  "invalid-json",
			setupMutator: func(m *MockMutator) {},
			setupMetrics: func(m *MockMetricsRecorder) {
				m.On("ObserveRequestDuration", mock.AnythingOfType("float64")).Once()
				m.On("RecordError", "decode").Once()
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:         "empty body records read error",
			requestBody:  "",
			setupMutator: func(m *MockMutator) {},
			setupMetrics: func(m *MockMetricsRecorder) {
				m.On("ObserveRequestDuration", mock.AnythingOfType("float64")).Once()
				m.On("RecordError", "read").Once()
			},
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMutator := new(MockMutator)
			mockMetrics := new(MockMetricsRecorder)
			tt.setupMutator(mockMutator)
			tt.setupMetrics(mockMetrics)

			h := NewHandlerWithMetrics(mockMutator, slog.Default(), mockMetrics)

			var body []byte
			if s, ok := tt.requestBody.(string); ok {
				body = []byte(s)
			} else {
				body, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest("POST", "/mutate", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			h.HandleMutate(w, req)

			resp := w.Result()
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)

			mockMutator.AssertExpectations(t)
			mockMetrics.AssertExpectations(t)
		})
	}
}

func TestHandler_HandleNonPodResource(t *testing.T) {
	mockMutator := new(MockMutator)
	mockMetrics := new(MockMetricsRecorder)

	// Non-pod resources should be allowed without calling mutator
	review := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID: "test-uid",
			Kind: metav1.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "ConfigMap", // Not a Pod
			},
			Object: runtime.RawExtension{
				Raw: []byte(`{"apiVersion":"v1","kind":"ConfigMap","metadata":{"name":"test"}}`),
			},
		},
	}

	mockMetrics.On("ObserveRequestDuration", mock.AnythingOfType("float64")).Once()

	h := NewHandlerWithMetrics(mockMutator, slog.Default(), mockMetrics)

	body, _ := json.Marshal(review)
	req := httptest.NewRequest("POST", "/mutate", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.HandleMutate(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var respReview admissionv1.AdmissionReview
	err := json.NewDecoder(w.Body).Decode(&respReview)
	require.NoError(t, err)
	assert.True(t, respReview.Response.Allowed)
	assert.Empty(t, respReview.Response.Patch)

	// Mutator should NOT have been called
	mockMutator.AssertNotCalled(t, "Mutate", mock.Anything)
	mockMetrics.AssertExpectations(t)
}

// Helper function to create a valid AdmissionReview
func createValidAdmissionReview(name, namespace string) admissionv1.AdmissionReview {
	return admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Namespace: namespace,
			Kind: metav1.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "Pod",
			},
			Object: runtime.RawExtension{
				Raw: []byte(`{"apiVersion":"v1","kind":"Pod","metadata":{"name":"` + name + `","namespace":"` + namespace + `"}}`),
			},
		},
	}
}
