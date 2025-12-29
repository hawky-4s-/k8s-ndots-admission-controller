# Task 1.3: Admission Handler Core

**Phase**: 1 - Core Foundation  
**Estimate**: 3-4 hours  
**Dependencies**: Task 1.2

## Objective

Implement the core admission handler that receives AdmissionReview requests, decodes Pod objects, and returns AdmissionReview responses with JSON patches.

## Deliverables

- [ ] Admission handler in `internal/admission/handler.go`
- [ ] AdmissionReview request/response handling
- [ ] Pod decoding and validation
- [ ] Basic JSONPatch response structure

## Implementation Details

### Handler Package (`internal/admission/handler.go`)

```go
package admission

import (
    "encoding/json"
    "io"
    "net/http"
    
    admissionv1 "k8s.io/api/admission/v1"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
    scheme       = runtime.NewScheme()
    codecs       = serializer.NewCodecFactory(scheme)
    deserializer = codecs.UniversalDeserializer()
)

type Handler struct {
    mutator *Mutator
    logger  *slog.Logger
}

func NewHandler(mutator *Mutator, logger *slog.Logger) *Handler {
    return &Handler{
        mutator: mutator,
        logger:  logger,
    }
}

func (h *Handler) HandleMutate(w http.ResponseWriter, r *http.Request) {
    // 1. Read request body
    // 2. Deserialize AdmissionReview
    // 3. Validate request
    // 4. Extract Pod from request
    // 5. Call mutator to generate patch
    // 6. Build and send AdmissionReview response
}
```

### Request Processing Flow

```go
func (h *Handler) HandleMutate(w http.ResponseWriter, r *http.Request) {
    // Read body
    body, err := io.ReadAll(r.Body)
    if err != nil {
        h.sendError(w, "failed to read request body", http.StatusBadRequest)
        return
    }
    
    // Deserialize AdmissionReview
    var admissionReview admissionv1.AdmissionReview
    if _, _, err := deserializer.Decode(body, nil, &admissionReview); err != nil {
        h.sendError(w, "failed to decode admission review", http.StatusBadRequest)
        return
    }
    
    // Validate request
    if admissionReview.Request == nil {
        h.sendError(w, "admission review request is nil", http.StatusBadRequest)
        return
    }
    
    // Process and respond
    response := h.mutate(admissionReview.Request)
    admissionReview.Response = response
    admissionReview.Response.UID = admissionReview.Request.UID
    
    // Send response
    respBytes, _ := json.Marshal(admissionReview)
    w.Header().Set("Content-Type", "application/json")
    w.Write(respBytes)
}
```

### Mutate Function

```go
func (h *Handler) mutate(req *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
    // Only handle Pod resources
    if req.Kind.Kind != "Pod" {
        return &admissionv1.AdmissionResponse{
            Allowed: true,
        }
    }
    
    // Decode Pod
    var pod corev1.Pod
    if err := json.Unmarshal(req.Object.Raw, &pod); err != nil {
        return &admissionv1.AdmissionResponse{
            Allowed: false,
            Result: &metav1.Status{
                Message: fmt.Sprintf("failed to decode pod: %v", err),
            },
        }
    }
    
    // Generate patch
    patch, err := h.mutator.Mutate(&pod)
    if err != nil {
        // Log error but allow pod (fail open)
        h.logger.Error("mutation failed", "error", err)
        return &admissionv1.AdmissionResponse{Allowed: true}
    }
    
    // No mutation needed
    if patch == nil {
        return &admissionv1.AdmissionResponse{Allowed: true}
    }
    
    // Return patch
    patchBytes, _ := json.Marshal(patch)
    patchType := admissionv1.PatchTypeJSONPatch
    
    return &admissionv1.AdmissionResponse{
        Allowed:   true,
        Patch:     patchBytes,
        PatchType: &patchType,
    }
}
```

### JSONPatch Structure

```go
type PatchOperation struct {
    Op    string      `json:"op"`
    Path  string      `json:"path"`
    Value interface{} `json:"value,omitempty"`
}

// Example patches:
// Add dnsConfig when absent
[]PatchOperation{{Op: "add", Path: "/spec/dnsConfig", Value: dnsConfig}}

// Add options when dnsConfig exists but options is nil
[]PatchOperation{{Op: "add", Path: "/spec/dnsConfig/options", Value: options}}

// Replace specific option
[]PatchOperation{{Op: "replace", Path: "/spec/dnsConfig/options/0", Value: option}}
```

## Error Handling

| Scenario | Behavior |
|----------|----------|
| Invalid request body | Return 400 Bad Request |
| Not a Pod resource | Allow (no mutation) |
| Pod decode failure | Return error response |
| Mutation error | Allow (fail open) - configurable |

## Acceptance Criteria

- [ ] Handler accepts AdmissionReview v1 requests
- [ ] Handler correctly decodes Pod objects
- [ ] Handler returns valid AdmissionReview responses
- [ ] Handler returns JSONPatch when mutation is needed
- [ ] Handler allows non-Pod resources without mutation
- [ ] Handler handles errors gracefully (fail open by default)
- [ ] Content-Type is set to `application/json`

## Testing

```go
func TestHandler_HandleMutate_ValidPod(t *testing.T) {
    // Test successful mutation of a valid pod
}

func TestHandler_HandleMutate_InvalidBody(t *testing.T) {
    // Test handling of invalid request body
}

func TestHandler_HandleMutate_NonPodResource(t *testing.T) {
    // Test that non-Pod resources are allowed without mutation
}
```

## Notes

- Use `admissionv1` not `admissionv1beta1` (deprecated)
- UID must be copied from request to response
- Consider adding request ID for tracing
