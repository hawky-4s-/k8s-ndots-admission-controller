package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRecorder(t *testing.T) {
	reg := prometheus.NewRegistry()
	recorder := NewRecorder(reg)

	require.NotNil(t, recorder)
}

func TestRecorder_RecordMutation(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		action    string
	}{
		{
			name:      "mutated action",
			namespace: "default",
			action:    "mutated",
		},
		{
			name:      "skipped action",
			namespace: "kube-system",
			action:    "skipped",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := prometheus.NewRegistry()
			recorder := NewRecorder(reg)

			recorder.RecordMutation(tt.namespace, tt.action)

			count := testutil.ToFloat64(recorder.mutationsTotal.WithLabelValues(tt.namespace, tt.action))
			assert.Equal(t, float64(1), count)
		})
	}
}

func TestRecorder_RecordError(t *testing.T) {
	tests := []struct {
		name      string
		errorType string
	}{
		{"decode error", "decode"},
		{"mutation error", "mutation"},
		{"marshal error", "marshal"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := prometheus.NewRegistry()
			recorder := NewRecorder(reg)

			recorder.RecordError(tt.errorType)

			count := testutil.ToFloat64(recorder.errorsTotal.WithLabelValues(tt.errorType))
			assert.Equal(t, float64(1), count)
		})
	}
}

func TestRecorder_ObserveRequestDuration(t *testing.T) {
	reg := prometheus.NewRegistry()
	recorder := NewRecorder(reg)

	recorder.ObserveRequestDuration(0.123)

	// Verify histogram has an observation
	count := testutil.CollectAndCount(recorder.requestDuration)
	assert.Equal(t, 1, count)
}

func TestRecorder_MultipleRecordings(t *testing.T) {
	reg := prometheus.NewRegistry()
	recorder := NewRecorder(reg)

	// Record multiple mutations
	recorder.RecordMutation("default", "mutated")
	recorder.RecordMutation("default", "mutated")
	recorder.RecordMutation("prod", "mutated")
	recorder.RecordMutation("default", "skipped")

	// Verify counts
	assert.Equal(t, float64(2), testutil.ToFloat64(recorder.mutationsTotal.WithLabelValues("default", "mutated")))
	assert.Equal(t, float64(1), testutil.ToFloat64(recorder.mutationsTotal.WithLabelValues("prod", "mutated")))
	assert.Equal(t, float64(1), testutil.ToFloat64(recorder.mutationsTotal.WithLabelValues("default", "skipped")))
}
