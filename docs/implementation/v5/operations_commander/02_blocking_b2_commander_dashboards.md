<!-- CLASSIFICATION: UNCLASSIFIED -->

# B2 Blocking Plan: Commander Dashboard Trio

## Scope

Implement Operations Commander dashboard set and route/layout wiring for Fusion, Multi-Domain, and Operator UI views.

## Dependencies

1. Prior phase required: B1 completed.
2. Next phases unblocked by this phase: B3 and B4.

## In-Scope

1. Create dedicated commander dashboard components.
2. Wire dashboard selection and rendering through app shell composition.
3. Preserve existing sensor health and coverage dashboard behavior.

## Out-of-Scope

1. Quick-action backend wiring.
2. Raw observation stream rendering details.
3. Test hardening beyond phase-required assertions.

## Required Implementation Targets

1. [web-cop-gpu/src/App.tsx](web-cop-gpu/src/App.tsx)
2. [web-cop-gpu/src/components/shell/AppShell.tsx](web-cop-gpu/src/components/shell/AppShell.tsx)
3. New commander dashboard components under [web-cop-gpu/src/components](web-cop-gpu/src/components)
4. Existing panels/timeline composition points under [web-cop-gpu/src/components/panels](web-cop-gpu/src/components/panels)
5. Browser E2E coverage under [web-cop-gpu/e2e](web-cop-gpu/e2e)

## Required Dashboard Minimums

1. Fusion dashboard contains map container, side panel container, and observation/fused layer mount points.
2. Multi-Domain dashboard contains domain KPI overlay and layer toggle controls.
3. Operator UI dashboard contains alert column, detail pane, and timeline pane.

## Execution Steps (Strict Order)

1. Define component boundaries and props contract for the three dashboards.
   Expected artifact: explicit component interface and role-safe rendering contracts.
2. Implement Fusion dashboard skeleton with map and side panel placeholders.
   Expected artifact: renderable Fusion view for commander role.
3. Implement Multi-Domain dashboard skeleton with KPI overlay and layer toggles.
   Expected artifact: renderable Multi-Domain view.
4. Implement Operator UI dashboard skeleton with three-pane layout.
   Expected artifact: renderable Operator UI view.
5. Wire App-level dashboard routing/selection to render commander trio based on role state.
   Expected artifact: commander can switch among all three dashboards.
6. Stop and validate checkpoint.
   Validation requirement: manual navigation across commander dashboards without breaking non-commander dashboards.
7. Add component tests for each dashboard render and switch behavior.
   Expected artifact: deterministic tests for panel presence and tab transitions.
8. Add browser E2E for commander dashboard navigation and expected composition visibility.
   Expected artifact: failing tests if any dashboard is absent or miswired.

## Verification

1. Run frontend unit tests for app shell and commander dashboard components.
   Command: pnpm --dir web-cop-gpu test --run AppShell dashboard
2. Run targeted browser E2E for commander navigation.
   Command: pnpm --dir web-cop-gpu playwright test -g "commander|dashboard|navigation"
3. Required assertions:
   All three commander dashboards are selectable.
   Expected panel composition exists in each dashboard.
   Sensor role dashboards are unchanged.
4. Coverage gate:
   Touched dashboard/shell logic must retain or exceed 80 percent line coverage target.

## Exit Criteria

1. Commander can switch among Fusion, Multi-Domain, and Operator UI dashboards.
2. No visual/layout regression in sensor health and coverage flows.
3. Tests capture dashboard composition and selection behavior.

## Handoff

1. Deliverables checklist:
   Three dashboard components.
   App wiring updates.
   Unit and E2E tests for render/switch paths.
2. Risks to record:
   Any shared state coupling that may block B3/B4 integration.
3. Evidence artifacts:
   Navigation screenshots for all three dashboards, test summaries, changed file list.
