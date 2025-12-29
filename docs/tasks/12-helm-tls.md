# Task 4.3: TLS/Certificate Management

**Phase**: 4 - Helm Chart  
**Estimate**: 3-4 hours  
**Dependencies**: Task 4.1, Task 4.2

## Objective

Implement TLS certificate management with cert-manager integration and self-signed fallback.

## Deliverables

- [ ] cert-manager Certificate and Issuer templates
- [ ] Self-signed certificate generation hook
- [ ] CA bundle injection into webhook config
- [ ] Secret management

## Options

### Option 1: cert-manager (Recommended)

```yaml
# values.yaml
tls:
  useCertManager: true
  certManager:
    issuerRef:
      name: selfsigned-issuer
      kind: Issuer
```

### Option 2: Self-signed with Hook

```yaml
tls:
  useCertManager: false
  selfSigned:
    enabled: true
```

Uses a pre-install hook Job to generate certificates.

## Acceptance Criteria

- [ ] cert-manager integration works
- [ ] Self-signed fallback generates valid certs
- [ ] CA bundle correctly injected into webhook
- [ ] Certs mounted correctly in deployment
