# Task 3.2: Structured Logging

**Phase**: 3 - Configuration & Observability  
**Estimate**: 1-2 hours  
**Dependencies**: Task 3.1

## Objective

Implement structured logging with configurable levels, JSON output for production, and mutation decision context.

## Deliverables

- [ ] Logger setup using `slog` (Go 1.21+)
- [ ] Configurable log levels
- [ ] JSON and text output formats
- [ ] Contextual logging for mutations

## Log Fields for Mutations

| Field | Description |
|-------|-------------|
| `namespace` | Pod namespace |
| `name` | Pod name |
| `mode` | Annotation mode |
| `action` | mutated, skipped |
| `reason` | Skip reason if applicable |
| `ndots_before` | Previous ndots value |
| `ndots_after` | New ndots value |

## Acceptance Criteria

- [ ] JSON logging in production
- [ ] Text logging for development
- [ ] Configurable log levels
- [ ] All mutations logged with context
