// CLASSIFICATION: UNCLASSIFIED
// src/App.tsx — Root application component (Phase 3: UI & Interaction)
//
// Wires together:
//   - Capability gate (Phase 0)
//   - Render Worker (Phase 1) + Data Worker (Phase 2)
//   - SolidJS overlay UI (Phase 3)
//
// Reference: docs/implementation/v4/phase3_ui_interaction.md §4 Signal Architecture

import { createEffect, createSignal, onCleanup, onMount, Show } from "solid-js";
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
import { dashboard, enforceRoleDashboardGuard, role } from "./signals/viewport";

// Spatial alert signals (Level 3 navigation)
import { setSpatialAlerts } from "./signals/spatial-alerts";

// Services
import { startAlertStream } from "./services/alerts";
import { fetchAuthToken } from "./services/auth";
import { fetchTrackDetail } from "./services/query";
import { fetchSensorStatuses } from "./services/sensor-health";

// Components
import { CoverageMapDashboard } from "./components/dashboard/CoverageMapDashboard";
import { FusionCommanderDashboard } from "./components/dashboard/FusionCommanderDashboard";
import { MultiDomainCommanderDashboard } from "./components/dashboard/MultiDomainCommanderDashboard";
import { OperatorUiCommanderDashboard } from "./components/dashboard/OperatorUiCommanderDashboard";
import { SensorHealthDashboard } from "./components/dashboard/SensorHealthDashboard";
import AlertSidebar from "./components/panels/AlertSidebar";
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
        setSelectedTrack({
          trackIdHash: msg.trackIdHash,
          x: msg.x,
          y: msg.y,
          source: "canvas",
        });
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
            setTrackDetailError(
              err instanceof Error ? err.message : "Fetch failed",
            );
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

  // ── UTC Clock for Header ────────────────────────────────────────────────
  const [headerTime, setHeaderTime] = createSignal(
    new Date().toISOString().slice(11, 19),
  );
  onMount(() => {
    const timer = setInterval(() => {
      setHeaderTime(new Date().toISOString().slice(11, 19));
    }, 1000);
    onCleanup(() => clearInterval(timer));
  });

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
        // Route coverage-gap alerts (description contains "gap") to the spatial alerts signal
        // so Level 3 navigation can highlight the affected sector.
        // Phase A: AlertPayload.trackId repurposed as the affected sensor ID for gap alerts
        // since gap alerts originate from sensors, not tracks.
        setSpatialAlerts(
          msg.alerts
            .filter((a) => a.description.toLowerCase().includes("gap"))
            .map((a) => ({
              alertId: a.alertId,
              // trackId carries the sensor ID for gap alerts emitted by the Coverage Analyzer
              affectedSensorId: a.trackId || "UNKNOWN",
              // Sector ID is encoded in the alert ID suffix (gap-<sectorId>-<timestamp>)
              // or falls back to the alert ID itself for Phase A.
              sectorId: a.alertId.startsWith("gap-")
                ? (a.alertId.split("-")[1] ?? a.alertId)
                : a.alertId,
              severity:
                a.severity === "CRITICAL" ||
                a.severity === "ELEVATED" ||
                a.severity === "WATCH"
                  ? a.severity
                  : ("WATCH" as const),
              description: a.description,
              lastContactUtc: new Date(a.detectedAtMs).toISOString(),
              acknowledged: a.acknowledged,
              areaPolygon: [],
            })),
        );
        break;

      case "token-expiring":
        // Fetch a refreshed JWT and forward it to the Data Worker.
        // Do not log the token value. (SDLC Rule 5)
        fetchAuthToken()
          .then((newToken) => {
            if (newToken && dataWorker) {
              const refreshMsg: TokenRefreshMessage = {
                type: "token-refresh",
                token: newToken,
              };
              dataWorker.postMessage(refreshMsg);
            }
          })
          .catch(() => {
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

    // Init Data Worker — pass URL and JWT token when available.
    // When VITE_WEBTRANSPORT_URL is undefined (local dev), the worker falls back to mock mode.
    const wtUrl = import.meta.env.VITE_WEBTRANSPORT_URL as string | undefined;

    // Transfer OffscreenCanvas to Render Worker
    if (canvasRef) {
      const offscreen = canvasRef.transferControlToOffscreen();
      const initMsg: RenderInitMessage = {
        type: "init",
        canvas: offscreen,
        sab,
        dataWorkerActive: !!wtUrl, // Re-enable mock tracks if no WT URL
      };
      // Transfer OffscreenCanvas only — SAB is shared, not transferred
      renderWorker.postMessage(initMsg, [offscreen]);
    }

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
        setFps(Math.round((frameCount * 1000) / (now - lastFpsTime)));
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

  // Sync dashboard signal to Render Worker
  onMount(() => {
    import("./signals/viewport").then(({ dashboard, viewport }) => {
      createEffect(() => {
        const current = dashboard();
        renderWorker?.postMessage({
          type: "set_dashboard",
          dashboard: current,
        });
      });

      createEffect(() => {
        const vp = viewport();
        renderWorker?.postMessage({
          type: "set_viewport",
          centerLat: vp.centerLat,
          centerLon: vp.centerLon,
          zoom: vp.zoom,
        });
      });
    });
  });

  // Level 3: Strategic View — Sync live coverage to WebGPU
  onMount(() => {
    const syncCoverage = async () => {
      if (!renderWorker) return;
      try {
        const statuses = await fetchSensorStatuses();
        const records = statuses
          .filter((s) => s.coverage)
          .map((s) => ({
            centerLon: s.coverage!.centerLon,
            centerLat: s.coverage!.centerLat,
            rangeNm: s.coverage!.rangeNm,
            bearingStart: s.coverage!.bearingStart,
            bearingEnd: s.coverage!.bearingEnd,
            recordType: 0, // Sector
            alertLevel: s.dlqCount > 100 ? 2 : s.dlqCount > 50 ? 1 : 0,
          }));
        renderWorker.postMessage({ type: "set_coverage", records });
      } catch (e) {
        // Silent fail for background sync
      }
    };

    const timer = setInterval(syncCoverage, 10000);
    syncCoverage();
    onCleanup(() => clearInterval(timer));
  });

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

  createEffect(() => {
    role();
    dashboard();
    enforceRoleDashboardGuard();
  });

  const isOperationsCommander = () => role() === "operations_commander";

  const showAlerts = () => role() === "sensor_operator";

  const showTimeline = () =>
    dashboard() === "analytics" && !isOperationsCommander();

  const mapCanvas = () => (
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
  );

  const renderCommanderDashboard = () => {
    if (dashboard() === "commander") {
      return <FusionCommanderDashboard mapContent={mapCanvas()} />;
    }
    if (dashboard() === "coverage") {
      return <MultiDomainCommanderDashboard mapContent={mapCanvas()} />;
    }
    return (
      <OperatorUiCommanderDashboard
        alertColumnContent={<AlertSidebar />}
        detailPaneContent={<TrackDetailPanel />}
        timelinePaneContent={<EventTimeline />}
      />
    );
  };

  const renderNonCommanderDashboard = () => (
    <Show
      when={dashboard() === "health"}
      fallback={
        <Show when={dashboard() === "coverage"} fallback={mapCanvas()}>
          <CoverageMapDashboard />
        </Show>
      }
    >
      <SensorHealthDashboard />
    </Show>
  );

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
          headerBar={
            <>
              {/* Left group: role and dashboard controls + Health Dashboard Title */}
              <div
                style={{
                  display: "flex",
                  "flex-direction": "row",
                  "align-items": "center",
                  gap: "16px",
                  flex: 1,
                }}
              >
                <div
                  style={{
                    display: "flex",
                    "align-items": "center",
                    gap: "12px",
                  }}
                >
                  <RoleSelector />
                  <DashboardSelector />
                </div>

                <Show when={dashboard() === "health"}>
                  <div
                    style={{
                      padding: "0 16px",
                      "border-left": "1px solid rgba(255,255,255,0.1)",
                      display: "flex",
                      "align-items": "center",
                      gap: "12px",
                    }}
                  >
                    <h1
                      style={{
                        "font-size": "1.1rem",
                        "font-weight": "600",
                        color: "#f8fafc",
                        margin: 0,
                        "letter-spacing": "0.02em",
                      }}
                    >
                      Sensor Health Dashboard
                    </h1>
                  </div>
                </Show>
              </div>

              {/* Right group: Connection and Time */}
              <div
                style={{
                  display: "flex",
                  "flex-direction": "row",
                  "align-items": "center",
                  gap: "20px",
                }}
              >
                <Show when={dashboard() === "health"}>
                  <div
                    style={{
                      "font-family": "monospace",
                      "font-size": "0.85rem",
                      color: "#94a3b8",
                      background: "rgba(0,0,0,0.2)",
                      padding: "4px 10px",
                      "border-radius": "4px",
                      border: "1px solid rgba(255,255,255,0.05)",
                    }}
                  >
                    <span style={{ color: "#64748b", "margin-right": "6px" }}>
                      UTC
                    </span>
                    <span style={{ color: "#e2e8f0" }}>{headerTime()}</span>
                  </div>
                </Show>
                <ConnectionIndicator />
              </div>
            </>
          }
          canvas={
            <Show
              when={isOperationsCommander()}
              fallback={renderNonCommanderDashboard()}
            >
              {renderCommanderDashboard()}
            </Show>
          }
          rightPanel={
            !isOperationsCommander() &&
            dashboard() !== "health" &&
            dashboard() !== "coverage" ? (
              <>
                <TrackDetailPanel />
                <Show when={showAlerts()}>
                  <AlertSidebar />
                </Show>
              </>
            ) : undefined
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
