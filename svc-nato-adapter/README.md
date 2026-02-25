<!-- CLASSIFICATION: UNCLASSIFIED -->

# svc-nato-adapter

NATO Interoperability Adapter — noop stub for STANAG 5516 / NFFI / MIP support.

## Overview

This service implements `NatoAdapterService` as a structurally complete noop. All RPCs
log the request and return success with empty payloads. No real STANAG 5516 or NFFI
encoding/decoding occurs in this v1 stub.

## Requirements Traceability

| Requirement | Feature | Use Case |
|---|---|---|
| CR-NATO-001 … CR-NATO-005 | FEAT-15 | UC011 |

## Configuration

All configuration is via environment variables:

| Variable | Default | Description |
|---|---|---|
| `RTSA_GRPC_PORT` | `50051` | gRPC server port |
| `RTSA_HEALTH_PORT` | `8081` | Health check HTTP port |
| `RTSA_SERVICE_NAME` | `svc-nato-adapter` | Service identifier |
| `RTSA_LOG_LEVEL` | `info` | Log level (debug/info/warn/error) |
| `RTSA_LOG_FORMAT` | `json` | Log format (json/text) |
| `RTSA_TLS_ENABLED` | `false` | Enable mTLS (required in production) |
| `RTSA_ENVIRONMENT` | `development` | Deployment environment |

## Running locally

```bash
cd svc-nato-adapter
go run ./cmd/server
```

## Testing

```bash
go test ./...
```
