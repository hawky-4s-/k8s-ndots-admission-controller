package admission

import (
	corev1 "k8s.io/api/core/v1"
)

// PatchOperation represents a JSON patch operation.
type PatchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

// PodMutator defines the interface for pod mutations.
type PodMutator interface {
	Mutate(pod *corev1.Pod) ([]PatchOperation, error)
}

// MetricsRecorder defines the interface for recording metrics.
type MetricsRecorder interface {
	RecordMutation(namespace, action string)
	RecordError(errorType string)
	ObserveRequestDuration(seconds float64)
}
