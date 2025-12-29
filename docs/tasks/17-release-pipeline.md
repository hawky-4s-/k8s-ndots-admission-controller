# Task 6.2: Release Pipeline

**Phase**: 6 - CI/CD & Documentation  
**Estimate**: 2-3 hours  
**Dependencies**: Task 6.1

## Objective

Set up release workflow for building multi-arch images and creating GitHub releases.

## Deliverables

- [ ] `.github/workflows/release.yaml`
- [ ] Multi-arch Docker builds (amd64, arm64)
- [ ] Container registry push
- [ ] GitHub release creation

## Triggers

- Tag push matching `v*`

## Steps

1. Run full CI
2. Build multi-arch images with buildx
3. Push to container registry
4. Create GitHub release with changelog
5. (Optional) Publish Helm chart

## Acceptance Criteria

- [ ] Multi-arch images built
- [ ] Images pushed to registry
- [ ] GitHub release created with artifacts
- [ ] Changelog included in release notes
