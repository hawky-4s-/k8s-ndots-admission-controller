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
| `image.repository` | Image repository | `k8s-ndots-admission-controller` |
| `image.tag` | Image tag | `latest` |
| `config.ndotsValue` | The ndots value to set | `2` |
| `config.annotationMode` | Mutation mode (`always`, `opt-in`, `opt-out`) | `opt-out` |
| `useCertManager` | Enable cert-manager integration | `true` |

See `values.yaml` for full configuration options.
