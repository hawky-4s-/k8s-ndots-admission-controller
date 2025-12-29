# Task 2.2: Annotation Mode Support

**Phase**: 2 - Mutation Logic  
**Estimate**: 2-3 hours  
**Dependencies**: Task 2.1

## Objective

Implement annotation-based mutation control with three modes: `always`, `opt-in`, and `opt-out`.

## Deliverables

- [ ] Annotation parsing logic
- [ ] Mode-based mutation decision
- [ ] Configurable annotation key
- [ ] Integration with Mutator

## Implementation Details

### Annotation Modes

| Mode | Behavior |
|------|----------|
| `always` | Mutate all pods regardless of annotation |
| `opt-in` | Mutate only when annotation == `"true"` |
| `opt-out` | Mutate all pods unless annotation == `"false"` |

### Configuration Extension

```go
type Config struct {
    // ... existing fields
    AnnotationKey  string // default: "change-ndots"
    AnnotationMode string // "always", "opt-in", "opt-out"
}

type AnnotationMode string

const (
    ModeAlways AnnotationMode = "always"
    ModeOptIn  AnnotationMode = "opt-in"
    ModeOptOut AnnotationMode = "opt-out"
)
```

### Annotation Checker

```go
// internal/admission/annotation.go

package admission

// AnnotationChecker determines if a pod should be mutated based on annotations
type AnnotationChecker struct {
    annotationKey string
    mode          AnnotationMode
}

func NewAnnotationChecker(key string, mode AnnotationMode) *AnnotationChecker {
    return &AnnotationChecker{
        annotationKey: key,
        mode:          mode,
    }
}

// ShouldMutate returns true if the pod should be mutated based on annotation mode
func (c *AnnotationChecker) ShouldMutate(pod *corev1.Pod) bool {
    switch c.mode {
    case ModeAlways:
        return true
        
    case ModeOptIn:
        return c.hasAnnotation(pod) && c.getAnnotationValue(pod) == "true"
        
    case ModeOptOut:
        if !c.hasAnnotation(pod) {
            return true // No annotation = mutate
        }
        return c.getAnnotationValue(pod) != "false"
        
    default:
        return true // Default to always
    }
}

func (c *AnnotationChecker) hasAnnotation(pod *corev1.Pod) bool {
    if pod.Annotations == nil {
        return false
    }
    _, exists := pod.Annotations[c.annotationKey]
    return exists
}

func (c *AnnotationChecker) getAnnotationValue(pod *corev1.Pod) string {
    if pod.Annotations == nil {
        return ""
    }
    return pod.Annotations[c.annotationKey]
}
```

### Updated Mutator

```go
type Mutator struct {
    ndotsValue        string
    annotationChecker *AnnotationChecker
    logger            *slog.Logger
}

func (m *Mutator) Mutate(pod *corev1.Pod) ([]PatchOperation, error) {
    // Check annotation-based skip
    if !m.annotationChecker.ShouldMutate(pod) {
        m.logger.Debug("skipping mutation due to annotation",
            "namespace", pod.Namespace,
            "name", pod.Name,
            "annotation", m.annotationChecker.annotationKey,
            "mode", m.annotationChecker.mode,
        )
        return nil, nil
    }
    
    // ... rest of mutation logic
}
```

### Mode Decision Matrix

| Mode | No Annotation | `"true"` | `"false"` | Other Value |
|------|--------------|----------|-----------|-------------|
| `always` | Mutate | Mutate | Mutate | Mutate |
| `opt-in` | Skip | Mutate | Skip | Skip |
| `opt-out` | Mutate | Mutate | Skip | Mutate |

### Example Pod Annotations

```yaml
# Opt-in: Request mutation
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
  annotations:
    change-ndots: "true"
spec:
  containers:
    - name: app
      image: nginx
---
# Opt-out: Prevent mutation
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
  annotations:
    change-ndots: "false"
spec:
  containers:
    - name: app
      image: nginx
```

### Custom Annotation Key

```yaml
# Helm values.yaml
annotationKey: "my-company.io/ndots-mutation"
annotationMode: "opt-out"
```

```yaml
# Pod using custom annotation
metadata:
  annotations:
    my-company.io/ndots-mutation: "false"
```

## Acceptance Criteria

- [ ] `always` mode mutates all pods regardless of annotations
- [ ] `opt-in` mode only mutates pods with annotation = "true"
- [ ] `opt-out` mode skips pods with annotation = "false"
- [ ] Annotation key is configurable
- [ ] Invalid mode values default to `always`
- [ ] Annotation values are case-sensitive ("true" vs "TRUE")
- [ ] Logging indicates skip reason

## Testing

```go
func TestAnnotationChecker_ShouldMutate_AlwaysMode(t *testing.T) {
    tests := []struct {
        name        string
        annotations map[string]string
        want        bool
    }{
        {"no annotation", nil, true},
        {"annotation true", map[string]string{"change-ndots": "true"}, true},
        {"annotation false", map[string]string{"change-ndots": "false"}, true},
    }
    // ... test implementation
}

func TestAnnotationChecker_ShouldMutate_OptInMode(t *testing.T) {
    tests := []struct {
        name        string
        annotations map[string]string
        want        bool
    }{
        {"no annotation", nil, false},
        {"annotation true", map[string]string{"change-ndots": "true"}, true},
        {"annotation false", map[string]string{"change-ndots": "false"}, false},
        {"annotation other", map[string]string{"change-ndots": "maybe"}, false},
    }
    // ... test implementation
}

func TestAnnotationChecker_ShouldMutate_OptOutMode(t *testing.T) {
    tests := []struct {
        name        string
        annotations map[string]string
        want        bool
    }{
        {"no annotation", nil, true},
        {"annotation true", map[string]string{"change-ndots": "true"}, true},
        {"annotation false", map[string]string{"change-ndots": "false"}, false},
        {"annotation other", map[string]string{"change-ndots": "maybe"}, true},
    }
    // ... test implementation
}

func TestAnnotationChecker_CustomKey(t *testing.T) {
    // Test with custom annotation key
}
```

## Notes

- Annotations are inherited from Pod templates in Deployments/StatefulSets
- Consider documenting annotation in README with examples
- Boolean string comparison should be exact ("true"/"false")
