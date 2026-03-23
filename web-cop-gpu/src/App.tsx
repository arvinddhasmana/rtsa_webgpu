// CLASSIFICATION: UNCLASSIFIED
// src/App.tsx — Root application component (Phase 3: UI & Interaction)

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
    setOpenTrackDetails,
    setSelectedTrack,
    setTrackDetail,
    setTrackDetailError,
    setTrackDetailLoading,
    setTrackOverlayPositions,
} from "./signals/track";
import {
    dashboard,
    enforceRoleDashboardGuard,
    mapStyle,
    role,
    setMapStyle,
    viewport,
} from "./signals/viewport";

// Spatial alert signals
import { setSpatialAlerts } from "./signals/spatial-alerts";

// Services
import { startAlertStream } from "./services/alerts";
import { fetchAuthToken } from "./services/auth";
import { fetchTrackDetail } from "./services/query";
import { startObservationStream } from "./services/sensor-observations";
import { allObservations } from "./signals/sensor-observations";

// Components
import { CoverageMapDashboard } from "./components/dashboard/CoverageMapDashboard";
import { FusionCommanderDashboard } from "./components/dashboard/FusionCommanderDashboard";
import { MultiDomainCommanderDashboard } from "./components/dashboard/MultiDomainCommanderDashboard";
import { OperatorUiCommanderDashboard } from "./components/dashboard/OperatorUiCommanderDashboard";
import { SensorHealthDashboard } from "./components/dashboard/SensorHealthDashboard";
import { LeafletMap } from "./components/map/LeafletMap";
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
    HoveredMessage,
    RenderInitMessage,
    RenderToMainMessage,
    TokenRefreshMessage,
} from "./workers/shared-protocol";

// Fps tracking
let frameCount = 0;
let lastFpsTime = performance.now();

export default function App() {
  const [caps, setCaps] = createSignal<Capabilities | null>(null);
  const [canvas, setCanvas] = createSignal<HTMLCanvasElement | null>(null);

  const [hoveredTrack, setHoveredTrack] = createSignal<HoveredMessage | null>(null);

  const [renderWorker, setRenderWorker] = createSignal<Worker | null>(null);
  const [dataWorker, setDataWorker] = createSignal<Worker | null>(null);
  let alertStreamController: AbortController | null = null;

  // ── Render Worker message handler ──────────────────────────────────────────

  function handleRenderMessage(event: MessageEvent<RenderToMainMessage>) {
    const msg = event.data;
    switch (msg.type) {
      case "status":
        if (msg.ready) {
          setConnecting(false);
          // Re-send the current viewport to the render worker now that it is
          // fully initialised. The initial set_viewport message sent when
          // setRenderWorker() was first called arrived before renderState was
          // ready and was silently dropped, leaving the camera at its (0,0)
          // default. This is especially critical now that the default viewport
          // is Persian Gulf (27°N, 54°E, zoom 6) — without this resend the
          // GPU camera stays at (0°,0°, scale=2) while Leaflet shows a
          // zoomed-in Persian Gulf, making all 150 West Asia tracks appear at
          // completely wrong pixel positions (effectively invisible on the map).
          const rw = renderWorker();
          if (rw) {
            const vp = viewport();
            rw.postMessage({
              type: "set_viewport",
              centerLat: vp.centerLat,
              centerLon: vp.centerLon,
              zoom: vp.zoom,
            });
          }
        }
        break;

      case "hovered": {
        setHoveredTrack(msg);
        break;
      }

      case "picked": {
        console.log("[App] Track picked from GPU:", msg);
        if (msg.trackIdHash === 0) {
          setSelectedTrack(null);
          setTrackDetail(null);
          return;
        }

        setSelectedTrack({
          trackIdHash: msg.trackIdHash,
          x: msg.x,
          y: msg.y,
          source: "canvas",
        });
        const hashHex = msg.trackIdHash.toString(16).padStart(8, "0");
        setTrackDetailLoading(true);
        setTrackDetailError(null);
        setTrackDetail(null);
        fetchTrackDetail(hashHex, msg as any)
          .then((detail) => {
            setTrackDetail(detail);
            setTrackDetailLoading(false);
            if (detail) {
              setOpenTrackDetails((curr) => {
                if (curr.some((t) => t.trackId === detail.trackId)) return curr;
                setTrackOverlayPositions((prev) => ({
                  ...prev,
                  [detail.trackId]: prev[detail.trackId] ?? { x: window.innerWidth / 2 - 400 + curr.length * 40, y: window.innerHeight / 2 - 300 + curr.length * 40 }
                }));
                return [...curr, detail];
              });
            }
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
        if (msg.fps > 0) setFps(msg.fps);
        setTrackCount(msg.trackCount);
        setVisibleCount(msg.visibleCount);
        break;
    }
  }

  // ── UTC Clock ───────────────────────────────────────────────────────────
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
        setSpatialAlerts(
          msg.alerts
            .filter((a) => a.description.toLowerCase().includes("gap"))
            .map((a) => ({
              alertId: a.alertId,
              affectedSensorId: a.trackId || "UNKNOWN",
              sectorId: a.alertId.startsWith("gap-")
                ? (a.alertId.split("-")[1] ?? a.alertId)
                : a.alertId,
              severity:
                a.severity === "CRITICAL" ||
                a.severity === "ELEVATED" ||
                a.severity === "WATCH"
                  ? a.severity
                  : "WATCH",
              description: a.description,
              lastContactUtc: new Date(a.detectedAtMs).toISOString(),
              acknowledged: a.acknowledged,
              areaPolygon: [],
            })),
        );
        break;

      case "token-expiring":
        fetchAuthToken().then((newToken) => {
          const dw = dataWorker();
          if (newToken && dw) {
            dw.postMessage({
              type: "token-refresh",
              token: newToken,
            } as TokenRefreshMessage);
          }
        });
        break;
    }
  }

  // ── Canvas Interaction ─────────────────────────────────────────────────────

  // ── Canvas/Map Interaction ──────────────────────────────────────────────────

  function handleMapClick(x: number, y: number) {
    const el = canvas();
    const rw = renderWorker();
    if (!el || !rw) return;
    const px = Math.round(x * devicePixelRatio);
    const py = Math.round(y * devicePixelRatio);
    rw.postMessage({ type: "select_track", x: px, y: py });
  }

  const [hoverPosition, setHoverPosition] = createSignal<{
    x: number;
    y: number;
  } | null>(null);

  function handleMapHover(x: number, y: number) {
    const el = canvas();
    const rw = renderWorker();
    if (!el || !rw) return;
    const px = Math.round(x * devicePixelRatio);
    const py = Math.round(y * devicePixelRatio);
    setHoverPosition({ x, y }); // Use logical pixels for component absolute positioning
    rw.postMessage({ type: "hover_track", x: px, y: py });
  }

  // ── Resize ────────────────────────────────────────────────────────────────
  let resizeObserver: ResizeObserver | null = null;
  function setupResizeObserver(canvasEl: HTMLCanvasElement) {
    resizeObserver = new ResizeObserver((entries) => {
      const entry = entries[0];
      const rw = renderWorker();
      if (!entry || !rw) return;
      const { width, height } = entry.contentRect;
      const dpr = devicePixelRatio;
      rw.postMessage({
        type: "resize",
        width: Math.round(width * dpr),
        height: Math.round(height * dpr),
      });
    });
    resizeObserver.observe(canvasEl);
  }

  // ── Initialise Workers ─────────────────────────────────────────────────────

  async function init(detected: Capabilities) {
    if (
      !detected.webgpu ||
      !detected.sharedArrayBuffer ||
      !detected.offscreenCanvas
    )
      return;

    const ringBuffer = allocateSAB();
    const sab = ringBuffer.sab;

    const rwMain = new Worker(
      new URL("./workers/render-worker.ts", import.meta.url),
      { type: "module" },
    );
    const dwMain = new Worker(
      new URL("./workers/data-worker.ts", import.meta.url),
      { type: "module" },
    );

    rwMain.addEventListener("message", handleRenderMessage);
    dwMain.addEventListener("message", handleDataMessage);

    setRenderWorker(rwMain);
    setDataWorker(dwMain);

    const wtUrl = import.meta.env.VITE_WEBTRANSPORT_URL as string | undefined;

    let lastCanvas: HTMLCanvasElement | null = null;
    createEffect(() => {
      const el = canvas();
      const worker = renderWorker();
      if (el && worker && el !== lastCanvas) {
        lastCanvas = el;
        const tryInit = () => {
          if (el !== canvas()) return; // Canvas changed again
          const dpr = devicePixelRatio;
          let w =
            el.offsetWidth || Math.round(el.getBoundingClientRect().width);
          let h =
            el.offsetHeight || Math.round(el.getBoundingClientRect().height);

          if (w === 0 || h === 0) {
            requestAnimationFrame(tryInit);
            return;
          }

          const initialWidth = Math.round(w * dpr);
          const initialHeight = Math.round(h * dpr);

          console.log(
            `[App] Canvas mounted. size: ${initialWidth}x${initialHeight}`,
          );
          const offscreen = el.transferControlToOffscreen();
          worker.postMessage(
            {
              type: "init",
              canvas: offscreen,
              sab,
              initialWidth,
              initialHeight,
              dataWorkerActive: !!wtUrl,
            } as RenderInitMessage,
            [offscreen],
          );
          setupResizeObserver(el);
        };
        requestAnimationFrame(tryInit);
      }
    });

    const token = wtUrl ? await fetchAuthToken() : undefined;
    setOperatorId(operatorIdFromToken(token));
    const dwActive = dataWorker();
    if (dwActive) {
      dwActive.postMessage({
        type: "init",
        sab,
        url: wtUrl,
        token,
      } as DataInitMessage);
    }

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

    console.log("[App] Initializing services...");
    startObservationStream();
    alertStreamController = startAlertStream();
  }

  // ── Sync State ───────────────────────────────────────────────────────────

  onMount(() => {
    createEffect(() => {
      const current = dashboard();
      const currentRole = role();
      const rw = renderWorker();
      if (rw) {
        console.log(`[App] Syncing dashboard mode: ${current}`);
        rw.postMessage({ type: "set_dashboard", dashboard: current });

        // Force HD map by default for Operations Commander
        if (currentRole === "operations_commander") {
          setMapStyle(1);
        }
      }
    });

    createEffect(() => {
      const style = mapStyle();
      const rw = renderWorker();
      if (rw) {
        rw.postMessage({ type: "set_map_style", mapStyle: style });
      }
    });

    createEffect(() => {
      const vp = viewport();
      const rw = renderWorker();
      if (rw) {
        rw.postMessage({
          type: "set_viewport",
          centerLat: vp.centerLat,
          centerLon: vp.centerLon,
          zoom: vp.zoom,
        });
      }
    });
  });

  // Observations sync
  createEffect(() => {
    const obs = allObservations();
    const rw = renderWorker();
    if (!rw) return;

    const records = obs.map((o: any) => ({
      id: o.id,
      lat: o.lat,
      lon: o.lon,
      type: o.type,
      confidence: o.confidence,
    }));
    rw.postMessage({ type: "set_observations", observations: records });
  });

  onMount(async () => {
    const detected = await checkCapabilities();
    setCaps(detected);
    await init(detected);
  });

  onCleanup(() => {
    renderWorker()?.terminate();
    dataWorker()?.terminate();
    alertStreamController?.abort();
    resizeObserver?.disconnect();
  });

  createEffect(() => {
    role();
    dashboard();
    enforceRoleDashboardGuard();
  });

  const isOperationsCommander = () => role() === "operations_commander";
  const showAlerts = () => role() === "sensor_operator";
  const showTimeline = () =>
    dashboard() === "analytics" && !isOperationsCommander();

  const renderMainViewport = () => {
    const currentDashboard = dashboard();
    const MapCanvas = (
      <canvas
        ref={setCanvas}
        id="gpu-canvas"
        style={{
          width: "100%",
          height: "100%",
          display: "block",
          "pointer-events": "none",
          "z-index": 1,
          position: "absolute",
          top: 0,
          left: 0,
        }}
      />
    );

    const MapContainer = (
      <div style={{ position: "relative", width: "100%", height: "100%" }}>
        <LeafletMap onMapClick={handleMapClick} onMapHover={handleMapHover} />
        {MapCanvas}
        <Show when={hoveredTrack() !== null && hoveredTrack()!.trackIdHash !== 0 && hoverPosition() !== null}>
          {(() => {
            const t = hoveredTrack()!;
            const AFFILIATION = ["UNKNOWN","PENDING","FRIENDLY","NEUTRAL","SUSPECT","HOSTILE"];
            const ENTITY_TYPE = ["UNSPEC","SURFACE","AIR","SUBSURFACE","LAND","CYBER"];
            const AFFIL_COLOR = ["#FBBF24","#80CCFF","#38BDFF","#57E688","#FF9933","#F87171"];
            const affil = t.threatLevel >= 0 && t.threatLevel <= 5 ? t.threatLevel : 0;
            const etype = t.entityType >= 0 && t.entityType <= 5 ? t.entityType : 0;
            const ctx   = t.context === 1 ? "CIVILIAN" : "MILITARY";
            const speedKts = (t.speed * 1.94384).toFixed(0);
            const altFt    = (t.altitude * 3.28084).toFixed(0);
            const color    = AFFIL_COLOR[affil]!;
            return (
              <div
                style={{
                  position: "absolute",
                  left: `${hoverPosition()!.x + 18}px`,
                  top: `${hoverPosition()!.y + 18}px`,
                  background: "rgba(8, 18, 36, 0.93)",
                  color: "#e2e8f0",
                  padding: "8px 12px",
                  "border-radius": "6px",
                  border: `1px solid ${color}`,
                  "z-index": 1000,
                  "pointer-events": "none",
                  "font-family": "'Courier New', monospace",
                  "font-size": "11px",
                  "line-height": "1.6",
                  "min-width": "170px",
                  "box-shadow": `0 0 8px ${color}55`,
                }}
              >
                <div style={{ color, "font-weight": "bold", "margin-bottom": "4px", "font-size": "12px" }}>
                  ▸ {AFFILIATION[affil]} {ENTITY_TYPE[etype]}
                </div>
                <div>CTX:&nbsp; <span style={{ color: "#94a3b8" }}>{ctx}</span></div>
                <div>SPD:&nbsp; <span style={{ color: "#94a3b8" }}>{speedKts} kts</span></div>
                <div>ALT:&nbsp; <span style={{ color: "#94a3b8" }}>{altFt} ft</span></div>
                {t.alertFlags > 0 && <div style={{ color: "#F87171", "margin-top": "4px" }}>⚠ ALERT ACTIVE</div>}
                <div style={{ color: "#475569", "margin-top": "4px", "font-size": "10px" }}>
                  ID: {t.trackIdHash.toString(16).padStart(8, "0").toUpperCase()}
                </div>
                <div style={{ color: "#475569", "font-size": "10px" }}>Click for full details</div>
              </div>
            );
          })()}
        </Show>
      </div>
    );

    if (currentDashboard === "health") return <SensorHealthDashboard />;
    if (currentDashboard === "commander")
      return <FusionCommanderDashboard mapContent={MapContainer} />;
    if (currentDashboard === "coverage") {
      return isOperationsCommander() ? (
        <MultiDomainCommanderDashboard mapContent={MapContainer} />
      ) : (
        <CoverageMapDashboard />
      );
    }
    if (!isOperationsCommander()) return MapContainer;

    return (
      <OperatorUiCommanderDashboard
        alertColumnContent={<AlertSidebar />}
        detailPaneContent={<TrackDetailPanel />}
      />
    );
  };

  return (
    <Show
      when={caps() !== null}
      fallback={<div style={{ padding: "2rem" }}>Initialising…</div>}
    >
      <Show
        when={
          caps()!.webgpu && caps()!.sharedArrayBuffer && caps()!.offscreenCanvas
        }
        fallback={renderDegradedNotice(caps()!)}
      >
        <AppShell
          headerBar={
            <>
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
          canvas={renderMainViewport()}
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
