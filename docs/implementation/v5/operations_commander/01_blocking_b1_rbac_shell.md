<!-- CLASSIFICATION: UNCLASSIFIED -->

# B1 Blocking Plan: RBAC Shell and Role-Dashboard Model

## Scope

Implement a complete five-role UI model and enforce role-scoped dashboard access/default behavior required for Operations Commander readiness.

## Dependencies

1. Prior phase: none.
2. Next phase unblocked by this phase: B2.

## In-Scope

1. Extend role/domain model to five roles.
2. Add role-to-default-dashboard mapping.
3. Add role-to-allowed-dashboards mapping.
4. Enforce dashboard selection guards when role changes.

## Out-of-Scope

1. Commander dashboard content implementation.
2. Alert quick-action behavior.
3. Fusion observation rendering.

## Required Implementation Targets

1. [web-cop-gpu/src/signals/viewport.ts](web-cop-gpu/src/signals/viewport.ts)
2. [web-cop-gpu/src/components/toolbar/RoleSelector.tsx](web-cop-gpu/src/components/toolbar/RoleSelector.tsx)
3. [web-cop-gpu/src/components/toolbar/DashboardSelector.tsx](web-cop-gpu/src/components/toolbar/DashboardSelector.tsx)
4. [web-cop-gpu/src/App.tsx](web-cop-gpu/src/App.tsx)
5. Related frontend unit tests under [web-cop-gpu/src](web-cop-gpu/src)
6. Browser E2E role landing assertions under [web-cop-gpu/e2e](web-cop-gpu/e2e)

## Required Symbols and Contracts

1. Extend current role union to include analyst, security officer, sensor operator, and nato liaison while preserving operations commander.
2. Add explicit role default dashboard constant map.
3. Add explicit role allowed dashboard constant map.

## Execution Steps (Strict Order)

1. Define role and dashboard enums/types in central signal/domain file.
   Expected artifact: single source of truth role/dashboard types and map constants.
2. Implement default dashboard selection by role on session init and role change.
   Expected artifact: deterministic role landing behavior.
3. Implement allowed dashboard guard that resets invalid selection to role default.
   Expected artifact: no invalid dashboard state after role switch.
4. Update role selector options and labels for all five roles.
   Expected artifact: selector displays all supported roles.
5. Update dashboard selector to show role-scoped choices only.
   Expected artifact: disallowed dashboards hidden or disabled by role policy.
6. Stop and validate checkpoint.
   Validation requirement: manual role switching proves deterministic defaulting and guard behavior.
7. Add/extend unit tests for role and dashboard selector behavior.
   Expected artifact: tests fail when two-role assumptions reappear.
8. Add browser E2E assertions for role default landing.
   Expected artifact: each role lands on mapped default dashboard route/container.

## Verification

1. Run frontend unit tests for role/dashboard modules.
   Command: pnpm --dir web-cop-gpu test --run RoleSelector DashboardSelector
2. Run targeted browser E2E for role landing.
   Command: pnpm --dir web-cop-gpu playwright test -g "role|dashboard|landing"
3. Required assertions:
   Each role has exactly one default dashboard.
   Changing role invalidates disallowed dashboard selection.
   No stale two-role labels remain in UI.
4. Coverage gate:
   Touched frontend files retain or improve line coverage with target 80 percent plus for changed logic.

## Exit Criteria

1. No stale two-role assumptions remain.
2. Operations commander defaults to Fusion container.
3. Role-scoped dashboard visibility is enforced in UI state and tests.

## Handoff

1. Deliverables checklist:
   Updated role/dashboard domain model.
   Updated selectors and app wiring.
   Unit and E2E test updates.
2. Risks to record:
   Any legacy dashboard IDs still referenced by old code paths.
3. Evidence artifacts:
   Test output summary, screenshot per role landing, changed file list.
