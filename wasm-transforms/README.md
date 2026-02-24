<!-- CLASSIFICATION: UNCLASSIFIED -->

# RTSA Wasm Data Transforms

> **Classification**: UNCLASSIFIED
> **Module**: 06 — Wasm Data Transforms
> **Phase**: P1 (Ingestion)

Redpanda broker-side WebAssembly transforms that validate message schema
compliance and enforce classification headers on all sensor, track, alert, and
feedback topics **before** consumer services read them.

## Transforms

| Transform | Input Topics | Validates |
|---|---|---|
| `sensor-validator` | `sensors.*` | Headers, classification, `SensorObservation` proto, `sensor_id`, `sensor_type` |
| `track-validator` | `tracks.fused.*` | Headers, classification, `FusedTrack` proto, `track_id` |
| `alert-validator` | `alerts.anomaly.*` | Headers, classification, `AnomalyAlert` proto, `alert_id` |
| `feedback-validator` | `feedback.operator.*` | Headers, classification, `OperatorFeedback` proto, `feedback_id` |

## Required Headers

Every message must carry:

| Header | Description |
|---|---|
| `rtsa-classification` | One of: `UNCLASSIFIED`, `PROTECTED_A`, `PROTECTED_B`, `PROTECTED_C`, `SECRET` |
| `rtsa-source-service` | Name of the producing service |
| `rtsa-trace-id` | Distributed trace identifier |
| `rtsa-timestamp` | ISO 8601 message timestamp |
| `rtsa-schema-version` | Schema version string |

Messages missing any header, or carrying an invalid classification value, are
routed to the appropriate DLQ topic.

## Validation Outcomes

- **Pass**: `rtsa-validated: true` header added; message forwarded to default output topic.
- **Fail**: `rtsa-validation-error: <reason>` header added; message routed to DLQ (`dlq.sensors`, `dlq.tracks`, `dlq.alerts`, `dlq.feedback`).

## Building

```bash
# Build all transforms
cd wasm-transforms
for dir in sensor-validator track-validator alert-validator feedback-validator; do
    (cd "$dir" && make build)
done
```

Each Makefile compiles with `GOOS=wasip1 GOARCH=wasm`.

## Testing

Tests run in native Go using mock implementations of the SDK interfaces — no
Wasm runtime required:

```bash
cd wasm-transforms
for dir in sensor-validator track-validator alert-validator feedback-validator; do
    (cd "$dir" && make test)
done
```

## Deploying

```bash
# From the repository root:
RTSA_REDPANDA_BROKERS=localhost:9092 ./scripts/deploy-transforms.sh
```

## Configuration

| Environment Variable | Default | Description |
|---|---|---|
| `RTSA_REDPANDA_BROKERS` | `localhost:9092` | Redpanda broker address(es) |
