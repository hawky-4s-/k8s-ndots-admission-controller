# Task 3.1: Configuration Management

**Phase**: 3 - Configuration & Observability  
**Estimate**: 2-3 hours  
**Dependencies**: Task 1.2

## Objective

Implement centralized configuration management with environment variable support, validation, and sensible defaults.

## Deliverables

- [ ] Config struct in `internal/config/config.go`
- [ ] Environment variable loading
- [ ] Configuration validation
- [ ] Default values

## Environment Variables Reference

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8443` | Webhook server port |
| `METRICS_PORT` | `8080` | Metrics server port |
| `TLS_CERT_PATH` | `/certs/tls.crt` | Path to TLS certificate |
| `TLS_KEY_PATH` | `/certs/tls.key` | Path to TLS private key |
| `NDOTS_VALUE` | `2` | ndots value to set |
| `ANNOTATION_KEY` | `change-ndots` | Annotation key for opt-in/out |
| `ANNOTATION_MODE` | `opt-out` | Mode: always, opt-in, opt-out |
| `NAMESPACE_INCLUDE` | (empty) | Comma-separated included namespaces |
| `NAMESPACE_EXCLUDE` | `kube-system,...` | Comma-separated excluded namespaces |
| `LOG_LEVEL` | `info` | Log level: debug, info, warn, error |

## Acceptance Criteria

- [ ] All configuration loaded from environment variables
- [ ] Sensible default values provided
- [ ] Validation catches invalid configuration
- [ ] Configuration is immutable after loading
