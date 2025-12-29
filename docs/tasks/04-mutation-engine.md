# Task 2.1: DNSConfig Mutation Engine

**Phase**: 2 - Mutation Logic  
**Estimate**: 3-4 hours  
**Dependencies**: Task 1.3

## Objective

Implement the core mutation logic that modifies Pod `spec.dnsConfig.options` to set the `ndots` value, handling all edge cases.

## Deliverables

- [ ] Mutator in `internal/admission/mutator.go`
- [ ] DNSConfig creation when absent
- [ ] Options array handling (add/update ndots)
- [ ] JSONPatch generation for all scenarios
- [ ] No-op detection when ndots already matches

## Implementation Details

### Mutator Package (`internal/admission/mutator.go`)

```go
package admission

import (
    corev1 "k8s.io/api/core/v1"
)

type Mutator struct {
    ndotsValue string
    config     *Config
    logger     *slog.Logger
}

func NewMutator(cfg *Config, logger *slog.Logger) *Mutator {
    return &Mutator{
        ndotsValue: strconv.Itoa(cfg.NdotsValue),
        config:     cfg,
        logger:     logger,
    }
}

func (m *Mutator) Mutate(pod *corev1.Pod) ([]PatchOperation, error) {
    // 1. Check if mutation should be skipped (handled in Task 2.2)
    // 2. Check current dnsConfig state
    // 3. Generate appropriate patch
    // 4. Return nil if no mutation needed
}
```

### Mutation Scenarios

#### Scenario 1: dnsConfig is nil

```go
// Pod has no dnsConfig at all
// Action: Add complete dnsConfig with options containing ndots

patch := []PatchOperation{{
    Op:   "add",
    Path: "/spec/dnsConfig",
    Value: corev1.PodDNSConfig{
        Options: []corev1.PodDNSConfigOption{{
            Name:  "ndots",
            Value: &m.ndotsValue,
        }},
    },
}}
```

#### Scenario 2: dnsConfig exists, options is nil

```go
// Pod has dnsConfig but no options array
// Action: Add options array with ndots

patch := []PatchOperation{{
    Op:   "add",
    Path: "/spec/dnsConfig/options",
    Value: []corev1.PodDNSConfigOption{{
        Name:  "ndots",
        Value: &m.ndotsValue,
    }},
}}
```

#### Scenario 3: options exists, no ndots

```go
// Pod has options but no ndots entry
// Action: Append ndots to options array

patch := []PatchOperation{{
    Op:   "add",
    Path: "/spec/dnsConfig/options/-",
    Value: corev1.PodDNSConfigOption{
        Name:  "ndots",
        Value: &m.ndotsValue,
    },
}}
```

#### Scenario 4: ndots exists with different value

```go
// Pod has ndots but with wrong value
// Action: Replace the ndots entry

// Find index of existing ndots
idx := findNdotsIndex(pod.Spec.DNSConfig.Options)

patch := []PatchOperation{{
    Op:   "replace",
    Path: fmt.Sprintf("/spec/dnsConfig/options/%d/value", idx),
    Value: m.ndotsValue,
}}
```

#### Scenario 5: ndots already has target value

```go
// Pod already has correct ndots value
// Action: No patch needed (return nil)

return nil, nil
```

### Helper Functions

```go
// findNdotsIndex returns the index of ndots option, or -1 if not found
func findNdotsIndex(options []corev1.PodDNSConfigOption) int {
    for i, opt := range options {
        if opt.Name == "ndots" {
            return i
        }
    }
    return -1
}

// getNdotsValue returns the current ndots value or empty string
func getNdotsValue(pod *corev1.Pod) string {
    if pod.Spec.DNSConfig == nil {
        return ""
    }
    for _, opt := range pod.Spec.DNSConfig.Options {
        if opt.Name == "ndots" && opt.Value != nil {
            return *opt.Value
        }
    }
    return ""
}

// shouldMutate determines if mutation is needed based on current state
func (m *Mutator) shouldMutate(pod *corev1.Pod) bool {
    currentValue := getNdotsValue(pod)
    return currentValue != m.ndotsValue
}
```

### Full Mutate Implementation

```go
func (m *Mutator) Mutate(pod *corev1.Pod) ([]PatchOperation, error) {
    // Check if mutation is needed
    if !m.shouldMutate(pod) {
        m.logger.Debug("skipping mutation, ndots already set",
            "namespace", pod.Namespace,
            "name", pod.Name,
            "current", getNdotsValue(pod),
        )
        return nil, nil
    }
    
    var patches []PatchOperation
    
    // Scenario 1: No dnsConfig
    if pod.Spec.DNSConfig == nil {
        patches = append(patches, PatchOperation{
            Op:   "add",
            Path: "/spec/dnsConfig",
            Value: map[string]interface{}{
                "options": []map[string]interface{}{{
                    "name":  "ndots",
                    "value": m.ndotsValue,
                }},
            },
        })
        return patches, nil
    }
    
    // Scenario 2: dnsConfig exists, options is nil
    if pod.Spec.DNSConfig.Options == nil {
        patches = append(patches, PatchOperation{
            Op:   "add",
            Path: "/spec/dnsConfig/options",
            Value: []map[string]interface{}{{
                "name":  "ndots",
                "value": m.ndotsValue,
            }},
        })
        return patches, nil
    }
    
    // Scenario 3 & 4: options exists
    idx := findNdotsIndex(pod.Spec.DNSConfig.Options)
    if idx == -1 {
        // Scenario 3: Append ndots
        patches = append(patches, PatchOperation{
            Op:   "add",
            Path: "/spec/dnsConfig/options/-",
            Value: map[string]interface{}{
                "name":  "ndots",
                "value": m.ndotsValue,
            },
        })
    } else {
        // Scenario 4: Replace existing ndots
        patches = append(patches, PatchOperation{
            Op:    "replace",
            Path:  fmt.Sprintf("/spec/dnsConfig/options/%d/value", idx),
            Value: m.ndotsValue,
        })
    }
    
    return patches, nil
}
```

## Acceptance Criteria

- [ ] Creates dnsConfig when pod has none
- [ ] Adds options array when dnsConfig exists without options
- [ ] Appends ndots when options exist without ndots
- [ ] Replaces ndots when value differs from target
- [ ] Returns nil patch when ndots already matches target
- [ ] Preserves existing dnsConfig settings (nameservers, searches)
- [ ] Generates valid RFC 6902 JSONPatch

## Testing

```go
func TestMutator_Mutate_NoDNSConfig(t *testing.T) {
    // Pod with no dnsConfig should get full dnsConfig added
}

func TestMutator_Mutate_EmptyOptions(t *testing.T) {
    // Pod with dnsConfig but nil options
}

func TestMutator_Mutate_AppendNdots(t *testing.T) {
    // Pod with options but no ndots entry
}

func TestMutator_Mutate_UpdateNdots(t *testing.T) {
    // Pod with different ndots value
}

func TestMutator_Mutate_NoOp(t *testing.T) {
    // Pod already has correct ndots - should return nil
}

func TestMutator_Mutate_PreservesOtherOptions(t *testing.T) {
    // Ensure other DNS options (timeout, attempts) are preserved
}
```

## Notes

- ndots value is always a string in PodDNSConfigOption
- JSONPatch paths must escape special characters (RFC 6901)
- Order of options array should be preserved
