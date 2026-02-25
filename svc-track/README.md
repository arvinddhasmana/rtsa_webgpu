<!-- CLASSIFICATION: UNCLASSIFIED -->

# svc-track — Real-Time Track Streaming Service

> **Module**: 10 — Track Service
> **Phase**: P4 (Presentation)
> **Classification**: UNCLASSIFIED
> **Feature**: FEAT-13 Situational Awareness UI
> **UC**: UC012 Situational Awareness UI

---

## Overview

`svc-track` is a presentation-layer microservice that:

1. **Consumes** fused track events from all 5 `tracks.fused.*` Redpanda topics
2. **Maintains** an in-memory cache of all active track state (indexed by `track_id`)
3. **Serves** real-time track updates over gRPC server-streaming to COP UI clients
4. **Enforces** classification filtering on every outbound message

It is a **stateless, in-memory service** — no ClickHouse dependency. On restart it rehydrates from the Redpanda consumer group offset.

---

## Architecture

```
Redpanda
  tracks.fused.surface ─┐
  tracks.fused.air      ├─ FusedTrackConsumer ─► TrackCache ─► onChange ─► StreamHandler
  tracks.fused.subsurface│                                                        │
  tracks.fused.land     │                                                    Fan-out to
  tracks.fused.cyber   ─┘                                               per-client channels
                                                                               │
COP Web App ─── gRPC StreamTracks ───────────────────────────────────► TrackUpdate stream
            ─── gRPC GetTrackDetails ──────────────────────────────────► FusedTrack
            ─── gRPC GetTrackHistory ──────────────────────────────────► TrackHistoryResponse
```

---

## gRPC API

| RPC | Type | Description |
|---|---|---|
| `StreamTracks` | Server-streaming | Initial snapshot + incremental updates |
| `GetTrackDetails` | Unary | Full track with source attribution |
| `GetTrackHistory` | Unary | Recent position history from cache |

Classification filtering is **mandatory** on all responses. Tracks with `classification > caller.clearance_level` are excluded (returned as NOT_FOUND for unary RPCs to avoid leaking existence).

---

## Filter Criteria (AND-combined)

| # | Criterion | Default if unset |
|---|---|---|
| 1 | `classification ≤ clearance_level` | Always enforced |
| 2 | `entity_type IN entity_types` | All types |
| 3 | `hostile_class IN hostile_classes` | All classifications |
| 4 | Position within `bounding_box` | Global |
| 5 | `confidence_score ≥ min_confidence` | 0.0 (all) |

---

## Configuration

All configuration via environment variables with `RTSA_` prefix:

| Variable | Default | Description |
|---|---|---|
| `RTSA_GRPC_PORT` | `50051` | gRPC server port |
| `RTSA_HEALTH_PORT` | `8081` | Health check HTTP port |
| `RTSA_METRICS_PORT` | `9090` | Prometheus metrics port |
| `RTSA_REDPANDA_BROKERS` | `localhost:19092` | Comma-separated broker list |
| `RTSA_CONSUMER_GROUP` | `track-service` | Kafka consumer group ID |
| `RTSA_TLS_ENABLED` | `false` | Enable mTLS (required for production) |
| `RTSA_TLS_CA_CERT` | `./certs/dev/ca.crt` | CA certificate path |
| `RTSA_TLS_SERVER_CERT` | `./certs/dev/server.crt` | Server certificate path |
| `RTSA_TLS_SERVER_KEY` | `./certs/dev/server.key` | Server key path |
| `RTSA_LOG_LEVEL` | `info` | Log level (debug/info/warn/error) |
| `RTSA_LOG_FORMAT` | `json` | Log format (json/text) |
| `RTSA_HISTORY_MAX_POINTS` | `100` | Max position history points per track |
| `RTSA_STREAM_CHANNEL_BUFFER` | `256` | Per-client update channel buffer size |

---

## Metrics

| Metric | Type | Labels |
|---|---|---|
| `rtsa_track_service_active_tracks` | Gauge | `entity_type`, `status` |
| `rtsa_track_service_stream_clients` | Gauge | — |
| `rtsa_track_service_updates_sent_total` | Counter | `entity_type`, `update_type` |
| `rtsa_track_service_cache_update_duration_seconds` | Histogram | — |

---

## Health Endpoints

| Path | Description |
|---|---|
| `GET :8081/healthz` | Liveness check |
| `GET :8081/readyz` | Readiness check |
| `GET :9090/metrics` | Prometheus metrics |

---

## Running Locally

```bash
# Prerequisites: Redpanda running (see deploy/docker-compose.yml)
cd svc-track
go run ./cmd/track

# Or with Docker:
docker build -t svc-track:dev -f Dockerfile ..
docker run -e RTSA_REDPANDA_BROKERS=localhost:19092 svc-track:dev
```

---

## Testing

```bash
# Unit tests (race detector):
cd svc-track
go test ./... -race -count=1 -v

# With coverage:
go test ./... -race -count=1 -coverprofile=coverage.out
go tool cover -html=coverage.out

# Integration tests (requires Docker):
cd tests/integration
go test ./... -tags=integration -v
```

---

## Security Notes

- **RULE-SEC-01**: No credentials in source code. All secrets via environment variables.
- **RULE-SEC-02**: Production MUST use mTLS (`RTSA_TLS_ENABLED=true` with valid certs).
- **RULE-SEC-03**: All track data from Redpanda is validated before cache insertion.
- **Classification filtering** is the last line of defence before data leaves this service.
- Never log raw track positions, sensor payloads, or classification levels.

---

## Dependencies

- `github.com/twmb/franz-go` — Kafka/Redpanda client
- `google.golang.org/grpc` — gRPC server
- `github.com/prometheus/client_golang` — Metrics
- `github.com/arvinddhasmana/RTSA_VS_Opus/gen/go` — Generated protobuf types
