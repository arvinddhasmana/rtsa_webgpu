<!-- CLASSIFICATION: UNCLASSIFIED -->

# B4 Blocking Plan: Fusion Raw Observation Transparency

## Scope

Implement UC016 fusion transparency behavior by rendering raw sensor observations alongside fused tracks, with role-appropriate side metrics and viewport-aware stream updates.

## Dependencies

1. Prior phase required: B2 completed.
2. Recommended sequence: execute after B3 for stable shared dashboard state behavior.
3. Next phase unblocked by this phase: P1.

## In-Scope

1. Frontend subscription to `StreamSensorObservations`.
2. Sensor icon mapping and legend parity.
3. Concurrent rendering of fused tracks and raw observations.
4. Debounced viewport bounding-box update behavior.
5. Fusion side panel KPIs: active counts, confidence buckets, contribution rates.

## Out-of-Scope

1. Backend stream protocol redesign.
2. Classification policy logic changes on backend.
3. New map engine architecture.

## Required Implementation Targets

1. [web-cop-gpu/src/services](web-cop-gpu/src/services)
2. [web-cop-gpu/src/components](web-cop-gpu/src/components)
3. [web-cop-gpu/src/signals/viewport.ts](web-cop-gpu/src/signals/viewport.ts)
4. [svc-track/internal/handler/stream_observations.go](svc-track/internal/handler/stream_observations.go) as contract reference
5. Browser E2E fixtures/specs under [web-cop-gpu/e2e](web-cop-gpu/e2e)

## Security and Data Handling Requirements

1. Treat stream payloads as untrusted and validate client-side schema before state mutation.
2. Do not expose hidden-classification hints in iconography, labels, or counts.
3. Keep logging structured and avoid raw payload dumps in INFO level logs.

## Execution Steps (Strict Order)

1. Implement typed service adapter for `StreamSensorObservations` parsing and lifecycle management.
   Expected artifact: reconnect-safe observation stream client with schema guard.
2. Implement sensor-type to icon map and legend model.
   Expected artifact: deterministic icon selection and legend rendering contract.
3. Add state container for fused and raw layers with independent visibility toggles.
   Expected artifact: concurrent layer rendering path.
4. Implement debounced bounding-box update requests on viewport movement.
   Expected artifact: controlled stream/filter update rate without flooding.
5. Build Fusion side panel KPI aggregation and rendering logic.
   Expected artifact: live counts and confidence/contribution summaries.
6. Stop and validate checkpoint.
   Validation requirement: manual session shows multiple sensor types and synchronized side metrics.
7. Add unit tests for icon mapping, bbox builder, and KPI aggregation logic.
   Expected artifact: deterministic pure-function coverage for mapping/aggregation.
8. Add integration/browser tests for stream parsing and rendering assertions.
   Expected artifact: fixture-driven deterministic checks for icon and count visibility.

## Verification

1. Run frontend unit tests for fusion service and map-layer utilities.
   Command: pnpm --dir web-cop-gpu test --run fusion observation viewport
2. Run integration tests for stream parsing/state update pipeline.
   Command: pnpm --dir web-cop-gpu test --run stream observation
3. Run targeted browser E2E for fusion raw observation rendering.
   Command: pnpm --dir web-cop-gpu playwright test -g "fusion|raw observation|legend|kpi"
4. Required assertions:
   Raw and fused layers render concurrently.
   At least two sensor icon types appear with correct legend labels.
   Viewport moves trigger debounced bbox update requests.
   Side panel metrics update from stream state.

## Exit Criteria

1. Fusion dashboard demonstrates raw plus fused transparency with live KPI side panel.
2. Deterministic tests fail when icon mapping, bbox updates, or KPI computation regress.
3. Security/data handling constraints are preserved.

## Handoff

1. Deliverables checklist:
   Observation stream adapter.
   Layer and legend rendering.
   KPI side panel logic.
   Unit, integration, and browser tests.
2. Risks to record:
   Performance hotspots from high update rates and large observation sets.
3. Evidence artifacts:
   Fusion screenshots showing icon diversity, KPI snapshots, test summaries, changed file list.
