# Task 5.3: E2E Tests

**Phase**: 5 - Testing  
**Estimate**: 4-5 hours  
**Dependencies**: Task 5.2, Task 4.3

## Objective

End-to-end tests in a real Kubernetes cluster (kind) testing all workload types.

## Deliverables

- [ ] E2E test framework in `test/e2e/`
- [ ] Kind cluster setup scripts
- [ ] Workload type tests

## Test Scenarios

- [ ] Pod created directly
- [ ] Pod via Deployment
- [ ] Pod via StatefulSet
- [ ] Pod via DaemonSet
- [ ] Pod via Job/CronJob
- [ ] Annotation opt-in/opt-out
- [ ] Namespace exclusion
- [ ] Webhook failure (failurePolicy test)

## Commands

```bash
make kind-create
make deploy
make test-e2e
make kind-delete
```

## Acceptance Criteria

- [ ] All workload types tested
- [ ] Pods have correct ndots value
- [ ] Annotation behavior verified
- [ ] Namespace exclusion works
