// CLASSIFICATION: UNCLASSIFIED
// src/App.tsx — Root application component (Phase 3: UI & Interaction)
//
// Wires together:
//   - Capability gate (Phase 0)
//   - Render Worker (Phase 1) + Data Worker (Phase 2)
//   - SolidJS overlay UI (Phase 3)
//
// Reference: docs/implementation/v4/phase3_ui_interaction.md §4 Signal Architecture

import { createSignal, onMount, onCleanup, Show } from "solid-js";
import { checkCapabilities, type Capabilities } from "./services/capabilities";
import { renderDegradedNotice } from "./components/DegradedNotice";
import { allocateSAB } from "./services/sab";

// Signals
import {
  setSelectedTrack,
  setTrackDetail,
  setTrackDetailLoading,
  setTrackDetailError,
} from "./signals/track";
import {
  setFps,
  setTrackCount,
  setVisibleCount,
  setLatencyMs,
  setDatagramsPerSec,
  setRecordsPerSec,
  setDecodeErrors,
} from "./signals/stats";
import { setWtConnected, setConnecting } from "./signals/connection";
import { updateAlerts } from "./signals/alerts";
import { role, dashboard } from "./signals/viewport";

// Services
import { fetchTrackDetail } from "./services/query";
import { startAlertStream } from "./services/alerts";

// Components
import { AppShell } from "./components/shell/AppShell";
import { RoleSelector } from "./components/toolbar/RoleSelector";
import { DashboardSelector } from "./components/toolbar/DashboardSelector";
import { ConnectionIndicator } from "./components/toolbar/ConnectionIndicator";
import { TrackDetailPanel } from "./components/panels/TrackDetailPanel";
import { AlertSidebar } from "./components/panels/AlertSidebar";
import { FeedbackForm } from "./components/panels/FeedbackForm";
import { SearchOverlay } from "./components/search/SearchOverlay";
import { EventTimeline } from "./components/timeline/EventTimeline";
import { StatusBar } from "./components/status/StatusBar";

// Worker message types
import type {
  RenderToMainMessage,
  DataToMainMessage,
  RenderInitMessage,
  DataInitMessage,
} from "./workers/shared-protocol";

// Fps tracking
let frameCount = 0;
let lastFpsTime = performance.now();

export default function App() {
  const [caps, setCaps] = createSignal<Capabilities | null>(null);

  let canvasRef: HTMLCanvasElement | undefined;
  let renderWorker: Worker | null = null;
  let dataWorker: Worker | null = null;
  let alertStreamController: AbortController | null = null;
  let fpsIntervalId: ReturnType<typeof setInterval> | null = null;

  // ── Render Worker message handler ──────────────────────────────────────────

  function handleRenderMessage(event: MessageEvent<RenderToMainMessage>) {
    const msg = event.data;
    switch (msg.type) {
      case "status":
        if (msg.ready) {
          setConnecting(false);
        }
        break;

      case "picked": {
        setSelectedTrack({ trackIdHash: msg.trackIdHash, x: msg.x, y: msg.y });
        // Attempt to fetch full track detail via gRPC using hash as track ID prefix
        const hashHex = msg.trackIdHash.toString(16).padStart(8, "0");
        setTrackDetailLoading(true);
        setTrackDetailError(null);
        setTrackDetail(null);
        fetchTrackDetail(hashHex)
          .then((detail) => {
            setTrackDetail(detail);
            setTrackDetailLoading(false);
          })
          .catch((err: unknown) => {
            setTrackDetailError(err instanceof Error ? err.message : "Fetch failed");
            setTrackDetailLoading(false);
          });
        break;
      }

      case "stats":
        // The render worker currently doesn't send stats yet; handled via local FPS counter
        setTrackCount(msg.trackCount);
        setVisibleCount(msg.visibleCount);
        break;

      default:
        break;
    }
  }

  // ── Data Worker message handler ────────────────────────────────────────────

  function handleDataMessage(event: MessageEvent<DataToMainMessage>) {
    const msg = event.data;
    switch (msg.type) {
      case "connection_status":
        setWtConnected(msg.connected);
        setLatencyMs(msg.latency);
        if (msg.connected) setConnecting(false);
        break;

      case "stats":
        setDatagramsPerSec(msg.datagramsReceived);
        setRecordsPerSec(msg.recordsDecoded);
        setDecodeErrors(msg.decodeErrors);
        break;

      case "alerts_updated":
        updateAlerts(msg.alerts);
        break;

      default:
        break;
    }
  }

  // ── Canvas click → pick buffer ─────────────────────────────────────────────

  function handleCanvasClick(e: MouseEvent) {
    if (!canvasRef || !renderWorker) return;
    const rect = canvasRef.getBoundingClientRect();
    const x = Math.round((e.clientX - rect.left) * devicePixelRatio);
    const y = Math.round((e.clientY - rect.top) * devicePixelRatio);
    renderWorker.postMessage({ type: "select_track", x, y });
  }

  // ── Resize observer ────────────────────────────────────────────────────────

  let resizeObserver: ResizeObserver | null = null;

  function setupResizeObserver(canvas: HTMLCanvasElement) {
    resizeObserver = new ResizeObserver((entries) => {
      const entry = entries[0];
      if (!entry || !renderWorker) return;
      const { width, height } = entry.contentRect;
      const dpr = devicePixelRatio;
      renderWorker.postMessage({
        type: "resize",
        width: Math.round(width * dpr),
        height: Math.round(height * dpr),
      });
    });
    resizeObserver.observe(canvas);
  }

  // ── Initialise workers ─────────────────────────────────────────────────────

  async function init(detected: Capabilities) {
    if (
      !detected.webgpu ||
      !detected.sharedArrayBuffer ||
      !detected.offscreenCanvas ||
      !detected.webtransport
    ) {
      return;
    }

    const ringBuffer = allocateSAB();
    const sab = ringBuffer.sab;

    // Spawn workers
    renderWorker = new Worker(
      new URL("./workers/render-worker.ts", import.meta.url),
      { type: "module" },
    );
    dataWorker = new Worker(
      new URL("./workers/data-worker.ts", import.meta.url),
      { type: "module" },
    );

    renderWorker.addEventListener("message", handleRenderMessage);
    dataWorker.addEventListener("message", handleDataMessage);

    // Transfer OffscreenCanvas to Render Worker
    if (canvasRef) {
      const offscreen = canvasRef.transferControlToOffscreen();
      const initMsg: RenderInitMessage = {
        type: "init",
        canvas: offscreen,
        sab,
      };
      // Transfer OffscreenCanvas only — SAB is shared, not transferred
      renderWorker.postMessage(initMsg, [offscreen]);
    }

    // Init Data Worker (stub mode — no URL for Phase 2 mock)
    const dataInit: DataInitMessage = { type: "init", sab };
    dataWorker.postMessage(dataInit);

    // Local FPS counter from requestAnimationFrame
    function rafLoop() {
      frameCount++;
      const now = performance.now();
      if (now - lastFpsTime >= 1000) {
        setFps(Math.round(frameCount * 1000 / (now - lastFpsTime)));
        frameCount = 0;
        lastFpsTime = now;
      }
      requestAnimationFrame(rafLoop);
    }
    requestAnimationFrame(rafLoop);

    // Setup resize observer
    if (canvasRef) {
      setupResizeObserver(canvasRef);
    }

    // Start gRPC alert stream
    alertStreamController = startAlertStream();
  }

  onMount(async () => {
    const detected = await checkCapabilities();
    setCaps(detected);
    await init(detected);
  });

  onCleanup(() => {
    renderWorker?.terminate();
    dataWorker?.terminate();
    alertStreamController?.abort();
    resizeObserver?.disconnect();
    if (fpsIntervalId !== null) clearInterval(fpsIntervalId);
  });

  // ── Dashboard layout helpers ───────────────────────────────────────────────

  const showAlerts = () =>
    role() === "sensor_operator" || role() === "operations_commander";

  const showTimeline = () =>
    role() === "operations_commander" || dashboard() === "analytics";

  return (
    <Show
      when={caps() !== null}
      fallback={<div style={{ padding: "2rem" }}>Initialising…</div>}
    >
      <Show
        when={
          caps()!.webgpu &&
          caps()!.sharedArrayBuffer &&
          caps()!.offscreenCanvas &&
          caps()!.webtransport
        }
        fallback={renderDegradedNotice(caps()!)}
      >
        <AppShell
          toolbar={
            <>
              <RoleSelector />
              <DashboardSelector />
              <ConnectionIndicator />
            </>
          }
          canvas={
            <canvas
              ref={canvasRef}
              id="gpu-canvas"
              onClick={handleCanvasClick}
              style={{
                width: "100%",
                height: "100%",
                display: "block",
                cursor: "crosshair",
              }}
            />
          }
          rightPanel={
            <>
              <TrackDetailPanel />
              <Show when={showAlerts()}>
                <AlertSidebar />
              </Show>
            </>
          }
          bottomPanel={
            <>
              <StatusBar />
              <Show when={showTimeline()}>
                <EventTimeline />
              </Show>
            </>
          }
          overlay={
            <>
              <FeedbackForm />
              <SearchOverlay />
            </>
          }
        />
      </Show>
    </Show>
  );
}
