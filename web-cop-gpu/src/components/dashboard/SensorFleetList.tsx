// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/SensorFleetList.tsx — Scrollable sensor fleet list
//
// Reference: docs/implementation/v5/sensordashboard_three_level_plan.md §B2

import { For, JSX, Show } from "solid-js";
import { SensorStatus } from "../../services/sensor-health";
import { statusColor } from "./dashboard-utils";

export interface SensorFleetListProps {
  sensors: SensorStatus[];
  selectedSensorId?: string;
  onSensorSelect?: (sensor: SensorStatus) => void;
  onSensorHover?: (sensor: SensorStatus | null) => void;
  compact?: boolean;
  maxHeight?: string;
}

function TypeIcon(iconProps: { type: string; size?: number }): JSX.Element {
  const size = iconProps.size ?? 14;
  const t = iconProps.type;

  if (t === "RADAR") {
    return (
      <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M19.07 4.93a10 10 0 0 0-14.14 0" />
        <path d="M16.24 7.76a6 6 0 0 0-8.48 0" />
        <circle cx="12" cy="12" r="2" />
      </svg>
    );
  }
  if (t === "EW/SIGINT" || t === "ELINT/COMINT") {
    return (
      <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <polyline points="22 12 18 12 15 21 9 3 6 12 2 12" />
      </svg>
    );
  }
  if (t === "AIS/BFT") {
    return (
      <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" />
        <polyline points="9 22 9 12 15 12 15 22" />
      </svg>
    );
  }
  if (t === "ISR") {
    return (
      <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="11" cy="11" r="8" />
        <line x1="21" y1="21" x2="16.65" y2="16.65" />
      </svg>
    );
  }
  if (t === "CYBER") {
    return (
      <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <rect x="2" y="3" width="20" height="14" rx="2" ry="2" />
        <line x1="8" y1="21" x2="16" y2="21" />
        <line x1="12" y1="17" x2="12" y2="21" />
      </svg>
    );
  }
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
      <circle cx="12" cy="12" r="10" />
    </svg>
  );
}

function StatusBadge(badgeProps: { status: string; compact?: boolean }): JSX.Element {
  const color = statusColor(badgeProps.status);
  if (badgeProps.compact) {
    return (
      <div style={{
        width: "8px",
        height: "8px",
        "border-radius": "50%",
        background: color,
        "flex-shrink": 0,
        "box-shadow": badgeProps.status === "CONNECTED" ? `0 0 6px ${color}` : "none",
      }} />
    );
  }
  return (
    <div style={{
      background: `${color}15`,
      color,
      border: `1px solid ${color}30`,
      padding: "1px 6px",
      "border-radius": "10px",
      "font-size": "0.58rem",
      "text-transform": "uppercase",
      "letter-spacing": "0.06em",
      "white-space": "nowrap",
    }}>
      {badgeProps.status}
    </div>
  );
}

/** Scrollable sensor fleet list — B2 */
export function SensorFleetList(props: SensorFleetListProps): JSX.Element {
  return (
    <div
      data-testid="sensor-fleet-list"
      style={{
        display: "flex",
        "flex-direction": "column",
        gap: "2px",
        "overflow-y": "auto",
        "max-height": props.maxHeight ?? "none",
      }}
    >
      <For each={props.sensors}>
        {(sensor) => {
          const isSelected = () => props.selectedSensorId === sensor.sensorId;
          const dotColor = statusColor(sensor.status);

          return (
            <div
              data-testid={`fleet-row-${sensor.sensorId}`}
              onClick={() => props.onSensorSelect?.(sensor)}
              onMouseEnter={() => props.onSensorHover?.(sensor)}
              onMouseLeave={() => props.onSensorHover?.(null)}
              style={{
                display: "flex",
                "align-items": "center",
                gap: props.compact ? "6px" : "8px",
                padding: props.compact ? "4px 6px" : "6px 10px",
                "border-radius": "6px",
                cursor: "pointer",
                background: isSelected()
                  ? "rgba(59, 130, 246, 0.12)"
                  : "transparent",
                border: isSelected()
                  ? "1px solid rgba(59, 130, 246, 0.25)"
                  : "1px solid transparent",
                transition: "background 0.1s ease, border 0.1s ease",
              }}
            >
              {/* Status dot */}
              <div style={{
                width: "8px",
                height: "8px",
                "border-radius": "50%",
                background: dotColor,
                "flex-shrink": 0,
                "box-shadow": sensor.status === "CONNECTED" ? `0 0 5px ${dotColor}` : "none",
              }} />

              {/* Type icon */}
              <div style={{ color: "#64748b", "flex-shrink": 0, "line-height": 0 }}>
                <TypeIcon type={sensor.sensorType} />
              </div>

              <Show when={!props.compact}>
                {/* Sensor ID */}
                <div style={{
                  flex: 1,
                  "font-size": "0.72rem",
                  color: isSelected() ? "#e2e8f0" : "#cbd5e1",
                  "font-family": "monospace",
                  overflow: "hidden",
                  "text-overflow": "ellipsis",
                  "white-space": "nowrap",
                }}>
                  {sensor.sensorId}
                </div>

                {/* Status badge */}
                <StatusBadge status={sensor.status} />

                {/* Obs rate */}
                <div style={{
                  "font-size": "0.65rem",
                  color: "#64748b",
                  "font-family": "monospace",
                  "min-width": "45px",
                  "text-align": "right",
                  "white-space": "nowrap",
                }}>
                  {sensor.eventsPerSecond.toFixed(1)} /s
                </div>
              </Show>

              <Show when={props.compact}>
                {/* Compact mode: only ID abbreviated */}
                <div style={{
                  "font-size": "0.65rem",
                  color: "#94a3b8",
                  "font-family": "monospace",
                  overflow: "hidden",
                  "text-overflow": "ellipsis",
                  "white-space": "nowrap",
                  "max-width": "80px",
                }}>
                  {sensor.sensorId}
                </div>
              </Show>
            </div>
          );
        }}
      </For>
    </div>
  );
}
