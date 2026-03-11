// CLASSIFICATION: UNCLASSIFIED
// src/App.tsx — Root application component (Phase 3: UI & Interaction)
//
// Wires together:
//   - Capability gate (Phase 0)
//   - Render Worker (Phase 1) + Data Worker (Phase 2)
//   - SolidJS overlay UI (Phase 3)
//
// Reference: docs/implementation/v4/phase3_ui_interaction.md §4 Signal Architecture

import { createSignal, onCleanup, onMount, Show } from "solid-js";
import { renderDegradedNotice } from "./components/DegradedNotice";
import { checkCapabilities, type Capabilities } from "./services/capabilities";
import { allocateSAB } from "./services/sab";

// Signals
import { updateAlerts } from "./signals/alerts";
import { operatorIdFromToken, setOperatorId } from "./signals/auth";
import { setConnecting, setWtConnected } from "./signals/connection";
import {
    setDatagramsPerSec,
    setDecodeErrors,
    setFps,
    setLatencyMs,
    setRecordsPerSec,
    setTrackCount,
    setVisibleCount,
} from "./signals/stats";
import {
    setSelectedTrack,
    setTrackDetail,
    setTrackDetailError,
    setTrackDetailLoading,
} from "./signals/track";
import { dashboard, role } from "./signals/viewport";

// Services
import { startAlertStream } from "./services/alerts";
import { fetchAuthToken } from "./services/auth";
import { fetchTrackDetail } from "./services/query";

// Components
import { AlertSidebar } from "./components/panels/AlertSidebar";
import { FeedbackForm } from "./components/panels/FeedbackForm";
import { TrackDetailPanel } from "./components/panels/TrackDetailPanel";
import { SearchOverlay } from "./components/search/SearchOverlay";
import { AppShell } from "./components/shell/AppShell";
import { StatusBar } from "./components/status/StatusBar";
import { EventTimeline } from "./components/timeline/EventTimeline";
import { ConnectionIndicator } from "./components/toolbar/ConnectionIndicator";
import { DashboardSelector } from "./components/toolbar/DashboardSelector";
import { RoleSelector } from "./components/toolbar/RoleSelector";

// Worker message types
import type {
    DataInitMessage,
    DataToMainMessage,
    RenderInitMessage,
    RenderToMainMessage,
    TokenRefreshMessage,
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
        // Use measured FPS from the Render Worker's FrameTimer when available.
        if (msg.fps > 0) setFps(msg.fps);
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

      case "token-expiring":
        // Fetch a refreshed JWT and forward it to the Data Worker.
        // Do not log the token value. (SDLC Rule 5)
        fetchAuthToken().then((newToken) => {
          if (newToken && dataWorker) {
            const refreshMsg: TokenRefreshMessage = { type: "token-refresh", token: newToken };
            dataWorker.postMessage(refreshMsg);
          }
        }).catch(() => {
          // Auth refresh failed — worker will continue with the existing token
          // until it expires and reconnection fails naturally.
        });
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
      const w = window as any;
      const isTestContext = w.__RTSA_TEST_TRACK_COUNT__ !== undefined;
      const initMsg: RenderInitMessage = {
        type: "init",
        canvas: offscreen,
        sab,
        dataWorkerActive: !isTestContext,
        testTrackCount: w.__RTSA_TEST_TRACK_COUNT__,
        testCameraScale: w.__RTSA_TEST_SCALE__,
      };
      // Transfer OffscreenCanvas only — SAB is shared, not transferred
      renderWorker.postMessage(initMsg, [offscreen]);
    }

    // Init Data Worker — pass URL and JWT token when available.
    // When VITE_WEBTRANSPORT_URL is undefined (local dev), the worker falls back to mock mode.
    const wtUrl = import.meta.env.VITE_WEBTRANSPORT_URL as string | undefined;
    const token = wtUrl ? await fetchAuthToken() : undefined;
    // Decode operator identity from JWT claims; falls back to "anonymous".
    setOperatorId(operatorIdFromToken(token));
    const dataInit: DataInitMessage = { type: "init", sab, url: wtUrl, token };
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

    // Start gRPC alert stream or inject mocks for E2E visual tests
    const w = window as any;
    if (w.__RTSA_TEST_TRACK_COUNT__ !== undefined) {
      updateAlerts([
        { alertId: "alert-1", trackId: "1", severity: "CRITICAL", description: "Mock Hostile Incursion", detectedAtMs: Date.now() - 5000, acknowledged: false },
        { alertId: "alert-2", trackId: "2", severity: "ELEVATED", description: "Mock Speed Anomaly", detectedAtMs: Date.now() - 15000, acknowledged: false },
        { alertId: "alert-3", trackId: "3", severity: "WATCH", description: "Mock Route Deviation", detectedAtMs: Date.now() - 45000, acknowledged: false },
      ]);
    } else {
      alertStreamController = startAlertStream();
    }
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
            <div style={{ position: "relative", width: "100%", height: "100%", "background-color": "#0a0f1a" }}>
              {/* Temporary raster map background for Phase 3 E2E test visualization */}
              <img
                src="https://a.tile.openstreetmap.org/0/0/0.png"
                style={{
                  position: "absolute",
                  top: "50%",
                  left: "50%",
                  transform: "translate(-50%, -50%) scale(4)",
                  opacity: 0.15,
                  "pointer-events": "none",
                  filter: "invert(1) hue-rotate(180deg)", /* dark mode styling */
                }}
                alt="map-background"
              />
              <canvas
                ref={canvasRef}
                id="gpu-canvas"
                onClick={handleCanvasClick}
                style={{
                  position: "absolute",
                  top: 0,
                  left: 0,
                  width: "100%",
                  height: "100%",
                  display: "block",
                  cursor: "crosshair",
                }}
              />
            </div>
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
