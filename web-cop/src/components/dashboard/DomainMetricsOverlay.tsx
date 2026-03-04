// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/DomainMetricsOverlay.tsx
// Floating domain KPI cards for Multi-Domain Dashboard.

import React, { useMemo, useState } from "react";
import { useAlertStore } from "../../stores/alertStore";
import { useSensorHealthStore } from "../../stores/sensorHealthStore";
import { useTrackStore } from "../../stores/trackStore";

const DOMAINS = [
  { key: "AIR",        label: "AIR",     color: "#38BDF8", keywords: ["air"], sensorTypes: ["RADAR", "EW", "ADS-B", "IFF"] },
  { key: "SURFACE",    label: "SURFACE", color: "#60A5FA", keywords: ["surface", "ship", "vessel", "boat"], sensorTypes: ["AIS", "RADAR", "CAMERA"] },
  { key: "SUBSURFACE", label: "SUB",     color: "#0D9488", keywords: ["sub"], sensorTypes: ["SONAR"] },
  { key: "LAND",       label: "LAND",    color: "#A3844B", keywords: ["land", "vehicle", "person"], sensorTypes: ["ISR", "ELINT", "CAMERA"] },
  { key: "SPACE",      label: "SPACE",   color: "#A78BFA", keywords: ["space", "sat"], sensorTypes: ["TELESCOPE"] },
  { key: "CYBER",      label: "CYBER",   color: "#34D399", keywords: ["cyber", "node"], sensorTypes: ["CYBER"] },
] as const;

function getDomain(entityType: string): string {
  const t = entityType.toLowerCase();
  for (const d of DOMAINS) {
    if (d.keywords.some((k) => t.includes(k))) return d.key;
  }
  return "UNKNOWN";
}

export const DomainMetricsOverlay: React.FC = () => {
  const currentTracksMap = useTrackStore((s) => s.tracks);
  const alerts = useAlertStore((s) => s.alerts);
  const sensors = useSensorHealthStore((s) => s.sensors);
  const [collapsed, setCollapsed] = useState(false);

  const metrics = useMemo(() => {
    const map: Record<string, { count: number; hostile: number; obsPerSec: number }> = {};
    DOMAINS.forEach((d) => {
      map[d.key] = { count: 0, hostile: 0, obsPerSec: 0 };
    });

    currentTracksMap.forEach((t) => {
      const domain = getDomain(t.entityType);
      if (map[domain]) {
        map[domain].count++;
        if (t.hostileClass === "HOSTILE") map[domain].hostile++;
      }
    });

    sensors.forEach((s) => {
      const type = s.sensorType.toUpperCase();
      DOMAINS.forEach((d) => {
        if ((d.sensorTypes as readonly string[]).includes(type)) {
          map[d.key].obsPerSec += s.eventsPerSecond;
        }
      });
    });

    return DOMAINS.map((d) => ({ ...d, ...map[d.key] }));
  }, [currentTracksMap, sensors]);

  const criticalAlerts = useMemo(
    () =>
      Array.from(alerts.values()).filter((a) => a.severity === "CRITICAL").length,
    [alerts]
  );

  const totalTracks = useMemo(() => currentTracksMap.size, [currentTracksMap]);
  const totalHostile = useMemo(
    () =>
      Array.from(currentTracksMap.values()).filter((t) => t.hostileClass === "HOSTILE")
        .length,
    [currentTracksMap]
  );
  const totalObsPerSec = useMemo(
    () =>
      Array.from(sensors.values()).reduce((sum, s) => sum + s.eventsPerSecond, 0),
    [sensors]
  );

  return (
    <div
      data-testid="domain-metrics-overlay"
      style={{
        position: "absolute",
        top: "12px",
        left: "12px",
        zIndex: 5,
        display: "flex",
        flexDirection: "column",
        gap: "6px",
        pointerEvents: "auto",
        userSelect: "none",
      }}
    >
      {/* Header summary row */}
      <div
        style={{
          backgroundColor: "rgba(15, 23, 42, 0.85)",
          backdropFilter: "blur(8px)",
          border: "1px solid #334155",
          borderRadius: "8px",
          padding: "8px 12px",
          display: "flex",
          alignItems: "center",
          gap: "12px",
          cursor: "pointer",
        }}
        onClick={() => setCollapsed((c) => !c)}
      >
        <span style={{ fontSize: "0.7rem", fontWeight: "bold", color: "#F59E0B" }}>
          MULTI-DOMAIN
        </span>
        <Pill label="TRACKS" value={totalTracks} color="#60A5FA" />
        <Pill label="HOSTILE" value={totalHostile} color="#EF4444" />
        <Pill label="OBS/S" value={Math.round(totalObsPerSec)} color="#34D399" />
        {criticalAlerts > 0 && (
          <Pill label="CRIT" value={criticalAlerts} color="#DC2626" pulse />
        )}
        <span
          style={{ marginLeft: "auto", fontSize: "0.7rem", color: "#64748B" }}
          title={collapsed ? "Expand" : "Collapse"}
        >
          {collapsed ? "▶" : "▼"}
        </span>
      </div>

      {/* Domain cards */}
      {!collapsed &&
        metrics
          .filter((m) => m.count > 0 || m.obsPerSec > 0)
          .map((m) => (
            <div
              key={m.key}
              style={{
                backgroundColor: "rgba(15, 23, 42, 0.82)",
                backdropFilter: "blur(8px)",
                border: `1px solid ${m.color}33`,
                borderLeft: `3px solid ${m.color}`,
                borderRadius: "6px",
                padding: "8px 12px",
                display: "flex",
                alignItems: "center",
                gap: "16px",
                minWidth: "220px",
              }}
            >
              <div style={{ width: "52px", display: "flex", flexDirection: "column" }}>
                <span style={{ fontSize: "0.65rem", fontWeight: "bold", color: m.color }}>
                  {m.label}
                </span>
                {m.hostile > 0 && (
                  <span style={{ fontSize: "0.6rem", color: "#EF4444", fontWeight: "bold" }}>
                    {m.hostile} HST
                  </span>
                )}
              </div>
              <div style={{ display: "flex", flexDirection: "column", minWidth: "40px" }}>
                <span style={{ fontSize: "0.6rem", color: "#94A3B8" }}>TRACKS</span>
                <span style={{ fontSize: "1.1rem", fontFamily: "monospace", color: "#F1F5F9" }}>
                  {m.count}
                </span>
              </div>
              <div style={{ display: "flex", flexDirection: "column", minWidth: "40px", marginLeft: "auto", textAlign: "right" }}>
                <span style={{ fontSize: "0.6rem", color: "#94A3B8" }}>OBS/S</span>
                <span style={{ fontSize: "1.1rem", fontFamily: "monospace", color: "#34D399" }}>
                  {m.obsPerSec >= 1000 ? (m.obsPerSec / 1000).toFixed(1) + 'k' : Math.round(m.obsPerSec)}
                </span>
              </div>
            </div>
          ))}
    </div>
  );
};

const Pill: React.FC<{
  label: string;
  value: number;
  color: string;
  pulse?: boolean;
}> = ({ label, value, color, pulse }) => (
  <div style={{ display: "flex", alignItems: "center", gap: "4px" }}>
    <span
      style={{
        fontSize: "0.6rem",
        color: "#64748B",
        textTransform: "uppercase",
      }}
    >
      {label}
    </span>
    <span
      style={{
        fontSize: "0.75rem",
        fontWeight: "bold",
        color,
        animation: pulse ? "pulse 1.2s infinite" : undefined,
      }}
    >
      {value}
    </span>
  </div>
);
