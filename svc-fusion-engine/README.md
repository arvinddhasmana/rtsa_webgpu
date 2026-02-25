<!-- CLASSIFICATION: UNCLASSIFIED -->

# svc-fusion-engine

Multi-Source Fusion Engine for the RTSA system. Consumes sensor observations from all `sensors.*` Redpanda topics, correlates them using spatial/temporal gating and weighted scoring, maintains fused track state with a Kalman filter, and produces `FusedTrack` events.

## Overview

| Component | Description |
|---|---|
| **Gating Filter** | Haversine distance + time-delta gate per entity type |
| **Correlation Scorer** | 4-component weighted score (pos 0.35, vel 0.25, type 0.20, temporal 0.20) |
| **Kalman Filter** | 4-state constant-velocity model `[lat, lon, vN, vE]` |
| **Track Manager** | Thread-safe in-memory track state (sync.RWMutex) |
| **Merger** | Bidirectional correlation for track merging |
| **Stale Monitor** | Background goroutine: ACTIVE→STALE (60s), STALE→DROPPED (300s) |

## Configuration

All configuration is via environment variables:

| Variable | Default | Description |
|---|---|---|
| `RTSA_FUSION_INPUT_TOPICS` | `sensors.radar,sensors.ew,...` | Comma-separated input topics |
| `RTSA_FUSION_CONSUMER_GROUP` | `fusion-engine` | Kafka consumer group |
| `RTSA_FUSION_GATE_SURFACE_DISTANCE` | `5.0` | Surface gate distance (NM) |
| `RTSA_FUSION_GATE_SURFACE_TIME` | `30` | Surface gate time delta (s) |
| `RTSA_FUSION_GATE_AIR_DISTANCE` | `20.0` | Air gate distance (NM) |
| `RTSA_FUSION_GATE_AIR_TIME` | `15` | Air gate time delta (s) |
| `RTSA_FUSION_GATE_SUB_DISTANCE` | `2.0` | Subsurface gate distance (NM) |
| `RTSA_FUSION_GATE_SUB_TIME` | `60` | Subsurface gate time delta (s) |
| `RTSA_FUSION_WEIGHT_POSITION` | `0.35` | Position correlation weight |
| `RTSA_FUSION_WEIGHT_VELOCITY` | `0.25` | Velocity correlation weight |
| `RTSA_FUSION_WEIGHT_TYPE` | `0.20` | Entity type correlation weight |
| `RTSA_FUSION_WEIGHT_TEMPORAL` | `0.20` | Temporal correlation weight |
| `RTSA_FUSION_AUTO_THRESHOLD` | `0.85` | Auto-correlate threshold |
| `RTSA_FUSION_TENTATIVE_THRESHOLD` | `0.60` | Tentative correlate threshold |
| `RTSA_FUSION_STALE_TIMEOUT` | `60` | Seconds until ACTIVE→STALE |
| `RTSA_FUSION_DROP_TIMEOUT` | `300` | Seconds until STALE→DROPPED |
| `RTSA_FUSION_STALE_CHECK_INTERVAL` | `10` | Stale check interval (s) |
| `RTSA_FUSION_OUTPUT_PREFIX` | `tracks.fused` | Output topic prefix |
| `RTSA_REDPANDA_BROKERS` | `localhost:9092` | Redpanda broker addresses |
| `RTSA_LOG_LEVEL` | `info` | Log level (`debug`\|`info`) |

## Output Topics

| Entity Type | Topic |
|---|---|
| SURFACE | `tracks.fused.surface` |
| AIR | `tracks.fused.air` |
| SUBSURFACE | `tracks.fused.subsurface` |
| LAND | `tracks.fused.land` |
| CYBER | `tracks.fused.cyber` |

## Running Locally

```bash
export RTSA_REDPANDA_BROKERS=localhost:9092
export RTSA_REDPANDA_TLS_ENABLED=false
export RTSA_LOG_LEVEL=debug
go run ./cmd/fusion-engine
```

## Testing

```bash
# Unit tests (≥80% coverage on domain logic)
go test ./internal/domain/... -race -count=1 -coverprofile=coverage.out
go tool cover -func=coverage.out

# All unit tests
go test ./... -race -count=1

# Integration tests (requires Redpanda running)
RTSA_INTEGRATION_TESTS=true RTSA_REDPANDA_BROKERS=localhost:19092 \
  go test -tags integration ./internal/integration/... -v
```

## Architecture

```
SensorObservation (Redpanda sensors.*)
        │
        ▼
  GatingFilter.FindCandidates()
  (Haversine distance + time delta)
        │
        ├── No candidates ──────────► CreateTrack → NEW event
        │
        └── Candidates exist
                │
                ▼
        CorrelationScorer.Score()
        (weighted 4-component score)
                │
                ├── score ≥ 0.85 ──► UpdateTrack (auto-correlate)
                ├── score ≥ 0.60 ──► UpdateTrack (tentative)
                └── score < 0.60 ──► CreateTrack → NEW event
                        │
                        ▼
               KalmanFilter.Predict() + Update()
                        │
                        ▼
               FindMergeCandidates() ──► MergeTracks()
                        │
                        ▼
               FusedTrack → tracks.fused.{entity_type}
```
