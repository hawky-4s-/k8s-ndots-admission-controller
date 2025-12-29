package admission

import "log/slog"

type NamespaceFilter struct {
	include map[string]bool
	exclude map[string]bool
	logger  *slog.Logger
}

func NewNamespaceFilter(include, exclude []string, logger *slog.Logger) *NamespaceFilter {
	f := &NamespaceFilter{
		include: make(map[string]bool),
		exclude: make(map[string]bool),
		logger:  logger,
	}

	for _, ns := range include {
		f.include[ns] = true
	}

	for _, ns := range exclude {
		f.exclude[ns] = true
	}

	return f
}

// ShouldMutate returns true if the namespace should be mutated.
func (f *NamespaceFilter) ShouldMutate(namespace string) bool {
	// Exclude takes priority
	if f.exclude[namespace] {
		f.logger.Debug("namespace excluded", "namespace", namespace)
		return false
	}

	// If include list is set, namespace must be in it
	if len(f.include) > 0 {
		allowed := f.include[namespace]
		if !allowed {
			f.logger.Debug("namespace not in include list", "namespace", namespace)
		}
		return allowed
	}

	// Default: allow
	return true
}
