# Lobstertank

**Control plane for managing multiple [OpenClaw](https://github.com/AdamPippert) gateways.**

Lobstertank provides a unified interface for registering, monitoring, and
orchestrating multiple OpenClaw agent gateway instances across heterogeneous
infrastructure — from local containers to OpenShift clusters to ephemeral
cloud sandboxes.

## Features

- **Gateway Registry** — Register, tag, and manage the lifecycle of OpenClaw
  gateway instances with health-check polling and TTL-based expiration.
- **Meta-Agent** — Fan-out prompts to multiple gateways simultaneously and
  aggregate responses through a single interface.
- **Transport Abstraction** — Connect to gateways over HTTPS, Tailscale,
  Headscale, or Cloudflare Tunnels via a pluggable provider model.
- **Auth Abstraction** — Authenticate with bearer tokens, mTLS, or OIDC.
  Each gateway can use a different auth method.
- **Secrets Management** — Built-in encrypted secrets store or first-class
  HashiCorp Vault integration via the SecretProvider interface.
- **Audit Logging** — Structured audit trail for every gateway operation,
  policy decision, and agent interaction.
- **Multi-Environment Deployment** — OCI-compatible container images that run
  under Podman, on Kubernetes, OpenShift, or standalone hosts.

## Quick Start

### Prerequisites

| Tool | Version |
|------|---------|
| Go | 1.22+ |
| Node.js | 20 LTS+ |
| Podman | 4.0+ |
| podman-compose | 1.0+ |

### Development

```bash
# Install dependencies and pre-commit hooks
make setup

# Run backend and frontend locally
make dev

# Run all linters and tests
make check
```

### Container Build

```bash
# Build OCI images with Podman
make container-backend
make container-frontend

# Run the full stack
make compose-up
```

## Architecture

```
┌─────────────┐     ┌──────────────────────────────────────────┐
│   Browser    │────▶│          Lobstertank Frontend             │
│   (React)    │     │          (SPA — port 3000)                │
└─────────────┘     └──────────────┬───────────────────────────┘
                                   │ REST API
                    ┌──────────────▼───────────────────────────┐
                    │         Lobstertank Backend               │
                    │         (Go — port 8080)                  │
                    │                                           │
                    │  ┌─────────┐ ┌─────────┐ ┌────────────┐ │
                    │  │ Gateway │ │  Auth    │ │  Secrets   │ │
                    │  │Registry │ │Provider  │ │ Provider   │ │
                    │  └────┬────┘ └─────────┘ └────────────┘ │
                    │       │      ┌─────────┐ ┌────────────┐ │
                    │       │      │Transport│ │   Audit    │ │
                    │       │      │Provider │ │   Logger   │ │
                    │       │      └────┬────┘ └────────────┘ │
                    │       │           │                       │
                    │  ┌────▼───────────▼────┐                 │
                    │  │    Meta-Agent       │                 │
                    │  │    (Fan-out)        │                 │
                    │  └────────────────────┘                 │
                    └──────────────┬───────────────────────────┘
                                   │
              ┌────────────────────┼────────────────────┐
              │                    │                    │
      ┌───────▼──────┐   ┌───────▼──────┐   ┌───────▼──────┐
      │  OpenClaw    │   │  OpenClaw    │   │  OpenClaw    │
      │  Gateway A   │   │  Gateway B   │   │  Gateway C   │
      │  (local)     │   │  (OpenShift) │   │  (DO droplet)│
      └──────────────┘   └──────────────┘   └──────────────┘
```

## Repository Structure

```
Lobstertank/
├── backend/            # Go API server
│   ├── cmd/            # Application entrypoints
│   └── internal/       # Private application packages
│       ├── audit/      # Structured audit logging
│       ├── auth/       # Authentication providers
│       ├── config/     # Configuration loading
│       ├── gateway/    # Gateway domain (models, registry, handlers)
│       ├── metaagent/  # Multi-gateway fan-out orchestration
│       ├── secrets/    # Secrets management providers
│       ├── server/     # HTTP server and routing
│       ├── store/      # Database abstraction (PostgreSQL, SQLite)
│       └── transport/  # Network transport providers
├── frontend/           # React + TypeScript SPA
│   └── src/
│       ├── api/        # API client
│       ├── components/ # React components
│       ├── hooks/      # Custom React hooks
│       └── types/      # TypeScript type definitions
├── deploy/             # OCI Containerfiles, compose, Helm
├── docs/               # Architecture, ADRs, API specs
│   ├── adr/            # Architecture Decision Records
│   └── api/            # OpenAPI specification
└── scripts/            # Development and CI scripts
```

## Documentation

- [Architecture Overview](docs/architecture.md)
- [API Specification](docs/api/openapi.yaml)
- [Contributing Guide](CONTRIBUTING.md)

### Architecture Decision Records

- [ADR-0001: Monorepo Structure](docs/adr/0001-monorepo-structure.md)
- [ADR-0002: Go Backend](docs/adr/0002-go-backend.md)
- [ADR-0003: Transport Provider Abstraction](docs/adr/0003-transport-abstraction.md)
- [ADR-0004: Secret Provider Abstraction](docs/adr/0004-secret-provider-abstraction.md)

## License

Proprietary. See [LICENSE](LICENSE) for details.
