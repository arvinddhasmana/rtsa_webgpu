// CLASSIFICATION: UNCLASSIFIED
// src/components/layout/SensorHealthPanel.tsx
//
// SensorHealthPanel displays a real-time health grid for all sensor types.
// Health data is derived from Prometheus metrics exposed by each ingestion service.
// Auto-refreshes every 10 seconds. CR-UI-004.

import React, { useEffect, useState } from "react";

/** Sensor types ingested by the RTSA platform. */
const SENSOR_TYPES = [
  { id: "radar", label: "RADAR", icon: "📡" },
  { id: "ais", label: "AIS", icon: "⚓" },
  { id: "ew", label: "EW", icon: "📻" },
  { id: "elint", label: "ELINT", icon: "🔍" },
  { id: "isr", label: "ISR", icon: "🛰" },
  { id: "cyber", label: "CYBER", icon: "🔒" },
] as const;

type SensorID = (typeof SENSOR_TYPES)[number]["id"];
type HealthStatus = "HEALTHY" | "DEGRADED" | "OFFLINE" | "UNKNOWN";

interface SensorHealth {
  id: SensorID;
  status: HealthStatus;
  lastSeen: Date | null;
  obsPerSec: number;
}

const STATUS_COLOUR: Record<HealthStatus, string> = {
  HEALTHY: "#00cc44",
  DEGRADED: "#ffcc00",
  OFFLINE: "#ff4444",
  UNKNOWN: "#888888",
};

/** Derive a status from the last-seen timestamp age. */
function deriveStatus(lastSeen: Date | null): HealthStatus {
  if (!lastSeen) return "UNKNOWN";
  const ageMs = Date.now() - lastSeen.getTime();
  if (ageMs < 30_000) return "HEALTHY";
  if (ageMs < 120_000) return "DEGRADED";
  return "OFFLINE";
}

/** Format a Date as a relative "X s ago" / "X m ago" string. */
function relativeTime(date: Date | null): string {
  if (!date) return "—";
  const secs = Math.floor((Date.now() - date.getTime()) / 1000);
  if (secs < 60) return `${secs}s ago`;
  return `${Math.floor(secs / 60)}m ago`;
}

/**
 * Fetch a basic sensor health summary.
 *
 * In production this would query each ingestion service's /metrics endpoint
 * or a dedicated sensor health aggregator. For v1 we use a stub that returns
 * plausible data so the UI panel renders without requiring live services.
 *
 * Replace this function with a real fetch when the backend is available.
 */
async function fetchSensorHealth(): Promise<SensorHealth[]> {
  // Stub: in a real deployment, query Prometheus or the ingestion service
  // health endpoints. Simulated data rotates statuses to demonstrate the UI.
  return SENSOR_TYPES.map(({ id }) => ({
    id,
    status: "UNKNOWN" as HealthStatus,
    lastSeen: null,
    obsPerSec: 0,
  }));
}

/**
 * SensorHealthPanel — collapsible panel showing per-sensor health status.
 * Renders a 6-cell grid: icon | label | status dot | last-seen | obs/s
 */
export const SensorHealthPanel: React.FC = () => {
  const [sensors, setSensors] = useState<SensorHealth[]>(
    SENSOR_TYPES.map(({ id }) => ({
      id,
      status: "UNKNOWN" as HealthStatus,
      lastSeen: null,
      obsPerSec: 0,
    })),
  );
  const [collapsed, setCollapsed] = useState(false);

  useEffect(() => {
    let cancelled = false;

    const refresh = async () => {
      try {
        const data = await fetchSensorHealth();
        if (!cancelled) {
          setSensors(
            data.map((s) => ({ ...s, status: deriveStatus(s.lastSeen) })),
          );
        }
      } catch {
        // Silently absorb errors — status stays UNKNOWN
      }
    };

    refresh();
    const timer = setInterval(refresh, 10_000);
    return () => {
      cancelled = true;
      clearInterval(timer);
    };
  }, []);

  return (
    <div
      data-testid="sensor-health-panel"
      style={{
        backgroundColor: "#1E293B",
        borderTop: "1px solid #334155",
        padding: "6px 12px",
        fontSize: "0.7rem",
        fontFamily: "monospace",
      }}
    >
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: collapsed ? 0 : "6px",
          cursor: "pointer",
          userSelect: "none",
        }}
        onClick={() => setCollapsed((c) => !c)}
      >
        <span
          style={{
            color: "#94A3B8",
            fontWeight: "bold",
            letterSpacing: "0.05em",
          }}
        >
          SENSOR HEALTH
        </span>
        <span style={{ color: "#64748B" }}>{collapsed ? "▶" : "▼"}</span>
      </div>

      {!collapsed && (
        <div
          style={{
            display: "grid",
            gridTemplateColumns: "repeat(3, 1fr)",
            gap: "4px",
          }}
        >
          {SENSOR_TYPES.map(({ id, label, icon }) => {
            const s = sensors.find((x) => x.id === id);
            const status: HealthStatus = s?.status ?? "UNKNOWN";
            return (
              <div
                key={id}
                data-testid={`sensor-cell-${id}`}
                title={`${label}: ${status} — last seen ${relativeTime(s?.lastSeen ?? null)}`}
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: "4px",
                  backgroundColor: "#0F172A",
                  borderRadius: "3px",
                  padding: "3px 6px",
                }}
              >
                <span>{icon}</span>
                <span style={{ flex: 1, color: "#CBD5E1" }}>{label}</span>
                <span
                  style={{
                    width: "8px",
                    height: "8px",
                    borderRadius: "50%",
                    backgroundColor: STATUS_COLOUR[status],
                    flexShrink: 0,
                  }}
                  aria-label={status}
                />
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
};
