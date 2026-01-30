# Contributing to Lobstertank

Thank you for your interest in contributing to Lobstertank. This document
provides guidelines and instructions for contributing.

## Prerequisites

- Go 1.22+
- Node.js 20 LTS+
- Podman 4.0+
- podman-compose 1.0+
- pre-commit

## Development Setup

```bash
# Clone and enter the repository
git clone https://github.com/AdamPippert/Lobstertank.git
cd Lobstertank

# Run the dev setup script
make setup

# Start the development environment
make dev
```

## Repository Structure

```
Lobstertank/
├── backend/          # Go API server
│   ├── cmd/          # Application entrypoints
│   └── internal/     # Private application packages
├── frontend/         # React + TypeScript SPA
├── deploy/           # OCI Containerfiles, compose, Helm
├── docs/             # Architecture, ADRs, API specs
└── scripts/          # Development and CI scripts
```

## Branching Strategy

- `main` — stable, release-ready code
- `feat/<name>` — new features
- `fix/<name>` — bug fixes
- `chore/<name>` — maintenance tasks

## Commit Messages

This project uses [Conventional Commits](https://www.conventionalcommits.org/).
A pre-commit hook enforces this.

```
feat: add gateway health check polling
fix: resolve token refresh race condition
docs: update ADR for secret provider selection
```

## Code Quality

### Backend (Go)

```bash
make lint-backend    # Run golangci-lint
make test-backend    # Run unit tests
make test-backend-integration  # Run integration tests
```

### Frontend (TypeScript/React)

```bash
make lint-frontend   # Run ESLint + Prettier check
make test-frontend   # Run Vitest
```

## Pull Request Process

1. Create a feature branch from `main`.
2. Make changes, ensuring tests pass locally (`make check`).
3. Open a pull request using the PR template.
4. Obtain at least one approving review from a code owner.
5. Merge via squash-and-merge.

## Security

- Never commit secrets, credentials, or private keys.
- Use the secret provider abstraction for all sensitive values.
- Report security vulnerabilities privately to the maintainers.

## Architecture Decision Records

Significant design decisions are documented as ADRs in `docs/adr/`. When
proposing a change that alters the architecture, include a new ADR in your PR.
