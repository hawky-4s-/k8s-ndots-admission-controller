package admission

import "strings"

type AnnotationMode string

const (
	ModeAlways AnnotationMode = "always"
	ModeOptIn  AnnotationMode = "opt-in"
	ModeOptOut AnnotationMode = "opt-out"
)

type AnnotationChecker struct {
	key  string
	mode AnnotationMode
}

func NewAnnotationChecker(key string, mode string) *AnnotationChecker {
	return &AnnotationChecker{
		key:  key,
		mode: AnnotationMode(strings.ToLower(mode)),
	}
}

// ShouldMutate determines if mutation is required based on annotations.
func (c *AnnotationChecker) ShouldMutate(annotations map[string]string) bool {
	switch c.mode {
	case ModeAlways:
		return true
	case ModeOptIn:
		if annotations == nil {
			return false
		}
		return annotations[c.key] == "true"
	case ModeOptOut:
		if annotations == nil {
			return true
		}
		return annotations[c.key] != "false"
	default:
		return true // Default to always behavior
	}
}
