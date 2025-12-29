# ndots-webhook Helm Chart

A Helm chart for deploying the ndots-webhook mutating admission controller.

## Installation

```bash
helm upgrade --install ndots . \
  --namespace ndots-system \
  --create-namespace
```

## Configuration

| Key | Description | Default |
|-----|-------------|---------|
| `image.repository` | Image repository | `ndots-webhook` |
| `image.tag` | Image tag | `latest` |
| `config.ndotsValue` | The ndots value to set | `2` |
| `config.annotationMode` | Mutation mode (`always`, `opt-in`, `opt-out`) | `opt-out` |
| `useCertManager` | Enable cert-manager integration | `true` |

See `values.yaml` for full configuration options.
