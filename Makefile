.PHONY: all build build-backend build-frontend lint lint-backend lint-frontend \
	test test-backend test-frontend check dev clean setup \
	container-backend container-frontend compose-up compose-down

# ──────────────────────────────────────────────
# Variables
# ──────────────────────────────────────────────
BACKEND_DIR    := backend
FRONTEND_DIR   := frontend
DEPLOY_DIR     := deploy
BIN_DIR        := bin
BINARY         := $(BIN_DIR)/lobstertank

GO             := go
CONTAINER_CMD  := podman

# ──────────────────────────────────────────────
# Aggregate targets
# ──────────────────────────────────────────────
all: lint test build

check: lint test

# ──────────────────────────────────────────────
# Build
# ──────────────────────────────────────────────
build: build-backend build-frontend

build-backend:
	cd $(BACKEND_DIR) && $(GO) build -o ../$(BINARY) ./cmd/lobstertank

build-frontend:
	cd $(FRONTEND_DIR) && npm ci && npm run build

# ──────────────────────────────────────────────
# Lint
# ──────────────────────────────────────────────
lint: lint-backend lint-frontend

lint-backend:
	cd $(BACKEND_DIR) && golangci-lint run ./...

lint-frontend:
	cd $(FRONTEND_DIR) && npm run lint

# ──────────────────────────────────────────────
# Test
# ──────────────────────────────────────────────
test: test-backend test-frontend

test-backend:
	cd $(BACKEND_DIR) && $(GO) test -race -coverprofile=coverage.out ./...

test-backend-integration:
	cd $(BACKEND_DIR) && $(GO) test -race -tags=integration ./...

test-frontend:
	cd $(FRONTEND_DIR) && npm test -- --run

# ──────────────────────────────────────────────
# Containers (OCI / Podman)
# ──────────────────────────────────────────────
container-backend:
	$(CONTAINER_CMD) build -f $(DEPLOY_DIR)/Containerfile.backend -t lobstertank-backend:dev .

container-frontend:
	$(CONTAINER_CMD) build -f $(DEPLOY_DIR)/Containerfile.frontend -t lobstertank-frontend:dev .

compose-up:
	cd $(DEPLOY_DIR) && podman-compose up -d

compose-down:
	cd $(DEPLOY_DIR) && podman-compose down

# ──────────────────────────────────────────────
# Development
# ──────────────────────────────────────────────
setup:
	@echo "Installing pre-commit hooks..."
	pre-commit install
	pre-commit install --hook-type commit-msg
	@echo "Installing frontend dependencies..."
	cd $(FRONTEND_DIR) && npm ci
	@echo "Verifying Go modules..."
	cd $(BACKEND_DIR) && $(GO) mod download
	@echo "Setup complete."

dev:
	@echo "Starting backend (hot-reload not included — use air or similar)..."
	cd $(BACKEND_DIR) && $(GO) run ./cmd/lobstertank &
	@echo "Starting frontend dev server..."
	cd $(FRONTEND_DIR) && npm run dev

# ──────────────────────────────────────────────
# Clean
# ──────────────────────────────────────────────
clean:
	rm -rf $(BIN_DIR)
	rm -rf $(FRONTEND_DIR)/dist
	cd $(BACKEND_DIR) && rm -f coverage.out
