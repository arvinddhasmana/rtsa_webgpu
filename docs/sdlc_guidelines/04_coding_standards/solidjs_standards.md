<!-- CLASSIFICATION: UNCLASSIFIED -->
# SolidJS Frontend Standards

> **Document**: RTSA SolidJS Coding Standards
> **Version**: 1.0
> **Classification**: UNCLASSIFIED
> **Last Updated**: 2026-03-05
> **Replaces**: `react_standards.md` (archived to `docs/archive/pre_v1_2026-03/`)
> **Prerequisite**: Load `general_coding.md` and `secure_coding.md` first

---

## 1. Technology Stack

| Technology | Version | Purpose |
|---|---|---|
| SolidJS | Latest stable | UI framework — fine-grained reactivity, JSX compilation |
| TypeScript | 5+ | Type safety |
| Vite | Latest stable | Build tool + dev server |
| `vite-plugin-solid` | Latest stable | SolidJS JSX transform |
| `@connectrpc/connect-web` | Latest stable | gRPC-Web client (cold path only) |
| `@connectrpc/protobuf-es` | Latest stable | Protobuf TypeScript runtime |
| Vitest | Latest stable | Unit testing |
| `@solidjs/testing-library` | Latest stable | Component testing |
| Playwright | Latest stable | E2E browser testing |

### Not In This Stack

The following are **not used** in the WebGPU COP codebase:
- React, react-dom, react-query
- Zustand, Redux, MobX (state management libraries)
- MapLibre GL, Leaflet, OpenLayers (map libraries)
- protobuf-ts (replaced by Rust Wasm decoder for hot path)

---

## 2. Project Structure

```
web-cop-gpu/
├── index.html                          # Entry point — canvas + SolidJS mount
├── vite.config.ts                      # Vite + SolidJS plugin config
├── tsconfig.json                       # TypeScript strict mode
├── package.json
├── public/
│   ├── icons/                          # Track icon atlas source images
│   └── fonts/                          # SDF font atlas files
├── src/
│   ├── index.tsx                       # SolidJS render() + capability gate
│   ├── App.tsx                         # Root component — canvas + overlay shell
│   ├── components/
│   │   ├── shell/                      # ClassificationBanner, AppShell
│   │   ├── toolbar/                    # RoleSelector, DashboardSelector, ConnectionIndicator
│   │   ├── panels/                     # TrackDetailPanel, AlertSidebar, FeedbackForm
│   │   ├── search/                     # SearchOverlay, QueryBuilder, ResultsView
│   │   ├── timeline/                   # EventTimeline
│   │   └── status/                     # StatusBar, FPSCounter, LatencyIndicator
│   ├── signals/
│   │   ├── track.ts                    # Track selection, detail signals
│   │   ├── alerts.ts                   # Alert list signal (from Data Worker)
│   │   ├── stats.ts                    # FPS, track count, latency signals
│   │   ├── connection.ts               # WebTransport + gRPC health signals
│   │   └── viewport.ts                 # Current map viewport (zoom, center)
│   ├── services/
│   │   ├── grpc-client.ts              # ConnectRPC transport (cold path)
│   │   ├── feedback.ts                 # Feedback gRPC calls
│   │   ├── alerts.ts                   # Alert gRPC calls
│   │   └── query.ts                    # Query gRPC calls (ClickHouse)
│   ├── workers/
│   │   ├── data-worker.ts              # WebTransport + Wasm decoder entry
│   │   ├── render-worker.ts            # WebGPU pipeline entry (OffscreenCanvas)
│   │   └── shared-protocol.ts          # postMessage types (Worker ↔ Main)
│   ├── wasm/
│   │   └── decoder.wasm                # Built Rust Wasm module (gitignored, build artifact)
│   ├── shaders/
│   │   ├── interpolation.wgsl          # Compute: position extrapolation
│   │   ├── culling.wgsl                # Compute: view-frustum culling
│   │   ├── track-icons.wgsl            # Vertex/fragment: instanced track quads
│   │   ├── trail.wgsl                  # Vertex/fragment: track trail lines
│   │   ├── halos.wgsl                  # Vertex/fragment: alert halo circles
│   │   ├── labels.wgsl                 # Vertex/fragment: SDF text
│   │   └── pick.wgsl                   # Fragment: pick buffer write
│   ├── gpu/
│   │   ├── device.ts                   # WebGPU adapter/device initialization
│   │   ├── buffers.ts                  # Buffer creation + management
│   │   ├── pipelines.ts                # Compute + render pipeline builders
│   │   ├── tile-renderer.ts            # Raster/vector tile rendering
│   │   └── atlas.ts                    # Icon + SDF font atlas loading
│   └── types/
│       ├── track.ts                    # Track domain types
│       ├── alert.ts                    # Alert domain types
│       └── protocol.ts                 # Worker message protocol types
├── tests/
│   ├── components/                     # SolidJS component tests
│   ├── services/                       # gRPC client tests
│   ├── workers/                        # Worker logic tests
│   └── e2e/                            # Playwright E2E tests
└── wasm-decoder/                       # Rust source for Wasm FlatBuffer decoder
    ├── Cargo.toml
    ├── src/
    │   └── lib.rs                      # Decoder entry point
    └── tests/
```

---

## 3. SolidJS Reactive Primitives

### 3.1 Core Rules

1. **Use `createSignal` for simple state** — never `useState` (React pattern does not exist)
2. **Use `createMemo` for derived values** — recalculates only when dependencies change
3. **Use `createEffect` for side effects** — runs when tracked signals change
4. **Use `createResource` for async data** — built-in loading/error states for gRPC calls
5. **Never destructure props** — destructuring breaks reactivity tracking

```typescript
// CLASSIFICATION: UNCLASSIFIED

// ✅ Correct — props accessed directly
function TrackCount(props: { count: number }) {
  return <span>{props.count} tracks</span>;
}

// ❌ Wrong — destructuring breaks reactivity
function TrackCount({ count }: { count: number }) {
  return <span>{count} tracks</span>;
}
```

### 3.2 Signal Naming

| Pattern | Convention | Example |
|---|---|---|
| Signal accessor | `camelCase` | `selectedTrack()` |
| Signal setter | `set` + PascalCase | `setSelectedTrack(track)` |
| Memo | `camelCase` | `visibleAlerts()` |
| Resource | `camelCase` | `timelineEvents()` |

### 3.3 State Management

SolidJS signals replace Zustand stores. There are **no global stores** — signals are defined in dedicated files under `src/signals/` and imported where needed.

```typescript
// CLASSIFICATION: UNCLASSIFIED
// src/signals/track.ts

import { createSignal } from "solid-js";
import type { TrackDetail } from "../types/track";

// Selected track — updated when operator clicks on WebGPU canvas
const [selectedTrack, setSelectedTrack] = createSignal<TrackDetail | null>(null);

export { selectedTrack, setSelectedTrack };
```

---

## 4. Component Patterns

### 4.1 Component Naming

- File name: `PascalCase.tsx` (e.g., `AlertSidebar.tsx`)
- Component function: Named export, PascalCase
- One component per file (small helper components may be co-located)

### 4.2 Component Structure

```typescript
// CLASSIFICATION: UNCLASSIFIED

import { createSignal, createEffect, Show, For } from "solid-js";
import type { Alert } from "../../types/alert";

interface AlertSidebarProps {
  alerts: () => Alert[];   // Signal accessor — NOT a raw value
  onAcknowledge: (id: string) => void;
}

export function AlertSidebar(props: AlertSidebarProps) {
  const [filter, setFilter] = createSignal<string>("all");

  const filtered = () =>
    props.alerts().filter((a) =>
      filter() === "all" ? true : a.severity === filter()
    );

  return (
    <aside class="alert-sidebar">
      <select onChange={(e) => setFilter(e.currentTarget.value)}>
        <option value="all">All</option>
        <option value="CRITICAL">Critical</option>
        <option value="ELEVATED">Elevated</option>
      </select>
      <For each={filtered()}>
        {(alert) => (
          <div class={`alert-item severity-${alert.severity}`}>
            <span>{alert.description}</span>
            <button onClick={() => props.onAcknowledge(alert.id)}>Ack</button>
          </div>
        )}
      </For>
    </aside>
  );
}
```

### 4.3 Control Flow Components

Use SolidJS built-in control flow — never ternaries for conditional rendering:

| Pattern | SolidJS | Notes |
|---|---|---|
| Conditional render | `<Show when={condition()} fallback={...}>` | Lazy — children not evaluated when false |
| List render | `<For each={items()}>` | Keyed by index, efficient updates |
| Index render | `<Index each={items()}>` | Keyed by position, for non-keyed lists |
| Error boundary | `<ErrorBoundary fallback={...}>` | Catches component errors |
| Suspense | `<Suspense fallback={...}>` | For async resources |
| Switch/Match | `<Switch><Match when={...}>` | Multi-branch conditional |
| Dynamic | `<Dynamic component={...}>` | Dynamic component selection |

---

## 5. Worker Communication

### 5.1 Message Protocol

All Worker ↔ Main Thread messages use typed `postMessage`:

```typescript
// CLASSIFICATION: UNCLASSIFIED
// src/workers/shared-protocol.ts

export type MainToRenderMessage =
  | { type: "viewport_change"; zoom: number; centerLat: number; centerLon: number }
  | { type: "select_track"; slotIndex: number }
  | { type: "init"; canvas: OffscreenCanvas; sab: SharedArrayBuffer };

export type RenderToMainMessage =
  | { type: "stats"; fps: number; trackCount: number; visibleCount: number }
  | { type: "picked"; trackId: string; slotIndex: number }
  | { type: "ready" };

export type DataWorkerToMainMessage =
  | { type: "alerts_updated"; alerts: Alert[] }
  | { type: "connection_status"; connected: boolean; latency: number };
```

### 5.2 Worker Bridge Pattern

```typescript
// CLASSIFICATION: UNCLASSIFIED
// Connecting Worker messages to SolidJS signals

import { createSignal, onCleanup } from "solid-js";
import type { RenderToMainMessage } from "../workers/shared-protocol";

export function useRenderWorker(worker: Worker) {
  const [fps, setFps] = createSignal(0);
  const [trackCount, setTrackCount] = createSignal(0);

  const handler = (e: MessageEvent<RenderToMainMessage>) => {
    if (e.data.type === "stats") {
      setFps(e.data.fps);
      setTrackCount(e.data.trackCount);
    }
  };

  worker.addEventListener("message", handler);
  onCleanup(() => worker.removeEventListener("message", handler));

  return { fps, trackCount };
}
```

---

## 6. gRPC-Web Integration (Cold Path)

All operator commands (feedback, alert ack, queries) use gRPC-Web via ConnectRPC:

```typescript
// CLASSIFICATION: UNCLASSIFIED
// src/services/grpc-client.ts

import { createConnectTransport } from "@connectrpc/connect-web";

export const transport = createConnectTransport({
  baseUrl: import.meta.env.VITE_API_GATEWAY_URL,
});
```

```typescript
// CLASSIFICATION: UNCLASSIFIED
// src/services/feedback.ts

import { createPromiseClient } from "@connectrpc/connect";
import { FeedbackService } from "../../gen/ts/rtsa/feedback/v1/feedback_connect";
import { transport } from "./grpc-client";

const client = createPromiseClient(FeedbackService, transport);

export async function submitFeedback(trackId: string, classification: string, justification: string) {
  return client.submitOperatorFeedback({
    trackId,
    classification,
    justification,
  });
}
```

---

## 7. Security Rules

### 7.1 XSS Prevention

SolidJS auto-escapes text content in JSX (like React). The dangerous equivalent is:

```typescript
// ❌ NEVER use innerHTML with untrusted data
<div innerHTML={untrustedHTML} />

// ✅ Use text content — auto-escaped
<div>{userProvidedText}</div>
```

**Rule**: `innerHTML` directive usage requires security review approval.

### 7.2 Classification Enforcement

- Classification banners are static, sourced from deployment config (`VITE_CLASSIFICATION_LEVEL`)
- No classification data is stored in browser signals or localStorage
- The WebGPU hot path drops records above operator clearance before writing to SharedArrayBuffer

### 7.3 Content Security Policy

```
Content-Security-Policy:
  default-src 'self';
  script-src 'self' 'wasm-unsafe-eval';
  worker-src 'self';
  connect-src 'self' https://*.rtsa.mil.ca;
  style-src 'self' 'unsafe-inline';
  img-src 'self' data: blob:;
```

Note: `'wasm-unsafe-eval'` is required for WebAssembly module instantiation.

---

## 8. Testing

### 8.1 Component Testing

```typescript
// CLASSIFICATION: UNCLASSIFIED

import { describe, it, expect } from "vitest";
import { render, screen } from "@solidjs/testing-library";
import { AlertSidebar } from "./AlertSidebar";

describe("AlertSidebar", () => {
  it("renders alert list", () => {
    const alerts = () => [
      { id: "1", severity: "CRITICAL", description: "Hostile track detected" },
    ];

    render(() => (
      <AlertSidebar alerts={alerts} onAcknowledge={() => {}} />
    ));

    expect(screen.getByText("Hostile track detected")).toBeDefined();
  });
});
```

### 8.2 Test Coverage Target

- **Components**: 80%+ line coverage via `@solidjs/testing-library` + Vitest
- **Services** (gRPC clients): Mock transport, verify request/response contracts
- **Workers**: Test message handling logic in isolation (no real WebGPU/WebTransport)
- **E2E**: Playwright tests against running dev stack (WebGPU requires real browser)

---

## 9. File Naming Conventions

| File Type | Convention | Example |
|---|---|---|
| Component | `PascalCase.tsx` | `AlertSidebar.tsx` |
| Signal module | `camelCase.ts` | `track.ts` |
| Service | `camelCase.ts` | `feedback.ts` |
| Worker entry | `kebab-case.ts` | `data-worker.ts` |
| WGSL shader | `kebab-case.wgsl` | `track-icons.wgsl` |
| Type definitions | `camelCase.ts` | `track.ts` |
| Test file | `*.test.ts` or `*.test.tsx` | `AlertSidebar.test.tsx` |

---

## 10. Cross-References

| Document | Path |
|---|---|
| General Coding Standards | `docs/sdlc_guidelines/04_coding_standards/general_coding.md` |
| Secure Coding | `docs/sdlc_guidelines/04_coding_standards/secure_coding.md` |
| WebGPU Guidelines | `docs/sdlc_guidelines/08_tech_specific/webgpu_guidelines.md` |
| WGSL Shader Standards | `docs/sdlc_guidelines/08_tech_specific/wgsl_shader_standards.md` |
| v1 Architecture — SolidJS Section | `docs/architecture/v1/RTSA_WebGPU_Architecture_v1.md` §7 |
| Component Design | `docs/architecture/component_design.md` |
