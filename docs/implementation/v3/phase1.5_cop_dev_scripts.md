<!-- CLASSIFICATION: UNCLASSIFIED -->
# Phase 1.5: COP Development Automation Scripts — Detail Implementation Plan

**Status**: 🔄 In Progress
**Goal**: Enable UI developers to run the full RTSA backend with live simulated data while developing the web-cop frontend locally with hot-reload.

---

## Problem

The existing `scripts/demo/run-demo.sh` starts **all** services including `web-cop` (which binds port 5173). UI developers need the backend running but want to use `npm run dev` locally for hot-reload. Manually orchestrating docker compose with `--scale web-cop=0`, topic init, seeding, and simulator launch is error-prone.

## Solution

Two new scripts in `scripts/cop-dev/`:

| Script | Purpose |
|---|---|
| `start-backend.sh` | One-command backend startup for UI development |
| `stop-backend.sh` | Flexible teardown (simulator only / services / full reset) |

---

## [NEW] `scripts/cop-dev/start-backend.sh`

### Usage

```bash
# Default: start everything with sensor-health scenario + seed data
bash scripts/cop-dev/start-backend.sh

# Custom scenario
bash scripts/cop-dev/start-backend.sh --scenario maritime-demo.yaml

# Skip simulator (backend only)
bash scripts/cop-dev/start-backend.sh --no-sim

# Include first-time env setup
bash scripts/cop-dev/start-backend.sh --setup

# Dry-run (print commands without executing)
bash scripts/cop-dev/start-backend.sh --dry-run
```

### Lifecycle

```
1. [Optional] Run setup-dev.sh           (--setup flag)
2. Start infrastructure                  (Redpanda, ClickHouse, Observability)
3. Initialize topics + ClickHouse schema  (init-topics.sh, init-clickhouse.sh)
4. Seed demo data                         (seed-demo-data.sh) — always by default
5. Start application services             (--scale web-cop=0)
6. Wait for health checks                 (all gRPC services SERVING)
7. [Optional] Start simulator             (background, --scenario flag)
8. Print developer connection info
```

### Key Design

- Sources `scripts/demo/_common.sh` for shared utilities
- Overrides `DC` variable to add `--scale web-cop=0`
- Seeds data by default (UI dev always needs data present)
- Simulator runs detached (`docker compose run -d`) so shell returns
- Prints clear developer-facing output with URLs

---

## [NEW] `scripts/cop-dev/stop-backend.sh`

### Usage

```bash
bash scripts/cop-dev/stop-backend.sh              # Stop simulator only (default)
bash scripts/cop-dev/stop-backend.sh --sim         # Same as default
bash scripts/cop-dev/stop-backend.sh --services    # Stop all containers, keep volumes
bash scripts/cop-dev/stop-backend.sh --reset       # Full teardown including volumes
```

---

## Developer Workflow

```bash
# Terminal 1: Backend
bash scripts/cop-dev/start-backend.sh
# → All services healthy, simulator feeding data, URLs printed

# Terminal 2: Frontend (hot-reload)
cd web-cop && npm run dev
# → Vite on :5173, gRPC via Envoy on :8443

# When done:
bash scripts/cop-dev/stop-backend.sh --reset
```

---

## Verification

1. Run `start-backend.sh` → all services show `healthy`
2. Run `npm run dev` in `web-cop/` → no port conflicts
3. Open `http://localhost:5173` → Sensor Operator role → live sensor data updating
4. Run `stop-backend.sh --reset` → clean teardown
5. Run `start-backend.sh` again → idempotent startup
