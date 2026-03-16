<!-- CLASSIFICATION: UNCLASSIFIED -->

# P3 Phase Plan: Documentation, Demo Alignment, and Traceability Closure

## Scope

Align user guide, demo walkthrough, and requirements traceability to implemented commander behavior, then produce handoff-ready completion package with residual risk disclosure.

## Dependencies

1. Prior phases required: B1, B2, B3, B4, P1, and P2 completed.
2. Final phase: no downstream implementation phase.

## In-Scope

1. Update Operations Commander user guide to match implemented labels/flows.
2. Update demo showcase instructions to exact actions and expected outputs.
3. Create requirements-to-implementation traceability matrix for commander scope.
4. Publish completion report with residual risks and deferred items.

## Out-of-Scope

1. Net-new commander features.
2. Broad non-commander documentation refresh.
3. CI or infrastructure redesign.

## Required Implementation Targets

1. [docs/user_guide/operations_commander](docs/user_guide/operations_commander)
2. [docs/demo/demo_setup_run_showcase.md](docs/demo/demo_setup_run_showcase.md)
3. [docs/business/requirements.md](docs/business/requirements.md)
4. [docs/business/feature_list.md](docs/business/feature_list.md)
5. [docs/implementation/v5/operations_commander](docs/implementation/v5/operations_commander)

## Required Traceability Coverage

1. CR-UI-009 to CR-UI-015
2. CR-ING-011
3. CR-FB-008
4. CR-HIS-008
5. UC012, UC016, UC010 commander paths

## Execution Steps (Strict Order)

1. Inventory final implemented behavior and test evidence from B1 through P2 handoffs.
   Expected artifact: behavior/evidence catalog used as documentation source.
2. Update commander user guide pages to exact current UI labels and workflows.
   Expected artifact: no mismatch between guide text and UI behavior.
3. Update demo showcase script with explicit click path and expected screen outcomes.
   Expected artifact: deterministic demo steps executable by new operator.
4. Build traceability matrix linking requirement IDs to code locations and tests.
   Expected artifact: reviewable matrix with no unmapped commander-critical requirements.
5. Create completion report summarizing done, deferred, residual risk, and next backlog.
   Expected artifact: handoff-ready closure document.
6. Stop and validate checkpoint.
   Validation requirement: one full manual walkthrough of updated demo script succeeds.
7. Perform final consistency pass across guide, demo, and traceability docs.
   Expected artifact: no contradictory terminology or stale references.

## Verification

1. Run markdown lint or repository doc consistency checks where available.
   Command: pnpm --dir web-cop-gpu exec markdownlint ../docs/\*_/_.md
2. Execute manual commander demo walkthrough once end-to-end.
   Command: follow updated [docs/demo/demo_setup_run_showcase.md](docs/demo/demo_setup_run_showcase.md) step-by-step.
3. Required assertions:
   Guide steps match actual UI labels and actions.
   Demo expected outcomes are observable and reproducible.
   Traceability matrix maps each required ID to implementation and tests.
   Residual risks and deferred items are explicit.

## Exit Criteria

1. Docs, demo, and implementation behavior are consistent.
2. Traceability matrix is complete for commander scope requirements.
3. Completion package is ready for reviewer handoff.

## Handoff

1. Deliverables checklist:
   Updated commander user guide.
   Updated demo setup/showcase guide.
   Commander traceability matrix.
   Completion report with residual risks.
2. Risks to record:
   Any remaining environment-specific demo prerequisites that can affect reproducibility.
3. Evidence artifacts:
   Links to updated docs, walkthrough notes, matrix file path, changed file list.
