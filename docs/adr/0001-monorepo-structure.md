# ADR-0001: Monorepo Structure

## Status

Accepted

## Context

Lobstertank comprises a Go backend, a React frontend, deployment manifests,
and documentation. We need to decide whether to use a monorepo or polyrepo
structure.

## Decision

Use a single monorepo containing all components:

```
backend/     — Go API server
frontend/    — React SPA
deploy/      — OCI Containerfiles, compose, Helm
docs/        — Architecture, ADRs, API specs
scripts/     — Development tooling
```

## Rationale

- **Portability**: A single `git clone` gives contributors everything needed
  to develop, test, and deploy.
- **Atomic changes**: Frontend and backend API changes can be reviewed and
  merged together.
- **Simplified CI**: One pipeline validates the entire system.
- **Enterprise alignment**: Easier to enforce consistent standards, linting,
  and code ownership across all components.

## Consequences

- The repository will grow larger over time. Sparse checkouts or build caching
  mitigate this.
- CI must be configured to detect changed paths and skip unaffected jobs.
- Contributors working on only one component still clone everything.
