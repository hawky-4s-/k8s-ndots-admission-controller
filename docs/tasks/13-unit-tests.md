# Task 5.1: Unit Tests

**Phase**: 5 - Testing  
**Estimate**: 4-5 hours  
**Dependencies**: Phase 2

## Objective

Comprehensive unit test coverage for mutation logic, annotation modes, and namespace filtering.

## Test Coverage Areas

### Mutation Engine
- [ ] dnsConfig absent → creates full config
- [ ] dnsConfig.options nil → adds options array
- [ ] ndots absent → appends to options
- [ ] ndots different value → updates value
- [ ] ndots matches target → no-op
- [ ] Other DNS options preserved

### Annotation Modes
- [ ] `always` mode with/without annotations
- [ ] `opt-in` mode: true, false, missing, other
- [ ] `opt-out` mode: true, false, missing, other

### Namespace Filtering
- [ ] Exclude list blocks namespaces
- [ ] Include list restricts to namespaces
- [ ] Exclude takes priority over include
- [ ] Empty lists allow all

## Acceptance Criteria

- [ ] 80%+ code coverage on mutation logic
- [ ] All edge cases covered with table-driven tests
- [ ] `go test -race` passes
