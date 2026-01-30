#!/usr/bin/env bash
# Lobstertank — development environment setup
# Usage: bash scripts/dev-setup.sh

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

echo "==> Lobstertank Development Setup"
echo ""

# ── Check prerequisites ─────────────────────
check_cmd() {
  if ! command -v "$1" &>/dev/null; then
    echo "ERROR: $1 is required but not found. Please install it first."
    exit 1
  fi
  echo "  ✓ $1 found: $(command -v "$1")"
}

echo "Checking prerequisites..."
check_cmd go
check_cmd node
check_cmd npm
check_cmd podman
echo ""

# ── Install pre-commit hooks ────────────────
if command -v pre-commit &>/dev/null; then
  echo "Installing pre-commit hooks..."
  pre-commit install
  pre-commit install --hook-type commit-msg
  echo ""
else
  echo "WARN: pre-commit not found. Skipping hook installation."
  echo "      Install with: pip install pre-commit"
  echo ""
fi

# ── Backend setup ───────────────────────────
echo "Setting up backend..."
cd "$REPO_ROOT/backend"
go mod download
echo "  Backend dependencies downloaded."
echo ""

# ── Frontend setup ──────────────────────────
echo "Setting up frontend..."
cd "$REPO_ROOT/frontend"
npm ci
echo "  Frontend dependencies installed."
echo ""

# ── Environment file ────────────────────────
if [ ! -f "$REPO_ROOT/.env" ]; then
  echo "Creating .env from .env.example..."
  cp "$REPO_ROOT/.env.example" "$REPO_ROOT/.env"
  echo "  .env created. Edit it to set your secrets."
else
  echo ".env already exists, skipping."
fi
echo ""

# ── Done ────────────────────────────────────
echo "==> Setup complete!"
echo ""
echo "Quick start:"
echo "  make dev        — Start backend + frontend"
echo "  make check      — Run linters + tests"
echo "  make compose-up — Run with Podman containers"
