<!-- CLASSIFICATION: UNCLASSIFIED -->

# Operations Commander Implementation Order

## Scope

This document is the canonical execution index for completing Operations Commander remediation in dependency order. Use it as the single source of sequencing, phase gates, and handoff payload requirements.

## Dependency Chain

1. B1: RBAC shell and role-dashboard model.
2. B2: Commander dashboard trio and route/render wiring.
3. B3: Alert quick-actions and assignment wiring.
4. B4: Fusion raw observation rendering.
5. P1: Detail, timeline, and tactical edge parity.
6. P2: Test hardening and CI regression gates.
7. P3: Docs/demo/traceability alignment.

## Global In-Scope

1. Commander role parity for UC012, UC016, and UC010 workflow paths.
2. Frontend role-aware dashboards and operator interactions.
3. Browser E2E and service-level test evidence that fails on regressions.
4. Traceability updates for CR-UI-009 through CR-UI-015, CR-ING-011, CR-FB-008, and CR-HIS-008.

## Global Out-of-Scope

1. Net-new backend domains unrelated to commander flow.
2. UI visual redesign outside required parity and usability fixes.
3. Infrastructure changes not required for feature delivery or testing.

## Required Handoff Payload After Every Phase

1. Changed file list.
2. Test commands executed.
3. Pass/fail summary.
4. Unresolved issues and blockers.
5. Evidence artifact paths.

## Global Stop-and-Validate Gates

1. Do not start the next phase unless previous exit criteria are satisfied.
2. Maintain existing role/dashboard behavior for non-commander flows.
3. Record any deferred item in phase handoff with rationale and impact.

## Global Completion Criteria

1. UC012, UC016, and UC010 commander paths are implemented end-to-end.
2. Browser E2E has deterministic assertions for quick-actions and dashboard behavior.
3. Traceability table is updated and linked to implementation/tests.
4. Completion package includes evidence for all phases and residual risks.

## Evidence Package Checklist

1. Unit/integration/E2E command transcripts or captured summaries.
2. Browser screenshots or video for commander dashboard and quick-actions.
3. Updated docs links for user guide and demo script.
4. Final traceability matrix and completion report.
