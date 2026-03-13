# Plan: Header Controls Migration + Sensor Deep Diagnostic View

## Context

- Repo: rtsa_webgpu, web-cop-gpu SolidJS + WebGPU frontend
- Current: AppShell has a 14rem left toolbar with RoleSelector + DashboardSelector + ConnectionIndicator
- Current: SensorHealthDashboard has its own DashboardSidebar (260px) for health filters
- Target 1: Move Role/Dashboard/Connection to top header bar; remove left toolbar pane
- Target 2: Click sensor card → open Deep Down Diagnostic view

## Architecture

### Current Layout

```
[Top ClassificationBanner]
[Left Toolbar 14rem: Role + Dashboard + Connection] | [Centre: SensorHealthDashboard: [Sidebar 260px] + [Grid]]
[Bottom ClassificationBanner]
```

### Target Layout

```
[Top ClassificationBanner]
[Header Bar: Role | Dashboard | ——————————————— Connection]
[Centre: SensorHealthDashboard: [Sidebar 260px] + [Grid ↔ DiagnosticView]]
[Bottom ClassificationBanner]
```

---

## Phase A — Header Controls Migration

### Step 1 — `src/components/shell/AppShell.tsx`

- Add `headerBar?: JSX.Element` to `AppShellProps`
- Render full-width horizontal header strip (≈52px, flex row) between top ClassificationBanner and main content row
- REMOVE the left toolbar pane (the `data-testid="app-toolbar"` div and `toolbar` prop) entirely
- Keep: canvas, rightPanel, bottomPanel, overlay props unchanged
- Add `data-testid="app-header"` to new header strip

### Step 2 — `src/App.tsx`

- Replace `toolbar={...}` with `headerBar={...}`
- Horizontal layout: left group → `RoleSelector` + `DashboardSelector`; right group → `ConnectionIndicator`
- Remove `toolbar` prop usage

### Step 3 — `tests/components/AppShell.test.tsx`

- Remove `app-toolbar` testId tests
- Add `app-header` testId test; test `headerBar` slot content renders inside header

---

## Phase B — Sensor Deep Diagnostic View

### Step 1 — `src/signals/sensor-filters.ts`

Add:

```ts
export const [selectedSensor, setSelectedSensor] =
  createSignal<SensorStatus | null>(null);
```

### Step 2 — `src/services/sensor-health.ts`

Add `SensorDiagnosticDetail` interface (extends `SensorStatus`):

```ts
export interface SensorDiagnosticDetail extends SensorStatus {
  dlqBreakdown: { reason: string; count: number }[];
  recentEvents: {
    timeUtc: string;
    event: string;
    severity: "info" | "warn" | "error";
  }[];
  subSensors: {
    id: string;
    status: "ACTIVE" | "DEGRADED" | "INACTIVE";
    location: string;
    lastSeenSeconds: number;
  }[];
  latencyMs: number;
  throughputHistory: number[]; // 20 data points
}
```

Add `fetchSensorDiagnostic(sensor: SensorStatus): Promise<SensorDiagnosticDetail>`:

- Enriches existing sensor data
- Deterministic seed-based mock for DLQ breakdown / events / sub-sensors (same pattern as sparklines)
- Function signature is API-ready for a future real `GetSensorDiagnostic` RPC — no backend changes needed now

### Step 3 — NEW `src/components/dashboard/SensorDiagnosticView.tsx`

Full-pane component replacing SensorGrid when a sensor is selected. Sections (top-to-bottom):

1. **Header row**: `«` Back button → `setSelectedSensor(null)` | sensor ID breadcrumb | type badge | status badge | "Last updated: ..."
2. **Metrics row**: 4 metric cards — Throughput obs/s | DLQ Count (red if >50) | Validation % | Latency ms
3. **Throughput history**: SVG area chart (20 data points), coloured by status, expected-range band overlay
4. **Sub-Sensors table**: ID | Status coloured badge | Location | Last Seen — hidden if empty
5. **DLQ Breakdown table**: Rejection Reason | Count | severity-coloured bar
6. **Recent Events timeline**: Zulu timestamp | severity icon | event text

Attributes:

- Root: `data-testid="sensor-diagnostic-view"`
- Back button: `data-testid="diagnostic-back-btn"`
- Uses `createResource(() => props.sensor, fetchSensorDiagnostic)`
- Classification header: `// CLASSIFICATION: UNCLASSIFIED`

### Step 4 — `src/components/dashboard/SensorStatusCard.tsx`

- Add `onSelect?: (sensor: SensorStatus) => void` prop (never destructure props — SolidJS reactivity)
- Add `onClick={() => props.onSelect?.(props.sensor)}` to card root div
- Add `cursor: "pointer"` to existing inline styles
- Add `data-testid={`sensor-card-${props.sensor.sensorId}`}` on root div for E2E

### Step 5 — `src/components/dashboard/SensorGrid.tsx`

- Add `onSensorSelect?: (sensor: SensorStatus) => void` to `SensorGridProps`
- Pass `onSelect={props.onSensorSelect}` to each `<SensorStatusCard />`

### Step 6 — `src/components/dashboard/SensorHealthDashboard.tsx`

- Import `selectedSensor`, `setSelectedSensor` from `../../signals/sensor-filters`
- Import `SensorDiagnosticView` from `./SensorDiagnosticView`
- Replace centre `<SensorGrid>` with:
  ```tsx
  <Show
    when={selectedSensor() !== null}
    fallback={
      <SensorGrid
        sensors={sensors() || []}
        onSensorSelect={setSelectedSensor}
      />
    }
  >
    <SensorDiagnosticView sensor={selectedSensor()!} />
  </Show>
  ```

---

## Phase C — Tests

### Unit Tests (Vitest)

**1. Update `tests/components/AppShell.test.tsx`**

- Remove: `app-toolbar` testId assertions
- Add: `app-header` testId renders; `headerBar` slot content rendered inside header

**2. Update `tests/components/SensorStatusCard.test.tsx`**

- Add: `onSelect` prop fires when card is clicked
- Add: `data-testid="sensor-card-RADAR-01"` present on rendered card

**3. Update `tests/components/SensorGrid.test.tsx`**

- Add: `onSensorSelect` propagated through to `SensorStatusCard` via `onSelect`

**4. NEW `tests/components/SensorDiagnosticView.test.tsx`**

- Renders sensor ID + type in header
- Back button click calls `setSelectedSensor(null)` (mock the signal)
- DLQ Breakdown section visible
- Recent Events section visible
- Sub-sensors section renders or is hidden when empty

### E2E Browser Tests (Playwright)

**5. Update `e2e/sensor-health.spec.ts`**

- Update `[data-testid="dashboard-selector"]` locators — now inside `[data-testid="app-header"]`
- Add: Role selector visible in header (`[data-testid="app-header"]` contains role-selector)
- Add: Clicking first `.sensor-card-hover` → `[data-testid="sensor-diagnostic-view"]` becomes visible
- Add: Diagnostic back button → diagnostic gone, sensor grid visible again
- Add: Screenshot `e2e/snapshots/sensor-health-with-header.png`

**6. NEW `e2e/sensor-diagnostic.spec.ts`**

```ts
test.describe("Sensor Diagnostic Deep Dive", () => {
  test("clicking a sensor card opens diagnostic view", ...);
  test("diagnostic view shows sensor ID and status", ...);
  test("DLQ breakdown section is visible", ...);
  test("recent events section is visible", ...);
  test("back button returns to sensor grid", ...);
  test("captures screenshot — diagnostic view", ...); // e2e/snapshots/sensor-diagnostic-full.png
  test("captures screenshot — back to grid",   ...); // e2e/snapshots/sensor-diagnostic-back.png
});
```

---

## Files

### Modified

| File                                                 | Change                                                 |
| ---------------------------------------------------- | ------------------------------------------------------ |
| `src/components/shell/AppShell.tsx`                  | Remove left toolbar pane; add `headerBar` slot         |
| `src/App.tsx`                                        | Use `headerBar`; horizontal layout of controls         |
| `src/signals/sensor-filters.ts`                      | Add `selectedSensor` signal                            |
| `src/services/sensor-health.ts`                      | Add `SensorDiagnosticDetail` + `fetchSensorDiagnostic` |
| `src/components/dashboard/SensorHealthDashboard.tsx` | Route grid ↔ diagnostic via `selectedSensor`           |
| `src/components/dashboard/SensorGrid.tsx`            | Add `onSensorSelect` prop                              |
| `src/components/dashboard/SensorStatusCard.tsx`      | Add `onSelect` prop + click handler + testid           |
| `tests/components/AppShell.test.tsx`                 | Update for `headerBar` prop                            |
| `tests/components/SensorStatusCard.test.tsx`         | Add click/onSelect test                                |
| `e2e/sensor-health.spec.ts`                          | Update selectors; add diagnostic flow tests            |

### Created

| File                                                | Purpose                               |
| --------------------------------------------------- | ------------------------------------- |
| `src/components/dashboard/SensorDiagnosticView.tsx` | Full diagnostic pane component        |
| `tests/components/SensorDiagnosticView.test.tsx`    | Unit tests for diagnostic view        |
| `e2e/sensor-diagnostic.spec.ts`                     | E2E diagnostic suite with screenshots |

---

## Decisions

- **DashboardSidebar stays**: Only the AppShell left toolbar is removed. The health-filter sidebar inside `SensorHealthDashboard` is unchanged.
- **Diagnostic is full-pane, not modal**: Replaces the sensor grid in the centre area to maximise diagnostic real-estate and avoid overlay scroll issues.
- **Diagnostic data is seeded mock**: DLQ breakdown, events, and sub-sensors use deterministic seed-based mocks (same pattern as sparklines). `fetchSensorDiagnostic` signature is API-ready for a future real `GetSensorDiagnostic` gRPC RPC.
- **No backend changes**: All changes are within `web-cop-gpu/`.
- **No props destructuring**: Per SolidJS standards — always use `props.xxx`.

---

## Phase D — Backend `GetSensorDiagnostic` RPC (Full E2E)

### What this is

`GetSensorDiagnostic` is a new unary gRPC RPC added to the existing `IngestionService` in
`proto/rtsa/ingestion/v1/ingestion_service.proto`. It returns deep operational data for a
single named sensor: throughput time-series, DLQ rejection breakdown by reason,
per-observation latency, sub-sensor listing, and recent event log.

Currently, `ListSensorStatuses` only returns aggregate counters (`total_received`,
`total_rejected`, `events_per_second`, `last_observation_time`). The `IngestionHandler` in
each of the 6 ingestion services already tracks per-sensor counters in a `sync.Map`
(`sensorState`), but **does not** persist DLQ reasons, latency samples, or event history.
This gap is what the full backend implementation closes.

The frontend `fetchSensorDiagnostic()` currently returns seed-based mock data. Once the RPC
exists, the frontend replaces the mock body with a real gRPC-Web call — the function
signature stays identical.

---

### D1 — Proto: Add RPC + messages to `ingestion_service.proto`

```proto
// Unary: get in-depth diagnostic data for a single sensor
// Deadline: 10s
rpc GetSensorDiagnostic(GetSensorDiagnosticRequest) returns (SensorDiagnosticResponse);
```

New messages:

```proto
message GetSensorDiagnosticRequest {
  string sensor_id = 1;
  // Number of throughput history samples to return (default: 20, max: 60)
  int32 history_samples = 2;
  // Number of recent events to return (default: 20, max: 100)
  int32 recent_events_limit = 3;
}

message ThroughputSample {
  google.protobuf.Timestamp sampled_at = 1;
  double events_per_second = 2;
}

message DLQReasonCount {
  string reason = 1;    // e.g. "invalid_timestamp", "coordinates_out_of_range"
  int64  count  = 2;
}

enum SensorEventSeverity {
  SENSOR_EVENT_SEVERITY_UNSPECIFIED = 0;
  SENSOR_EVENT_SEVERITY_INFO  = 1;
  SENSOR_EVENT_SEVERITY_WARN  = 2;
  SENSOR_EVENT_SEVERITY_ERROR = 3;
}

message SensorEvent {
  google.protobuf.Timestamp occurred_at = 1;
  string                    event_text  = 2;
  SensorEventSeverity       severity    = 3;
}

message SubSensorStatus {
  string sensor_id        = 1;
  string status           = 2;   // "ACTIVE" | "DEGRADED" | "INACTIVE"
  string location_label   = 3;   // free-text geo label
  google.protobuf.Timestamp last_seen_at = 4;
}

message SensorDiagnosticResponse {
  // --- core fields (same as SensorStatusResponse) ---
  string                       sensor_id             = 1;
  rtsa.common.v1.SensorType    sensor_type           = 2;
  bool                         connected             = 3;
  int64                        total_received        = 4;
  int64                        total_accepted        = 5;
  int64                        total_rejected        = 6;
  google.protobuf.Timestamp    last_observation_time = 7;
  double                       events_per_second     = 8;
  // --- new diagnostic fields ---
  double                       latency_ms            = 9;   // EWMA of ingest-to-produce latency
  double                       validation_pass_rate  = 10;  // 0.0–100.0 percent
  repeated ThroughputSample    throughput_history    = 11;
  repeated DLQReasonCount      dlq_breakdown         = 12;
  repeated SensorEvent         recent_events         = 13;
  repeated SubSensorStatus     sub_sensors           = 14;
  optional SensorCoverage      coverage              = 15;
}
```

Run `buf generate` after editing the proto to regenerate Go + TS stubs.

---

### D2 — Expand `sensorState` in all 6 ingestion handlers

Each ingestion service has its own `IngestionHandler` in `internal/handler/ingestion.go`.
The `sensorState` struct currently holds only `totalReceived atomic.Int64` and
`lastObsTime atomic.Value`. Expand it (same pattern in all 6 services):

```go
type sensorState struct {
    totalReceived  atomic.Int64
    totalAccepted  atomic.Int64
    totalRejected  atomic.Int64
    lastObsTime    atomic.Value        // time.Time
    latencyEwmaNs  atomic.Int64       // EWMA numerator × 1000 for int arithmetic
    dlqReasons     sync.Map           // map[string]*atomic.Int64 (reason → count)
    throughputMu   sync.Mutex
    throughputRing []throughputSample // fixed-capacity ring buffer, cap=60
    throughputHead int
    eventsMu       sync.Mutex
    eventRing      []sensorEvent      // fixed-capacity ring buffer, cap=100
    eventHead      int
}

type throughputSample struct {
    sampledAt time.Time
    eps       float64
}

type sensorEvent struct {
    occurredAt time.Time
    text       string
    severity   ingestionv1.SensorEventSeverity
}
```

Key changes to `IngestSingleObservation`:

1. **Before validate** — record `t0 := time.Now()`.
2. **On rejection** — `ss.totalRejected.Add(1)`; extract reason string; increment
   `ss.dlqReasons` counter for that reason string; append event `WARN "observation rejected: <reason>"`.
3. **After produce** — compute `latencyNs := time.Since(t0).Nanoseconds()`; update EWMA in
   `ss.latencyEwmaNs`; append event `INFO "observation accepted: <obs_id>"`; `ss.totalAccepted.Add(1)`.

Add a background goroutine per sensor (or a shared ticker in the handler) that samples
`ss.totalReceived / elapsedSeconds` every 30 seconds and appends to `throughputRing`.

---

### D3 — Implement `GetSensorDiagnostic` in all 6 handlers

```go
func (h *IngestionHandler) GetSensorDiagnostic(
    ctx context.Context,
    req *ingestionv1.GetSensorDiagnosticRequest,
) (*ingestionv1.SensorDiagnosticResponse, error) {
    if req.GetSensorId() == "" {
        return nil, status.Error(codes.InvalidArgument, "sensor_id is required")
    }

    raw, ok := h.sensors.Load(req.GetSensorId())
    if !ok {
        return nil, status.Errorf(codes.NotFound, "sensor %q not found", req.GetSensorId())
    }
    ss := raw.(*sensorState)
    lastTime := ss.lastObsTime.Load().(time.Time)

    // --- throughput history ---
    limit := int(req.GetHistorySamples())
    if limit <= 0 || limit > 60 { limit = 20 }
    samples := ss.snapshotThroughput(limit)

    // --- DLQ breakdown ---
    var dlqBreakdown []*ingestionv1.DLQReasonCount
    ss.dlqReasons.Range(func(k, v interface{}) bool {
        dlqBreakdown = append(dlqBreakdown, &ingestionv1.DLQReasonCount{
            Reason: k.(string),
            Count:  v.(*atomic.Int64).Load(),
        })
        return true
    })

    // --- recent events ---
    evtLimit := int(req.GetRecentEventsLimit())
    if evtLimit <= 0 || evtLimit > 100 { evtLimit = 20 }
    events := ss.snapshotEvents(evtLimit)

    // --- latency EWMA ---
    latencyMs := float64(ss.latencyEwmaNs.Load()) / 1e6

    // --- validation pass rate ---
    total := ss.totalReceived.Load()
    var passRate float64 = 100.0
    if total > 0 {
        passRate = float64(ss.totalAccepted.Load()) / float64(total) * 100.0
    }

    resp := &ingestionv1.SensorDiagnosticResponse{
        SensorId:            req.GetSensorId(),
        SensorType:          commonv1.SensorType_SENSOR_TYPE_RADAR,  // per-service constant
        Connected:           time.Since(lastTime) < 30*time.Second,
        TotalReceived:       ss.totalReceived.Load(),
        TotalAccepted:       ss.totalAccepted.Load(),
        TotalRejected:       ss.totalRejected.Load(),
        EventsPerSecond:     eventsPerSecond(ss),
        LatencyMs:           latencyMs,
        ValidationPassRate:  passRate,
        ThroughputHistory:   samples,
        DlqBreakdown:        dlqBreakdown,
        RecentEvents:        events,
        Coverage:            h.coverage,
    }
    if !lastTime.IsZero() {
        resp.LastObservationTime = timestamppb.New(lastTime)
    }
    return resp, nil
}
```

All 6 ingestion services (`svc-radar`, `svc-aw`, `svc-elint`, `svc-isr`, `svc-ais`,
`svc-cyber`) get the same implementation differing only in the `SensorType` constant.
Extract the shared tracking logic into `pkg/ingestion/sensorstate.go` to avoid 6-way
duplication — each service embeds `*ingestion.SensorStateTracker`.

---

### D4 — `buf generate`: Regenerate Go + TS stubs

```bash
cd /home/arvind/workspace/rtsa_webgpu
buf generate
```

Produces:

- `gen/go/rtsa/ingestion/v1/ingestion_service_grpc.pb.go` — new `GetSensorDiagnostic` server interface method and client call
- `gen/ts/rtsa/ingestion/v1/ingestion_service_connect.ts` — new TS client method

---

### D5 — Frontend: Replace mock with real gRPC-Web call

In `web-cop-gpu/src/services/sensor-health.ts`, replace the mock body of
`fetchSensorDiagnostic` with:

```typescript
export async function fetchSensorDiagnostic(
  sensor: SensorStatus,
): Promise<SensorDiagnosticDetail> {
  const target = ingestionTargetHeader(sensor.sensorType); // maps type → x-ingestion-target
  const opts = target
    ? { headers: new Headers({ "x-ingestion-target": target }) }
    : undefined;

  const resp = await client.getSensorDiagnostic(
    { sensorId: sensor.sensorId, historySamples: 20, recentEventsLimit: 20 },
    opts,
  );

  return {
    ...sensor,
    latencyMs: resp.latencyMs,
    throughputHistory: resp.throughputHistory.map((s) => s.eventsPerSecond),
    dlqBreakdown: resp.dlqBreakdown.map((d) => ({
      reason: d.reason,
      count: Number(d.count),
    })),
    recentEvents: resp.recentEvents.map((e) => ({
      timeUtc: e.occurredAt?.toDate().toISOString() ?? "",
      event: e.eventText,
      severity: protoSeverityToTs(e.severity),
    })),
    subSensors: resp.subSensors.map((s) => ({
      id: s.sensorId,
      status: s.status as "ACTIVE" | "DEGRADED" | "INACTIVE",
      location: s.locationLabel,
      lastSeenSeconds: s.lastSeenAt
        ? Math.round((Date.now() - s.lastSeenAt.toDate().getTime()) / 1000)
        : -1,
    })),
  };
}
```

---

### D6 — Envoy: No changes required

Envoy already routes `x-ingestion-target` to the correct ingestion cluster. The new
`GetSensorDiagnostic` RPC flows through the existing gRPC-Web transcoding config.
Verify in `deploy/envoy/envoy.yaml` that the ingestion cluster routes cover all gRPC
methods (they do — they route by cluster, not by method name).

---

### D7 — Simulator: Inject invalid observations to populate DLQ breakdown

The `sensor-health-demo.yaml` scenario already drives throughput. To populate
`dlq_breakdown` with varied rejection reasons, extend the simulator scenario:

```yaml
# Append to sensor-health-demo.yaml
invalid_injection:
  enabled: true
  rate: 0.04 # 4% of observations are intentionally invalid
  reasons:
    invalid_timestamp: 0.40 # 40% of invalid obs have future timestamp
    coordinates_out_of_range: 0.35 # 35% — lat/lon = 0,0
    missing_sensor_id: 0.15 # 15% — empty sensor_id
    schema_mismatch: 0.10 # 10% — wrong sensor_data type
```

Implement in `tools/simulator/internal/generator/anomaly.go` — an `InvalidObservationInjector`
that wraps the regular generator and probabilistically corrupts outgoing observations.
The injector applies only the field mutation; the existing gRPC sender sends it
unchanged, causing the ingestion validator to reject it and record the reason in `dlqReasons`.

---

### D8 — Tests

#### D8a — Unit tests: handler `GetSensorDiagnostic` (all 6 services)

In `svc-radar-ingestion/internal/integration/ingestion_integration_test.go`, add:

```go
// TestGetSensorDiagnostic_AfterIngestCycle validates that after sending N valid and M
// invalid observations, GetSensorDiagnostic returns correct aggregates.
func TestGetSensorDiagnostic_AfterIngestCycle(t *testing.T) {
    h := newTestHandler(t, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET)
    ctx := context.Background()

    // Send 10 valid
    for i := 0; i < 10; i++ {
        obs := validObs("RADAR-DIAG-01")
        _, err := h.IngestSingleObservation(ctx, obs)
        require.NoError(t, err)
    }

    // Send 5 invalid (bad latitude)
    for i := 0; i < 5; i++ {
        obs := invalidLatObs("RADAR-DIAG-01")
        _, err := h.IngestSingleObservation(ctx, obs)
        require.NoError(t, err)  // gRPC err == nil; rejection in ack
    }

    resp, err := h.GetSensorDiagnostic(ctx,
        &ingestionv1.GetSensorDiagnosticRequest{
            SensorId: "RADAR-DIAG-01",
            HistorySamples: 20,
        })
    require.NoError(t, err)
    assert.Equal(t, int64(15), resp.TotalReceived)
    assert.Equal(t, int64(10), resp.TotalAccepted)
    assert.Equal(t, int64(5),  resp.TotalRejected)
    assert.InDelta(t, 66.67, resp.ValidationPassRate, 0.1)
    require.Len(t,  resp.DlqBreakdown, 1)
    assert.Equal(t, "validation failed: latitude out of range", resp.DlqBreakdown[0].Reason)
    assert.Equal(t, int64(5), resp.DlqBreakdown[0].Count)
    assert.NotEmpty(t, resp.RecentEvents)
}
```

#### D8b — Integration test: full-stack sensor-health pipeline

In `tests/integration/ingestion_test.go`, add `TestIT_DIAG01_GetSensorDiagnosticFullStack`:

1. Spin up handler with nil Redpanda producers (in-process integration style)
2. Ingest mixed valid/invalid observations with varied rejection reasons
3. Call `GetSensorDiagnostic`; assert `dlq_breakdown` contains all injected reason types
4. Assert `throughput_history` is populated after a sample tick
5. Assert `recent_events` count matches total ingest count (up to cap)

#### D8c — E2E test: full pipeline with simulator

In `tests/e2e/`, add `sensor_diagnostic_e2e_test.go` (requires `//go:build e2e`):

```
Simulator (sensor-health-demo.yaml)
  → gRPC IngestSingleObservation (10s of live traffic incl. invalid obs)
  → GetSensorDiagnostic("RADAR-NORTH-01") via gRPC client
  → Assert: connected=true, total_received>0, dlq_breakdown non-empty, recent_events non-empty
```

Start backend with: `bash scripts/cop-dev/start-backend.sh --scenario sensor-health-demo.yaml`

Run with: `go test -tags=e2e ./tests/e2e/...`

#### D8d — E2E browser test: diagnostic view with live data

Update `web-cop-gpu/e2e/sensor-diagnostic.spec.ts` — add a `test.describe("with live backend")`
block that:

1. Starts the dev server pointing at the live backend (`BASE_URL=http://localhost:5174`)
2. Waits for sensor cards to load (real data from the simulator)
3. Clicks the first sensor card
4. Asserts `[data-testid="sensor-diagnostic-view"]` visible
5. Asserts DLQ breakdown section has at least one row with a real reason
6. Asserts throughput history chart SVG has `path` elements
7. Screenshots: `e2e/snapshots/sensor-diagnostic-live-data.png`

Run with: `BASE_URL=http://localhost:5174 pnpm exec playwright test e2e/sensor-diagnostic.spec.ts --headed`

---

### D9 — Shared package: `pkg/ingestion/sensorstate.go`

To avoid duplicating the expanded `sensorState` struct and its methods across 6 services,
extract the tracking logic into a new shared package:

```
pkg/ingestion/
  sensorstate.go  — SensorStateTracker struct, ring buffers, EWMA, dlqReasons
  sensorstate_test.go — unit tests for ring buffer correctness, thread safety
```

Each ingestion handler embeds `*ingestion.SensorStateTracker` and calls its methods
(`RecordAccepted(latencyNs int64)`, `RecordRejected(reason string)`, etc.) rather than
duplicating the atomic/sync logic.

---

### D10 — Audit: emit diagnostic query event

`GetSensorDiagnostic` is a READ operation on sensor state. Emit an audit event per
ITSG-33 AU-2:

```go
h.auditEmitter.Emit(ctx, audit.AuditParams{
    EventType:    audit.EventResourceRead,
    ActorID:      operatorIDFromContext(ctx),  // extracted from JWT via gRPC metadata
    ActorType:    auditv1.ActorType_ACTOR_TYPE_OPERATOR,
    ResourceType: "sensor_diagnostic",
    ResourceID:   req.GetSensorId(),
    Action:       auditv1.AuditAction_AUDIT_ACTION_READ,
})
```

---

## Updated Files (full E2E)

### Additions from Phase D

| File                                                       | Change                                                              |
| ---------------------------------------------------------- | ------------------------------------------------------------------- |
| `proto/rtsa/ingestion/v1/ingestion_service.proto`          | Add `GetSensorDiagnostic` RPC + 6 new messages                      |
| `pkg/ingestion/sensorstate.go`                             | NEW — shared `SensorStateTracker` (ring buffers, EWMA, DLQ reasons) |
| `pkg/ingestion/sensorstate_test.go`                        | NEW — unit tests for SensorStateTracker                             |
| `svc-radar-ingestion/internal/handler/ingestion.go`        | Embed tracker; implement `GetSensorDiagnostic`                      |
| `svc-ais-ingestion/internal/handler/ingestion.go`          | Same                                                                |
| `svc-ew-ingestion/internal/handler/ingestion.go`           | Same                                                                |
| `svc-elint-ingestion/internal/handler/ingestion.go`        | Same                                                                |
| `svc-isr-ingestion/internal/handler/ingestion.go`          | Same                                                                |
| `svc-cyber-ingestion/internal/handler/ingestion.go`        | Same                                                                |
| `svc-*/internal/integration/ingestion_integration_test.go` | Add D8a tests (×6 services)                                         |
| `tools/simulator/internal/generator/anomaly.go`            | Add `InvalidObservationInjector`                                    |
| `tools/simulator/scenarios/sensor-health-demo.yaml`        | Add `invalid_injection` block                                       |
| `tests/integration/ingestion_test.go`                      | Add `TestIT_DIAG01_GetSensorDiagnosticFullStack`                    |
| `tests/e2e/sensor_diagnostic_e2e_test.go`                  | NEW — full-stack E2E test                                           |
| `web-cop-gpu/src/services/sensor-health.ts`                | Replace mock `fetchSensorDiagnostic` with real gRPC-Web call        |
| `web-cop-gpu/e2e/sensor-diagnostic.spec.ts`                | Add live-backend test block + screenshot                            |

---

## Updated Verification

```bash
# 1. Proto lint + generate
buf lint && buf generate

# 2. Shared package unit tests
go test ./pkg/ingestion/...

# 3. All ingestion service unit + integration tests
go test -tags=integration ./svc-radar-ingestion/...
go test -tags=integration ./svc-ais-ingestion/...
go test -tags=integration ./svc-ew-ingestion/...
go test -tags=integration ./svc-elint-ingestion/...
go test -tags=integration ./svc-isr-ingestion/...
go test -tags=integration ./svc-cyber-ingestion/...

# 4. Full-stack integration tests (requires Redpanda testcontainers)
go test -tags=integration ./tests/integration/...

# 5. Frontend unit tests
cd web-cop-gpu && pnpm test

# 6. Start full backend with simulator (sensor-health-demo scenario)
bash scripts/cop-dev/start-backend.sh --scenario sensor-health-demo.yaml

# 7. E2E Go (requires running backend)
go test -tags=e2e ./tests/e2e/...

# 8. E2E browser tests (requires running backend + dev server)
cd web-cop-gpu
BASE_URL=http://localhost:5174 pnpm exec playwright test --headed
# Screenshots saved to e2e/snapshots/
```

All tests must pass. No silent failures. Screenshots must be committed as proof.

---

## Exclusions (out of scope)

- Sensor Coverage Map overlay (UC017 §6)
- Live toast notification on sensor-goes-offline transition (UC017 §7a)

---

## Verification

```bash
# 1. Unit tests
cd web-cop-gpu && pnpm test

# 2. E2E browser tests (with headed Chromium, screenshots saved)
cd web-cop-gpu && pnpm exec playwright test

# 3. Full backend stack with health demo data (optional live validation)
bash scripts/cop-dev/start-backend.sh --scenario sensor-health-demo.yaml
```

All tests must pass. Screenshots must be written to `e2e/snapshots/`. No silent failures.

---

## Reference Documents

- `docs/business/usecases/UC017_sensor_health_monitoring.md`
- `docs/user_guide/sensor_operator/01_sensor_health.md`
- `docs/user_guide/sensor_operator/02_data_quality.md`
- `docs/implementation/v5/ui_images/diagnostic_deep_dive.png` ← open this for exact layout reference
- `docs/sdlc_guidelines/04_coding_standards/solidjs_standards.md`
- `docs/sdlc_guidelines/08_tech_specific/webgpu_guidelines.md`
- `docs/lets_get_started/data_simulation_in_dev_test_demo.md`
