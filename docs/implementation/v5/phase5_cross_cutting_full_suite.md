<!-- CLASSIFICATION: UNCLASSIFIED -->

# Phase 5 — Cross-Cutting Validation & Full Suite

> **Use Cases**: All (UC001–UC017)
> **Prerequisites**: Phases 1–4 complete — all individual phase tests passing
> **Common Guidelines**: See [00_common_guidelines.md](00_common_guidelines.md)
> **Classification**: UNCLASSIFIED
> **Date**: 2026-03-10

---

## 1. Objectives

- Run the **complete test suite** end-to-end across all tiers
- Verify no test regressions from batch fixes in Phases 1–4
- Validate cross-cutting concerns: security, audit, classification, performance
- Run full demo scenario to validate E2E user experience
- Produce final test results summary
- Generate walkthrough documenting what was accomplished

---

## 2. Full Backend Test Suite

### 2.1 Complete Go Unit Tests

```bash
# All 14 Go services + pkg — with coverage
bash scripts/dev/test-go.sh
```

**Expected**: ALL modules pass. Per-module coverage in `.test-results/<timestamp>/unit/`.

**Coverage verification**:

| Module | Target |
|--------|--------|
| `svc-fusion-engine` (domain) | ≥90% |
| `svc-anomaly-detection` (domain) | ≥90% |
| `svc-feedback` (trust scoring) | ≥95% |
| `svc-feedback` (anti-poisoning) | ≥95% |
| `svc-radar-ingestion` (domain) | ≥90% |
| All other services | ≥80% |
| `pkg/*` | ≥80% |

### 2.2 Complete Integration Tests

```bash
# IT01–IT14 (all 14 integration tests)
bash scripts/dev/test-integration.sh
```

**Expected**: ALL pass. **IT14 is MANDATORY** — classification propagation chain.

### 2.3 Complete E2E Tests

```bash
# E2E01–E2E03. Neg01–Neg04
bash scripts/dev/test-e2e.sh
```

**Expected**: ALL pass. Check `.test-results/<timestamp>/e2e/failures.txt` is empty.

### 2.4 Complete Benchmarks

```bash
# B01–B04
bash scripts/dev/test-bench.sh
```

**Expected**:

| Benchmark | Threshold |
|-----------|-----------|
| B01 | Ingestion ≥ 1,000 obs/sec |
| B02 | Fusion p95 ≤ 100ms |
| B03 | Anomaly p95 ≤ 150ms |
| B04 | Query 100-row p95 ≤ 500ms |

### 2.5 Full Suite (Single Command)

```bash
# Runs everything: unit → integration → E2E → benchmarks
# Produces .test-results/<timestamp>/SUMMARY.txt
bash scripts/dev/test-all.sh
```

**Expected**: SUMMARY.txt shows all stages PASSED:

```
  Stage                                   Tests  Passed  Failed    Status
  ──────────────────────────────────────  ─────  ───────  ───────  ────────
  Unit — Go                                 NNN      NNN        0    PASSED
  Unit — Frontend                           NNN      NNN        0    PASSED
  Integration                               NNN      NNN        0    PASSED
  E2E                                       NNN      NNN        0    PASSED
  Benchmarks                                NNN      NNN        0    PASSED
```

---

## 3. Full Frontend Test Suite

### 3.1 Vitest Unit Tests

```bash
cd web-cop-gpu && pnpm test
```

**Expected**: All component tests pass.

### 3.2 Vitest Coverage

```bash
cd web-cop-gpu && pnpm test:coverage
```

**Expected**: ≥80% line coverage for SolidJS components.

### 3.3 Playwright E2E — Headless (Tier 1)

```bash
cd web-cop-gpu && pnpm test:e2e
```

**Expected**: All 7 spec files pass:
- `cold-boot.spec.ts` — App loads, classification banner, capability gate
- `role-access.spec.ts` — All 5 roles, correct dashboard per role
- `alerts-feedback.spec.ts` — Alert panel, quick-actions, feedback form
- `visual-regression.spec.ts` — FPS display, latency, visual consistency
- `reconnection.spec.ts` — Connection loss + recovery
- `degraded-mode.spec.ts` — WebGPU unavailable fallback
- `security-audit.spec.ts` — CSP headers, no eval(), Wasm integrity

### 3.4 Playwright E2E — Headed (Tier 1, Visual)

```bash
cd web-cop-gpu && pnpm test:e2e -- --headed
```

**Purpose**: Visual confirmation that UI renders correctly during test execution.

---

## 4. Cross-Cutting Validation

### 4.1 Security & Classification

| # | Validation | Test/Command |
|---|-----------|-------------|
| 1 | mTLS enforced on all gRPC channels | Health check + service logs |
| 2 | Classification propagation end-to-end | IT14 (MANDATORY) |
| 3 | Classification violation rejected | Neg03 |
| 4 | CSP headers present and correct | `security-audit.spec.ts` |
| 5 | No `eval()` or `Function()` in frontend bundle | `security-audit.spec.ts` |
| 6 | No hardcoded URLs to external services | Code search |
| 7 | No `console.log` in production code | Lint check |

```bash
# Verify no console.log in production code
cd web-cop-gpu && pnpm lint

# Verify no hardcoded external URLs
grep -rn "http://" web-cop-gpu/src/ --include="*.ts" --include="*.tsx" \
  | grep -v "localhost" | grep -v "CLASSIFICATION"
```

### 4.2 Audit Trail Completeness

| # | Validation | Test |
|---|-----------|------|
| 1 | All state-changing operations produce audit events | `tests/integration/audit_test.go` |
| 2 | Audit log has no TTL (ITSG-33 AU-11) | IT12 |
| 3 | Track creation → audit event | IT04 |
| 4 | Feedback submission → audit event | E2E03 |
| 5 | Alert assignment → audit event | Unit tests (svc-alert) |
| 6 | NATO nomination → audit event | Unit tests (svc-nato-adapter) |

### 4.3 Performance

All benchmarks from B01–B04 validated in section 2.4. Additionally:

```bash
# UI performance — verify FPS ≥ 55 with mock data
# This is verified via headed E2E test visual inspection
# StatusBar should show FPS ≥ 55 and latency values
```

### 4.4 Data Pipeline End-to-End

```bash
# Verify Redpanda topics exist
docker exec rtsa-redpanda rpk topic list

# Verify ClickHouse tables + materialized views
docker exec rtsa-clickhouse clickhouse-client --database rtsa \
  --query "SHOW TABLES"
docker exec rtsa-clickhouse clickhouse-client --database rtsa \
  --query "SELECT name FROM system.tables WHERE name LIKE 'mv_%'"
```

---

## 5. Full Demo Scenario Validation

Run the complete demo to validate the E2E user experience:

```bash
# 1. Fresh start
bash scripts/demo/stop-demo.sh --volumes 2>/dev/null || true

# 2. Start multi-domain demo with seed data
bash scripts/demo/run-demo.sh multi-domain --seed

# 3. Health check
bash scripts/dev/health-check.sh

# 4. Verify web COP is reachable
curl -sf http://localhost:5173/ > /dev/null && echo "[✓] web-cop-gpu" \
  || echo "[✗] web-cop-gpu not responding"
```

### 5.1 Role-by-Role Verification (Browser Subagent)

Open `http://localhost:5173` in Chrome 113+. For each role:

| # | Role | Verify | Demo Guide Section |
|---|------|--------|--------------------|
| 1 | Any | Classification banner top + bottom | §6.1 Scene 1 |
| 2 | Sensor Operator | Sensor Health: cards updating, coverage overlays | §6.2 Scene 2 |
| 3 | Operations Commander → Fusion | Fused + raw icons, FusionSidePanel | §6.3 Scene 3 |
| 4 | Operations Commander → Operator UI | Alert panel, timeline, quick-actions | §6.4 Scene 4 |
| 5 | Operations Commander → Multi-Domain | Full-screen map, domain metrics, coverage | §6.5 Scene 5 |
| 6 | Intelligence Analyst → Forensics | Time-range query, entity timeline | §6.6 Scene 6 |
| 7 | Security Officer → Audit | Feedback entries, trust scores | §6.7 Scene 7 |
| 8 | NATO Liaison → NATO Exchange | Link connectivity, nomination queue | §6.8 Scene 8 |

### 5.2 Clean Teardown

```bash
bash scripts/demo/stop-demo.sh --volumes
```

---

## 6. Pre-PR Check

Run the local CI mirror to verify everything passes the full gating pipeline:

```bash
bash scripts/dev/pre-pr-check.sh
```

**Expected**: All 5 gates pass (secrets/lint, build, unit tests + coverage, security, integration).

---

## 7. Issue Log

_Issues found during Phase 5 cross-cutting validation are recorded here and fixed in batch._

| # | Area | Severity | File(s) | Issue Description | Requirement |
|---|------|----------|---------|-------------------|-------------|
| | | | | _(to be populated during execution)_ | |

---

## 8. Batch Fix Execution

After the Issue Log is populated:

1. Fix all BLOCKING issues first
2. Fix WARNING issues
3. Fix IMPROVEMENT items
4. Re-run `bash scripts/dev/test-all.sh` to confirm zero failures
5. Re-run `cd web-cop-gpu && pnpm test && pnpm test:e2e` to confirm frontend is green

---

## 9. Final Deliverables

After all fixes are applied and verified:

1. **Test Results Summary**: `.test-results/<timestamp>/SUMMARY.txt` with all stages PASSED
2. **Walkthrough Document**: `docs/implementation/v5/walkthrough.md` summarizing:
   - What was validated per use case
   - What issues were found and fixed per phase
   - Test results evidence
   - UI screenshots/recordings
   - Remaining known limitations (if any)
3. **UI Images**: `docs/implementation/v5/ui_images/` — high-fidelity images of all dashboards + proposed improvements

---

## 10. Phase 5 Completion Criteria

- [ ] `test-all.sh` SUMMARY.txt: all stages PASSED, zero failures
- [ ] Frontend Vitest: all pass, ≥80% coverage
- [ ] Frontend Playwright headless: all 7 specs pass
- [ ] Frontend Playwright headed: visual verification complete
- [ ] IT14 MANDATORY classification chain: passes
- [ ] All 4 benchmarks pass thresholds (B01–B04)
- [ ] All negative tests pass (Neg01–Neg04)
- [ ] Security audit spec passes
- [ ] No `console.log` in production code
- [ ] Full demo scenario runs successfully
- [ ] All 8 role-by-role demo scenes verified
- [ ] pre-pr-check.sh passes all 5 gates
- [ ] All Issue Log items resolved
- [ ] Walkthrough document created
- [ ] Final test results archived
