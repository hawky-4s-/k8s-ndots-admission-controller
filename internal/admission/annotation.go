package admission

import (
	"strconv"
	"strings"
)

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
// This is kept for backward compatibility or simple checks.
func (c *AnnotationChecker) ShouldMutate(annotations map[string]string) bool {
	switch c.mode {
	case ModeAlways:
		return true
	case ModeOptIn:
		if annotations == nil {
			return false
		}
		// Use ParseBool for robustness
		if b, err := strconv.ParseBool(annotations[c.key]); err == nil {
			return b
		}
		return false
	case ModeOptOut:
		if annotations == nil {
			return true
		}
		// If explicit "false", return false. Else true.
		if b, err := strconv.ParseBool(annotations[c.key]); err == nil {
			return b
		}
		return true
	default:
		return true // Default to always behavior
	}
}

// Evaluate determines if mutation is required based on Pod and Namespace annotations,
// considering precedence rules (Pod > Namespace > Default).
func (c *AnnotationChecker) Evaluate(podAnnotations, nsAnnotations map[string]string) bool {
	if c.mode == ModeAlways {
		return true
	}

	// Helper to check explicit status
	// Returns: 1 (True), -1 (False), 0 (Neutral/Unset)
	checkExplicit := func(anns map[string]string) int {
		if anns == nil {
			return 0
		}
		val, ok := anns[c.key]
		if !ok {
			return 0
		}

		if b, err := strconv.ParseBool(val); err == nil {
			if b {
				return 1
			}
			return -1
		}
		return 0 // Value set but not a boolean? Treat as unset/neutral.
	}

	// 1. Check Pod Annotations
	podRes := checkExplicit(podAnnotations)
	if podRes == 1 {
		return true
	}
	if podRes == -1 {
		return false
	}

	// 2. Check Namespace Annotations
	nsRes := checkExplicit(nsAnnotations)
	if nsRes == 1 {
		return true
	}
	if nsRes == -1 {
		return false
	}

	// 3. Defaults
	switch c.mode {
	case ModeOptIn:
		return false
	case ModeOptOut:
		return true
	default:
		return true
	}
}
