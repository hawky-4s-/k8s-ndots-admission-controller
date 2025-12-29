package admission

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNamespaceFilter_ShouldMutate(t *testing.T) {
	logger := slog.Default()

	tests := []struct {
		name      string
		include   []string
		exclude   []string
		namespace string
		want      bool
	}{
		// Exclude priority
		{"exclude matches -> skip", nil, []string{"kube-system"}, "kube-system", false},
		{"exclude matches (with include set) -> skip", []string{"kube-system"}, []string{"kube-system"}, "kube-system", false},

		// Include logic
		{"include set, match -> mutate", []string{"prod"}, nil, "prod", true},
		{"include set, no match -> skip", []string{"prod"}, nil, "dev", false},

		// Default (no lists)
		{"no lists -> mutate", nil, nil, "any", true},

		// Include empty (treat as all allowed except exclude)
		{"include empty, exclude no match -> mutate", nil, []string{"kube-system"}, "prod", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewNamespaceFilter(tt.include, tt.exclude, logger)
			got := f.ShouldMutate(tt.namespace)
			assert.Equal(t, tt.want, got)
		})
	}
}
