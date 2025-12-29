package admission

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

type Handler struct {
	mutator PodMutator
	logger  *slog.Logger
	metrics MetricsRecorder
}

func NewHandler(mutator PodMutator, logger *slog.Logger) *Handler {
	return &Handler{
		mutator: mutator,
		logger:  logger,
	}
}

// NewHandlerWithMetrics creates a new handler with metrics recording.
func NewHandlerWithMetrics(mutator PodMutator, logger *slog.Logger, metrics MetricsRecorder) *Handler {
	return &Handler{
		mutator: mutator,
		logger:  logger,
		metrics: metrics,
	}
}

var (
	scheme       = runtime.NewScheme()
	codecs       = serializer.NewCodecFactory(scheme)
	deserializer = codecs.UniversalDeserializer()
)

// HandleMutate handles the admission review request.
func (h *Handler) HandleMutate(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		if h.metrics != nil {
			h.metrics.ObserveRequestDuration(time.Since(start).Seconds())
		}
	}()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("failed to read body", "error", err)
		h.recordError("read")
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		h.logger.Error("empty body")
		h.recordError("read")
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	// Verify content type? Usually application/json.
	// For now, assume json/yaml parsable by deserializer.

	var admissionReview admissionv1.AdmissionReview
	if _, _, err := deserializer.Decode(body, nil, &admissionReview); err != nil {
		h.logger.Error("failed to decode admission review", "error", err)
		h.recordError("decode")
		http.Error(w, "failed to decode admission review", http.StatusBadRequest)
		return
	}

	if admissionReview.Request == nil {
		h.logger.Error("admission review request is nil")
		h.recordError("decode")
		http.Error(w, "admission review request is nil", http.StatusBadRequest)
		return
	}

	response := h.mutate(admissionReview.Request)
	admissionReview.Response = response
	admissionReview.Response.UID = admissionReview.Request.UID

	respBytes, err := json.Marshal(admissionReview)
	if err != nil {
		h.logger.Error("failed to marshal response", "error", err)
		h.recordError("marshal")
		http.Error(w, "failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(respBytes)
}

// recordError safely records an error if metrics is configured.
func (h *Handler) recordError(errorType string) {
	if h.metrics != nil {
		h.metrics.RecordError(errorType)
	}
}

// recordMutation safely records a mutation if metrics is configured.
func (h *Handler) recordMutation(namespace, action string) {
	if h.metrics != nil {
		h.metrics.RecordMutation(namespace, action)
	}
}

// Internal helper for logic
func (h *Handler) mutate(req *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
	// Only handle Pod resources
	if req.Kind.Kind != "Pod" {
		return &admissionv1.AdmissionResponse{
			Allowed: true,
		}
	}

	var pod corev1.Pod
	if err := json.Unmarshal(req.Object.Raw, &pod); err != nil {
		h.recordError("decode")
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: fmt.Sprintf("failed to decode pod: %v", err),
			},
		}
	}

	namespace := req.Namespace
	if namespace == "" {
		namespace = pod.Namespace
	}

	patch, err := h.mutator.Mutate(&pod)
	if err != nil {
		h.logger.Error("mutation failed", "error", err)
		h.recordError("mutation")
		// Fail open or closed? Plan said fail open usually, but let's allow it with error log
		return &admissionv1.AdmissionResponse{
			Allowed: true,
		}
	}

	if patch == nil {
		h.logger.Info("skipped mutation",
			"namespace", namespace,
			"name", pod.Name,
			"reason", "no changes needed",
		)
		h.recordMutation(namespace, "skipped")
		return &admissionv1.AdmissionResponse{
			Allowed: true,
		}
	}

	patchBytes, err := json.Marshal(patch)
	if err != nil {
		h.logger.Error("failed to marshal patch", "error", err)
		h.recordError("marshal")
		return &admissionv1.AdmissionResponse{
			Allowed: true,
		}
	}

	h.logger.Info("mutated pod",
		"namespace", namespace,
		"name", pod.Name,
	)
	h.recordMutation(namespace, "mutated")

	patchType := admissionv1.PatchTypeJSONPatch
	return &admissionv1.AdmissionResponse{
		Allowed:   true,
		Patch:     patchBytes,
		PatchType: &patchType,
	}
}
