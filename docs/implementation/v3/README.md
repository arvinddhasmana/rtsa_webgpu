<!-- CLASSIFICATION: UNCLASSIFIED -->
# RTSA COP Web Application — v3.0 Implementation Plan

> **Approach**: Build role-by-role, starting with Sensor Operator (foundation), then Operations Commander (most shared components), then remaining roles. Each phase is end-to-end: components → integration → tests → validation against live Docker backend.

---

## Confirmed Design Decisions

| Decision | Resolution |
|---|---|
| **Backend connectivity** | Wire gRPC calls to real services. Use existing demo Docker Compose + simulator scripts with simulated sensor data. |
| **Analyst Dashboard naming** | Rename second analyst view from `operator` → `intel-search` to match user guide. |
| **SensorHealthPanel removal** | Remove legacy bottom panel; consolidate into dashboard-only view. |
| **Timeline Scrubber** | Include in Phase 2 (Ops Commander), not deferred to Phase 4. |
| **CSS design system** | Standardize on CSS custom properties via `design-system.css`. No Tailwind. |
| **Mockup selection** | Variant C (Split-Pane Table + Map) with: (1) DLQ pie chart popup icon on inline expand, (2) resizable panes + Escape reset. |

---

## Phase Overview

| Phase | Role | Status | Detail Document |
|---|---|---|---|
| **Phase 1** | Sensor Operator | ✅ Complete | [phase1_sensor_operator.md](phase1_sensor_operator.md) |
| **Phase 1.5** | Dev Automation Scripts | 🔄 In Progress | [phase1.5_cop_dev_scripts.md](phase1.5_cop_dev_scripts.md) |
| **Phase 2** | Operations Commander | ⬜ Planned | [phase2_operations_commander.md](phase2_operations_commander.md) |
| **Phase 3** | Intelligence Analyst | ⬜ Planned | [phase3_intelligence_analyst.md](phase3_intelligence_analyst.md) |
| **Phase 4** | Security Officer & NATO Liaison | ⬜ Planned | [phase4_security_nato.md](phase4_security_nato.md) |

---

## Architecture

```
External Sensors --gRPC--> Ingestion Services --> Redpanda sensors.* topics
                                                          |
                                                          v
                                                    Fusion Engine
                                                          |
                                          tracks.fused.* Redpanda topics
                                               /                  \
                                    Anomaly Detection           Track Service
                                         |                           |
                                   alerts.* topics           gRPC-Web streams
                                         |                           |
                                   Alert Service              Envoy Gateway
                                         |                     (port 8443)
                                   Redpanda Connect                 |
                                   (ETL to ClickHouse)       Browser (Vite)
                                         |                   (port 5173)
                                   ClickHouse OLAP
```

### Development Workflow

```bash
# Terminal 1: Start backend (all services except web-cop)
bash scripts/cop-dev/start-backend.sh

# Terminal 2: Start frontend with hot-reload
cd web-cop && npm run dev

# Browser: http://localhost:5173 → connected to live gRPC backend via Envoy :8443
```

---

## Verification Strategy

| Layer | Tool | Threshold |
|---|---|---|
| Unit Tests | Vitest | ≥80% line coverage per phase |
| E2E Tests | Playwright | All role workflows covered |
| TypeScript | `npx tsc --noEmit` | Zero errors |
| Backend Health | `scripts/dev/health-check.sh` | All services `[✓]` |
