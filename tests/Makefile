# CLASSIFICATION: UNCLASSIFIED
# tests/Makefile — Test targets for Module 17 Integration & E2E Tests

.PHONY: test-integration test-e2e test-bench test-all help

## help: Show this help
help:
@grep -E '^##[^#]' $(MAKEFILE_LIST) | sed 's/## //'

## test-integration: Run integration tests (requires Docker)
test-integration:
RTSA_INTEGRATION_TESTS=true go test -v -tags integration -timeout 10m \
./integration/...

## test-e2e: Run end-to-end tests (requires full Docker Compose stack)
test-e2e:
docker compose -f docker-compose.test.yml up -d --wait
RTSA_INTEGRATION_TESTS=true go test -v -tags e2e -timeout 15m \
./e2e/...
docker compose -f docker-compose.test.yml down -v

## test-bench: Run performance benchmarks with threshold assertions
test-bench:
RTSA_INTEGRATION_TESTS=true go test -v -tags integration \
-bench=. -benchtime=10s -timeout 10m \
./benchmark/...

## test-all: Run all tests (integration + e2e + benchmarks)
test-all: test-integration test-e2e test-bench

## test-unit: Run unit-level integration tests only (no Docker infrastructure required)
test-unit:
RTSA_INTEGRATION_TESTS=true go test -v -tags integration -timeout 5m \
-run "TestIT0[1-7]_|TestFeedback|TestQuery" \
./integration/...
