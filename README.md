# Kubernetes ndots Admission Controller

A Mutating Admission Controller that injects or updates the `ndots` configuration in `Pod.spec.dnsConfig`. This helps improve DNS resolution performance for applications running in Kubernetes, especially when communicating with external services.

## Features

- **Automatic Injection**: Sets `ndots` value in Pod DNS configuration.
- **Configurable Modes**:
    - `opt-in`: Only mutate pods with annotation `change-ndots: "true"`.
    - `opt-out`: Mutate all pods except those with annotation `change-ndots: "false"`.
    - `always`: Mutate all pods regardless of annotations.
- **Namespace Filtering**: configurable list of included/excluded namespaces.
- **Critical Namespace Protection**: automatically excludes `kube-system` and other critical namespaces.
- **Helm Chart**: Easy deployment with Cert Manager integration.
- **Observability**: Prometheus metrics and structured logging.

## Installation

### Prerequisites

- Kubernetes 1.25+
- Helm 3.0+
- [Cert Manager](https://cert-manager.io/) (recommended for TLS)

### Install with Helm

1. Add the repository (if applicable) or clone this repo:
   ```bash
   git clone https://github.com/hawky-4s-/k8s-ndots-admission-controller.git
   cd k8s-ndots-admission-controller
   ```

2. Install the chart:
   ```bash
   helm upgrade --install ndots ./charts/ndots-webhook \
     --namespace ndots-system \
     --create-namespace
   ```

## Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `config.ndotsValue` | The ndots value to set | `2` |
| `config.annotationKey` | Annotation key for control | `change-ndots` |
| `config.annotationMode` | Mode: `always`, `opt-in`, `opt-out` | `opt-out` |
| `config.excludedNamespaces` | List of namespaces to ignore | `[kube-system, kube-public]` |
| `useCertManager` | Use cert-manager for TLS | `true` |

### Annotation Modes

- **opt-out** (Default): Mutations happen automatically. To skip a pod, add:
  ```yaml
  metadata:
    annotations:
      change-ndots: "false"
  ```
- **opt-in**: No mutations happen by default. To enable for a pod, add:
  ```yaml
  metadata:
    annotations:
      change-ndots: "true"
  ```

## Examples

### Deployment with Opt-Out

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  template:
    metadata:
      annotations:
        change-ndots: "false" # Prevents ndots modification
    spec:
      containers:
        - name: app
          image: nginx
```

## Monitoring

Metrics are exposed on port `8080` at `/metrics`.

| Metric | Description |
|--------|-------------|
| `ndots_admission_requests_total` | Total admission requests processed |
| `ndots_pod_mutations_total` | Total number of pod mutations performed |
| `ndots_admission_duration_seconds` | Latency of admission requests |

## Development

This project follows strict development guidelines. See [AGENTS.md](./AGENTS.md) for details.

### Prerequisites

- Go 1.25+
- Docker
- Kind (for local clusters)

### Common Commands

```bash
# Run unit tests
make test

# Run linting
make lint

# Run E2E tests (requires Kind)
make e2e

# Build binary
make build

# Build Docker image
make docker-build
```

## License

This project is licensed under the [MIT License](LICENSE).
