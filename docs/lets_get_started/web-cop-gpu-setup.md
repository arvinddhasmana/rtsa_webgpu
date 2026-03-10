<!-- CLASSIFICATION: UNCLASSIFIED -->

# web-cop-gpu — Developer Onboarding Guide

> **Document**: Phase 4 Developer Onboarding — WebGPU COP
> **Version**: 1.0
> **Classification**: UNCLASSIFIED
> **Prerequisite**: Complete `getting_started.md` first to set up the full development stack.

---

## 1. Overview

`web-cop-gpu` is the production WebGPU Common Operating Picture (COP) frontend. It uses:

- **SolidJS** — reactive UI framework (no virtual DOM)
- **WebGPU** — GPU-accelerated track rendering (50k tracks @ 60 FPS)
- **WebTransport** — QUIC-based binary datagram delivery (hot path)
- **SharedArrayBuffer** — zero-copy track data between workers
- **Rust/Wasm** — `wasm-decoder` for FlatBuffer record decoding in the Data Worker
- **gRPC-Web (ConnectRPC)** — cold path for queries, feedback, alerts

---

## 2. Prerequisites

| Tool        | Version | Purpose                  |
| ----------- | ------- | ------------------------ |
| Node.js     | 20 LTS+ | Build toolchain          |
| pnpm        | 9+      | Package manager          |
| Rust        | 1.77+   | Wasm decoder compilation |
| wasm-pack   | 0.12+   | Rust → Wasm build tool   |
| Chrome/Edge | 117+    | WebGPU-capable browser   |

> **Note**: Firefox does not yet support WebGPU. Use Chrome or Edge for development.

---

## 3. Setup Steps

### 3.1 Install Node dependencies

```bash
cd web-cop-gpu
pnpm install
```

### 3.2 Build the Wasm Decoder (Rust)

```bash
cd web-cop-gpu/wasm-decoder
wasm-pack build --target web --out-dir ../src/wasm-decoder-pkg
```

### 3.3 Start the Dev Server

```bash
cd web-cop-gpu
pnpm dev
```

The dev server starts at **http://localhost:5174**. It sets the required security headers:

- `Cross-Origin-Opener-Policy: same-origin`
- `Cross-Origin-Embedder-Policy: require-corp`

> Both headers are mandatory for `SharedArrayBuffer` to be available. Without them, the app will fall back to degraded mode.

### 3.4 Open in Chrome with WebGPU enabled

Navigate to **http://localhost:5174** in Chrome or Edge. WebGPU is enabled by default in recent versions.

If WebGPU is not available, you will see the **Degraded Mode** notice listing the missing capabilities.

---

## 4. Project Structure

```
web-cop-gpu/
├── src/
│   ├── App.tsx                    # Root component (wires workers + UI)
│   ├── index.tsx                  # Entry point
│   ├── gpu/
│   │   ├── buffers.ts             # Pre-allocated GPU buffer management
│   │   ├── renderer.ts            # Per-frame render orchestration (10-step pipeline)
│   │   ├── pipelines.ts           # WebGPU render + compute pipelines
│   │   ├── bind-groups.ts         # GPUBindGroup allocation
│   │   ├── uniforms.ts            # Uniform buffer write + view-projection
│   │   ├── pick.ts                # Pick buffer (track click detection)
│   │   ├── atlas.ts               # Texture atlas (icons, SDF font)
│   │   ├── map-tiles.ts           # Background map tile rendering
│   │   ├── mock-data.ts           # Mock SAB data for offline development
│   │   ├── frame-timer.ts         # GPU timestamp query + JS frame timing (Phase 4)
│   │   └── lod.ts                 # Level-of-Detail system (Phase 4)
│   ├── shaders/                   # WGSL shaders
│   ├── workers/
│   │   ├── render-worker.ts       # OffscreenCanvas render loop (GPU thread)
│   │   ├── data-worker.ts         # WebTransport + Wasm decode (data thread)
│   │   └── shared-protocol.ts     # Typed postMessage protocol
│   ├── services/
│   │   ├── capabilities.ts        # Browser capability detection
│   │   ├── sab.ts                 # SharedArrayBuffer layout
│   │   ├── grpc-client.ts         # ConnectRPC client setup
│   │   ├── query.ts               # Track query service
│   │   ├── alerts.ts              # Alert stream service
│   │   └── feedback.ts            # Feedback service
│   ├── signals/                   # SolidJS global signals
│   ├── components/                # SolidJS overlay UI components
│   └── types/                     # TypeScript type declarations
├── tests/                         # Vitest unit tests
├── e2e/                           # Playwright E2E tests (Phase 4)
├── wasm-decoder/                  # Rust Wasm decoder source
├── vite.config.ts
├── vitest.config.ts
└── playwright.config.ts
```

---

## 5. Running Tests

### Unit Tests (Vitest)

```bash
cd web-cop-gpu
pnpm test
```

### Unit Tests with Coverage

```bash
pnpm test:coverage
```

Target: **≥ 80% line coverage** per file (SDLC requirement).

### E2E Tests (Playwright)

First, start the dev server in a separate terminal:

```bash
pnpm dev
```

Then run the E2E suite:

```bash
pnpm test:e2e
```

Run with a browser UI (headed mode):

```bash
pnpm test:e2e:headed
```

### Visual Regression Tests

Capture golden screenshots (first run):

```bash
pnpm test:e2e:update-snapshots
```

Compare against golden screenshots (CI):

```bash
pnpm test:e2e:visual
```

> Golden screenshots are stored in `e2e/snapshots/`. They must be committed to the repository.

---

## 6. Key Architectural Decisions

### SharedArrayBuffer Layout

The SAB is divided into three regions:

| Region         | Offset | Size     | Purpose                         |
| -------------- | ------ | -------- | ------------------------------- |
| Header         | 0      | 4 KB     | Atomic counters (active_count)  |
| Dirty Bitfield | 4096   | 8 KB     | Bit per slot for changed tracks |
| Track Data     | 12288  | variable | 128-byte records, up to 65536   |

### Zero Per-Frame Allocation Rule

All GPU buffers are allocated **once at startup** to maximum capacity (`MAX_TRACKS = 65,536`). No `createBuffer()` calls occur during the render loop. This is enforced by the webgpu_guidelines.md §4.4.

### LOD System

The LOD system (`src/gpu/lod.ts`) automatically reduces rendering complexity at low zoom levels:

- **Full** (`scale ≥ 0.5`): all effects, all tracks
- **Medium** (`scale ≥ 0.1`): no trails, labels disabled above 10k tracks, max 20k instances
- **Minimal** (`scale < 0.1`): icons only, max 10k instances

### Frame Timer

The `FrameTimer` (`src/gpu/frame-timer.ts`) wraps GPU timestamp queries to measure per-pass timing. It provides:

- Smoothed 60-frame rolling averages
- JS main-thread wall-clock measurement
- Graceful fallback when `timestamp-query` is unsupported

---

## 7. Security Requirements

All developers must follow these rules:

1. **No hardcoded tokens/secrets** — use environment variables only.
2. **COOP/COEP headers must stay enabled** — required for SharedArrayBuffer.
3. **CSP must not be weakened** — `'unsafe-inline'` scripts are prohibited.
4. **No `innerHTML` with user-controlled data** — use DOM APIs or SolidJS JSX.
5. **Never log raw sensor payloads or track data** at INFO level or above.
6. **Never destructure SolidJS props** — breaks signal reactivity.

---

## 8. Performance Targets

| Metric            | Target       | Measurement            |
| ----------------- | ------------ | ---------------------- |
| Total frame time  | ≤ 8 ms       | FrameTimer.smoothed    |
| SAB read + upload | ≤ 2 ms       | FrameTimer.sabUploadMs |
| Compute passes    | ≤ 1 ms       | FrameTimer.computeMs   |
| Render passes     | ≤ 4 ms       | FrameTimer.renderMs    |
| Main thread CPU   | < 20%        | Chrome DevTools        |
| Track throughput  | 50k @ 60 FPS | StatusBar FPS display  |

---

## 9. Troubleshooting

| Symptom                          | Cause                                  | Fix                                                    |
| -------------------------------- | -------------------------------------- | ------------------------------------------------------ |
| Degraded notice on startup       | WebGPU or SAB unavailable              | Use Chrome 117+ with COOP/COEP headers                 |
| SharedArrayBuffer unavailable    | Missing COOP/COEP headers              | Run via `pnpm dev` (not file://)                       |
| Wasm decoder fails to load       | wasm-pack build not run                | Run `wasm-pack build` in `wasm-decoder/`               |
| WebTransport connection fails    | Backend not running or TLS error       | Start `pkg/webtransport` server with dev TLS certs     |
| GPU buffer size errors           | Track count exceeds MAX_TRACKS         | Track count is capped at 65,536 (SAB layout)           |
| Visual regression failures in CI | Rendering differences between machines | Update snapshots with `pnpm test:e2e:update-snapshots` |
