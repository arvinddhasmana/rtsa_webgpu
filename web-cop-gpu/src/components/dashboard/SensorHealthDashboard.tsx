// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/SensorHealthDashboard.tsx — Main Health Dashboard
//
// Reference: docs/business/usecases/UC017_sensor_health_monitoring.md

import {
    createEffect,
    createResource,
    createSignal,
    For,
    onCleanup,
    Show,
} from "solid-js";
import {
    fetchSensorStatuses,
    SensorStatus,
} from "../../services/sensor-health";
import {
    cardView,
    selectedSensor,
    setSelectedSensor,
} from "../../signals/sensor-filters";
import { spatialAlerts } from "../../signals/spatial-alerts";
import { dashboard } from "../../signals/viewport";
import { CriticalAlertsPanel } from "./CriticalAlertsPanel";
import { DashboardSidebar } from "./DashboardSidebar";
import { DraggableOverlayCard } from "./DraggableOverlayCard";
import { SensorDetailHoverPanel } from "./SensorDetailHoverPanel";
import { SensorDiagnosticView } from "./SensorDiagnosticView";
import { SensorFleetList } from "./SensorFleetList";
import { SensorGrid } from "./SensorGrid";
import { SensorHealthDiagnosticCard } from "./SensorHealthDiagnosticCard";
import { SensorOverviewMap } from "./SensorOverviewMap";
import { statusColor } from "./dashboard-utils";

// ── Icons ──────────────────────────────────────────────────────────────────

function IconLayoutV() {
  return (
    <svg
      width="14"
      height="14"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      stroke-width="2"
    >
      <rect x="3" y="3" width="18" height="8" rx="1" />
      <rect x="3" y="13" width="18" height="8" rx="1" />
    </svg>
  );
}

function IconLayoutH() {
  return (
    <svg
      width="14"
      height="14"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      stroke-width="2"
    >
      <rect x="3" y="3" width="8" height="18" rx="1" />
      <rect x="13" y="3" width="8" height="18" rx="1" />
    </svg>
  );
}

function IconMaximize() {
  return (
    <svg
      width="12"
      height="12"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      stroke-width="2"
    >
      <path d="M8 3H5a2 2 0 0 0-2 2v3m18 0V5a2 2 0 0 0-2-2h-3m0 18h3a2 2 0 0 0 2-2v-3M3 16v3a2 2 0 0 0 2 2h3" />
    </svg>
  );
}

function IconMinimize() {
  return (
    <svg
      width="12"
      height="12"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      stroke-width="2"
    >
      <path d="M8 3v3a2 2 0 0 1-2 2H3m18 0h-3a2 2 0 0 1-2-2V3m0 18v-3a2 2 0 0 1 2-2h3M3 16h3a2 2 0 0 1 2 2v3" />
    </svg>
  );
}

function IconSwap() {
  return (
    <svg
      width="14"
      height="14"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      stroke-width="2"
    >
      <path d="M7 16V4m0 0L3 8m4-4l4 4" />
      <path d="M17 8v12m0 0l4-4m-4 4l-4-4" />
    </svg>
  );
}

/**
 * Sensor Health Monitoring Dashboard.
 * Orchestrates data fetching, resizable split layout, and filtering.
 */
export function SensorHealthDashboard() {
  const [sensors, { refetch }] = createResource(fetchSensorStatuses);
  const [hoveredSensorId, setHoveredSensorId] = createSignal<
    string | undefined
  >(undefined);
  const [hoveredSensor, setHoveredSensor] =
    createSignal<Parameters<typeof SensorDetailHoverPanel>[0]["sensor"]>(null);

  // ── Layout state ──
  /** Split ratio 0-100: percentage size of pane A */
  const [splitRatio, setSplitRatio] = createSignal(40);
  /** "vertical" = top/bottom split; "horizontal" = left/right split */
  const [layout, setLayout] = createSignal<"vertical" | "horizontal">(
    "vertical",
  );
  /** Whether pane A (status cards) and pane B (coverage) are swapped */
  const [swapped, setSwapped] = createSignal(false);
  /** Which pane is fullscreen (null = neither) */
  const [fullscreen, setFullscreen] = createSignal<null | "a" | "b">(null);
  /** Open diagnostic overlays keyed by sensor */
  const [openDiagnostics, setOpenDiagnostics] = createSignal<SensorStatus[]>(
    [],
  );
  const [overlayPositions, setOverlayPositions] = createSignal<
    Record<string, { x: number; y: number }>
  >({});

  function basePosition(index: number): { x: number; y: number } {
    const root = document.getElementById("sensor-health-dashboard-root");
    const pane = document.getElementById("health-split-container");
    if (root && pane) {
      const rootR = root.getBoundingClientRect();
      const paneR = pane.getBoundingClientRect();
      return {
        x: Math.max(0, paneR.left - rootR.left) + 24 + index * 28,
        y: Math.max(0, paneR.top - rootR.top) + 24 + index * 18,
      };
    }
    return { x: 24 + index * 28, y: 24 + index * 18 };
  }

  function openDiagnosticCard(sensor: SensorStatus) {
    setOpenDiagnostics((curr) => {
      if (curr.some((s) => s.sensorId === sensor.sensorId)) {
        return curr;
      }
      setOverlayPositions((prev) => ({
        ...prev,
        [sensor.sensorId]: prev[sensor.sensorId] ?? basePosition(curr.length),
      }));
      return [...curr, sensor];
    });
  }

  function closeDiagnosticCard(sensorId: string) {
    setOpenDiagnostics((curr) => curr.filter((s) => s.sensorId !== sensorId));
  }

  function closeAllDiagnostics() {
    setOpenDiagnostics([]);
  }

  function updateOverlayPosition(
    sensorId: string,
    pos: { x: number; y: number },
  ) {
    setOverlayPositions((prev) => ({ ...prev, [sensorId]: pos }));
  }

  function autoArrangeDiagnostics() {
    const root = document.getElementById("sensor-health-dashboard-root");
    const pane = document.getElementById("health-split-container");
    const items = openDiagnostics();
    if (!root || !pane || items.length === 0) return;

    const rootR = root.getBoundingClientRect();
    const paneR = pane.getBoundingClientRect();

    // Calculate offset relative to the dashboard root
    const offX = Math.max(0, paneR.left - rootR.left);
    const offY = Math.max(0, paneR.top - rootR.top);
    const w = paneR.width;
    const h = paneR.height;
    if (w <= 0 || h <= 0) return;

    const gap = 16;
    const cardW = 440; // Use a fixed width for consistent arrangement
    const cardH = 580; // Estimated height

    const columns = Math.max(1, Math.floor((w - 20) / (cardW + gap)));
    const arranged: Record<string, { x: number; y: number }> = {};

    items.forEach((sensor, idx) => {
      const col = idx % columns;
      const row = Math.floor(idx / columns);
      arranged[sensor.sensorId] = {
        x: offX + 10 + col * (cardW + gap),
        y: offY + 10 + row * (cardH + gap),
      };
    });
    setOverlayPositions(arranged);
  }

  function syncOverlayFrame() {
    // No-op: overlay now uses inset:0 and auto-arrange reads DOM directly
  }

  // ── Divider drag ──
  const [draggingDivider, setDraggingDivider] = createSignal(false);

  function onDividerMouseDown(e: MouseEvent) {
    setDraggingDivider(true);
    e.preventDefault();
  }

  function onMouseMove(e: MouseEvent) {
    if (!draggingDivider()) return;
    const container = document.getElementById("health-split-container");
    if (!container) return;
    const rect = container.getBoundingClientRect();
    if (layout() === "vertical") {
      const ratio = ((e.clientY - rect.top) / rect.height) * 100;
      setSplitRatio(Math.min(Math.max(ratio, 15), 80));
    } else {
      const ratio = ((e.clientX - rect.left) / rect.width) * 100;
      setSplitRatio(Math.min(Math.max(ratio, 15), 80));
    }
  }

  function onMouseUp() {
    setDraggingDivider(false);
  }

  function onKeyDown(e: KeyboardEvent) {
    if (e.key === "Escape" && openDiagnostics().length > 0) {
      e.preventDefault();
      closeAllDiagnostics();
    }
  }

  window.addEventListener("mousemove", onMouseMove);
  window.addEventListener("mouseup", onMouseUp);
  window.addEventListener("keydown", onKeyDown);
  window.addEventListener("resize", syncOverlayFrame);

  // Auto-refresh every 10 seconds as per requirements
  const timer = setInterval(refetch, 10000);
  onCleanup(() => {
    clearInterval(timer);
    window.removeEventListener("mousemove", onMouseMove);
    window.removeEventListener("mouseup", onMouseUp);
    window.removeEventListener("keydown", onKeyDown);
    window.removeEventListener("resize", syncOverlayFrame);
  });

  // Clear the selected sensor when dashboard changes
  createEffect(() => {
    void dashboard();
    setSelectedSensor(null);
    closeAllDiagnostics();
    setOverlayPositions({});
  });

  createEffect(() => {
    void layout();
    void splitRatio();
    void swapped();
    void fullscreen();
    void selectedSensor();
  });

  // ── Computed pane sizes ──
  const paneASize = () => {
    if (fullscreen() === "a") return "100%";
    if (fullscreen() === "b") return "0%";
    return `${splitRatio()}%`;
  };

  // ── Toolbar button style ──
  const tbBtn = (active?: boolean) => ({
    display: "inline-flex",
    "align-items": "center",
    gap: "4px",
    background: active ? "rgba(59,130,246,0.18)" : "rgba(255,255,255,0.04)",
    border: active
      ? "1px solid rgba(59,130,246,0.4)"
      : "1px solid rgba(255,255,255,0.1)",
    "border-radius": "6px",
    color: active ? "#60a5fa" : "#64748b",
    padding: "3px 8px",
    cursor: "pointer",
    "font-size": "0.58rem",
    "font-family": "monospace",
    "letter-spacing": "0.05em",
    transition: "all 0.15s ease",
  });

  // ── Pane panels ──

  const StatusCardsPane = () => (
    <div
      style={{
        display: "flex",
        "flex-direction": "column",
        height: "100%",
        overflow: "hidden",
      }}
    >
      {/* Pane header */}
      <div
        style={{
          padding: "4px 14px",
          "border-bottom": "1px solid rgba(255,255,255,0.04)",
          display: "flex",
          "align-items": "center",
          gap: "6px",
          "flex-shrink": 0,
          background: "rgba(0,0,0,0.1)",
        }}
      >
        <span
          style={{
            "font-size": "0.58rem",
            "font-weight": "700",
            "text-transform": "uppercase",
            "letter-spacing": "0.1em",
            color: "#475569",
          }}
        >
          Sensor Status
        </span>
        <span style={{ "font-size": "0.55rem", color: "#1e3a5f" }}>
          Real-time health overview
        </span>
        <div style={{ flex: 1 }} />
        {/* Fullscreen toggle for this pane */}
        <button
          title={
            fullscreen() === (swapped() ? "b" : "a") ? "Restore" : "Fullscreen"
          }
          onClick={() =>
            setFullscreen((f) =>
              f === (swapped() ? "b" : "a") ? null : swapped() ? "b" : "a",
            )
          }
          style={tbBtn(fullscreen() === (swapped() ? "b" : "a"))}
        >
          <Show
            when={fullscreen() === (swapped() ? "b" : "a")}
            fallback={<IconMaximize />}
          >
            <IconMinimize />
          </Show>
        </button>
      </div>

      <div style={{ flex: 1, overflow: "hidden", "min-height": 0 }}>
        <SensorGrid
          sensors={sensors() || []}
          cardView={cardView()}
          onSensorSelect={(s) => {
            setHoveredSensorId(s.sensorId);
            setHoveredSensor(s);
            setSelectedSensor(s);
          }}
          onOpenDiagnostic={(s) => openDiagnosticCard(s)}
        />
      </div>
    </div>
  );

  const CoveragePane = () => (
    <div
      style={{
        display: "flex",
        "flex-direction": "column",
        height: "100%",
        overflow: "hidden",
        background: "rgba(10, 15, 28, 0.45)",
        "backdrop-filter": "blur(20px)",
      }}
    >
      {/* Pane header */}
      <div
        style={{
          padding: "4px 14px",
          "border-bottom": "1px solid rgba(255,255,255,0.04)",
          display: "flex",
          "align-items": "center",
          gap: "6px",
          "flex-shrink": 0,
          background: "rgba(0,0,0,0.1)",
        }}
      >
        <span
          style={{
            "font-size": "0.58rem",
            "font-weight": "700",
            "text-transform": "uppercase",
            "letter-spacing": "0.1em",
            color: "#475569",
          }}
        >
          Sensor Coverage Map
        </span>
        <div style={{ flex: 1 }} />
        {/* Fullscreen toggle for this pane */}
        <button
          title={
            fullscreen() === (swapped() ? "a" : "b") ? "Restore" : "Fullscreen"
          }
          onClick={() =>
            setFullscreen((f) =>
              f === (swapped() ? "a" : "b") ? null : swapped() ? "a" : "b",
            )
          }
          style={tbBtn(fullscreen() === (swapped() ? "a" : "b"))}
        >
          <Show
            when={fullscreen() === (swapped() ? "a" : "b")}
            fallback={<IconMaximize />}
          >
            <IconMinimize />
          </Show>
        </button>
      </div>

      <div
        style={{ flex: 1, overflow: "hidden", "min-height": 0, padding: "8px" }}
      >
        <SensorOverviewMap
          sensors={sensors() || []}
          spatialAlerts={spatialAlerts()}
          hoveredSensorId={hoveredSensorId()}
          onSensorClick={(s) => {
            setHoveredSensorId(s.sensorId);
            setHoveredSensor(s);
          }}
          height={400}
          fleetListPanel={
            <SensorFleetList
              sensors={sensors() || []}
              selectedSensorId={hoveredSensorId()}
              onSensorSelect={(s) => {
                setHoveredSensorId(s.sensorId);
                setHoveredSensor(s);
              }}
              onSensorHover={(s) => {
                if (s) {
                  setHoveredSensorId(s.sensorId);
                  setHoveredSensor(s);
                }
              }}
              maxHeight="240px"
            />
          }
          criticalAlertsPanel={
            <CriticalAlertsPanel
              spatialAlerts={spatialAlerts()}
              maxHeight="180px"
            />
          }
          sensorDetailPanel={
            <Show
              when={hoveredSensor() !== null}
              fallback={
                <div
                  style={{
                    padding: "16px 8px",
                    color: "#334155",
                    "font-size": "0.68rem",
                    "font-family": "monospace",
                    "text-align": "center",
                    "line-height": "1.6",
                  }}
                >
                  Select a sensor from the fleet list or click a footprint on
                  the map.
                </div>
              }
            >
              <SensorDetailHoverPanel sensor={hoveredSensor()} width="100%" />
            </Show>
          }
        />
      </div>
    </div>
  );

  // Determine pane order based on swapped state
  const PaneA = () => (swapped() ? <CoveragePane /> : <StatusCardsPane />);
  const PaneB = () => (swapped() ? <StatusCardsPane /> : <CoveragePane />);

  return (
    <div
      id="sensor-health-dashboard-root"
      style={{
        display: "flex",
        height: "100%",
        width: "100%",
        background:
          "radial-gradient(circle at 0% 0%, rgba(30, 58, 138, 0.15) 0%, transparent 50%), radial-gradient(circle at 100% 100%, rgba(88, 28, 135, 0.15) 0%, transparent 50%)",
        overflow: "hidden",
        position: "relative",
      }}
    >
      <DashboardSidebar sensors={sensors() || []} />

      <div
        style={{
          flex: 1,
          display: "flex",
          "flex-direction": "column",
          overflow: "hidden",
          "min-width": 0,
          position: "relative",
        }}
      >
        <Show
          when={!sensors.error}
          fallback={
            <div style={{ padding: "40px", color: "#f87171" }}>
              Error loading sensor data: {sensors.error?.message}
            </div>
          }
        >
          {/* Level 2: Sensor Diagnostic Detail — replaces dashboard when sensor selected */}
          <Show when={selectedSensor() !== null}>
            <SensorDiagnosticView sensor={selectedSensor()!} />
          </Show>

          {/* Level 1 Dashboard — visible when no sensor selected */}
          <Show when={selectedSensor() === null}>
            <div
              style={{
                display: "flex",
                "flex-direction": "column",
                flex: 1,
                overflow: "hidden",
                "min-height": 0,
              }}
            >
              {/* ── Split layout toolbar ── */}
              <div
                style={{
                  display: "flex",
                  "align-items": "center",
                  gap: "6px",
                  padding: "4px 12px",
                  "border-bottom": "1px solid rgba(255,255,255,0.05)",
                  "flex-shrink": 0,
                  background: "rgba(0,0,0,0.15)",
                }}
              >
                <span
                  style={{
                    "font-size": "0.55rem",
                    color: "#334155",
                    "text-transform": "uppercase",
                    "letter-spacing": "0.08em",
                    "font-family": "monospace",
                  }}
                >
                  Layout
                </span>

                {/* Toggle layout direction */}
                <button
                  data-testid="layout-toggle-vertical"
                  title="Top / Bottom split"
                  onClick={() => {
                    setLayout("vertical");
                    setFullscreen(null);
                  }}
                  style={tbBtn(layout() === "vertical")}
                >
                  <IconLayoutV />
                  T/B
                </button>
                <button
                  data-testid="layout-toggle-horizontal"
                  title="Left / Right split"
                  onClick={() => {
                    setLayout("horizontal");
                    setFullscreen(null);
                  }}
                  style={tbBtn(layout() === "horizontal")}
                >
                  <IconLayoutH />
                  L/R
                </button>

                <div
                  style={{
                    width: "1px",
                    height: "14px",
                    background: "rgba(255,255,255,0.06)",
                  }}
                />

                {/* Swap panes */}
                <button
                  data-testid="layout-swap"
                  title="Swap panes"
                  onClick={() => {
                    setSwapped((s) => !s);
                    setFullscreen(null);
                  }}
                  style={tbBtn(swapped())}
                >
                  <IconSwap />
                  Swap
                </button>

                <div
                  style={{
                    width: "1px",
                    height: "14px",
                    background: "rgba(255,255,255,0.06)",
                  }}
                />

                {/* Restore from fullscreen */}
                <Show when={fullscreen() !== null}>
                  <button
                    data-testid="layout-restore"
                    title="Restore split view"
                    onClick={() => setFullscreen(null)}
                    style={tbBtn(false)}
                  >
                    <IconMinimize />
                    Restore Split
                  </button>
                </Show>

                <div style={{ flex: 1 }} />

                <span
                  style={{
                    "font-size": "0.52rem",
                    color: "#1e3a5f",
                    "font-family": "monospace",
                  }}
                >
                  Drag divider to resize · ⤢ for fullscreen
                </span>
              </div>

              {/* ── Split container ── */}
              <div
                id="health-split-container"
                style={{
                  flex: 1,
                  display: "flex",
                  "flex-direction": layout() === "vertical" ? "column" : "row",
                  overflow: "hidden",
                  "min-height": 0,
                  "user-select": draggingDivider() ? "none" : "auto",
                  position: "relative",
                }}
              >
                {/* ── Pane A ── */}
                <div
                  style={{
                    [layout() === "vertical" ? "height" : "width"]: paneASize(),
                    "flex-shrink": 0,
                    overflow: "hidden",
                    transition: draggingDivider()
                      ? "none"
                      : "width 0.15s ease, height 0.15s ease",
                    display: fullscreen() === "b" ? "none" : "flex",
                    "flex-direction": "column",
                  }}
                >
                  <PaneA />
                </div>

                {/* ── Resizable divider ── */}
                <Show when={fullscreen() === null}>
                  <div
                    data-testid="split-divider"
                    onMouseDown={onDividerMouseDown}
                    style={{
                      [layout() === "vertical" ? "height" : "width"]: "5px",
                      "flex-shrink": 0,
                      background: draggingDivider()
                        ? "rgba(59,130,246,0.6)"
                        : "rgba(255,255,255,0.06)",
                      cursor:
                        layout() === "vertical" ? "row-resize" : "col-resize",
                      position: "relative",
                      transition: "background 0.15s ease",
                      display: "flex",
                      "align-items": "center",
                      "justify-content": "center",
                    }}
                  >
                    {/* Grip dots */}
                    <div
                      style={{
                        display: "flex",
                        "flex-direction":
                          layout() === "vertical" ? "row" : "column",
                        gap: "3px",
                        opacity: 0.4,
                        "pointer-events": "none",
                      }}
                    >
                      <div
                        style={{
                          width: "3px",
                          height: "3px",
                          "border-radius": "50%",
                          background: "#94a3b8",
                        }}
                      />
                      <div
                        style={{
                          width: "3px",
                          height: "3px",
                          "border-radius": "50%",
                          background: "#94a3b8",
                        }}
                      />
                      <div
                        style={{
                          width: "3px",
                          height: "3px",
                          "border-radius": "50%",
                          background: "#94a3b8",
                        }}
                      />
                    </div>
                  </div>
                </Show>

                {/* ── Pane B ── */}
                <div
                  style={{
                    flex: 1,
                    overflow: "hidden",
                    "min-height": 0,
                    "min-width": 0,
                    display: fullscreen() === "a" ? "none" : "flex",
                    "flex-direction": "column",
                  }}
                >
                  <PaneB />
                </div>
              </div>
            </div>
          </Show>
        </Show>
      </div>

      <Show when={openDiagnostics().length > 0}>
        <div
          data-testid="diagnostic-overlay-layer"
          style={{
            position: "absolute",
            inset: "0",
            "pointer-events": "none",
            "z-index": 120,
          }}
        >
          <Show when={openDiagnostics().length > 0}>
            <button
              data-testid="diagnostic-auto-arrange"
              onClick={() => autoArrangeDiagnostics()}
              title="Auto-arrange diagnostic cards"
              style={{
                position: "absolute",
                top: "8px",
                right: "8px",
                background: "rgba(255,255,255,0.06)",
                border: "1px solid rgba(255,255,255,0.14)",
                color: "#94a3b8",
                padding: "4px 8px",
                "border-radius": "8px",
                "font-size": "0.6rem",
                "font-family": "monospace",
                cursor: "pointer",
                "pointer-events": "auto",
                "box-shadow": "0 10px 30px rgba(0,0,0,0.35)",
              }}
            >
              Auto Arrange
            </button>
          </Show>

          <For each={openDiagnostics()}>
            {(sensor, idx) => {
              const pos =
                overlayPositions()[sensor.sensorId] ?? basePosition(idx());
              return (
                <DraggableOverlayCard
                  title={`Diag · ${sensor.sensorId}`}
                  position={pos}
                  onPositionChange={(p) =>
                    updateOverlayPosition(sensor.sensorId, p)
                  }
                  onClose={() => closeDiagnosticCard(sensor.sensorId)}
                  width="clamp(380px, 30vw, 500px)"
                  minWidth="360px"
                  maxHeight="66vh"
                  accentColor={statusColor(sensor.status)}
                  zIndex={200 + idx()}
                  constrainToParent={true}
                >
                  <SensorHealthDiagnosticCard
                    sensor={sensor}
                    onClose={() => closeDiagnosticCard(sensor.sensorId)}
                  />
                </DraggableOverlayCard>
              );
            }}
          </For>
        </div>
      </Show>

      <style>{`
        .refresh-btn:hover {
            background: rgba(59, 130, 246, 0.2);
            color: white;
        }
      `}</style>
    </div>
  );
}
