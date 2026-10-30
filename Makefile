# ==============================================================================
# ERP Purchasing Microservice - Developer Makefile
# ==============================================================================

SHELL          := /bin/bash
BINARY_NAME    := bin/purchasing
CMD_PATH       := ./cmd/server
SQLC_BIN       := $(shell which sqlc 2>/dev/null || echo $(HOME)/go/bin/sqlc)
DOCKER_COMPOSE := docker compose

.PHONY: help run build clean test test-all test-coverage vet fmt sqlc \
	docker-build docker-up docker-down docker-logs \
	db-up db-down db-wait db-shell db-logs ci

ci: fmt vet test ## Run standard CI pipeline checks locally (fmt, vet, test)

help: ## Display this help screen
	@echo "========================================================================"
	@echo " ERP Purchasing Microservice - Make Commands"
	@echo "========================================================================"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.env:
	@if [ ! -f .env ]; then \
		echo "Creating .env from .env.example..."; \
		cp .env.example .env; \
	fi

run: ## Run with explicit environment, reviewed schema and public verification key
	go run $(CMD_PATH)

build: ## Build statically linked production binary
	@mkdir -p bin
	CGO_ENABLED=0 go build -ldflags="-w -s" -o $(BINARY_NAME) $(CMD_PATH)
	@echo "Built binary at $(BINARY_NAME)"

clean: ## Remove build artifacts and test reports
	@rm -rf bin/ coverage.out coverage.html
	@echo "Cleaned build artifacts"

test: ## Run service unit tests
	go test -v -count=1 ./internal/service/... ./internal/logger/... ./internal/metrics/... ./internal/middleware/... ./internal/api/... ./internal/config/... ./internal/docs/...

test-all: ## Run tests; database cases skip without the explicitly owned disposable fixture
	go test -v -count=1 ./...

test-coverage: ## Run unit tests with HTML coverage report
	@go test -coverprofile=coverage.out ./internal/service/...
	@go tool cover -html=coverage.out -o coverage.html
	@go tool cover -func=coverage.out | tail -n 1
	@echo "Coverage HTML report generated at coverage.html"

vet: ## Run go vet analysis
	go vet ./...

fmt: ## Format Go source files
	go fmt ./...

sqlc: ## Regenerate sqlc queries
	@if [ -x "$(SQLC_BIN)" ]; then \
		$(SQLC_BIN) generate; \
		echo "sqlc code generation complete"; \
	else \
		echo "sqlc binary not found. Please install sqlc: go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest"; \
		exit 1; \
	fi

docker-build: ## Build docker image
	docker build -t purchasing-service:latest .

docker-up: ## Start the full stack (Postgres + Purchasing Service)
	$(DOCKER_COMPOSE) up -d --build

docker-down: ## Stop the full docker compose stack
	$(DOCKER_COMPOSE) down

docker-logs: ## Follow service logs from docker compose
	$(DOCKER_COMPOSE) logs -f purchasing-service

db-up: ## Start only PostgreSQL in docker for local development
	@echo "Starting PostgreSQL container..."
	@if ! $(DOCKER_COMPOSE) up -d postgres 2>/tmp/docker_compose_err.log; then \
		cat /tmp/docker_compose_err.log; \
		if grep -q "permission denied" /tmp/docker_compose_err.log; then \
			echo ""; \
			echo "==================> Docker Permission Issue <=================="; \
			echo "Your user does not have permission to access /var/run/docker.sock."; \
			echo "If you were recently added to the 'docker' group, try:"; \
			echo "  1) Run 'newgrp docker' in your terminal"; \
			echo "  2) Or log out and log back in to refresh group permissions"; \
			echo "==============================================================="; \
		fi; \
		rm -f /tmp/docker_compose_err.log; \
		exit 1; \
	fi; \
	rm -f /tmp/docker_compose_err.log

db-wait: ## Ensure PostgreSQL is running and ready to accept connections
	@if (echo > /dev/tcp/127.0.0.1/5433) 2>/dev/null; then \
		echo "PostgreSQL is already running on port 5433."; \
	else \
		$(MAKE) db-up; \
		echo "Waiting for PostgreSQL to be ready on port 5433..."; \
		ready=0; \
		for i in $$(seq 1 30); do \
			if (echo > /dev/tcp/127.0.0.1/5433) 2>/dev/null && $(DOCKER_COMPOSE) exec -T postgres pg_isready -U postgres -d purchasing >/dev/null 2>&1; then \
				ready=1; \
				break; \
			fi; \
			sleep 0.5; \
		done; \
		if [ $$ready -eq 1 ]; then \
			echo "PostgreSQL is ready."; \
		else \
			echo "Error: Timed out waiting for PostgreSQL container to become ready." >&2; \
			exit 1; \
		fi; \
	fi

db-down: ## Stop PostgreSQL container
	$(DOCKER_COMPOSE) stop postgres

db-logs: ## Follow PostgreSQL container logs
	$(DOCKER_COMPOSE) logs -f postgres

db-shell: ## Open interactive psql shell inside postgres container
	$(DOCKER_COMPOSE) exec postgres psql -U postgres -d purchasing
