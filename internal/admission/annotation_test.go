package admission

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnnotationChecker_ShouldMutate(t *testing.T) {
	key := "test.io/ndots"

	tests := []struct {
		name        string
		mode        string
		annotations map[string]string
		want        bool
	}{
		// Always mode
		{"always - no annotations", "always", nil, true},
		{"always - true annotation", "always", map[string]string{key: "true"}, true},
		{"always - false annotation", "always", map[string]string{key: "false"}, true},

		// Opt-in mode
		{"opt-in - no annotations", "opt-in", nil, false},
		{"opt-in - true annotation", "opt-in", map[string]string{key: "true"}, true},
		{"opt-in - false annotation", "opt-in", map[string]string{key: "false"}, false},
		{"opt-in - other annotation", "opt-in", map[string]string{key: "foo"}, false},

		// Opt-out mode
		{"opt-out - no annotations", "opt-out", nil, true},
		{"opt-out - true annotation", "opt-out", map[string]string{key: "true"}, true},
		{"opt-out - false annotation", "opt-out", map[string]string{key: "false"}, false},
		{"opt-out - other annotation", "opt-out", map[string]string{key: "foo"}, true},

		// Case insensitivity checks for mode
		{"mode case insensitive", "OPT-OUT", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := NewAnnotationChecker(key, tt.mode)
			got := checker.ShouldMutate(tt.annotations)
			assert.Equal(t, tt.want, got)
		})
	}
}
