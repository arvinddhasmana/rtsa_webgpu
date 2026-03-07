<!-- CLASSIFICATION: UNCLASSIFIED -->

# B5 — WebTransport ↔ Data Worker Integration Gaps

> **Batch**: B5 of 7
> **Theme**: WebTransport Integration
> **Priority**: BLOCKING — Real backend connection is non-functional; app always runs in mock mode
> **Agent Profile**: `greatest-ever-dev-forworkflow`
> **Source**: v4 Implementation Review R-014 through R-017

---

## Context

The Data Worker (`web-cop-gpu/src/workers/data-worker.ts`) is responsible for establishing a QUIC WebTransport connection to the Go backend and writing decoded FlatBuffer track records into the SharedArrayBuffer. Three gaps prevent this from working: the WebTransport URL is never passed to the worker, no JWT authentication token is sent with the connection request, and the Wasm decoder module path may not survive the production build.

---

## Issue R-014 — BLOCKING: Data Worker never receives WebTransport URL

**File**: `web-cop-gpu/src/App.tsx`

**Problem**: The `DataInitMessage` sent to the Data Worker does not include the `url` field. The worker never receives a WebTransport endpoint and permanently runs in mock mode.

**Required Fix**: Read the WebTransport URL from the Vite environment variable `VITE_WEBTRANSPORT_URL` and include it in the init message:

```typescript
const dataInit: DataInitMessage = {
  type: "init",
  sab,
  url: import.meta.env.VITE_WEBTRANSPORT_URL, // undefined in dev → mock mode
};
```

When `VITE_WEBTRANSPORT_URL` is undefined (local dev without a backend), the Data Worker must detect `url === undefined` and fall back to mock mode. Verify this fallback is already implemented in `data-worker.ts`; if not, add the conditional.

Also ensure `web-cop-gpu/.env.example` (or `web-cop-gpu/.env`) documents this variable:

```
VITE_WEBTRANSPORT_URL=https://localhost:4433/tracks
```

---

## Issue R-015 — BLOCKING: No JWT token sent with WebTransport connection

**File**: `web-cop-gpu/src/workers/data-worker.ts`

**Problem**: The Data Worker calls `new WebTransport(url)` without appending an authentication token. The Go WebTransport server (`pkg/webtransport/server.go`) validates a JWT in the `token` query parameter and returns HTTP 401 without it.

**Required Fix**:

The Data Worker cannot perform gRPC calls directly (no DOM, no Fetch API by default in all browsers). The recommended pattern is:

1. Before launching the Data Worker, the **main thread** (`web-cop-gpu/src/App.tsx`) acquires a JWT via the gRPC-Web cold path (use the existing auth service or create `web-cop-gpu/src/services/auth.ts` if it doesn't exist).
2. The JWT is passed to the Data Worker as part of the `DataInitMessage`:
   ```typescript
   const dataInit: DataInitMessage = {
     type: "init",
     sab,
     url: import.meta.env.VITE_WEBTRANSPORT_URL,
     token: await fetchAuthToken(), // calls gRPC auth service
   };
   ```
3. The Data Worker appends the token: `new WebTransport(`${url}?token=${token}`)`.

**Token refresh**: Implement a `refreshToken` postMessage protocol:

- When the token nears expiry (e.g., 60 seconds before the 5-minute TTL), the Data Worker sends `{ type: "token-expiring" }` to the main thread.
- The main thread fetches a new token and sends `{ type: "token-refresh", token: newJwt }` to the worker.
- The worker reconnects with the new token.

Read `pkg/webtransport/server.go` and the auth middleware to confirm the exact query parameter name (`token`) and TTL.

---

## Issue R-016 — WARNING: Wasm decoder module may not bundle in production build

**File**: `web-cop-gpu/src/workers/data-worker.ts`

**Problem**: The dynamic import:

```typescript
const mod = await import(
  /* @vite-ignore */ new URL(
    "../../wasm-decoder/pkg/wasm_decoder.js",
    import.meta.url,
  ).href
);
```

The `wasm-decoder/` directory is outside `src/`, which means Vite may not include it in `dist/` during `pnpm build`.

**Required Fix** (choose the simplest option for v4):

**Option A (recommended)**: Add a `vite-plugin-static-copy` (or equivalent) entry in `web-cop-gpu/vite.config.ts` to copy `wasm-decoder/pkg/` → `public/wasm/` at build time, then update the import path:

```typescript
new URL("/wasm/wasm_decoder.js", import.meta.url).href;
```

**Option B**: Add a pre-build script in `web-cop-gpu/package.json` (`"prebuild": "cp -r ../wasm-decoder/pkg public/wasm"`) and update the import path to `/wasm/wasm_decoder.js`.

Whichever option is chosen:

1. Update `web-cop-gpu/.gitignore` to exclude the copied Wasm output from the source tree (add `public/wasm/`).
2. Ensure the CI pipeline runs the Wasm build step before `pnpm build`.

**Verification**: Run `cd web-cop-gpu && pnpm build` and confirm that `dist/` contains the Wasm `.js` and `.wasm` files.

---

## Issue R-017 — WARNING: No WebTransport fallback to gRPC-Web streaming

**This issue requires architecture design discussion before implementation. Do NOT implement it in this batch.**

Document the gap as a TODO comment in `web-cop-gpu/src/workers/data-worker.ts` at the top of the `startWebTransport()` function (or equivalent connection initiator):

```typescript
// TODO(R-017): After N consecutive WebTransport failures, fall back to
// gRPC-Web server-streaming for the hot path (required for enterprise proxy
// environments). See v4 implementation review R-017 for specification.
```

This preserves the context without implementing unspecified behaviour.

---

## Implementation Rules

1. Read `web-cop-gpu/src/App.tsx`, `web-cop-gpu/src/workers/data-worker.ts`, and `pkg/webtransport/server.go` before editing.
2. **Never** hardcode JWT tokens, secrets, or URLs. All connection parameters come from environment variables or runtime signals.
3. Do not log JWT token values — this would be a security violation (`SDLC Rule 5: No PII in logs`).
4. The `DataInitMessage` TypeScript type must be updated to include the optional `token?: string` and `url?: string` fields — check the type definition file and update it.
5. All files must retain the `// CLASSIFICATION: UNCLASSIFIED` header.
6. Follow `docs/sdlc_guidelines/08_tech_specific/webtransport_guidelines.md` and `secure_coding.md`.

---

## Unit Tests Required

- Test that when `url` is `undefined` in the `DataInitMessage`, the Data Worker enters mock mode and does not attempt a `WebTransport` connection.
- Test that the WebTransport URL construction correctly appends the token: `url?token=<jwt>`.
- Mock the `WebTransport` global in tests to avoid network calls.
- Test that `pnpm build` produces `dist/assets/wasm_decoder*.wasm` (can be a shell assertion in `Makefile` or CI step).

---

## Verification Steps

1. `cd web-cop-gpu && pnpm test` — all unit tests must pass.
2. `cd web-cop-gpu && pnpm build` — confirm Wasm files appear in `dist/`.
3. With the backend running via `bash scripts/cop-dev/start-backend.sh` and `VITE_WEBTRANSPORT_URL` set: open the dev server → ConnectionIndicator must turn green → latency must appear in the StatusBar.
4. Without the env var set: app must start in mock mode with no errors in the console.

---

## PR Instructions

- PR title: `fix(web-cop-gpu): wire WebTransport URL and JWT auth into Data Worker, fix Wasm build (B5)`
- Label: `ai-orchestrator`
- After the PR is merged, move this file from `.github/instructions/todo/B5-WebTransport-Integration.md` to `.github/instructions/done/B5-WebTransport-Integration.md`.
