# ==============================================================================
# ERP Purchasing Microservice - Developer Makefile
# ==============================================================================

BINARY_NAME    := bin/purchasing
CMD_PATH       := ./cmd/server
SQLC_BIN       := $(shell which sqlc 2>/dev/null || echo $(HOME)/go/bin/sqlc)
DOCKER_COMPOSE := docker compose

.PHONY: help run build clean test test-all test-coverage vet fmt sqlc docker-build docker-up docker-down docker-logs db-up db-down db-shell ci

ci: fmt vet test ## Run standard CI pipeline checks locally (fmt, vet, test)

help: ## Display this help screen
	@echo "========================================================================"
	@echo " ERP Purchasing Microservice - Make Commands"
	@echo "========================================================================"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

run: ## Run service locally (loads .env)
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

test-all: ## Run all tests including database integration tests
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
	$(DOCKER_COMPOSE) up -d postgres

db-down: ## Stop PostgreSQL container
	$(DOCKER_COMPOSE) stop postgres

db-shell: ## Open interactive psql shell inside postgres container
	$(DOCKER_COMPOSE) exec postgres psql -U postgres -d purchasing
