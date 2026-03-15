// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/SensorGrid.tsx — Filtered sensor card grid
//
// Reference: docs/business/usecases/UC017_sensor_health_monitoring.md

import { createMemo, createSignal, For, Show } from "solid-js";
import { Portal } from "solid-js/web";
import { SensorStatus } from "../../services/sensor-health";
import { selectedStatuses, selectedTypes } from "../../signals/sensor-filters";
import { SensorHealthDiagnosticCard } from "./SensorHealthDiagnosticCard";
import { SensorStatusCard } from "./SensorStatusCard";

interface SensorGridProps {
  sensors: SensorStatus[];
  /** "full" (default) = rich dual-sparkline cards; "compact" = condensed metric cards */
  cardView?: "full" | "compact";
  onSensorSelect?: (sensor: SensorStatus) => void;
}

const ROW_OPTIONS = [1, 2, 3, 4] as const;

/**
 * Responsive sensor card grid with:
 * - Configurable row count via combobox (1–4 rows)
 * - Horizontal scroll for overflow sensors
 * - Inline diagnostic overlay triggered from each card's scope button
 *
 * Never destructure props — breaks SolidJS reactivity.
 */
export function SensorGrid(props: SensorGridProps) {
  const [rows, setRows] = createSignal<number>(2);
  const [diagnoseSensor, setDiagnoseSensor] = createSignal<SensorStatus | null>(
    null,
  );

  const filteredSensors = createMemo(() =>
    props.sensors.filter(
      (s) =>
        selectedStatuses().includes(s.status) &&
        selectedTypes().includes(s.sensorType),
    ),
  );

  const cardMinWidth = () =>
    props.cardView === "compact"
      ? "clamp(240px, 18vw, 310px)"
      : "clamp(280px, 22vw, 355px)";

  return (
    <div
      style={{
        display: "flex",
        "flex-direction": "column",
        height: "100%",
        overflow: "hidden",
        "background-color": "rgba(10, 15, 28, 0.2)",
      }}
    >
      {/* ── Toolbar: rows combobox + sensor count ── */}
      <div
        style={{
          display: "flex",
          "align-items": "center",
          gap: "10px",
          padding: "6px 24px",
          "border-bottom": "1px solid rgba(255,255,255,0.04)",
          "flex-shrink": 0,
        }}
      >
        <span
          style={{
            "font-size": "0.58rem",
            color: "#475569",
            "text-transform": "uppercase",
            "letter-spacing": "0.08em",
            "font-family": "monospace",
          }}
        >
          Rows
        </span>
        <select
          value={rows()}
          onChange={(e) => setRows(Number(e.currentTarget.value))}
          style={{
            background: "rgba(10, 15, 28, 0.9)",
            border: "1px solid rgba(255,255,255,0.14)",
            "border-radius": "6px",
            color: "#94a3b8",
            "font-size": "0.65rem",
            "font-family": "monospace",
            padding: "3px 24px 3px 8px",
            cursor: "pointer",
            outline: "none",
            appearance: "none",
            "-webkit-appearance": "none",
            "background-image":
              "url(\"data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='10' height='6' viewBox='0 0 10 6'%3E%3Cpath d='M0 0l5 6 5-6z' fill='%2364748b'/%3E%3C/svg%3E\")",
            "background-repeat": "no-repeat",
            "background-position": "right 7px center",
            transition: "border-color 0.2s",
          }}
        >
          <For each={ROW_OPTIONS}>
            {(n) => (
              <option value={n}>
                {n} {n === 1 ? "Row" : "Rows"}
              </option>
            )}
          </For>
        </select>

        {/* Divider */}
        <div
          style={{
            width: "1px",
            height: "14px",
            background: "rgba(255,255,255,0.06)",
          }}
        />

        <span
          style={{
            "font-size": "0.58rem",
            color: "#334155",
            "font-family": "monospace",
          }}
        >
          {filteredSensors().length} sensor
          {filteredSensors().length !== 1 ? "s" : ""}
        </span>

        <Show when={filteredSensors().length > rows() * 4}>
          <span
            style={{
              "font-size": "0.55rem",
              color: "#1e3a5f",
              "font-family": "monospace",
              display: "flex",
              "align-items": "center",
              gap: "4px",
            }}
          >
            <svg
              width="10"
              height="10"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
            >
              <polyline points="9 18 15 12 9 6" />
              <polyline points="15 18 21 12 15 6" />
            </svg>
            scroll for more
          </span>
        </Show>
      </div>

      {/* ── Card grid ── */}
      <Show
        when={filteredSensors().length > 0}
        fallback={
          <div
            style={{
              flex: 1,
              display: "flex",
              "flex-direction": "column",
              "align-items": "center",
              "justify-content": "center",
              color: "#64748b",
              gap: "1rem",
            }}
          >
            <svg
              width="48"
              height="48"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="1"
              opacity="0.5"
            >
              <circle cx="12" cy="12" r="10" />
              <line x1="8" y1="12" x2="16" y2="12" />
            </svg>
            <div style={{ "font-size": "1.1rem" }}>
              No sensors match the selected filters
            </div>
            <div style={{ "font-size": "0.8rem" }}>
              Try adjusting your status or type selections in the sidebar
            </div>
          </div>
        }
      >
        {/* Horizontal scroll container */}
        <div
          class="sensor-grid-scroll"
          style={{
            flex: 1,
            "min-height": 0,
            "overflow-x": "auto",
            "overflow-y": "hidden",
            padding: "16px 24px 20px",
          }}
        >
          {/* Grid — column-flow so cards fill top→bottom then left→right */}
          <div
            style={{
              display: "grid",
              "grid-template-rows": `repeat(${rows()}, 1fr)`,
              "grid-auto-flow": "column",
              "grid-auto-columns": cardMinWidth(),
              gap: "14px",
              height: "100%",
              width: "max-content",
              "min-width": "100%",
            }}
          >
            <For each={filteredSensors()}>
              {(sensor) => (
                <SensorStatusCard
                  sensor={sensor}
                  compact={props.cardView === "compact"}
                  onSelect={props.onSensorSelect}
                  onDiagnose={(s) => setDiagnoseSensor(s)}
                />
              )}
            </For>
          </div>
        </div>
      </Show>

      {/* ── Diagnostic overlay — rendered in Portal to escape backdrop-filter ancestors ── */}
      <Show when={diagnoseSensor() !== null}>
        <Portal>
          {/* Semi-transparent backdrop */}
          <div
            onClick={() => setDiagnoseSensor(null)}
            style={{
              position: "fixed",
              top: 0,
              left: 0,
              right: 0,
              bottom: 0,
              background: "rgba(0, 0, 0, 0.6)",
              "backdrop-filter": "blur(5px)",
              "-webkit-backdrop-filter": "blur(5px)",
              "z-index": 998,
              animation: "sensorGridBackdropIn 0.2s ease",
            }}
          />
          <SensorHealthDiagnosticCard
            sensor={diagnoseSensor()!}
            onClose={() => setDiagnoseSensor(null)}
          />
        </Portal>
      </Show>

      <style>{`
        .sensor-grid-scroll::-webkit-scrollbar {
          height: 5px;
        }
        .sensor-grid-scroll::-webkit-scrollbar-track {
          background: rgba(255,255,255,0.02);
          border-radius: 3px;
        }
        .sensor-grid-scroll::-webkit-scrollbar-thumb {
          background: rgba(255,255,255,0.1);
          border-radius: 3px;
        }
        .sensor-grid-scroll::-webkit-scrollbar-thumb:hover {
          background: rgba(255,255,255,0.18);
        }
        @keyframes sensorGridBackdropIn {
          from { opacity: 0; }
          to   { opacity: 1; }
        }
      `}</style>
    </div>
  );
}
