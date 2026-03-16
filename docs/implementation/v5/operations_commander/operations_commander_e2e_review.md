<!-- CLASSIFICATION: UNCLASSIFIED -->

# Operations Commander End-to-End Implementation Review (v5)

> **Classification**: UNCLASSIFIED
> **Date**: 2026-03-15
> **Reviewer**: Meanest Ever Reviewer (AI Agent)
> **Scope**: Operations Commander role capabilities, code implementation, and test coverage

---

## 1. Review Scope and Inputs

### 1.1 Required Documents Reviewed

- `docs/user_guide/operations_commander/README.md`
- `docs/user_guide/operations_commander/01_situational_awareness.md`
- `docs/user_guide/operations_commander/02_anomaly_alerts.md`
- `docs/user_guide/operations_commander/03_operator_feedback.md`
- `docs/user_guide/operations_commander/04_tactical_edge.md`
- `docs/demo/demo_setup_run_showcase.md`
- `docs/business/requirements.md`
- `docs/business/feature_list.md`
- `docs/business/usecases/UC008_multi_source_fusion.md`
- `docs/business/usecases/UC009_anomaly_detection.md`
- `docs/business/usecases/UC010_operator_feedback.md`
- `docs/business/usecases/UC012_situational_awareness_ui.md`
- `docs/business/usecases/UC013_historical_query.md`
- `docs/business/usecases/UC016_fusion_dashboard.md`
- `docs/architecture/v1/RTSA_WebGPU_Architecture_v1.md`
- `docs/sdlc_guidelines/00_master_policy.md`
- `docs/sdlc_guidelines/01_security_compliance/security_classification.md`
- `docs/sdlc_guidelines/01_security_compliance/itsg33_controls.md`
- `docs/sdlc_guidelines/01_security_compliance/nist800_53_controls.md`
- `docs/sdlc_guidelines/04_coding_standards/general_coding.md`
- `docs/sdlc_guidelines/04_coding_standards/go_standards.md`
- `docs/sdlc_guidelines/04_coding_standards/secure_coding.md`
- `docs/sdlc_guidelines/04_coding_standards/solidjs_standards.md`
- `docs/sdlc_guidelines/05_testing/testing_strategy.md`

### 1.2 Code and Test Areas Reviewed

- Frontend: `web-cop-gpu/src/**`, `web-cop-gpu/tests/**`, `web-cop-gpu/e2e/**`
- Backend services: `svc-track`, `svc-alert`, `svc-query`, `svc-feedback`, integration/e2e tests in `tests/**`
- Demo scripts: `scripts/demo/**`
- API contracts: `proto/rtsa/**`

### 1.3 Test Execution Performed (Representative)

- `go test ./svc-track/internal/handler ./svc-alert/internal/handler ./svc-query/internal/handler ./svc-feedback/internal/handler` → **PASS**
- `pnpm test tests/components/RoleSelector.test.tsx tests/components/DashboardSelector.test.tsx tests/components/FeedbackForm.test.tsx` (in `web-cop-gpu`) → **PASS**

Note: Full integration/E2E/browser suites were reviewed by test-case inspection, not fully executed in this pass.

---

## 2. Executive Verdict

**Overall status: PARTIALLY IMPLEMENTED with major functional gaps for Operations Commander role.**

Backend building blocks exist for several v2.0 features (`StreamSensorObservations`, `AssignAlert`, `GetEventTimeline`), but Commander-specific end-user workflows described in user guide/demo/business docs are **not fully implemented in the current frontend** and are **not covered by robust browser E2E assertions**.

---

## 3. High-Severity Findings (Blocking)

### BLOCKING-1 — Two-Level RBAC shell for 5 roles is not implemented

- **Expected**: 5 roles with role-specific default and Level-2 dashboards (Commander, Analyst, Security Officer, Sensor Operator, NATO Liaison)
- **Actual**: Frontend role model supports only `sensor_operator` and `operations_commander`
- **Evidence**:
  - `web-cop-gpu/src/signals/viewport.ts`
  - `web-cop-gpu/src/components/toolbar/RoleSelector.tsx`
- **Requirement impact**: CR-UI-009, CR-UI-010, UC012

### BLOCKING-2 — Operations Commander dashboards (Fusion/Multi-Domain/Operator UI) are missing as explicit UI implementations

- **Expected** (from user guide + demo + UC016/UC012):
  - Fusion Dashboard (default) with raw+fused map layers and FusionSidePanel
  - Multi-Domain Dashboard with domain KPI overlay
  - Operator UI dashboard with alert quick actions and timeline focus
- **Actual**:
  - No `FusionDashboard`, `MultiDomainDashboard`, or `OperatorDashboard` components in `web-cop-gpu/src`
  - `App.tsx` routes only to sensor health, coverage map, or generic canvas shell
- **Evidence**:
  - `web-cop-gpu/src/App.tsx`
  - `web-cop-gpu/src/components/dashboard/`
- **Requirement impact**: CR-UI-011/012/013, UC012, UC016

### BLOCKING-3 — Alert quick actions required by Commander workflow are not implemented

- **Expected**: `[Inspect] [Confirm] [Reject] [Assign]`
- **Actual**: Alert sidebar provides only `Ack` and conditional `View on Map` for coverage-gap alerts
- **Evidence**:
  - `web-cop-gpu/src/components/panels/AlertSidebar.tsx`
- **Requirement impact**: CR-UI-014, CR-FB-008, UC010, UC012

### BLOCKING-4 — `AssignAlert` RPC exists in backend but is not wired from frontend Commander flow

- **Expected**: Commander can assign alerts in UI; assignment emits audit event
- **Actual**:
  - Backend supports `AssignAlert`
  - Frontend has no `assignAlert` service call and no assignment UI path
- **Evidence**:
  - Backend: `proto/rtsa/inference/v1/alert_service.proto`, `svc-alert/internal/handler/assign.go`
  - Frontend missing usage: `web-cop-gpu/src/**` (no assign alert invocation)
- **Requirement impact**: CR-FB-008, UC010

### BLOCKING-5 — Fusion raw observation stream capability exists backend, but Commander UI does not render raw sensor observations as required

- **Expected**: raw sensor icons rendered with fused tracks (UC016)
- **Actual**:
  - Backend `StreamSensorObservations` handler exists
  - No Fusion dashboard layer in frontend to render raw observation icons by sensor type
- **Evidence**:
  - Backend: `svc-track/internal/handler/stream_observations.go`
  - Frontend: absence of Fusion dashboard components/hooks consuming sensor observation stream
- **Requirement impact**: CR-ING-011, CR-UI-011, UC016

---

## 4. Medium-Severity Findings

### WARNING-1 — Feedback form is incomplete relative to Commander workflow

- `FeedbackType.CONFIRM_ANOMALY` exists in API (`proto/rtsa/common/v1/types.proto`) but is missing from UI options.
- Current form options: confirm hostile/friendly, reclassify, reject anomaly.
- Evidence: `web-cop-gpu/src/components/panels/FeedbackForm.tsx`
- Impact: UC010 user journey mismatch and demo script mismatch.

### WARNING-2 — Entity detail panel does not include required sub-components from business/use case docs

- Expected: `SourceAttribution`, `EntityTimeline`, `FeedbackForm` in detail context
- Actual:
  - `TrackDetailPanel` displays scalar track fields + “Submit Feedback” button
  - Timeline is shown separately in bottom panel via `EventTimeline`
  - No source attribution table component
- Evidence:
  - `web-cop-gpu/src/components/panels/TrackDetailPanel.tsx`
  - `web-cop-gpu/src/components/timeline/EventTimeline.tsx`
- Impact: CR-UI-015, UC012, UC013

### WARNING-3 — Tactical edge UX expectations from Operations Commander guide are only partially implemented

- Generic connected/disconnected indicator exists.
- Explicit edge-mode banners/flows described in guide are not found (`EDGE MODE`, `DISCONNECTED — LOCAL DATA ONLY`, sync status panel).
- Evidence:
  - Present: `web-cop-gpu/src/components/toolbar/ConnectionIndicator.tsx`
  - Missing expected textual indicators in `web-cop-gpu/src/**`
- Impact: CR-UI-006, Tactical Edge guide alignment

### WARNING-4 — UI design-system requirement mismatch (Inter + glassmorphism)

- App shell currently enforces `monospace` font family.
- Evidence: `web-cop-gpu/src/components/shell/AppShell.tsx`
- Impact: CR-UI-019 deviation from business requirements.

---

## 5. Low-Severity / Test Quality Findings

### SUGGESTION-1 — Browser E2E tests are largely smoke/non-blocking for Commander-critical features

- Many assertions allow pass without proving required behavior (conditional `if (isVisible)` / boolean-only checks).
- No browser E2E coverage for:
  - Fusion dashboard raw+fused rendering
  - Multi-domain domain KPI overlay
  - Operator quick actions `[Inspect]/[Confirm]/[Reject]/[Assign]`
  - `AssignAlert` flow and audit-visible assignment state
- Evidence:
  - `web-cop-gpu/e2e/alerts-feedback.spec.ts`
  - `web-cop-gpu/e2e/role-access.spec.ts`

### SUGGESTION-2 — Several integration tests validate topic I/O patterns, not full service-to-service behavior

- `tests/integration/*` for anomaly/feedback/fusion often produce and consume directly on topics, which is useful but does not fully validate end-to-end RPC/UI behavior for Commander.
- Evidence:
  - `tests/integration/anomaly_test.go`
  - `tests/integration/feedback_test.go`
  - `tests/integration/fusion_test.go`

### SUGGESTION-3 — E2E Go tests for Commander scenarios are infrastructure-topic oriented and include outdated UC comments

- E2E tests validate topic flow, not Commander UI workflows.
- Example mismatch: `forensics_query_test.go` comment references UC009 as “historical forensics query”.
- Evidence:
  - `tests/e2e/forensics_query_test.go`
  - `tests/e2e/alert_workflow_test.go`
  - `tests/e2e/feedback_workflow_test.go`

---

## 6. UC Traceability Status (Operations Commander)

| Use Case | Expected Commander Capability                                                | Current Status                 | Evidence Summary                                                                  |
| -------- | ---------------------------------------------------------------------------- | ------------------------------ | --------------------------------------------------------------------------------- |
| UC008    | Consume fused tracks and visualize fusion outputs                            | **Partial**                    | Backend flow exists; Commander-specific fusion visualization missing in UI        |
| UC009    | Anomaly alerts with triage actions                                           | **Partial**                    | Alert streaming/ack exists; required quick actions missing                        |
| UC010    | Submit confirm/reject/reclassify feedback + assign alerts                    | **Partial**                    | SubmitFeedback present; confirm anomaly option missing; assign flow missing in UI |
| UC012    | Full situational UI with 5-role two-level RBAC and role dashboards           | **Not Implemented (per spec)** | Current UI has only 2 roles and no commander dashboard trio                       |
| UC013    | Unified event timeline in commander context                                  | **Partial**                    | Backend + timeline component exist; not integrated as spec’d in detail panel flow |
| UC016    | Fusion Dashboard default for Commander with raw sensor overlays + side panel | **Not Implemented (frontend)** | Backend stream exists, UI components absent                                       |

---

## 7. What Is Implemented and Verified

- Classification banners top and bottom: implemented (`AppShell`)
- Connection health indicator: implemented (`ConnectionIndicator`)
- Track detail retrieval path + feedback modal trigger: implemented
- Backend handlers present and unit-tested for:
  - `StreamSensorObservations`
  - `AssignAlert`
  - `GetEventTimeline`
  - `SubmitFeedback`
- Representative tests executed in this review pass: all selected unit tests passed

---

## 8. Required Remediation Plan (Priority Order)

1. Implement full 5-role RBAC shell and role-aware Level-2 dashboards per CR-UI-009/010.
2. Implement Operations Commander dashboards explicitly:
   - Fusion Dashboard (default)
   - Multi-Domain Dashboard
   - Operator UI Dashboard
3. Implement required alert quick actions in Commander UI and wire to backend RPCs:
   - Inspect → open detail context
   - Confirm/Reject → `FeedbackService.SubmitFeedback`
   - Assign → `AlertService.AssignAlert`
4. Implement Fusion-side raw sensor observation rendering from `StreamSensorObservations` with icon mapping by sensor type.
5. Add missing detail panel sub-components (`SourceAttribution`, in-context `EntityTimeline`).
6. Expand browser E2E suite to hard-assert Commander workflows; remove non-blocking conditional pass patterns.
7. Align tactical edge UI indicators with user guide text and expected behavior.

---

## 9. Final Assessment

The current system has **substantial backend capability** and **partial generic UI capability**, but the documented end-user Operations Commander experience (as defined in user guide, demo script, and business use cases) is **not yet fully realized**. This should be treated as a **CHANGES REQUIRED** state before claiming UC012/UC016 Commander readiness.
