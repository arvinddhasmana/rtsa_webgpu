<!-- CLASSIFICATION: UNCLASSIFIED -->
# RTSA Dashboard UI Modernization — Implementation Plan

> Lead Architect Review | 2026-02-27

## Goal

Modernize the RTSA COP Web Application to deliver three premium, mission-critical dashboard views (Operator UI, Fusion Dashboard, Multi-Domain Dashboard) via a Two-Level Role-Based Architecture. Includes backend API enhancements to supply the data these dashboards require.

---

## Proposed Changes

### Phase 1 — Backend API Enhancements

These must land first as they unblock the UI work.

#### [NEW] `proto/rtsa/entity/v1/track_service.proto` — `StreamSensorObservations` RPC
- Server-streaming RPC returning raw `SensorObservation` data from `sensors.*` Redpanda topics
- Enables the Fusion Dashboard to render individual sensor tracks alongside fused tracks

#### [MODIFY] `proto/rtsa/query/v1/query_service.proto` — `GetEventTimeline` RPC
- Unified timeline query across `tracks_fused`, `anomaly_detections`, `operator_feedback`, `audit_log` in ClickHouse
- Returns `repeated TimelineEvent` ordered by `event_time`

#### [MODIFY] `proto/rtsa/ingestion/v1/ingestion_service.proto` — Sensor coverage geometry
- Add `SensorCoverage` message (coverage polygon, range, bearing sector, sensor position)
- Add `ListSensorStatuses` bulk RPC

#### [MODIFY] `proto/rtsa/inference/v1/alert_service.proto` — `AssignAlert` RPC
- New RPC for operator-to-operator alert assignment

#### [NEW] ClickHouse Materialized Views
- `mv_active_tracks_by_domain` — 10s granularity
- `mv_sensor_throughput_5min` — rolling sensor observation rates
- `mv_alert_ack_latency` — time-to-acknowledge by severity

#### [MODIFY] `svc-track` — New consumer group for raw sensor streaming
- Consume from `sensors.*` topics, apply classification filter, stream to UI

---

### Phase 2 — Design System & Collapsible Pane Architecture

#### [MODIFY] `web-cop/src/index.css`
- Replace `Courier New` with `Inter` (Google Fonts)
- Define CSS custom properties for glassmorphism, blue/amber accent palette
- Add NVG and High-Contrast theme variables

#### [MODIFY] `web-cop/src/stores/uiStore.ts`
- Add `activeDashboardView` state: `'fusion' | 'multi-domain' | 'operator' | 'forensics' | 'audit' | 'sensor-health' | 'nato-exchange'`
- Add `setDashboardView(view)` action
- Add the missing roles `sensor_operator` and `nato_liaison` to `ActiveRole`

#### [MODIFY] `web-cop/src/components/layout/RoleSelector.tsx`
- Add all 5 roles (Commander, Analyst, Security, Sensor Operator, NATO Liaison)

#### [NEW] `web-cop/src/components/layout/DashboardSelector.tsx`
- Level-2 view switcher that renders role-appropriate view options
- Commander: Fusion (default) / Multi-Domain / Operator UI
- Analyst: Forensics (default) / Intelligence Search
- Security: Audit & Feedback
- Sensor Operator: Sensor Health
- NATO Liaison: NATO Exchange

#### [MODIFY] `web-cop/src/components/layout/MainLayout.tsx`
- Refactor to render `DashboardSelector` in the toolbar
- Route to dashboard-specific layout components based on `activeRole` + `activeDashboardView`
- Add universal collapsible pane system (chevron toggle on every panel header)

---

### Phase 3 — Dashboard Implementations

#### Operator UI
| Component | Path | Description |
|---|---|---|
| [NEW] `OperatorDashboard.tsx` | `components/layout/` | Map with `backdrop-blur` overlay; Timeline + Alert panels |
| [NEW] `TimelineView.tsx` | `components/timeline/` | Chronological event timeline using `GetEventTimeline` RPC |
| [MODIFY] `AlertCard.tsx` | `components/alerts/` | Add `[Inspect]`, `[Confirm]`, `[Reject]`, `[Assign]` buttons |

#### Fusion Dashboard
| Component | Path | Description |
|---|---|---|
| [NEW] `FusionDashboard.tsx` | `components/layout/` | Full map + collapsible `FusionSidePanel` |
| [NEW] `FusionSidePanel.tsx` | `components/fusion/` | Real-time metrics: active tracks, confidence scores, sensor contributions |
| [NEW] `useSensorStream.ts` | `hooks/` | Hook consuming `StreamSensorObservations` for raw sensor icons on map |

#### Multi-Domain Dashboard
| Component | Path | Description |
|---|---|---|
| [NEW] `MultiDomainDashboard.tsx` | `components/layout/` | Maximized map + floating metric overlays |
| [NEW] `DomainMetricsOverlay.tsx` | `components/dashboard/` | Domain-split KPIs (Air/Surface/Sub/Land/Cyber counts, rates) |
| [NEW] `SensorCoverageLayer.tsx` | `components/map/` | Renders radar fan sectors, EW arcs from `SensorCoverage` data |

#### Shared Enhancements
| Component | Path | Description |
|---|---|---|
| [MODIFY] `DetailPanel.tsx` | `components/detail/` | Implement `SourceAttribution`, `EntityTimeline`, `FeedbackForm` |
| [NEW] `MapLayerToggle.tsx` | `components/map/` | Floating Layers button exposing `layerVisibility` toggles |
| [MODIFY] `SearchOverlay.tsx` | `components/layout/` | Connect to `TrackStore` for real entity lookup |

---

### Phase 4 — Polish & Remaining Roles

- Implement Sensor Operator view (Sensor Health dashboard with coverage map)
- Implement NATO Liaison view (Exchange status, Manual Track Nomination)
- Complete all keyboard shortcuts (`M`, `F`, `Tab`, `Ctrl+Z`)
- Add micro-animations (panel transitions, track icon hover, smooth collapse)
- Implement Reduced Bandwidth Mode (edge auto-detection)

---

## Verification Plan

### Automated Tests
```bash
cd web-cop && npm run test        # Unit tests
cd web-cop && npm run test:e2e    # Playwright E2E
```
- Add Playwright tests for role switching, dashboard view switching, and panel collapse/expand
- Add tests for alert quick-action button visibility and click behavior

### Manual Verification
1. Start the stack: `docker compose -f deploy/docker-compose.yml up -d`
2. Open `http://localhost:5173`
3. Verify each role → each dashboard view combination
4. Verify all panes collapse and expand correctly
5. Verify glassmorphism, typography, and accent colors render correctly
6. Verify blurred map background in Operator UI
7. Verify raw sensor icons appear alongside fused tracks in Fusion Dashboard

### Phase 1 Testing Strategy (SDLC Compliance)
As requested, all newly implemented Phase 1 backend code will be rigorously tested according to `docs/sdlc_guidelines`.

1. **Unit Testing**:
   - Create `svc-track/internal/consumer/sensor_consumer_test.go`
   - Create `svc-track/internal/handler/stream_observations_test.go`
   - Use table-driven tests and mock dependencies (e.g., mock the Redpanda `kgo.Client`).
   - Ensure line coverage is ≥85% for handlers and ≥90% for domain logic.
2. **Integration Testing**:
   - Use the existing Docker environment: `make docker-up` to start infrastructure (Redpanda, ClickHouse, Observability).
   - Use `make test` and `make integration-test` to validate functionalities.
3. **Build Verification**:
   - Verify the services compile successfully without regressions using `make build`.
