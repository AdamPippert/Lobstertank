# ADR-0002: Go Backend

## Status

Accepted

## Context

The Lobstertank backend needs to manage gateway connections, handle concurrent
health checks, and integrate with Kubernetes-native tooling. We evaluated Go,
TypeScript (Node.js), Python, and Rust.

## Decision

Use Go for the backend API server.

## Rationale

- **Single binary deployment**: No runtime dependencies, simplifies container
  images and distribution.
- **Concurrency model**: Goroutines and channels naturally fit the fan-out
  pattern needed by the meta-agent.
- **Kubernetes ecosystem**: Client-go, controller-runtime, and Helm are all
  Go-native.
- **Strong typing**: Catches interface mismatches at compile time.
- **Enterprise adoption**: Widely used in infrastructure tooling
  (Kubernetes, Terraform, Prometheus), making it a familiar choice for
  platform engineering teams.

## Consequences

- Frontend and backend use different languages, requiring contributors to
  know both Go and TypeScript.
- Go's error handling is verbose but explicit.
- No ORM by default — SQL queries or a lightweight query builder will be used.
