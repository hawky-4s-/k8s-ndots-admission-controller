package admission

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnnotationChecker_Evaluate(t *testing.T) {
	tests := []struct {
		name           string
		mode           string
		podAnn         map[string]string
		nsAnn          map[string]string
		expected       bool
	}{
		// Mode: Always
		{
			name:     "Always mode - no annotations",
			mode:     "always",
			podAnn:   nil,
			nsAnn:    nil,
			expected: true,
		},

		// Mode: Opt-In
		{
			name:     "Opt-in - Pod true, NS empty -> True",
			mode:     "opt-in",
			podAnn:   map[string]string{"change-ndots": "true"},
			nsAnn:    nil,
			expected: true,
		},
		{
			name:     "Opt-in - Pod TRUE (case insensitive), NS empty -> True",
			mode:     "opt-in",
			podAnn:   map[string]string{"change-ndots": "TRUE"},
			nsAnn:    nil,
			expected: true,
		},
		{
			name:     "Opt-in - Pod 1 (bool), NS empty -> True",
			mode:     "opt-in",
			podAnn:   map[string]string{"change-ndots": "1"},
			nsAnn:    nil,
			expected: true,
		},
		{
			name:     "Opt-in - Pod false, NS true -> False (Pod wins)",
			mode:     "opt-in",
			podAnn:   map[string]string{"change-ndots": "false"},
			nsAnn:    map[string]string{"change-ndots": "true"},
			expected: false,
		},
		{
			name:     "Opt-in - Pod empty, NS true -> True",
			mode:     "opt-in",
			podAnn:   nil,
			nsAnn:    map[string]string{"change-ndots": "true"},
			expected: true,
		},
		{
			name:     "Opt-in - Pod empty, NS false -> False",
			mode:     "opt-in",
			podAnn:   nil,
			nsAnn:    map[string]string{"change-ndots": "false"},
			expected: false,
		},
		{
			name:     "Opt-in - Both empty -> False",
			mode:     "opt-in",
			podAnn:   nil,
			nsAnn:    nil,
			expected: false,
		},

		// Mode: Opt-Out
		{
			name:     "Opt-out - Pod false, NS empty -> False",
			mode:     "opt-out",
			podAnn:   map[string]string{"change-ndots": "false"},
			nsAnn:    nil,
			expected: false,
		},
		{
			name:     "Opt-out - Pod FALSE (case insensitive), NS empty -> False",
			mode:     "opt-out",
			podAnn:   map[string]string{"change-ndots": "FALSE"},
			nsAnn:    nil,
			expected: false,
		},
		{
			name:     "Opt-out - Pod 0 (bool), NS empty -> False",
			mode:     "opt-out",
			podAnn:   map[string]string{"change-ndots": "0"},
			nsAnn:    nil,
			expected: false,
		},
		{
			name:     "Opt-out - Pod true, NS false -> True (Pod wins)",
			mode:     "opt-out",
			podAnn:   map[string]string{"change-ndots": "true"},
			nsAnn:    map[string]string{"change-ndots": "false"},
			expected: true,
		},
		{
			name:     "Opt-out - Pod empty, NS false -> False",
			mode:     "opt-out",
			podAnn:   nil,
			nsAnn:    map[string]string{"change-ndots": "false"},
			expected: false,
		},
		{
			name:     "Opt-out - Pod empty, NS true -> True",
			mode:     "opt-out",
			podAnn:   nil,
			nsAnn:    map[string]string{"change-ndots": "true"},
			expected: true,
		},
		{
			name:     "Opt-out - Both empty -> True",
			mode:     "opt-out",
			podAnn:   nil,
			nsAnn:    nil,
			expected: true,
		},

		// Edge cases
		{
			name:     "Invalid boolean string treats as neutral",
			mode:     "opt-in",
			podAnn:   map[string]string{"change-ndots": "invalid"},
			nsAnn:    map[string]string{"change-ndots": "true"},
			expected: true, // Pod "invalid" -> neutral, NS "true" -> True
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := NewAnnotationChecker("change-ndots", tt.mode)
			result := checker.Evaluate(tt.podAnn, tt.nsAnn)
			assert.Equal(t, tt.expected, result)
		})
	}
}
