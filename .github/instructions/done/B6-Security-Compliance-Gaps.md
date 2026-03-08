<!-- CLASSIFICATION: UNCLASSIFIED -->

# B6 — Security and Compliance Gaps

> **Batch**: B6 of 7
> **Theme**: Security Compliance Gaps
> **Priority**: WARNING/BLOCKING — SDLC policy violations and STANAG colour mapping errors
> **Agent Profile**: `greatest-ever-dev-forworkflow`
> **Source**: v4 Implementation Review R-018 through R-022

---

## Context

This batch addresses security and compliance violations identified across both the Go backend (`pkg/flatbuf/serializer.go`) and the SolidJS frontend (`web-cop-gpu/src/`). Issues include hardcoded operator identity strings (authentication gap), `console.log` statements in production code (logging policy violation), and alert flags that are hardcoded to zero in the backend serializer (rendering halos impossible in a live demo).

---

## Issue R-018 — BLOCKING (demo): `alertFlags` always 0 in Go serializer

**File**: `pkg/flatbuf/serializer.go`

**Problem**: `alertFlags := uint32(0) // Phase 3 will wire alert state; keep zero for now` — every track sent to the frontend has `alert_flags = 0`, so no track ever shows a pulsing alert halo on the GPU canvas regardless of threat level.

**Required Fix**: Derive `alertFlags` from the `threat_level` field of the incoming `TrackUpdate` proto message as a temporary heuristic until Phase 3 wires proper alert state:

```go
alertFlags := uint32(0)
if update.ThreatLevel >= 4 { // Suspect (4) or Hostile (5)
    alertFlags = 1
}
```

Read the `TrackUpdate` proto definition in `proto/rtsa/` to confirm the field name and type before editing.

This is sufficient for the demo to show halos on suspect/hostile tracks without requiring Phase 3 alert infrastructure.

**Test**: `go test ./pkg/flatbuf/... -race` — add a test case where `ThreatLevel = 5` results in `alertFlags = 1` in the serialized output.

---

## Issue R-019 — WARNING: Threat level colour mapping inconsistent with STANAG APP-6

**Files**:

- `pkg/flatbuf/serializer.go` or `pkg/flatbuf/layout.go` — Go threat-level constants
- `web-cop-gpu/src/shaders/trail.wgsl` — `threat_color` function
- `web-cop-gpu/src/shaders/halos.wgsl` — halo colour mapping

**Problem**: The Go serializer maps `FRIENDLY → 2`, `NEUTRAL → 3`. The WGSL `trail_color` maps `3 → amber`. This produces **amber trails for neutral tracks**, which contradicts STANAG APP-6 (neutral = green). Additionally, friendly tracks `(2 → green)` accidentally receive the neutral colour meaning.

**Required Fix**:

1. Add a block comment in `pkg/flatbuf/layout.go` (or a dedicated `threat_level.go` constant file) that explicitly documents the mapping:

   ```go
   // ThreatLevel colour mapping (matches WGSL threat_color() function)
   // 0 = Unknown  → grey
   // 1 = Pending  → blue
   // 2 = Friendly → green
   // 3 = Neutral  → green   ← APP-6: neutral is green, not amber
   // 4 = Suspect  → amber/orange
   // 5 = Hostile  → red
   ```

2. Update the WGSL `threat_color` function in `trail.wgsl` and `halos.wgsl` so that **both level 2 (Friendly) and level 3 (Neutral) return green**. Level 4 (Suspect) should return amber/orange, and level 5 (Hostile) should return red. This is the STANAG APP-6 standard.

3. Friendly tracks must NOT show alert halos — ensure the halo render pass checks `alert_flags > 0`, not `threat_level >= 2`. Halos are for alerted tracks only, independent of classification.

Read `docs/sdlc_guidelines/08_tech_specific/wgsl_shader_standards.md` before editing WGSL files.

---

## Issue R-020 — WARNING: `console.log` statements in production code

**Files with violations**:

- `web-cop-gpu/src/workers/render-worker.ts` (multiple lines)
- `web-cop-gpu/src/workers/data-worker.ts`
- `web-cop-gpu/src/services/alerts.ts`

**SDLC Rule**: `solidjs_standards.md §7.1` — No `console.log` in production code.

**Required Fix**: For each `console.log` / `console.warn` call:

1. If the log is for error/failure reporting (e.g., WebGPU init failure, Wasm load failure), wrap it with a dev-mode guard:

   ```typescript
   if (import.meta.env.DEV) {
     console.error("[render-worker] GPU device lost:", reason);
   }
   ```

2. If the log is for debug tracing that should never appear in production, remove it entirely.

3. **Never** log: JWT tokens, track payloads, user identities, or any structured data that could contain PII.

4. Run `cd web-cop-gpu && pnpm lint` after changes — confirm no console usage ESLint warnings remain.

---

## Issue R-021 — WARNING: `FeedbackForm` hardcoded `operatorId`

**File**: `web-cop-gpu/src/components/panels/FeedbackForm.tsx`

**Problem**: `operatorId: "operator"` is hardcoded. In a real session, the operator identity comes from the authenticated JWT claims.

**Required Fix**:

1. Create or use an existing SolidJS context/signal that holds the current authenticated user identity (e.g., `AuthContext` or `userSignal`). Do not create a new abstraction if one already exists — search `web-cop-gpu/src/` for existing auth context patterns first.
2. Replace the hardcoded `"operator"` string with the value from the auth context/signal.
3. Make the `operatorId` field gracefully handle the unauthenticated case (e.g., use `"anonymous"` if the JWT has not yet been acquired, but do not submit feedback without a valid session in production).

**Constraint**: Do not destructure props in SolidJS components — access `props.xyz` directly per `solidjs_standards.md`.

---

## Issue R-022 — WARNING: `AlertSidebar` hardcoded `operatorId`

**File**: `web-cop-gpu/src/components/panels/AlertSidebar.tsx`

**Problem**: `await acknowledgeAlert(props.alert.alertId, "operator")` — same hardcoded identity pattern as R-021.

**Required Fix**: Apply the same fix as R-021 — source the `operatorId` from the auth context/signal. The `acknowledgeAlert` call must always pass the authenticated operator identity, not a literal string.

---

## Implementation Rules

1. Read every file before editing. Search the codebase for existing auth context patterns before creating anything new.
2. **Never** hardcode secrets, credentials, tokens, or user identifiers.
3. Do not introduce new `console.log` calls while fixing R-020.
4. All files must retain the `// CLASSIFICATION: UNCLASSIFIED` header.
5. Go code must not use `panic()` in non-test files.
6. Follow `docs/sdlc_guidelines/04_coding_standards/secure_coding.md` and `solidjs_standards.md`.
7. WGSL changes must be verified against `docs/sdlc_guidelines/08_tech_specific/wgsl_shader_standards.md`.

---

## Tests Required

**Go tests**:

- `go test ./pkg/flatbuf/... -race` — test `alertFlags` derivation from `ThreatLevel`.
- Add a table-driven test covering all threat levels (0–5) and verifying the Go output colour constant matches the WGSL mapping.

**TypeScript tests**:

- Test that `FeedbackForm` renders with the correct `operatorId` from the auth context, not the hardcoded literal.
- Test that `AlertSidebar` acknowledgement calls `acknowledgeAlert` with the auth context identity.

**Lint check**:

- `cd web-cop-gpu && pnpm lint` — zero console-usage warnings.

---

## Verification Steps

1. `go test ./pkg/flatbuf/... -race` — all tests pass.
2. `cd web-cop-gpu && pnpm test` — all unit tests pass.
3. `cd web-cop-gpu && pnpm lint` — no console-log ESLint violations.
4. `cd web-cop-gpu && pnpm dev` — open the app, observe that suspect/hostile mock tracks show pulsing halos and the correct STANAG colour (amber/red), while friendly/neutral tracks show green icons without halos.

---

## PR Instructions

- PR title: `fix: wire alert flags, correct STANAG colours, gate console.log, replace hardcoded operatorId (B6)`
- Label: `ai-orchestrator`
- After the PR is merged, move this file from `.github/instructions/todo/B6-Security-Compliance-Gaps.md` to `.github/instructions/done/B6-Security-Compliance-Gaps.md`.
