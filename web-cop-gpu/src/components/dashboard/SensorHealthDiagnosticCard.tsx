// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/SensorHealthDiagnosticCard.tsx — Glassmorphism diagnostic overlay card
//
// Appears as a floating card overlay when a sensor card is hovered or focused,
// providing a compact health summary before navigating to the full L2 diagnostic view.
//
// Reference: docs/implementation/v5/sensordashboard_three_level_plan.md §B3
// Reference: docs/business/usecases/UC017_sensor_health_monitoring.md

import { createResource, createSignal, JSX, Show } from "solid-js";
import { fetchSensorDiagnostic, SensorStatus } from "../../services/sensor-health";
import { setSelectedSensor } from "../../signals/sensor-filters";
import { statusColor } from "./dashboard-utils";

export interface SensorHealthDiagnosticCardProps {
  sensor: SensorStatus;
  onClose?: () => void;
}

function healthLabel(score: number): string {
  if (score >= 85) return "NOMINAL";
  if (score >= 60) return "DEGRADED";
  return "CRITICAL";
}

function healthColor(score: number): string {
  if (score >= 85) return "#4ade80";
  if (score >= 60) return "#fbbf24";
  return "#f87171";
}

/**
 * Compact glassmorphism overlay card showing a sensor health summary.
 * Appears on hover/focus over a SensorStatusCard in the L1 dashboard.
 * Clicking "View Details" navigates to the full L2 SensorDiagnosticView.
 *
 * Never destructure props — breaks SolidJS reactivity.
 */
export function SensorHealthDiagnosticCard(props: SensorHealthDiagnosticCardProps): JSX.Element {
  const [diag] = createResource(() => props.sensor, fetchSensorDiagnostic);
  const [btnHovered, setBtnHovered] = createSignal(false);

  const sColor = () => statusColor(props.sensor.status);

  function validationColor(rate: number): string {
    if (rate >= 95) return "#4ade80";
    if (rate >= 80) return "#fbbf24";
    return "#f87171";
  }

  function dlqColor(count: number): string {
    if (count > 50) return "#f87171";
    if (count > 10) return "#fbbf24";
    return "#4ade80";
  }

  const MetricRow = (metricProps: { label: string; value: JSX.Element }): JSX.Element => (
    <div
      style={{
        display: "flex",
        "justify-content": "space-between",
        "align-items": "center",
        padding: "4px 0",
        "border-bottom": "1px solid rgba(255,255,255,0.04)",
      }}
    >
      <span
        style={{
          "font-size": "0.62rem",
          color: "#64748b",
          "text-transform": "uppercase",
          "letter-spacing": "0.06em",
          "font-family": "monospace",
        }}
      >
        {metricProps.label}
      </span>
      <span style={{ "font-size": "0.7rem", "font-family": "monospace", color: "#e2e8f0" }}>
        {metricProps.value}
      </span>
    </div>
  );

  return (
    <div
      data-testid="sensor-health-diagnostic-card"
      style={{
        position: "absolute",
        top: "50%",
        left: "50%",
        transform: "translate(-50%, -50%)",
        "z-index": 50,
        width: "clamp(260px, 22vw, 340px)",
        background: "rgba(10, 15, 28, 0.85)",
        "backdrop-filter": "blur(30px)",
        "-webkit-backdrop-filter": "blur(30px)",
        border: `1px solid ${sColor()}40`,
        "border-top": `2px solid ${sColor()}`,
        "border-radius": "14px",
        padding: "16px",
        "box-shadow": `0 20px 60px rgba(0,0,0,0.6), 0 0 30px ${sColor()}15`,
        display: "flex",
        "flex-direction": "column",
        gap: "12px",
        color: "#f1f5f9",
        "font-family": "monospace",
        animation: "diagnosticCardIn 0.2s cubic-bezier(0.4, 0, 0.2, 1)",
      }}
    >
      <style>{`
        @keyframes diagnosticCardIn {
          from { opacity: 0; transform: translate(-50%, calc(-50% + 8px)); }
          to   { opacity: 1; transform: translate(-50%, -50%); }
        }
      `}</style>

      {/* ── Header ── */}
      <div style={{ display: "flex", "align-items": "flex-start", "justify-content": "space-between", gap: "8px" }}>
        <div style={{ display: "flex", "flex-direction": "column", gap: "2px", "min-width": 0 }}>
          <div
            style={{
              "font-size": "0.85rem",
              "font-weight": "700",
              overflow: "hidden",
              "text-overflow": "ellipsis",
              "white-space": "nowrap",
              color: "#f8fafc",
            }}
          >
            {props.sensor.sensorId}
          </div>
          <div style={{ display: "flex", gap: "6px", "align-items": "center" }}>
            <span style={{ "font-size": "0.6rem", color: "#64748b" }}>
              {props.sensor.sensorType}
            </span>
            <span
              style={{
                "font-size": "0.6rem",
                "font-weight": "700",
                color: sColor(),
                "text-transform": "uppercase",
              }}
            >
              {props.sensor.status}
            </span>
          </div>
        </div>
        {/* Close button */}
        <button
          data-testid="diagnostic-card-close-btn"
          onClick={() => props.onClose?.()}
          style={{
            background: "rgba(255,255,255,0.06)",
            border: "1px solid rgba(255,255,255,0.1)",
            color: "#64748b",
            cursor: "pointer",
            "border-radius": "6px",
            width: "22px",
            height: "22px",
            display: "flex",
            "align-items": "center",
            "justify-content": "center",
            "font-size": "0.75rem",
            "flex-shrink": 0,
          }}
        >
          ✕
        </button>
      </div>

      {/* ── Health Score ── */}
      <Show
        when={diag()}
        fallback={
          <div style={{ color: "#334155", "font-size": "0.65rem", "text-align": "center", padding: "4px 0" }}>
            Loading diagnostics…
          </div>
        }
      >
        {(d) => (
          <div
            style={{
              display: "flex",
              "align-items": "center",
              gap: "12px",
              background: "rgba(255,255,255,0.03)",
              "border-radius": "10px",
              padding: "10px 12px",
              border: "1px solid rgba(255,255,255,0.05)",
            }}
          >
            <div
              style={{
                "font-size": "1.8rem",
                "font-weight": "800",
                color: healthColor(d().healthScore),
                "line-height": 1,
              }}
            >
              {d().healthScore}
            </div>
            <div style={{ display: "flex", "flex-direction": "column", gap: "2px" }}>
              <div style={{ "font-size": "0.58rem", color: "#64748b", "text-transform": "uppercase", "letter-spacing": "0.08em" }}>
                Health Score
              </div>
              <div
                style={{
                  "font-size": "0.65rem",
                  "font-weight": "700",
                  color: healthColor(d().healthScore),
                  "text-transform": "uppercase",
                  "letter-spacing": "0.05em",
                }}
              >
                {healthLabel(d().healthScore)}
              </div>
            </div>
          </div>
        )}
      </Show>

      {/* ── Key Metrics ── */}
      <div style={{ display: "flex", "flex-direction": "column", gap: "0" }}>
        <MetricRow
          label="Throughput"
          value={
            <span>
              {props.sensor.eventsPerSecond}{" "}
              <span style={{ color: "#475569", "font-size": "0.58rem" }}>obs/s</span>
            </span>
          }
        />
        <MetricRow
          label="Validation"
          value={
            <span style={{ color: validationColor(props.sensor.validationPassRate) }}>
              {props.sensor.status === "OFFLINE" ? "N/A" : `${props.sensor.validationPassRate}%`}
            </span>
          }
        />
        <MetricRow
          label="DLQ Count"
          value={
            <span style={{ color: dlqColor(props.sensor.dlqCount) }}>
              {props.sensor.dlqCount}
            </span>
          }
        />
        <Show when={diag()}>
          {(d) => (
            <MetricRow
              label="Uptime"
              value={
                <span style={{ color: d().connectionUptimePct >= 95 ? "#4ade80" : "#fbbf24" }}>
                  {d().connectionUptimePct.toFixed(1)}%
                </span>
              }
            />
          )}
        </Show>
        <Show when={diag()}>
          {(d) => (
            <MetricRow
              label="Avg Latency"
              value={
                <span>
                  {d().avgLatencyMs}{" "}
                  <span style={{ color: "#475569", "font-size": "0.58rem" }}>ms</span>
                </span>
              }
            />
          )}
        </Show>
      </div>

      {/* ── View Full Details Button ── */}
      <button
        data-testid="diagnostic-card-view-full-btn"
        onClick={() => setSelectedSensor(props.sensor)}
        onMouseEnter={() => setBtnHovered(true)}
        onMouseLeave={() => setBtnHovered(false)}
        style={{
          background: btnHovered()
            ? `${sColor()}25`
            : `linear-gradient(135deg, ${sColor()}20, ${sColor()}10)`,
          border: `1px solid ${btnHovered() ? sColor() + "70" : sColor() + "40"}`,
          "box-shadow": btnHovered() ? `0 0 12px ${sColor()}20` : "none",
          color: sColor(),
          cursor: "pointer",
          "border-radius": "8px",
          padding: "8px 12px",
          "font-size": "0.65rem",
          "font-weight": "700",
          "text-transform": "uppercase",
          "letter-spacing": "0.08em",
          "font-family": "monospace",
          transition: "all 0.2s",
          width: "100%",
          display: "flex",
          "align-items": "center",
          "justify-content": "center",
          gap: "6px",
        }}
      >
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
          <polyline points="9 18 15 12 9 6" />
        </svg>
        View Full Diagnostics
      </button>
    </div>
  );
}
