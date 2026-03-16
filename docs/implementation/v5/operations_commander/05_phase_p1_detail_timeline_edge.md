<!-- CLASSIFICATION: UNCLASSIFIED -->

# P1 Phase Plan: Detail Panel, Timeline, and Tactical Edge Parity

## Scope

Close commander parity gaps for entity detail context, source attribution, integrated timeline behavior, and tactical edge/disconnected UX indicators.

## Dependencies

1. Prior phases required: B3 and B4 completed.
2. Next phase unblocked by this phase: P2.

## In-Scope

1. Add `SourceAttribution` section in detail panel.
2. Integrate contextual `EntityTimeline` in detail flow while retaining global timeline where needed.
3. Integrate `FeedbackForm` with selected alert/track context.
4. Add tactical edge/disconnected banners and sync status indicators aligned to guide behavior.

## Out-of-Scope

1. New analytics algorithms.
2. New backend timeline protocol.
3. Styling overhauls outside parity requirements.

## Required Implementation Targets

1. [web-cop-gpu/src/components/panels/TrackDetailPanel.tsx](web-cop-gpu/src/components/panels/TrackDetailPanel.tsx)
2. [web-cop-gpu/src/components/timeline/EventTimeline.tsx](web-cop-gpu/src/components/timeline/EventTimeline.tsx)
3. [web-cop-gpu/src/components/panels/FeedbackForm.tsx](web-cop-gpu/src/components/panels/FeedbackForm.tsx)
4. Related state wiring in [web-cop-gpu/src/App.tsx](web-cop-gpu/src/App.tsx)
5. Browser E2E specs under [web-cop-gpu/e2e](web-cop-gpu/e2e)

## Execution Steps (Strict Order)

1. Define detail panel data contract for source attribution fields and timeline linkage.
   Expected artifact: typed view model for detail context.
2. Implement `SourceAttribution` rendering with empty/loading/error states.
   Expected artifact: deterministic section in detail panel for source contributions.
3. Embed context-aware timeline component in detail flow.
   Expected artifact: entity timeline synchronized with selected alert/track.
4. Bind feedback form defaults to selected entity and active alert state.
   Expected artifact: in-context feedback payload prepopulation.
5. Implement tactical edge state indicators and disconnected banners.
   Expected artifact: visible status model aligned to guide expectations.
6. Stop and validate checkpoint.
   Validation requirement: manual walkthrough confirms source attribution, context timeline, and edge states.
7. Add component tests for conditional rendering and context wiring.
   Expected artifact: tests fail on missing sections or wrong context propagation.
8. Add browser E2E for disconnected and degraded state indicators.
   Expected artifact: deterministic assertions for indicator visibility and copy.

## Verification

1. Run targeted frontend unit tests for detail/timeline/feedback components.
   Command: pnpm --dir web-cop-gpu test --run TrackDetailPanel EventTimeline FeedbackForm
2. Run browser E2E for tactical edge and context behavior.
   Command: pnpm --dir web-cop-gpu playwright test -g "detail|timeline|edge|disconnected|feedback"
3. Required assertions:
   Source attribution section appears with expected content states.
   Entity timeline context tracks selected entity.
   Feedback form uses active context values.
   Edge/disconnected indicators reflect simulated connectivity states.

## Exit Criteria

1. CR-UI-015 parity is met for detail/timeline/source attribution behavior.
2. Tactical edge UX indicators and copy align with user guide expectations.
3. Automated tests enforce new behavior.

## Handoff

1. Deliverables checklist:
   Detail panel enhancements.
   Context timeline integration.
   Feedback form contextual wiring.
   Edge/disconnected indicators.
   Tests and evidence.
2. Risks to record:
   State synchronization complexity between global and context timelines.
3. Evidence artifacts:
   UI screenshots for source attribution and edge states, test summaries, changed file list.
