<!-- CLASSIFICATION: UNCLASSIFIED -->
# Phase 2 — Backend Integration

> **Document**: v4 Implementation — Phase 2
> **Version**: 1.0
> **Classification**: UNCLASSIFIED
> **Status**: Not Started
> **Prerequisite Phases**: Phase 0 (Foundation)
> **Parallel With**: Phase 1 (Core Rendering)
> **Architecture Reference**: `docs/architecture/v1/RTSA_WebGPU_Architecture_v1.md` §3, §9

---

## 1. Objective

Build the backend hot-path pipeline — Go FlatBuffer serializer, Go WebTransport server, Data Worker WebTransport client, and priority shedding — to deliver real-time track updates to the browser via QUIC datagrams.

---

## 2. Deliverables

| # | Deliverable | Description |
|---|---|---|
| B2-1 | FlatBuffer schema | `.fbs` files in `proto/rtsa/flatbuf/v1/` |
| B2-2 | Go FlatBuffer serializer | `pkg/flatbuf/` — Protobuf → FlatBuffer conversion |
| B2-3 | Go WebTransport server | `pkg/webtransport/` — session management + datagram sending |
| B2-4 | Session authentication | JWT validation for WebTransport sessions |
| B2-5 | Classification filtering | Server-side track filtering by operator clearance |
| B2-6 | Priority shedding | Load-based track priority dropping |
| B2-7 | Data Worker integration | Browser WebTransport client + Wasm decoder pipeline |
| B2-8 | Connection lifecycle | Connect, reconnect with exponential backoff, health reporting |
| B2-9 | Monitoring | OpenTelemetry metrics for WebTransport server |
| B2-10 | Tests | Go unit tests, integration tests, cross-language round-trip |

---

## 3. Detailed Tasks

### B2-1: FlatBuffer Schema

- Create `proto/rtsa/flatbuf/v1/track_update.fbs` per `docs/sdlc_guidelines/08_tech_specific/flatbuffers_guidelines.md` §2.3
- 128-byte GPU-aligned record layout
- Namespace: `rtsa.flatbuf.v1`
- Generate Go and Rust code from `.fbs`

### B2-2: Go FlatBuffer Serializer

- Implement `pkg/flatbuf/serializer.go` per `flatbuffers_guidelines.md` §3
- Convert Protobuf `TrackUpdate` → FlatBuffer `TrackUpdate`
- Reuse `flatbuffers.Builder` across calls (zero allocation per message)
- Hash `track_id` to `track_id_hash` using FNV-1a
- Map `track_type` + `threat_level` → `icon_index` via lookup table
- Project last 5 position history entries to trail ring buffer

### B2-3: Go WebTransport Server

- Implement `pkg/webtransport/server.go` per `docs/sdlc_guidelines/08_tech_specific/webtransport_guidelines.md` §3
- Use `quic-go/webtransport-go`
- Listen on configurable port (default 4443)
- mTLS with CSE-approved cipher suites
- Origin validation from allowed origins list
- Max concurrent sessions configurable

### B2-4: Session Authentication

- JWT token validation on WebTransport upgrade
- Token includes: `operator_id`, `clearance_level`, `expiry`
- Short-lived tokens (5 min TTL), refreshed via gRPC cold path
- Reject sessions with expired or invalid tokens

### B2-5: Classification Filtering

- Server-side filter: `track.ClassificationLevel > claims.ClearanceLevel → drop`
- Reference: `webtransport_guidelines.md` §7.4
- Log filtered track count (not content) for audit

### B2-6: Priority Shedding

- Implement per `webtransport_guidelines.md` §5
- Priority levels: P0 (Hostile) → P3 (Friendly)
- Under congestion: always deliver P0/P1, drop P2/P3
- Detect congestion via QUIC flow control backpressure
- Metric: `webtransport_datagrams_dropped_total{reason="congestion"}`

### B2-7: Data Worker Integration

- Connect Data Worker to real WebTransport server
- Replace mock data writer (from Phase 0) with real Wasm decoder pipeline
- Datagram → Wasm decode → write to SAB slot
- Batch support: up to 9 records per datagram (1152 bytes < MTU)

### B2-8: Connection Lifecycle

- Implement reconnection with exponential backoff per `webtransport_guidelines.md` §6.3
- Report connection status to main thread via `postMessage`
- Notify SolidJS `ConnectionIndicator` component

### B2-9: Monitoring

- Add OpenTelemetry metrics per `webtransport_guidelines.md` §8
- Add audit events per `webtransport_guidelines.md` §7.5
- Session open/close events → Redpanda audit topic

### B2-10: Tests

| Test | Type | Framework |
|---|---|---|
| FlatBuffer serializer round-trip | Unit | Go `testing` |
| FlatBuffer field offset verification | Unit | Go `testing` |
| WebTransport session management | Unit | Go `testing` + mock |
| Classification filtering | Unit | Go `testing` |
| Priority shedding | Unit | Go `testing` |
| Cross-language round-trip (Go → Wasm) | Integration | CI pipeline |
| Full WebTransport E2E | Integration | Go test server + browser |

---

## 4. Integration Points

### 4.1 Redpanda Consumer

The WebTransport server consumes from the `track-fused` Redpanda topic, the same topic used by the existing gRPC cold-path track service. It adds a new consumer group without modifying existing consumers.

### 4.2 Docker Compose Addition

```yaml
# deploy/docker-compose.services.yml (addition)
svc-webtransport:
  build:
    context: .
    dockerfile: Dockerfile.service
    args:
      SERVICE: svc-webtransport
  ports:
    - "4443:4443/udp"   # QUIC
    - "4443:4443/tcp"   # HTTP/3 fallback
  environment:
    - WEBTRANSPORT_LISTEN_ADDR=:4443
    - REDPANDA_BROKERS=redpanda:9092
  depends_on:
    - redpanda
```

### 4.3 Envoy / Proxy Configuration

- The existing Envoy proxy must be configured to forward QUIC traffic to the WebTransport service
- Alternatively, the WebTransport server can be exposed directly (it handles its own TLS)
- Reference: `docs/architecture/deployment_architecture.md` §9.5

---

## 5. Done Gate

| Criteria | Verification |
|---|---|
| FlatBuffer `.fbs` schema compiles (Go + Rust) | `flatc` generation in CI |
| Go serializer produces valid 128-byte records | Round-trip unit test |
| WebTransport server accepts connections with valid JWT | Integration test |
| WebTransport server rejects invalid JWT | Integration test |
| Classification filtering drops high-classification tracks | Unit test |
| Priority shedding drops low-priority under congestion | Unit test |
| Data Worker receives datagrams and writes to SAB | E2E test |
| Reconnection recovers after server restart | E2E test |
| OpenTelemetry metrics visible in Grafana | Manual verification |
| Audit events appear in Redpanda audit topic | Integration test |
| Go test coverage ≥ 80% on new code | `go test -cover` |
