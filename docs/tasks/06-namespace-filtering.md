# Task 2.3: Namespace Filtering

**Phase**: 2 - Mutation Logic  
**Estimate**: 1-2 hours  
**Dependencies**: Task 2.1

## Objective

Implement namespace-based filtering to include or exclude specific namespaces from mutation, with default exclusion of critical system namespaces.

## Deliverables

- [ ] Namespace filter in `internal/admission/namespace.go`
- [ ] Include/exclude list support
- [ ] Default exclusion of `kube-system`
- [ ] Integration with Mutator

## Implementation Details

### Configuration Extension

```go
type Config struct {
    // ... existing fields
    NamespaceInclude []string // If set, only these namespaces are mutated
    NamespaceExclude []string // These namespaces are never mutated
}

// Default excluded namespaces
var DefaultExcludedNamespaces = []string{
    "kube-system",
    "kube-public",
    "kube-node-lease",
}
```

### Namespace Filter

```go
// internal/admission/namespace.go

package admission

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

// ShouldMutate returns true if the namespace should be mutated
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
```

### Updated Mutator

```go
type Mutator struct {
    ndotsValue        string
    annotationChecker *AnnotationChecker
    namespaceFilter   *NamespaceFilter
    logger            *slog.Logger
}

func (m *Mutator) Mutate(pod *corev1.Pod) ([]PatchOperation, error) {
    // Check namespace filter first
    if !m.namespaceFilter.ShouldMutate(pod.Namespace) {
        m.logger.Debug("skipping mutation due to namespace filter",
            "namespace", pod.Namespace,
            "name", pod.Name,
        )
        return nil, nil
    }
    
    // Check annotation-based skip
    if !m.annotationChecker.ShouldMutate(pod) {
        // ... existing annotation logic
    }
    
    // ... rest of mutation logic
}
```

### Filter Priority

1. **Exclude list** - Always checked first; if namespace is excluded, skip mutation
2. **Include list** - If non-empty, namespace must be in list
3. **Default** - If neither list applies, allow mutation

### Configuration Examples

#### Exclude Only (Recommended)

```yaml
# Helm values.yaml
namespaceExclude:
  - kube-system
  - kube-public
  - istio-system
  - monitoring
```

#### Include Only

```yaml
# Helm values.yaml
namespaceInclude:
  - production
  - staging
```

#### Combined (Include with Exceptions)

```yaml
# Only mutate production and staging, but never monitoring
namespaceInclude:
  - production
  - staging
namespaceExclude:
  - monitoring  # Exception within included namespaces
```

### Webhook Configuration Integration

The namespace filter works at the application level. For improved efficiency, also configure the webhook itself:

```yaml
# MutatingWebhookConfiguration
webhooks:
  - name: ndots.admission.k8s.io
    namespaceSelector:
      matchExpressions:
        - key: kubernetes.io/metadata.name
          operator: NotIn
          values:
            - kube-system
            - kube-public
```

> **Note**: Application-level filtering provides more flexibility; webhook-level filtering improves performance by not even calling the webhook.

## Acceptance Criteria

- [ ] Pods in excluded namespaces are never mutated
- [ ] Include list restricts mutation to only listed namespaces
- [ ] Exclude takes priority over include
- [ ] Default excludes `kube-system`, `kube-public`, `kube-node-lease`
- [ ] Empty include list means "all namespaces" (minus excludes)
- [ ] Logging indicates namespace-based skip

## Testing

```go
func TestNamespaceFilter_Exclude(t *testing.T) {
    filter := NewNamespaceFilter(nil, []string{"kube-system"}, logger)
    
    assert.False(t, filter.ShouldMutate("kube-system"))
    assert.True(t, filter.ShouldMutate("default"))
    assert.True(t, filter.ShouldMutate("production"))
}

func TestNamespaceFilter_Include(t *testing.T) {
    filter := NewNamespaceFilter([]string{"production", "staging"}, nil, logger)
    
    assert.True(t, filter.ShouldMutate("production"))
    assert.True(t, filter.ShouldMutate("staging"))
    assert.False(t, filter.ShouldMutate("development"))
}

func TestNamespaceFilter_ExcludeTakesPriority(t *testing.T) {
    filter := NewNamespaceFilter(
        []string{"production", "monitoring"},
        []string{"monitoring"},
        logger,
    )
    
    assert.True(t, filter.ShouldMutate("production"))
    assert.False(t, filter.ShouldMutate("monitoring")) // Excluded wins
}

func TestNamespaceFilter_EmptyLists(t *testing.T) {
    filter := NewNamespaceFilter(nil, nil, logger)
    
    assert.True(t, filter.ShouldMutate("any-namespace"))
}
```

## Notes

- Consider glob/regex patterns for namespace matching (future enhancement)
- Namespace labels could be used for more dynamic filtering
- Document the interaction between app-level and webhook-level filtering
