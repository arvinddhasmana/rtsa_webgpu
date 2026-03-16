<!-- CLASSIFICATION: UNCLASSIFIED -->

# P2 Phase Plan: Test Hardening and CI Regression Gates

## Scope

Convert commander coverage from permissive smoke checks into deterministic assertions that fail on capability regressions across unit, browser E2E, integration, and E2E layers.

## Dependencies

1. Prior phases required: B1, B2, B3, B4, and P1 completed.
2. Next phase unblocked by this phase: P3.

## In-Scope

1. Replace weak boolean-only assertions in browser E2E.
2. Add dedicated commander feature specs.
3. Add deterministic fixtures/mocks for alert and observation streams.
4. Align integration/E2E naming and comments to UC mappings.

## Out-of-Scope

1. New product feature implementation.
2. Cross-repo CI platform redesign.
3. Visual snapshot baselining not required for core commander gates.

## Required Implementation Targets

1. [web-cop-gpu/e2e](web-cop-gpu/e2e)
2. [web-cop-gpu/src](web-cop-gpu/src) test files for component/service units
3. [tests/integration](tests/integration)
4. [tests/e2e](tests/e2e)
5. Test scripts and documentation references in [tests/README.md](tests/README.md)

## Required New or Updated Browser Specs

1. commander-dashboard-switching
2. fusion-raw-observation-rendering
3. operator-quick-actions
4. assign-alert-flow
5. detail-source-attribution-timeline

## Execution Steps (Strict Order)

1. Audit existing browser specs and tag permissive assertions for replacement.
   Expected artifact: checklist mapping each weak assertion to strict expected-state assertion.
2. Implement deterministic fixtures/mocks for alerts, tracks, and raw observations.
   Expected artifact: stable inputs for repeatable test outcomes.
3. Add or rewrite commander-specific browser specs listed above.
   Expected artifact: high-signal tests that fail when capabilities are missing.
4. Strengthen frontend unit tests around role routing, quick-actions, and fusion transparency logic.
   Expected artifact: unit coverage for core commander state and actions.
5. Update integration and e2e test naming/comments with explicit UC mappings.
   Expected artifact: traceable tests aligned to requirements.
6. Stop and validate checkpoint.
   Validation requirement: deliberately break one commander feature and verify test failure occurs.
7. Finalize CI gating recommendation and thresholds for commander-specific suites.
   Expected artifact: documented blocking/non-blocking gate policy with rationale.

## Verification

1. Run frontend unit tests with coverage.
   Command: pnpm --dir web-cop-gpu test --run
   Command: pnpm --dir web-cop-gpu test --coverage
2. Run browser E2E suite headless.
   Command: pnpm --dir web-cop-gpu playwright test
3. Run representative backend integration and e2e tests.
   Command: go test ./tests/integration/...
   Command: go test ./tests/e2e/...
4. Required assertions:
   Commander dashboard behavior failures cause test failure.
   Missing quick-action wiring causes test failure.
   Missing source attribution/timeline parity causes test failure.
   Fusion raw observation rendering regressions cause test failure.
5. Coverage gate:
   Maintain or exceed 80 percent line coverage for touched commander-related test targets where measurable.

## Exit Criteria

1. Commander capability regressions produce hard, deterministic failures in automated tests.
2. Browser and unit test suites are stable under deterministic fixtures.
3. UC-mapped test traceability is explicit and reviewable.

## Handoff

1. Deliverables checklist:
   Updated browser specs.
   Fixture/mocking strategy.
   Unit/integration/e2e hardening updates.
   CI gate recommendation.
2. Risks to record:
   Potential flakiness from environment-coupled services.
3. Evidence artifacts:
   Coverage summaries, test run summaries, failure-injection proof, changed file list.
