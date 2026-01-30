# ADR-0004: Secret Provider Abstraction

## Status

Accepted

## Context

Lobstertank stores credentials for authenticating with OpenClaw gateways.
Users range from individual developers (who want simplicity) to enterprises
(who mandate HashiCorp Vault or similar). The system must support both without
exposing secrets in API responses or logs.

## Decision

Introduce a `SecretProvider` interface with URI-style secret references:

```go
type Provider interface {
    Resolve(ctx context.Context, ref string) (string, error)
    Store(ctx context.Context, ref string, value string) error
    Delete(ctx context.Context, ref string) error
}
```

Ship two implementations:
1. **Builtin** — AES-256-GCM encrypted in-memory store.
2. **Vault** — HashiCorp Vault KV v2 (planned).

Gateway records reference secrets by URI (e.g., `builtin://gw-a/token`)
rather than storing plaintext values.

## Rationale

- **No plaintext exposure**: Secrets are resolved server-side only when
  needed for outbound requests.
- **Audit trail**: Secret access can be logged independently.
- **User choice**: Small deployments use the builtin provider; regulated
  environments plug in Vault.
- **Consistent API**: The gateway model does not change regardless of which
  provider backs the secrets.

## Consequences

- The builtin provider stores secrets in memory. A persistence layer
  (encrypted at rest in the database) is needed for production.
- Vault integration requires network access to a Vault server and
  appropriate policies.
- Secret rotation must be handled per-provider; no unified rotation API
  exists yet.
