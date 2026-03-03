# CLASSIFICATION: UNCLASSIFIED
# RTSA Project Root Makefile

.PHONY: help proto-gen proto-lint proto-breaking build test test-coverage \
        integration-test test-bench lint docker-up docker-down docker-logs \
        docker-up-all docker-down-all docker-logs-all \
        init-topics init-clickhouse health-check clean

# ──────────────────────────────────────────
# Help
# ──────────────────────────────────────────
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-25s\033[0m %s\n", $$1, $$2}'

# ──────────────────────────────────────────
# Protobuf
# ──────────────────────────────────────────
proto-gen: ## Generate Go and TypeScript code from .proto files
	buf generate

proto-lint: ## Lint protobuf files
	buf lint

proto-breaking: ## Check for breaking changes in protobuf
	buf breaking --against '.git#branch=main'

# ──────────────────────────────────────────
# Build
# ──────────────────────────────────────────
SERVICES := svc-radar-ingestion svc-ew-ingestion svc-elint-ingestion \
            svc-isr-ingestion svc-ais-ingestion svc-cyber-ingestion \
            svc-fusion-engine svc-anomaly-detection svc-feedback \
            svc-track svc-alert svc-query svc-audit

build: ## Build all services
	@for svc in $(SERVICES); do \
		echo "Building $$svc..."; \
		cd $$svc && go build ./... && cd ..; \
	done

# ──────────────────────────────────────────
# Test
# ──────────────────────────────────────────
test: ## Run all unit tests with race detector
	@for svc in $(SERVICES); do \
		echo "Testing $$svc..."; \
		(cd $$svc && go test -race -count=1 -coverprofile=coverage.out ./...); \
	done
	@cd pkg && go test -race -count=1 -coverprofile=coverage.out ./...

test-coverage: ## Run tests and show coverage summary
	@for svc in $(SERVICES); do \
		echo "=== $$svc ==="; \
		(cd $$svc && go test -race -count=1 -coverprofile=coverage.out ./... && \
		go tool cover -func=coverage.out | tail -1); \
	done

integration-test: ## Run integration tests (requires docker stack running)
	go test -race -count=1 -tags=integration ./tests/integration/...

test-bench: ## Run performance benchmarks (B01–B04) via test-bench.sh
	./scripts/dev/test-bench.sh

# ──────────────────────────────────────────
# Lint
# ──────────────────────────────────────────
lint: ## Run golangci-lint on all services
	@for svc in $(SERVICES); do \
		echo "Linting $$svc..."; \
		cd $$svc && golangci-lint run ./... && cd ..; \
	done
	@cd pkg && golangci-lint run ./...

# ──────────────────────────────────────────
# Docker — Infrastructure Only
# ──────────────────────────────────────────
docker-up: ## Start infrastructure stack (Redpanda, ClickHouse, observability)
	docker compose -f deploy/docker-compose.yml up -d

docker-down: ## Stop infrastructure stack
	docker compose -f deploy/docker-compose.yml down

docker-logs: ## Follow infrastructure logs
	docker compose -f deploy/docker-compose.yml logs -f

# ──────────────────────────────────────────
# Docker — Full Stack (Infrastructure + Services)
# ──────────────────────────────────────────
docker-up-all: ## Start full stack (infrastructure + RTSA services)
	docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.services.yml up -d --build

docker-down-all: ## Stop full stack
	docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.services.yml down

docker-logs-all: ## Follow full stack logs
	docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.services.yml logs -f

# ──────────────────────────────────────────
# Initialization
# ──────────────────────────────────────────
init-topics: ## Create Redpanda topics
	./scripts/dev/init-topics.sh

init-clickhouse: ## Initialize ClickHouse schema
	./scripts/dev/init-clickhouse.sh

health-check: ## Run health check
	./scripts/dev/health-check.sh

# ──────────────────────────────────────────
# Clean
# ──────────────────────────────────────────
clean: ## Remove build artifacts and coverage files
	@for svc in $(SERVICES); do \
		rm -f $$svc/coverage.out; \
	done
	rm -f pkg/coverage.out
	rm -rf gen/
