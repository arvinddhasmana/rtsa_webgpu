<!-- CLASSIFICATION: UNCLASSIFIED -->

# MIL-2525 Symbology West Asia Demo — Implementation Plan

## Scope

Implement MIL-STD-2525C / NATO APP-6 standard symbology for fused tracks rendered in
the RTSA WebGPU COP, with mock data focused on the West Asia / Iran / Persian Gulf threat
scenario.  The deliverable is a fully functional demo that renders 40 mock tracks using
correct silhouette shapes (domain) and colour fills (affiliation) over the West Asia
bounding box (lon 44°E–63°E, lat 24°N–37°N).

## Context

This plan was produced in response to a review of the existing icons strategy for fused
tracks following the RTSA Symbology Implementation Specification.  The decision was to
use the existing WebGPU stack and SVG overlay layer rather than introducing an external
symbology library.

## In-Scope

1. `TrackSymbol` SVG component (pure SVG, no canvas/WebGPU).
2. `track-symbol.ts` type definitions: `TrackDomain`, `TrackAffiliation`, `TrackContext`.
3. Mock data generator aligned with `TrackDomain` / `TrackAffiliation` enumerations.
4. Unit tests for `TrackSymbol` covering all domains, affiliations, and contexts.
5. Critical-blocker fixes required for the demo test suite to pass.

## Out-of-Scope

1. WGSL shader changes (handled separately in `track-icons.wgsl`).
2. Backend ingestion changes.
3. NATO STANAG 5516 binary encoding changes.

---

## Completed Work

| Item | File | Status |
|------|------|--------|
| `TrackDomain` enum (AIR/SURFACE/SUBSURFACE/LAND/SPACE/CYBER) | `src/types/track-symbol.ts` | ✅ Done |
| `TrackAffiliation` enum (UNKNOWN/PENDING/FRIENDLY/NEUTRAL/SUSPECT/HOSTILE) | `src/types/track-symbol.ts` | ✅ Done |
| `TrackContext` enum (REAL/EXERCISE/SIMULATION/TEST) | `src/types/track-symbol.ts` | ✅ Done |
| `TrackSymbolProps` interface | `src/types/track-symbol.ts` | ✅ Done |
| `TrackSymbol` SVG component (shape + colour + context modifiers) | `src/components/symbols/TrackSymbol.tsx` | ✅ Done |
| Mock data generator — West Asia geographic bounding box | `src/gpu/mock-data.ts` | ✅ Done |
| Domain-weighted distribution (35% Air / 25% Surface / …) | `src/gpu/mock-data.ts` | ✅ Done |
| Affiliation-weighted distribution (30% Hostile / 20% Suspect / …) | `src/gpu/mock-data.ts` | ✅ Done |
| Unit tests — TrackSymbol (27 assertions) | `tests/components/TrackSymbol.test.tsx` | ✅ Done |
| B1: RBAC shell — five-role model + role-dashboard guard | `src/signals/viewport.ts` | ✅ Done |
| B2: Commander dashboard trio (Fusion / Multi-Domain / Operator UI) | `src/components/dashboard/` | ✅ Done |
| B3: Alert quick-actions (Inspect / Confirm / Reject / Assign) | `src/components/panels/AlertSidebar.tsx` | ✅ Done |
| B4: Fusion raw observation subscription and KPI side panel | `src/components/dashboard/FusionCommanderDashboard.tsx` | ✅ Done |

---

## Critical Blockers (Fix First)

These items block the demo test suite from running.  They must be resolved before any
further development or CI merge.

### CB-1 — `@bufbuild/protobuf` unresolvable from gen/ts in Vitest (**FIXED**)

**Symptom:** Three component test suites fail at transform time with:

```
Error: Failed to resolve import "@bufbuild/protobuf" from
"../gen/ts/rtsa/common/v1/types_pb.ts"
```

**Affected tests:**
- `src/components/dashboard/CommanderDashboards.test.tsx`
- `tests/components/AlertSidebar.test.tsx`
- `tests/components/TrackDetailPanel.test.tsx`

**Root cause:** The `@gen` alias in `vitest.config.ts` resolves module paths to
`../gen/ts/`, which lives outside the `web-cop-gpu` project directory.  When Vite
processes files from that directory it cannot resolve `@bufbuild/protobuf` via normal
`node_modules` traversal.

**Fix:** Add an explicit `@bufbuild/protobuf` entry to the `resolve.alias` table in
`vitest.config.ts`, pointing to `node_modules/@bufbuild/protobuf` inside the
`web-cop-gpu` directory.  This gives Vite a deterministic path regardless of the
context file location.

**Status:** ✅ Fixed in this PR.

### CB-2 — `window.innerWidth` accessed at signal initialisation in `FusionCommanderDashboard` (**FIXED**)

**Symptom:** `createSignal({ x: window.innerWidth - 420, y: 80 })` runs at module
evaluation time, before any DOM is mounted.  In JSDOM (Vitest environment) and in
SSR contexts `window.innerWidth` is `0`, placing the draggable toast at `x = -420`.
The value is not a test failure by itself today, but it produces an off-screen
floating panel in any headless render and will cause sub-pixel pick failures in
future viewport tests.

**Fix:** Replace the immediate `window.innerWidth` access with a lazy signal
initialisation that reads the value inside an `onMount` callback, defaulting to `0`
and correcting to `window.innerWidth - 420` after mount.

**Status:** ✅ Fixed in this PR.

---

## Pending Items

> Items below were not completed in this implementation session.
> They are tracked here for the next agent or developer to action.

### P-1 — Uniforms `makeViewProjection` test failures (pre-existing)

**File:** `tests/gpu/uniforms.test.ts`

**Symptom:** Three assertions fail — the test expects the view-projection matrix to
match a simple orthographic projection with `sx = sy = 1` at `scale=1`, but the
actual `makeViewProjection` implementation produces a Web Mercator tile-based scale
(`~20.48`) that does not match the test expectation.

**Priority:** Medium.  Does not block the demo but must be resolved before the CI
regression gate in phase P2.

**Recommended action:** Either update the test to match the actual tile-space scale
behaviour, or re-examine whether `makeViewProjection` is intended to return a
normalised NDC matrix (test expectation) vs a tile-scale matrix (current implementation).

### P-2 — West Asia mock alert data in `startMockAlertStream`

**File:** `src/services/alerts.ts`

**Symptom:** The mock alert stream generates generic alert descriptions
(`"[MOCK] Real-time sensor anomaly detected"`) with no West Asia geographic context.
The demo would be more compelling with alerts that reference Persian Gulf sectors,
Strait of Hormuz, or Iranian airspace.

**Recommended action:** Add a pool of West Asia–specific mock alert descriptions and
rotate through them in `startMockAlertStream`.

### P-3 — `accumulated_alerts_logic_here` function rename

**File:** `src/services/alerts.ts` — line 174

**Symptom:** The helper function `accumulated_alerts_logic_here()` is named with a
non-idiomatic snake_case style that does not match the camelCase conventions used
throughout the codebase.

**Recommended action:** Rename to `getCurrentAlerts()` or inline the `alerts()` call
directly at the call site.

### P-4 — Phase P1: Detail, timeline, and tactical edge parity

**Reference plan:** `docs/implementation/v5/operations_commander/05_phase_p1_detail_timeline_edge.md`

Not started.  Requires B3 and B4 to be stable.

### P-5 — Phase P2: Test hardening and CI regression gates

**Reference plan:** `docs/implementation/v5/operations_commander/06_phase_p2_test_hardening.md`

Not started.  Requires P1 to be complete.

### P-6 — Phase P3: Docs, demo script, and traceability alignment

**Reference plan:** `docs/implementation/v5/operations_commander/07_phase_p3_docs_demo_traceability.md`

Not started.  Requires P2 to be complete.

### P-7 — Browser E2E test suite for commander dashboard navigation

**Files:** `web-cop-gpu/e2e/`

No Playwright E2E tests exist for commander-role dashboard switching, alert
quick-actions, or Fusion raw-observation rendering.  Required by B2, B3, and B4 exit
criteria.

### P-8 — `operators.ts` service — no unit tests

**File:** `src/services/operators.ts`

The operator search and assignment services lack unit tests.  The B3 exit criteria
require deterministic assertions for the assign flow, which depends on correct
operator-list behaviour.
