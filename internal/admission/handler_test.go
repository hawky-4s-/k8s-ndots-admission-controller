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
