# Task 6.1: GitHub Actions CI

**Phase**: 6 - CI/CD & Documentation  
**Estimate**: 2-3 hours  
**Dependencies**: Task 5.1

## Objective

Set up GitHub Actions CI workflow for linting, testing, building, and security scanning.

## Deliverables

- [ ] `.github/workflows/ci.yaml`
- [ ] Lint job with golangci-lint
- [ ] Test job with coverage
- [ ] Build job
- [ ] Security scanning

## Workflow Jobs

1. **lint**: golangci-lint
2. **test**: go test with coverage report
3. **build**: go build + docker build
4. **security**: gosec + trivy scan

## Triggers

- Push to `main`
- Pull requests
- Manual dispatch

## Acceptance Criteria

- [ ] All jobs run on PRs
- [ ] Coverage report uploaded
- [ ] Build artifacts cached
- [ ] Security issues reported
