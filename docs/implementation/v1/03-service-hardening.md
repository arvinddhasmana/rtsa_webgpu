<!-- CLASSIFICATION: UNCLASSIFIED -->

# v1 Module 03 — Service Hardening

> **Module**: 03-service-hardening
> **Phase**: P2 (can run in parallel with Module 02)
> **Dependencies**: Module 01 (Infrastructure Fixes)
> **Agent**: `@greatest-ever-developer`
> **Estimated Effort**: 3 days
> **Traceability**: CR-SEC-002, CR-SEC-003, CR-SEC-008, CR-UI-004, CR-UI-007, FEAT-02, FEAT-13

---

## 1. Objective

Harden existing services by wiring mTLS in `svc-track`, replacing `noopAuditEmitter` stubs with real `pkg/audit.Emitter` in `svc-feedback` and `svc-query`, enforcing classification headers on all source files, adding sensor coverage health endpoints (CR-UI-004), and implementing NVG-compatible dark mode (CR-UI-007).

---

## 2. Task 1: Wire mTLS in svc-track (CR-SEC-002)

### Problem

`svc-track/cmd/track/main.go` line ~90 contains:

```go
// TODO: Replace insecure credentials with mTLS when certificates are available.
// RULE-SEC-02: production deployments MUST use mTLS.
```

When `cfg.TLSEnabled` is `false`, the server uses `grpc.Creds(insecure.NewCredentials())`. The TLS-enabled branch exists but is conditional and the TODO marks it as incomplete.

### Instructions

1. In `svc-track/cmd/track/main.go`, update the TLS configuration:
   - Load CA cert, server cert, and server key from config paths (`RTSA_TLS_CA_CERT`, `RTSA_TLS_SERVER_CERT`, `RTSA_TLS_SERVER_KEY`)
   - Create `tls.Config` with `tls.RequireAndVerifyClientCert` for mTLS
   - Use `credentials.NewTLS()` to create gRPC server option
   - Remove the TODO comment once implemented
   - Keep the `cfg.TLSEnabled` toggle for development convenience (dev can still use insecure), but log a WARNING when TLS is disabled

2. Add `loadTLSCredentials()` helper function if not already present — check if it exists in the TLS-enabled branch

3. Reference implementation: `svc-radar-ingestion` uses `pkg/redpanda.ConnectionOptions.TLSEnabled` — follow similar pattern for gRPC server TLS

### Test Cases

| #   | Test                                   | Expected                                    |
| --- | -------------------------------------- | ------------------------------------------- |
| T01 | TLS enabled with valid certs           | Server starts, accepts mTLS connections     |
| T02 | TLS enabled with missing cert files    | Startup fails with clear error              |
| T03 | TLS disabled (dev mode)                | Server starts with insecure, WARNING logged |
| T04 | Client without cert when mTLS required | Connection refused                          |

### Files to Modify

- `svc-track/cmd/track/main.go` — update TLS wiring, remove TODO
- `svc-track/internal/config/config.go` — ensure TLS config fields exist

---

## 3. Task 2: Wire Audit Emitter in svc-feedback (CR-SEC-003)

### Problem

`svc-feedback/cmd/feedback/main.go` defines a `noopAuditEmitter` struct:

```go
type noopAuditEmitter struct{ logger *zap.Logger }

func (n *noopAuditEmitter) EmitAudit(_ context.Context, event *auditv1.AuditEvent) error {
    n.logger.Info("audit event emitted (noop)", ...)
    return nil
}
```

This is used where a real audit emitter should produce events to the `audit.events` Redpanda topic.

### Instructions

1. In `svc-feedback/cmd/feedback/main.go`:
   - Import `"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/audit"`
   - Replace `&noopAuditEmitter{logger: logger}` with `audit.NewEmitter(rawProducer, "svc-feedback", logger)` where `rawProducer` is a `*redpanda.Producer` instance
   - If the service doesn't already create a `redpanda.Producer`, add one for audit events (or reuse the existing feedback producer if the API supports it)
   - Remove the `noopAuditEmitter` type definition and constructor

2. Register the audit producer in the shutdown manager

3. Verify audit events are emitted for:
   - Feedback submission (SUBMIT action)
   - Feedback validation pass/fail (VALIDATE action)
   - Anti-poisoning rejection (REJECT action)
   - Rate limit exceeded (RATE_LIMIT action)

### Test Cases

| #   | Test                                  | Expected                                     |
| --- | ------------------------------------- | -------------------------------------------- |
| T05 | Submit feedback → audit event emitted | Audit event produced to `audit.events` topic |
| T06 | Anti-poisoning reject → audit event   | Event contains rejection reason              |
| T07 | Rate limit exceeded → audit event     | Event contains operator_id and limit details |

### Files to Modify

- `svc-feedback/cmd/feedback/main.go` — replace noop with real emitter
- `svc-feedback/internal/handler/feedback.go` — verify audit calls exist (may already be wired via interface)

---

## 4. Task 3: Wire Audit Emitter in svc-query (CR-SEC-003)

### Problem

`svc-query/cmd/query/main.go` similarly defines a `noopAuditEmitter` for read operations. Per ITSG-33 AU-3, data access events should be audited.

### Instructions

1. Same pattern as Task 2 — replace `noopAuditEmitter` with `audit.NewEmitter()`
2. Audit events should capture:
   - Query execution (QUERY action with query type: tracks/anomalies/audit)
   - Classification-filtered results (classified data access attempt)
   - Query guardrail violations (time range exceeded, too many rows)

### Test Cases

| #   | Test                                  | Expected                                      |
| --- | ------------------------------------- | --------------------------------------------- |
| T08 | Execute track query → audit event     | Event with action=QUERY, resource_type=tracks |
| T09 | Classification filter applied → audit | Event captures clearance level used           |
| T10 | Guardrail violation → audit event     | Event with outcome=FAILURE                    |

### Files to Modify

- `svc-query/cmd/query/main.go` — replace noop with real emitter

---

## 5. Task 4: Classification Header Enforcement

### Problem

SDLC guideline `general_coding.md` requires every source file to start with a classification header comment as its first line: `// CLASSIFICATION: UNCLASSIFIED` (Go), `{/* CLASSIFICATION: UNCLASSIFIED */}` (TSX), `<!-- CLASSIFICATION: UNCLASSIFIED -->` (Markdown).

### Instructions

1. Run a scan across all source files to find any missing classification headers:

```bash
# Go files missing header
find . -name "*.go" -not -path "./gen/*" -not -path "*/vendor/*" | while read f; do
  head -1 "$f" | grep -q "CLASSIFICATION" || echo "MISSING: $f"
done

# TypeScript/TSX files missing header
find web-cop/src -name "*.ts" -o -name "*.tsx" | while read f; do
  head -1 "$f" | grep -q "CLASSIFICATION" || echo "MISSING: $f"
done

# Proto files missing header
find proto -name "*.proto" | while read f; do
  head -1 "$f" | grep -q "CLASSIFICATION" || echo "MISSING: $f"
done
```

2. Add the appropriate header to any files missing it:
   - Go: `// CLASSIFICATION: UNCLASSIFIED` as line 1
   - TypeScript/TSX: `// CLASSIFICATION: UNCLASSIFIED` as line 1
   - Proto: `// CLASSIFICATION: UNCLASSIFIED` as line 1
   - YAML: `# CLASSIFICATION: UNCLASSIFIED` as line 1
   - SQL: `-- CLASSIFICATION: UNCLASSIFIED` as line 1
   - Markdown: `<!-- CLASSIFICATION: UNCLASSIFIED -->` as line 1
   - Shell: After shebang, `# CLASSIFICATION: UNCLASSIFIED` as line 2

### Validation

```bash
# Zero output expected
find . -name "*.go" -not -path "./gen/*" -not -path "*/vendor/*" -exec head -1 {} \; | grep -cv "CLASSIFICATION"
```

---

## 6. Task 5: Sensor Coverage Health Endpoint (CR-UI-004)

### Problem

CR-UI-004 requires "Display sensor coverage and health status" in the COP UI. Currently no service exposes per-sensor health information that the UI can consume.

### Instructions

#### Backend: Add `GetSensorStatus` support

The `IngestionService` proto already defines a `GetSensorStatus` RPC:

```protobuf
rpc GetSensorStatus(GetSensorStatusRequest) returns (SensorStatusResponse);
```

1. **Check if this RPC is implemented** in `svc-radar-ingestion/internal/handler/ingestion.go`. If not, implement it:
   - Return sensor type, last observation timestamp, total observations processed, error count, and health status (HEALTHY/DEGRADED/OFFLINE)
   - Track these metrics in the handler (counters already likely exist for Prometheus)

2. **Add a `SensorHealthPanel` component** to `web-cop/src/components/`:
   - Calls `GetSensorStatus` for each of the 6 sensor types via Envoy
   - Displays a grid showing: sensor type icon, last update time, observations/sec rate, health indicator (green/yellow/red)
   - Auto-refreshes every 10 seconds
   - Add to `MainLayout.tsx` as a collapsible sidebar panel

3. **Add Envoy route** for `IngestionService` from each ingestion service (or create a sensor health aggregator service)

### Simpler Alternative (if aggregator is too complex for v1)

Instead of per-service gRPC calls, expose a **Prometheus-based sensor health summary**:

- Each ingestion service already emits Prometheus metrics
- Add a `SensorHealthPanel` that reads from Prometheus via the existing Grafana/Prometheus stack
- Display last-seen timestamps and throughput per sensor type

### Test Cases

| #   | Test                                | Expected                              |
| --- | ----------------------------------- | ------------------------------------- |
| T11 | GetSensorStatus for active sensor   | Returns HEALTHY with recent timestamp |
| T12 | GetSensorStatus for idle sensor     | Returns DEGRADED or OFFLINE           |
| T13 | SensorHealthPanel renders 6 sensors | All 6 sensor types shown with status  |

---

## 7. Task 6: NVG-Compatible Dark Mode (CR-UI-007)

### Problem

CR-UI-007 (SHOULD priority) requires NVG-compatible dark mode. The COP web app has existing dark theme support but not specifically NVG-optimized.

### Instructions

1. **Create `web-cop/src/styles/nvg-theme.css`** with:
   - Green-on-black palette (NVG standard): `#00FF00` text on `#000000` background
   - Muted green variants: `#003300` for secondary backgrounds, `#006600` for borders
   - No white, blue, or red elements (NVG compatibility)
   - Minimum 4.5:1 contrast ratio per WCAG 2.1 AA
   - Map tiles: dark/satellite map style or solid dark background

2. **Add theme toggle in `uiStore`**:

   ```typescript
   // CLASSIFICATION: UNCLASSIFIED
   interface UIState {
     theme: "light" | "dark" | "nvg";
     setTheme: (theme: "light" | "dark" | "nvg") => void;
   }
   ```

3. **Apply theme class to root element** in `App.tsx`:

   ```html
   <div className="{`app-root" theme-${theme}`}></div>
   ```

4. **Add theme selector** to the layout header — simple dropdown or toggle button

5. **Test with map tiles**: Ensure MapLibre GL uses a dark tile source when NVG mode is active. Options:
   - CartoDB Dark Matter tiles
   - MapTiler Dark style
   - Solid dark background with vector overlays only

### CSS Variables Pattern

```css
/* CLASSIFICATION: UNCLASSIFIED */
.theme-nvg {
  --color-bg-primary: #000000;
  --color-bg-secondary: #001100;
  --color-bg-tertiary: #002200;
  --color-text-primary: #00ff00;
  --color-text-secondary: #00cc00;
  --color-text-muted: #009900;
  --color-border: #006600;
  --color-accent: #00ff00;
  --color-danger: #ff6600; /* Amber — NVG-visible alternative to red */
  --color-warning: #ffcc00; /* Yellow-green */
  --color-success: #00ff00;
  --color-track-friendly: #00ff00;
  --color-track-hostile: #ff6600;
  --color-track-neutral: #00cccc;
  --color-track-unknown: #999999;
}
```

### Test Cases

| #   | Test                                 | Expected                                   |
| --- | ------------------------------------ | ------------------------------------------ |
| T14 | Toggle to NVG mode                   | All UI elements use green-on-black palette |
| T15 | NVG mode: no white/blue/red elements | Visual inspection pass                     |
| T16 | NVG mode: map tiles are dark         | Dark or satellite tiles displayed          |
| T17 | Theme persists across page reload    | `localStorage` or store persistence        |
| T18 | Classification banner visible in NVG | Green text on darker green background      |

---

## 8. Test Summary

| #   | Test                        | Module Task | Component        |
| --- | --------------------------- | ----------- | ---------------- |
| T01 | mTLS with valid certs       | Task 1      | svc-track        |
| T02 | mTLS with missing certs     | Task 1      | svc-track        |
| T03 | TLS disabled warning        | Task 1      | svc-track        |
| T04 | mTLS client rejection       | Task 1      | svc-track        |
| T05 | Feedback submit audit event | Task 2      | svc-feedback     |
| T06 | Anti-poison reject audit    | Task 2      | svc-feedback     |
| T07 | Rate limit audit event      | Task 2      | svc-feedback     |
| T08 | Track query audit event     | Task 3      | svc-query        |
| T09 | Classification filter audit | Task 3      | svc-query        |
| T10 | Guardrail violation audit   | Task 3      | svc-query        |
| T11 | Active sensor status        | Task 5      | svc-\*-ingestion |
| T12 | Idle sensor status          | Task 5      | svc-\*-ingestion |
| T13 | SensorHealthPanel renders   | Task 5      | web-cop          |
| T14 | NVG mode activation         | Task 6      | web-cop          |
| T15 | NVG no white/blue/red       | Task 6      | web-cop          |
| T16 | NVG dark map tiles          | Task 6      | web-cop          |
| T17 | Theme persistence           | Task 6      | web-cop          |
| T18 | NVG classification banner   | Task 6      | web-cop          |

---

## 9. Validation Criteria

- [ ] `grep -r "noopAuditEmitter" svc-*/` returns zero results
- [ ] `grep -r "TODO.*mTLS\|TODO.*insecure" svc-track/` returns zero results
- [ ] Classification header scan shows zero missing files
- [ ] `make test` passes for svc-track, svc-feedback, svc-query
- [ ] Audit events appear in `audit.events` topic when feedback is submitted
- [ ] Audit events appear when queries are executed
- [ ] NVG theme toggle works in web-cop dev server
- [ ] Sensor health panel renders in COP UI (at least mock data)

---

## 10. Agent Invocation

```
@greatest-ever-developer Implement v1 Module 03 from docs/implementation/v1/03-service-hardening.md

Context:
- Read docs/implementation/v1/00-v1-overview.md for v1 scope and traceability
- Read docs/implementation/00-implementation-overview.md for global conventions
- Read docs/sdlc_guidelines/04_coding_standards/secure_coding.md for mTLS requirements
- READ svc-track/cmd/track/main.go — the TLS TODO is around line 90
- READ svc-feedback/cmd/feedback/main.go — noopAuditEmitter is at the top of the file
- READ svc-query/cmd/query/main.go — noopAuditEmitter pattern
- READ pkg/audit/emitter.go — this is the real audit emitter to wire in
- READ svc-radar-ingestion/cmd/radar-ingestion/main.go — reference for audit.NewEmitter() usage
- For NVG mode, reference web-cop/src/stores/uiStore.ts for existing theme state
- For sensor health, check proto/rtsa/ingestion/v1/ingestion_service.proto for GetSensorStatus RPC

Tasks (in priority order):
1. Wire mTLS in svc-track (remove TODO, add TLS toggle)
2. Replace noopAuditEmitter in svc-feedback with real pkg/audit.Emitter
3. Replace noopAuditEmitter in svc-query with real pkg/audit.Emitter
4. Scan all source files for missing classification headers and add them
5. Add sensor coverage health endpoint (or Prometheus-based panel)
6. Add NVG dark mode CSS + theme toggle to web-cop

Deliverables:
1. svc-track with mTLS wiring (TLS toggle preserved for dev convenience)
2. svc-feedback and svc-query with real audit emitters
3. All source files have classification headers
4. Sensor health panel component in web-cop
5. NVG dark mode theme with CSS variables
6. Unit tests for mTLS loading, audit emission, theme toggle
7. All existing tests continue to pass
```
