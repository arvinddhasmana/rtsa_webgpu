<!-- CLASSIFICATION: UNCLASSIFIED -->

# svc-webtransport

RTSA **hot-path** fan-out service. It bridges the fused-track event stream to the
browser Common Operating Picture (COP) with the lowest possible latency:

```
Redpanda (tracks.fused.*)  ──►  svc-webtransport  ──►  WebTransport / QUIC datagrams  ──►  SolidJS + WebGPU COP
     Protobuf TrackUpdate         decode + serialize            128-byte records
```

- **Consumes** Protobuf `rtsa.entity.v1.TrackUpdate` messages from the
  `tracks.fused` topic(s) via the shared `pkg/redpanda` consumer.
- **Serializes** each update into a fixed **128-byte, GPU-aligned** binary record
  using `pkg/flatbuf` — the exact record the Wasm decoder writes into the
  browser `SharedArrayBuffer`.
- **Broadcasts** records to every authenticated WebTransport session using
  `pkg/webtransport` (JWT auth, clearance filtering, priority shedding, QUIC
  datagram batching).

A single consumer goroutine owns the stateful serializer; fan-out to each
session is **non-blocking** — a saturated browser drops records rather than
back-pressuring the hot path for every other operator.

## Configuration

All configuration is via `RTSA_*` environment variables. **No secrets are ever
embedded in source.** The JWT signing secret is injected from Key Vault through
the Secrets Store CSI driver and the process **fails closed** if it is missing.

| Variable                    | Default                   | Purpose                                     |
| --------------------------- | ------------------------- | ------------------------------------------- |
| `RTSA_WT_JWT_SECRET`        | _(required)_              | HMAC secret validating operator JWTs        |
| `RTSA_WT_LISTEN_ADDR`       | `:4443`                   | WebTransport/QUIC listener (UDP)            |
| `RTSA_TLS_SERVER_CERT`      | `./certs/dev/server.crt`  | TLS certificate (WebTransport requires TLS) |
| `RTSA_TLS_SERVER_KEY`       | `./certs/dev/server.key`  | TLS private key                             |
| `RTSA_TLS_CA_CERT`          | `./certs/dev/ca.crt`      | CA bundle for Redpanda mTLS                 |
| `RTSA_WT_ALLOWED_ORIGINS`   | _(empty = dev allow-all)_ | CSV of permitted browser origins            |
| `RTSA_WT_MAX_SESSIONS`      | `0` (unlimited)           | Concurrent session cap                      |
| `RTSA_WT_DATAGRAM_BATCH`    | `9`                       | 128-byte records per datagram (1..9)        |
| `RTSA_WT_SUBSCRIBER_BUFFER` | `1024`                    | Per-session broadcast channel buffer        |
| `RTSA_REDPANDA_BROKERS`     | `localhost:19092`         | CSV of broker addresses                     |
| `RTSA_REDPANDA_TLS_ENABLED` | `false`                   | Enable broker mTLS                          |
| `RTSA_CONSUMER_GROUP`       | `webtransport-service`    | Kafka consumer group                        |
| `RTSA_WT_TOPICS`            | `tracks.fused`            | CSV of fused-track topics                   |
| `RTSA_WT_START_OFFSET`      | `latest`                  | `latest` (hot path) or `earliest`           |
| `RTSA_HEALTH_PORT`          | `8081`                    | HTTP `/healthz` + `/readyz`                 |
| `RTSA_METRICS_PORT`         | `9090`                    | Prometheus `/metrics`                       |
| `RTSA_OTEL_ENDPOINT`        | _(empty = off)_           | OTLP collector gRPC endpoint                |
| `RTSA_LOG_LEVEL`            | `info`                    | `debug` / `info` / `warn` / `error`         |

## Endpoints

| Endpoint   | Port (default) | Description                         |
| ---------- | -------------- | ----------------------------------- |
| `/wt`      | 4443/udp       | WebTransport session upgrade (QUIC) |
| `/healthz` | 8081           | Liveness                            |
| `/readyz`  | 8081           | Readiness (broker reachable)        |
| `/metrics` | 9090           | Prometheus metrics                  |

Emitted metrics: `rtsa_wt_source_records_total`,
`rtsa_wt_source_decode_errors_total`, `rtsa_wt_source_broadcast_dropped_total`
plus the WebTransport server's session/datagram counters.

## Build & test

```bash
# From the workspace root (uses go.work):
go build ./svc-webtransport/...
go test  ./svc-webtransport/...

# Container image:
docker build -f svc-webtransport/Dockerfile -t rtsa/svc-webtransport .
```
