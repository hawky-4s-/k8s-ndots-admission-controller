# k8s-ndots-admission-controller Helm Chart

A Helm chart for deploying the k8s-ndots-admission-controller mutating admission controller.

## Installation

```bash
helm upgrade --install ndots . \
  --namespace ndots-system \
  --create-namespace
```

## Configuration

| Key | Description | Default |
|-----|-------------|---------|
| `image.repository` | Image repository | `hawky4s/k8s-ndots-admission-controller` |
| `image.tag` | Image tag | `""` (chart appVersion) |
| `ndots.value` | The ndots value to set | `2` |
| `ndots.annotationMode` | Mutation mode (`always`, `opt-in`, `opt-out`) | `opt-out` |
| `tls.useCertManager` | Enable cert-manager integration | `true` |
| `metrics.enabled` | Enable metrics endpoint | `true` |
| `metrics.serviceMonitor.enabled` | Enable Prometheus ServiceMonitor | `false` |

> **Note**: If `metrics.enabled` is `true` and `metrics.serviceMonitor.enabled` is `false`, the Service will be automatically annotated with `prometheus.io/scrape: "true"` and `prometheus.io/port`.

See `values.yaml` for full configuration options.
