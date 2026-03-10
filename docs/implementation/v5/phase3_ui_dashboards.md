<!-- CLASSIFICATION: UNCLASSIFIED -->

# Phase 3 — UI & Dashboard Validation

> **Use Cases**: UC012 (Situational Awareness UI), UC016 (Fusion Dashboard), UC017 (Sensor Health Monitoring)
> **Prerequisites**: Phase 2 complete — all intelligence/detection tests passing
> **Common Guidelines**: See [00_common_guidelines.md](00_common_guidelines.md)
> **Classification**: UNCLASSIFIED
> **Date**: 2026-03-10

---

## 1. Objectives

- Verify the Two-Level RBAC shell: role selection drives dashboard selection (5 roles, correct tabs per role)
- Verify all dashboards render correctly with required components
- Verify Fusion Dashboard raw sensor observation layer + side panel
- Verify Operator UI alert quick-actions, event timeline, and entity detail panel
- Verify Multi-Domain Dashboard domain metrics + sensor coverage overlays
- Verify Sensor Health Dashboard cards + coverage map
- **Generate high-fidelity images** of the existing UI for each dashboard/view
- **Suggest UI improvements**: new images for gaps found, improvements on existing UI
- Run all frontend unit tests (Vitest), Playwright E2E tests (headless → headed), and backend service tests
- Compile all issues into Issue Log, then fix in batch

---

## 2. UI Image Generation

### 2.1 Capture Existing UI State

For each dashboard/view, generate a high-fidelity mockup image and store under `docs/implementation/v5/ui_images/`.

| # | Image | Description | Filename |
|---|-------|-------------|----------|
| 1 | Login / Role Selection | Role selector dropdown with 5 roles visible | `01_role_selection.webp` |
| 2 | Ops Commander — Fusion Dashboard | Map with fused + raw sensor icons, FusionSidePanel expanded | `02_fusion_dashboard.webp` |
| 3 | Ops Commander — Operator UI | Blurred map, glass-morphism alert panel, event timeline | `03_operator_dashboard.webp` |
| 4 | Ops Commander — Multi-Domain | Full-screen map, domain metrics overlay, sensor coverage | `04_multi_domain_dashboard.webp` |
| 5 | Sensor Operator — Sensor Health | Sensor status cards, coverage map overlays | `05_sensor_health_dashboard.webp` |
| 6 | Intelligence Analyst — Forensics | Time-range query panel, map with historical tracks | `06_forensics_dashboard.webp` |
| 7 | Security Officer — Audit | Feedback entries, trust scores, audit log | `07_audit_dashboard.webp` |
| 8 | NATO Liaison — NATO Exchange | Link connectivity, nomination queue, inbound tracks | `08_nato_exchange_dashboard.webp` |
| 9 | Entity Detail Panel | Track info, source attribution, entity timeline, feedback form | `09_entity_detail_panel.webp` |
| 10 | Alert Card with Quick Actions | [Inspect] [Confirm] [Reject] [Assign] buttons | `10_alert_quick_actions.webp` |
| 11 | Classification Banners | Top + bottom classification banners | `11_classification_banners.webp` |

### 2.2 Suggest Improvements

After capturing existing state, identify UI gaps and suggest improvements. Create "proposed" images for:
- Missing visual elements (e.g., bottom classification banner if absent)
- Improved data density or layout
- Better glassmorphism panel styling
- Enhanced map layer toggle UX
- Improved mobile/responsive breakpoints

Store proposed images with `_proposed` suffix (e.g., `03_operator_dashboard_proposed.webp`).

---

## 3. UC012 — Situational Awareness UI (Two-Level RBAC Shell)

**Spec**: `docs/business/usecases/UC012_situational_awareness_ui.md`
**Feature**: FEAT-13 + FEAT-16 + FEAT-17 + FEAT-18
**Requirements**: CR-UI-001 through CR-UI-020

### 3.1 Role → Dashboard Mapping

Dashboard selection is **dependent on the selected Role**. Verify this mapping:

| Role | Default Dashboard (Level 1) | Available Tabs (Level 2) |
|------|-----------------------------|--------------------------|
| Operations Commander | Fusion Dashboard | Fusion, Multi-Domain, Operator UI |
| Intelligence Analyst | Forensics | Forensics, Intel Search |
| Security Officer | Audit & Feedback | _(single view)_ |
| Sensor Operator | Sensor Health | _(single view)_ |
| NATO Liaison | NATO Exchange | _(single view)_ |

### 3.2 Code Review Checklist

| # | Review Item | File(s) | Requirement |
|---|-------------|---------|-------------|
| 1 | RoleSelector shows all 5 roles | `web-cop-gpu/src/components/toolbar/RoleSelector.tsx` | CR-UI-009 |
| 2 | DashboardSelector tabs change based on selected role | `web-cop-gpu/src/components/toolbar/DashboardSelector.tsx` | CR-UI-010 |
| 3 | `uiStore.activeDashboardView` signal drives correct component rendering | `web-cop-gpu/src/signals/` | CR-UI-010 |
| 4 | Classification banners rendered at BOTH top and bottom | `web-cop-gpu/src/components/shell/AppShell.tsx` | CR-UI-005 |
| 5 | Classification banners show correct level text and colour | `ClassificationBanner.tsx` | CR-UI-005 |
| 6 | Map view renders entity tracks in real-time | `web-cop-gpu/src/components/` (map-related) | CR-UI-001 |
| 7 | Track icons colour-coded by hostile class: RED=Hostile, BLUE=Friendly, GREEN=Neutral, YELLOW=Unknown | map rendering / WGSL | CR-UI-001 |
| 8 | Track trails (last N positions, fading) | trail render pass | CR-UI-001 |
| 9 | Alert panel renders with severity sorting (CRITICAL first) | `AlertSidebar.tsx` | CR-UI-002 |
| 10 | Quick actions on alerts: [Inspect] [Confirm] [Reject] [Assign] | `AlertSidebar.tsx` | CR-UI-014 |
| 11 | [Inspect] opens Entity Detail Panel | `TrackDetailPanel.tsx` | CR-UI-015 |
| 12 | Entity Detail Panel has: SourceAttribution, EntityTimeline, FeedbackForm | detail panel components | CR-UI-015 |
| 13 | [Confirm] calls `FeedbackService.SubmitFeedback` with `CONFIRM_ANOMALY` | feedback service call | CR-UI-014 |
| 14 | [Reject] calls `FeedbackService.SubmitFeedback` with `REJECT_ANOMALY` | feedback service call | CR-UI-014 |
| 15 | [Assign] opens operator picker, calls `AlertService.AssignAlert` | alert service call | CR-FB-008 |
| 16 | Panel collapse/expand toggle on all panels | panel components | CR-UI-018 |
| 17 | Map Layer Toggle: track labels, trails, sensor coverage, geo-fences, MGRS grid | layer controls | CR-UI-020 |
| 18 | Inter typeface loaded and applied | CSS / fonts | CR-UI-019 |
| 19 | Glassmorphism panel aesthetic (backdrop-filter, transparency) | CSS styles | CR-UI-019 |
| 20 | NVG-compatible dark mode | theme/CSS | CR-UI-007 |
| 21 | StatusBar shows FPS, latency, track count | `StatusBar.tsx` | performance |
| 22 | Ctrl+K search overlay | `SearchOverlay.tsx` | UC012 |
| 23 | Connection indicator (green/red/orange) | `ConnectionIndicator.tsx` | UC012 |
| 24 | `data-testid` attributes on all components (v4 review R-001 through R-004) | all components | E2E testing |
| 25 | No `console.log` in production code (v4 review R-020) | all components | SDLC §7.1 |

### 3.3 Track Service Backend Review

| # | Review Item | File(s) |
|---|-------------|---------|
| 1 | `StreamTracks` gRPC handler — streams fused tracks to UI | `svc-track/internal/handler/` |
| 2 | `StreamSensorObservations` gRPC handler — streams raw sensor observations (v2.0) | `svc-track/internal/handler/` |
| 3 | `GetTrackDetails` unary RPC | handler |
| 4 | Dual consumer groups: `track-service` (fused) + `track-svc-sensor-stream` (raw) | consumer config |
| 5 | Bounding box + classification filtering on streams | handler logic |

### 3.4 Tests to Run

```bash
# Backend: Track Service
cd svc-track && go test -race -count=1 -v ./...
cd svc-track && go test -race -coverprofile=coverage.out ./... && \
  go tool cover -func=coverage.out | tail -1

# Frontend: Vitest unit tests
cd web-cop-gpu && pnpm test

# Frontend coverage (target: ≥80%)
cd web-cop-gpu && pnpm test:coverage

# Frontend E2E — Headless first
cd web-cop-gpu && pnpm test:e2e
# Specs: cold-boot, role-access, alerts-feedback, visual-regression, reconnection, degraded-mode, security-audit

# Frontend E2E — Headed (visual verification via browser subagent)
cd web-cop-gpu && pnpm test:e2e -- --headed
```

### 3.5 Expected Outcomes

- All 10 component unit tests pass
- Vitest coverage ≥80% for SolidJS components
- All 7 Playwright E2E specs pass headless
- Headed E2E confirms visual correctness
- Role → Dashboard mapping works correctly
- All `data-testid` attributes present (v4 review items resolved)

---

## 4. UC016 — Fusion Dashboard

**Spec**: `docs/business/usecases/UC016_fusion_dashboard.md`
**Feature**: FEAT-16
**Requirements**: CR-UI-011, CR-ING-011

### 4.1 Code Review Checklist

| # | Review Item | File(s) | Requirement |
|---|-------------|---------|-------------|
| 1 | Distinct raw sensor icons per type: Radar ◇, EW △, SIGINT ◻, AIS ○, ISR ⬡ | map rendering | CR-UI-011 |
| 2 | Fused tracks shown as filled circles alongside raw observations | map rendering | CR-UI-011 |
| 3 | FusionSidePanel shows: active track count by domain | `FusionSidePanel.tsx` or equivalent | FEAT-16 |
| 4 | FusionSidePanel: Top-5 confidence tracks | side panel | FEAT-16 |
| 5 | FusionSidePanel: confidence histogram (High≥0.8=green, Med=blue, Low=amber, Tentative=red) | side panel | FEAT-16 |
| 6 | FusionSidePanel: sensor contribution chart | side panel | FEAT-16 |
| 7 | Collapsible side panel with chevron toggle | side panel | CR-UI-018 |
| 8 | `StreamSensorObservations` data flowing to map layer | data hook + map | CR-ING-011 |
| 9 | Bounding box filter limits raw observations to visible map extent | stream filter | FEAT-16 |
| 10 | Classification filtering at stream boundary | stream filter | CR-SEC-001 |

### 4.2 Tests to Run

```bash
# Check for FusionDashboard/FusionSidePanel component tests
ls -la web-cop-gpu/tests/components/

# Backend: verify StreamSensorObservations handler
cd svc-track && go test -race -v -run "SensorObservation" ./...
```

### 4.3 Expected Outcomes

- Raw sensor icons visually distinct per type on the map
- FusionSidePanel updates in real-time with track counts and confidence metrics
- StreamSensorObservations data flows through correctly

---

## 5. UC017 — Sensor Health Monitoring

**Spec**: `docs/business/usecases/UC017_sensor_health_monitoring.md`
**Feature**: FEAT-19
**Requirements**: CR-UI-016, CR-ING-012

### 5.1 Code Review Checklist

| # | Review Item | File(s) | Requirement |
|---|-------------|---------|-------------|
| 1 | SensorHealthDashboard renders per-sensor status cards | `SensorHealthDashboard.tsx` or equivalent | CR-UI-016 |
| 2 | Card shows: sensor_id, type icon, events/sec, last observation time, data quality score | card component | CR-UI-016 |
| 3 | Card colour: green (connected), amber (stale >30s), red (no data >2min) | card styling | CR-UI-016 |
| 4 | Click sensor card → centres map on sensor position | card click handler | FEAT-19 |
| 5 | Coverage map overlays: radar fan sectors (blue), EW arcs (amber), ISR polygons (green) | `SensorCoverageLayer.tsx` | CR-ING-012 |
| 6 | Coverage disappears when sensor goes offline | coverage reactivity | FEAT-19 |
| 7 | `ListSensorStatuses` bulk RPC includes `SensorCoverage` geometry | backend handler | CR-ING-012 |
| 8 | Status cards update every 10 seconds | polling/stream interval | FEAT-19 |

### 5.2 Tests to Run

```bash
# Check for SensorHealthDashboard component tests
ls -la web-cop-gpu/tests/components/

# Backend: verify ListSensorStatuses handler
# Check each ingestion service for sensor status endpoints
for svc in svc-radar-ingestion svc-ew-ingestion svc-elint-ingestion \
           svc-isr-ingestion svc-ais-ingestion svc-cyber-ingestion; do
  echo "=== $svc sensor status ==="
  (cd "$svc" && go test -race -v -run "SensorStatus" ./... 2>/dev/null || echo "No SensorStatus tests found")
done
```

### 5.3 Expected Outcomes

- Sensor status cards render correctly with colour coding
- Coverage map overlays match sensor types
- ListSensorStatuses returns SensorCoverage geometry

---

## 6. Operator UI Dashboard (Part of UC012)

**Feature**: FEAT-18
**Requirements**: CR-UI-013, CR-UI-014, CR-UI-015

### 6.1 Code Review Checklist

| # | Review Item | File(s) | Requirement |
|---|-------------|---------|-------------|
| 1 | Blurred map background with frosted-glass panels | Operator dashboard styling | CR-UI-013 |
| 2 | Alert panel with severity sorting (CRITICAL red-pulse, ELEVATED amber) | `AlertSidebar.tsx` | CR-UI-013 |
| 3 | Quick-action buttons: [Inspect] [Confirm] [Reject] [Assign] | alert card | CR-UI-014 |
| 4 | Event Timeline: chronological events from `GetEventTimeline` | `TimelineView.tsx` | CR-UI-013 |
| 5 | Timeline shows: track state changes, anomaly events, feedback, audit events | timeline rendering | CR-HIS-008 |
| 6 | Success toast on Confirm/Reject | feedback flow | UX |

---

## 7. Multi-Domain Dashboard (Part of UC012)

**Feature**: FEAT-17
**Requirements**: CR-UI-012

### 7.1 Code Review Checklist

| # | Review Item | File(s) | Requirement |
|---|-------------|---------|-------------|
| 1 | Full-screen map (all panels collapsed by default) | Multi-domain layout | CR-UI-012 |
| 2 | Domain Metrics Overlay: Air, Surface, Subsurface, Land, Cyber with track count, obs rate, hostile count | `DomainMetricsOverlay.tsx` | CR-UI-012 |
| 3 | Sensor coverage overlays: blue fan sectors (radar), amber circles (EW), green polygons (ISR) | `SensorCoverageLayer.tsx` | CR-UI-012 |
| 4 | Layer toggle button (bottom-right) | layer control | CR-UI-020 |
| 5 | Collapsed alert strip at bottom, expandable | alert panel | CR-UI-018 |

---

## 8. Visual Verification via Browser Subagent

After headless tests pass, verify visually:

1. Start dev server: `cd web-cop-gpu && pnpm dev`
2. Open `http://localhost:5173` in browser
3. For each role:
   - Select role from dropdown
   - Verify correct default dashboard loads
   - Verify dashboard tabs match role (Operations Commander has 3 tabs; others have 1–2)
   - Capture screenshot of each dashboard
4. Verify classification banners top + bottom
5. Verify glassmorphism styling on panels
6. Verify Inter typeface applied

---

## 9. Issue Log

_Issues found during Phase 3 review are recorded here and fixed in batch._

| # | UC | Severity | File(s) | Issue Description | Requirement |
|---|----|----------|---------|-------------------|-------------|
| | | | | _(to be populated during execution)_ | |

---

## 10. Batch Fix Execution

After the Issue Log is populated:

1. Fix all BLOCKING issues first
2. Fix WARNING issues
3. Fix IMPROVEMENT items (including UI suggestions)
4. Generate proposed UI images for improvements
5. Re-run ALL tests from sections 3–8 to confirm no regressions

---

## 11. Phase 3 Completion Criteria

- [ ] UC012: Two-Level RBAC shell — role selection correctly drives dashboard tabs
- [ ] UC012: All 25 code review items checked
- [ ] UC016: FusionDashboard renders raw + fused icons with correct side panel
- [ ] UC017: SensorHealthDashboard cards + coverage map verified
- [ ] Operator UI: Alert quick-actions, event timeline, entity detail panel verified
- [ ] Multi-Domain: Domain metrics overlay + sensor coverage verified
- [ ] All 10 Vitest component tests pass; ≥80% coverage
- [ ] All 7 Playwright E2E specs pass (headless)
- [ ] Headed E2E visual verification complete
- [ ] High-fidelity UI images generated and stored in `docs/implementation/v5/ui_images/`
- [ ] UI improvement suggestions documented with proposed images
- [ ] All Issue Log items resolved
- [ ] No silent test failures in any stage
