# Task 5.2: Integration Tests

**Phase**: 5 - Testing  
**Estimate**: 3-4 hours  
**Dependencies**: Task 5.1, Task 4.1

## Objective

Integration tests using envtest to test against a real Kubernetes API server.

## Deliverables

- [ ] envtest setup in `test/integration/`
- [ ] Webhook integration tests
- [ ] Admission request/response flow tests

## Test Scenarios

- [ ] Full admission flow with actual AdmissionReview
- [ ] Webhook configuration is applied
- [ ] Pod creation triggers mutation
- [ ] Multiple concurrent admissions

## Setup

```go
var testEnv *envtest.Environment

func TestMain(m *testing.M) {
    testEnv = &envtest.Environment{
        WebhookInstallOptions: envtest.WebhookInstallOptions{
            Paths: []string{"../../deploy/kubernetes"},
        },
    }
    // Start envtest...
}
```

## Acceptance Criteria

- [ ] envtest environment sets up correctly
- [ ] Webhook receives actual AdmissionReview requests
- [ ] Pod mutations are applied correctly
