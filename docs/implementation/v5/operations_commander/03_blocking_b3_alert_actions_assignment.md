<!-- CLASSIFICATION: UNCLASSIFIED -->

# B3 Blocking Plan: Alert Quick-Actions and Assignment

## Scope

Implement Operations Commander and operator quick-actions for alert triage with full frontend-service wiring to feedback and assignment RPCs.

## Dependencies

1. Prior phase required: B2 completed.
2. Related backend contracts: existing `AssignAlert` and `SubmitFeedback` flows.
3. Next phase unblocked by this phase: P1.

## In-Scope

1. Alert actions: Inspect, Confirm, Reject, Assign.
2. Confirm/Reject wiring to `FeedbackService.SubmitFeedback`.
3. Assign flow wiring to `AlertService.AssignAlert`.
4. Optimistic UI, rollback, error display, and retry.

## Out-of-Scope

1. New backend RPC design.
2. Fusion observation rendering internals.
3. Non-commander alert workflow redesign.

## Required Implementation Targets

1. [web-cop-gpu/src/components/panels/AlertSidebar.tsx](web-cop-gpu/src/components/panels/AlertSidebar.tsx)
2. [web-cop-gpu/src/components/panels/TrackDetailPanel.tsx](web-cop-gpu/src/components/panels/TrackDetailPanel.tsx)
3. [web-cop-gpu/src/services/alerts.ts](web-cop-gpu/src/services/alerts.ts)
4. [web-cop-gpu/src/services/feedback.ts](web-cop-gpu/src/services/feedback.ts)
5. Browser E2E under [web-cop-gpu/e2e](web-cop-gpu/e2e)
6. Reference backend contract tests in [svc-alert/internal/handler/assign.go](svc-alert/internal/handler/assign.go) and [svc-feedback/internal/handler/feedback_test.go](svc-feedback/internal/handler/feedback_test.go)

## Required Contracts and Behavior

1. Confirm and Reject must map to correct feedback type values and payload fields.
2. Assign must capture assignee input and send correct alert and assignee identifiers.
3. Inspect must focus entity detail context and selected track state.
4. RPC failure must rollback optimistic state and expose retry.

## Execution Steps (Strict Order)

1. Add action affordances to alert card UI for Inspect, Confirm, Reject, and Assign.
   Expected artifact: all four actions visible for eligible alerts.
2. Implement service adapters for Confirm/Reject feedback payload construction.
   Expected artifact: typed request builders with validation guards.
3. Implement assign modal/picker and assignment request dispatch.
   Expected artifact: assignment action updates local state optimistically.
4. Implement inspect behavior to focus detail panel and track selection.
   Expected artifact: alert-to-detail context handoff works consistently.
5. Add centralized error state and retry handlers per action type.
   Expected artifact: rollback and retry supported on RPC failure.
6. Stop and validate checkpoint.
   Validation requirement: manual run proves each action transitions to expected UI state.
7. Add unit tests for payload correctness and optimistic/rollback behavior.
   Expected artifact: tests fail for wrong RPC payloads or missing rollback.
8. Add browser E2E for full alert action flows.
   Expected artifact: deterministic post-action assertions for status and assignment.

## Verification

1. Run frontend unit tests for alert sidebar and service adapters.
   Command: pnpm --dir web-cop-gpu test --run AlertSidebar alerts feedback
2. Run targeted browser E2E for operator quick-actions and assignment.
   Command: pnpm --dir web-cop-gpu playwright test -g "quick-action|assign|confirm|reject|inspect"
3. Run backend alert/feedback handler tests as contract confidence checks.
   Command: go test ./svc-alert/internal/handler ./svc-feedback/internal/handler
4. Required assertions:
   All four actions are visible and enabled where expected.
   Confirm/Reject generate correct feedback payload semantics.
   Assign action results in visible assigned state.
   Failures trigger rollback and retry path.

## Exit Criteria

1. All four quick-actions are implemented and functional in operator dashboard flow.
2. Assignment state is observable in UI and persists after refresh/state sync.
3. Deterministic unit and E2E assertions prevent regression.

## Handoff

1. Deliverables checklist:
   Updated alert UI actions.
   Service-layer request/response handling.
   Error and retry UX.
   Unit and E2E coverage.
2. Risks to record:
   Assignee identity source assumptions and stale list handling.
3. Evidence artifacts:
   Action flow screenshots, test summaries, failure/rollback proof, changed file list.
