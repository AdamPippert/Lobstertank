# Lobstertank Architecture

## Overview

Lobstertank is a control plane for managing multiple OpenClaw agent gateway
instances. It provides gateway registration, health monitoring, policy
enforcement, and multi-gateway orchestration through a single interface.

## System Boundaries

```
                    ┌──────────────────────┐
                    │     Lobstertank      │
                    │   Control Plane      │
                    │                      │
  Operators ───────▶│  Frontend (React)    │
                    │         │            │
                    │   REST API (Go)      │
                    │    │    │    │       │
                    │ Registry│  Auth      │
                    │    │  Audit  Secrets │
                    │    │    │    │       │
                    └────┼────┼────┼───────┘
                         │    │    │
           ┌─────────────┼────┼────┼─────────────┐
           │             │    │    │             │
    ┌──────▼──┐   ┌──────▼──┐│ ┌──▼───────┐
    │ Gateway │   │ Gateway ││ │ Secret   │
    │ A       │   │ B       ││ │ Store    │
    │(local)  │   │(OCP)    ││ │(Vault/  │
    └─────────┘   └─────────┘│ │ builtin)│
                              │ └──────────┘
                       ┌──────▼──┐
                       │ Gateway │
                       │ C       │
                       │(DO/Dayt)│
                       └─────────┘
```

## Core Components

### Gateway Registry

Manages the lifecycle of registered gateway instances:
- CRUD operations on gateway records
- Status tracking (online, offline, degraded, unknown)
- TTL-based expiration for ephemeral gateways (sandboxes)
- Label-based organization

### Transport Provider

Abstracts how Lobstertank connects to gateways. Each gateway can use a
different transport:

| Transport   | Status       | Use Case                     |
|-------------|-------------|------------------------------|
| HTTPS       | Implemented  | Default, works everywhere     |
| Tailscale   | Stub         | Private tailnet connectivity  |
| Headscale   | Stub         | Self-hosted tailnet           |
| Cloudflare  | Stub         | Tunnel-based access           |

### Auth Provider

Handles authentication for both inbound requests (operators accessing
Lobstertank) and outbound connections (Lobstertank accessing gateways):

- **Inbound**: Bearer token (implemented), OIDC (planned)
- **Outbound**: Per-gateway token, mTLS, or OIDC

### Secrets Provider

Manages sensitive credentials used for gateway authentication:

- **Builtin**: AES-256-GCM encrypted in-memory store (implemented)
- **Vault**: HashiCorp Vault KV v2 integration (planned)

Secret references use URI-style addressing:
`builtin://gateway-a/token` or `vault://secret/data/gateway-a`

### Meta-Agent

Orchestrates interactions across multiple gateways:
- **Fan-out**: Send a prompt to N gateways concurrently
- **Aggregation**: Collect and return all responses
- Future: routing, load balancing, response merging

### Audit Logger

Structured JSON audit trail for every significant operation:
- Gateway registration, update, deletion
- Health check results
- Meta-agent fan-out requests
- Authentication events

## Data Flow

1. Operator opens the Lobstertank UI (React SPA)
2. UI fetches gateway list via `GET /api/v1/gateways`
3. Operator selects a gateway from the dropdown
4. Actions (health check, prompt) flow through the REST API
5. Backend resolves transport + auth for the target gateway
6. Backend communicates with the OpenClaw gateway instance
7. Response returns through the same path
8. All operations are audit-logged

## Deployment Models

- **Development**: `go run` + `npm run dev` (Vite proxy to backend)
- **Container**: OCI images via `podman build` + `podman-compose`
- **Kubernetes**: Helm chart (planned)
- **OpenShift**: Helm + OpenShift-specific values (planned)

## Security Model

- All API endpoints (except `/healthz`) require authentication
- Secrets are never exposed in API responses
- Secrets are encrypted at rest (AES-256-GCM) in the builtin provider
- Audit log captures all state-changing operations
- Container images run as non-root
- TLS 1.2 minimum for all outbound gateway connections
