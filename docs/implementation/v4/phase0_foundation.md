<!-- CLASSIFICATION: UNCLASSIFIED -->

# Phase 0 — Foundation

> **Document**: v4 Implementation — Phase 0
> **Version**: 1.0
> **Classification**: UNCLASSIFIED
> **Status**: Not Started
> **Prerequisite Phases**: None (first phase)
> **Parallel With**: —
> **Architecture Reference**: `docs/architecture/v1/RTSA_WebGPU_Architecture_v1.md` §2, §3, §6

---

## 1. Objective

Establish the `web-cop-gpu/` project scaffold, build toolchain, browser capability gate, SharedArrayBuffer ring buffer, and Rust Wasm FlatBuffer decoder — everything needed before rendering or network code can begin.

---

## 2. Deliverables

| #    | Deliverable             | Description                                                        |
| ---- | ----------------------- | ------------------------------------------------------------------ |
| F0-1 | Project scaffold        | `web-cop-gpu/` with Vite, SolidJS, TypeScript, Vitest              |
| F0-2 | Build pipeline          | Vite config with `vite-plugin-solid`, Wasm loader, Worker bundling |
| F0-3 | Capability gate         | Browser feature detection → allow / degrade flow                   |
| F0-4 | SharedArrayBuffer setup | Ring buffer allocation, slot metadata, COOP/COEP headers           |
| F0-5 | Rust Wasm decoder       | FlatBuffer → SAB decode module + build pipeline                    |
| F0-6 | Data Worker shell       | Worker with WebTransport stub, SAB write loop                      |
| F0-7 | Render Worker shell     | Worker with OffscreenCanvas, mock render loop                      |
| F0-8 | Unit tests              | Wasm decoder tests, capability gate tests                          |

---

## 3. Detailed Tasks

### F0-1: Project Scaffold

Create `web-cop-gpu/` alongside existing `web-cop/`:

```
web-cop-gpu/
├── index.html
├── package.json
├── vite.config.ts
├── tsconfig.json
├── src/
│   ├── index.tsx
│   ├── App.tsx
│   ├── components/
│   ├── signals/
│   ├── services/
│   ├── workers/
│   ├── gpu/
│   ├── shaders/
│   └── types/
├── tests/
└── wasm-decoder/
```

**Dependencies** (see `docs/architecture/dependency_graph.md` §3):

- `solid-js`, `vite`, `vite-plugin-solid`, `typescript`
- `vitest`, `@solidjs/testing-library`, `playwright`
- `@connectrpc/connect-web`, `@connectrpc/protobuf-es` (cold path)

### F0-2: Build Pipeline

- Configure Vite for SolidJS JSX transform
- Configure Worker bundling (`new Worker(new URL(...), { type: "module" })`)
- Configure Wasm module import (`.wasm` files in `src/wasm/`)
- Dev server must set COOP/COEP headers (see F0-4)

### F0-3: Capability Gate

Implement `checkCapabilities()` per `docs/sdlc_guidelines/08_tech_specific/webgpu_guidelines.md` §3.2:

```typescript
// Required capabilities
const caps = await checkCapabilities();
if (
  !caps.webgpu ||
  !caps.sharedArrayBuffer ||
  !caps.offscreenCanvas ||
  !caps.webtransport
) {
  renderDegradedNotice(caps);
  return;
}
```

Display a static HTML notice listing missing capabilities if any check fails.

### F0-4: SharedArrayBuffer Setup

- Allocate SAB with `maxTracks × 128` bytes (default 65,536 slots = 8 MB)
- Create metadata header (write pointer, track count) as `Uint32Array` view
- Ensure dev server and production proxy set:
  - `Cross-Origin-Opener-Policy: same-origin`
  - `Cross-Origin-Embedder-Policy: require-corp`
- Reference: `docs/architecture/data_architecture.md` §12, `docs/architecture/security_architecture.md` §14

### F0-5: Rust Wasm Decoder

- Scaffold `web-cop-gpu/wasm-decoder/` Rust crate
- Implement `decode_track_update()` per `docs/sdlc_guidelines/08_tech_specific/flatbuffers_guidelines.md` §4
- Build with `wasm-pack build --target web --release`
- Output to `web-cop-gpu/src/wasm/decoder.wasm`
- Unit tests in `wasm-decoder/tests/`

### F0-6: Data Worker Shell

- Create `src/workers/data-worker.ts`
- Accept SAB via `postMessage` on init
- Stub WebTransport connection (write mock 128-byte records to SAB for testing)
- Report connection status back to main thread
- Reference: `docs/sdlc_guidelines/08_tech_specific/webtransport_guidelines.md` §6

### F0-7: Render Worker Shell

- Create `src/workers/render-worker.ts`
- Accept OffscreenCanvas + SAB via `postMessage`
- Basic `requestAnimationFrame`-equivalent loop (16ms `setInterval`)
- No WebGPU yet — just read SAB and log track count (proves data flow)

### F0-8: Unit Tests

| Test                      | Framework         | What it verifies                                |
| ------------------------- | ----------------- | ----------------------------------------------- |
| Wasm decoder round-trip   | Rust `cargo test` | Correct field values at correct offsets         |
| Wasm decoder bounds check | Rust `cargo test` | Returns false for out-of-bounds slot index      |
| Capability gate           | Vitest            | Returns correct booleans for mocked `navigator` |
| SAB allocation            | Vitest            | Correct byte length, header layout              |

---

## 4. Done Gate

| Criteria                                                     | Verification                    |
| ------------------------------------------------------------ | ------------------------------- |
| `web-cop-gpu/` builds with `pnpm build`                      | CI green                        |
| Capability gate renders degrade notice when features missing | Vitest test passes              |
| SharedArrayBuffer allocated and writable from Worker         | Integration test                |
| Wasm decoder compiles and passes all Rust tests              | `cargo test` green              |
| Data Worker writes mock records to SAB                       | Manual verification + unit test |
| Render Worker reads SAB and logs track count                 | Console log verified            |
| Dev server sets COOP/COEP headers                            | Browser DevTools check          |
| Code coverage ≥ 80% on new TypeScript code                   | Vitest coverage report          |

---

## 5. Risks & Mitigations

| Risk                                    | Mitigation                                                        |
| --------------------------------------- | ----------------------------------------------------------------- |
| `wasm-pack` not in CI                   | Add Rust toolchain to CI Docker image                             |
| COOP/COEP breaks existing assets        | Use `credentialless` for COEP if `require-corp` blocks CDN assets |
| SharedArrayBuffer unavailable in Safari | Capability gate degrades; Safari is out of initial scope          |
