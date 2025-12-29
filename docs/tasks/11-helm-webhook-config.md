# Task 4.2: MutatingWebhookConfiguration Template

**Phase**: 4 - Helm Chart  
**Estimate**: 2-3 hours  
**Dependencies**: Task 4.1

## Objective

Create the MutatingWebhookConfiguration template with namespace selectors and configurable rules.

## Deliverables

- [ ] `mutatingwebhookconfiguration.yaml` template
- [ ] Namespace selector support
- [ ] Configurable failurePolicy
- [ ] CA bundle injection annotation

## Key Configuration

```yaml
webhooks:
  - name: ndots.admission.k8s.io
    rules:
      - apiGroups: [""]
        apiVersions: ["v1"]
        resources: ["pods"]
        operations: ["CREATE"]
    namespaceSelector:
      matchExpressions:
        - key: kubernetes.io/metadata.name
          operator: NotIn
          values: {{ .Values.namespaceExclude }}
    failurePolicy: {{ .Values.webhook.failurePolicy }}
    timeoutSeconds: {{ .Values.webhook.timeoutSeconds }}
```

## Acceptance Criteria

- [ ] Webhook configuration is valid
- [ ] Namespace selectors work correctly
- [ ] failurePolicy configurable (Fail/Ignore)
- [ ] CA bundle injection works with cert-manager
