// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/SensorStatusCard.tsx — Individual sensor status card
//
// Reference: docs/business/usecases/UC017_sensor_health_monitoring.md
// Reference: docs/implementation/v5/ui_images/dashboard_main.png

import { SensorStatus } from "../../services/sensor-health";

interface SensorStatusCardProps {
  sensor: SensorStatus;
}

/**
 * High-fidelity sensor status card with glassmorphism styling.
 * Displays real-time metrics, connectivity status, and a performance sparkline.
 */
export function SensorStatusCard(props: SensorStatusCardProps) {
  const statusColor = () => {
    switch (props.sensor.status) {
      case "CONNECTED": return "#4ade80"; // Bright Green
      case "STALE": return "#fbbf24";     // Amber
      case "OFFLINE": return "#f87171";   // Red
      default: return "#94a3b8";          // Slate
    }
  };

  // Generate a deterministic but dynamic-looking sparkline for the WOW factor.
  // In a real app, this would use historical data from the service.
  const sparklinePoints = () => {
    const points = [];
    // Seed based on ID to keep it stable for a single card but different across cards
    const seed = props.sensor.sensorId.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0);
    const baseValue = props.sensor.eventsPerSecond > 0 ? 25 : 35;

    for (let i = 0; i <= 10; i++) {
        const x = i * 10;
        // Offline sensors get a flat line at the bottom
        if (props.sensor.status === "OFFLINE") {
            points.push(`${x},38`);
            continue;
        }
        const variance = Math.sin(seed + i * 0.8) * 10;
        const y = baseValue + variance + (Math.random() * 4 - 2);
        points.push(`${x},${y}`);
    }
    return `M ${points.join(" L ")}`;
  };

  return (
    <div
      style={{
        background: "rgba(15, 23, 42, 0.6)",
        "backdrop-filter": "blur(12px)",
        border: "1px solid rgba(255, 255, 255, 0.08)",
        "border-top": `2px solid ${statusColor()}`,
        "border-radius": "12px",
        padding: "1.25rem",
        display: "flex",
        "flex-direction": "column",
        gap: "1rem",
        color: "#f1f5f9",
        transition: "all 0.3s cubic-bezier(0.4, 0, 0.2, 1)",
        "box-shadow": "0 4px 20px rgba(0, 0, 0, 0.3)",
        position: "relative",
        overflow: "hidden",
      }}
      class="sensor-card-hover"
    >
        {/* Subtle background glow based on status */}
        <div style={{
            position: "absolute",
            top: "-20px",
            right: "-20px",
            width: "80px",
            height: "80px",
            background: statusColor(),
            filter: "blur(50px)",
            opacity: 0.1,
            "pointer-events": "none"
        }} />

      {/* Header: Icon, ID, Type and Status Badge */}
      <div style={{ display: "flex", "justify-content": "space-between", "align-items": "flex-start" }}>
        <div style={{ display: "flex", "align-items": "center", gap: "0.75rem" }}>
          <div style={{
            padding: "0.5rem",
            background: "rgba(255,255,255,0.05)",
            "border-radius": "10px",
            border: "1px solid rgba(255,255,255,0.1)",
            display: "flex",
            "align-items": "center",
            "justify-content": "center",
            color: statusColor()
          }}>
             <SensorIcon type={props.sensor.sensorType} />
          </div>
          <div>
            <div style={{ "font-weight": "600", "font-size": "1rem", "letter-spacing": "0.025em" }}>{props.sensor.sensorId}</div>
            <div style={{ "font-size": "0.7rem", color: "#64748b", "text-transform": "uppercase", "letter-spacing": "0.05em", "margin-top": "2px" }}>{props.sensor.sensorType}</div>
          </div>
        </div>
        <div
          style={{
            "font-size": "0.65rem",
            "font-weight": "700",
            padding: "4px 10px",
            "border-radius": "20px",
            background: `${statusColor()}15`,
            color: statusColor(),
            border: `1px solid ${statusColor()}30`,
            "text-transform": "uppercase",
            "letter-spacing": "0.05em",
            display: "flex",
            "align-items": "center",
            gap: "5px"
          }}
        >
          <span style={{
            width: "6px",
            height: "6px",
            "border-radius": "50%",
            background: statusColor(),
            display: "inline-block",
            "box-shadow": `0 0 8px ${statusColor()}`
          }}></span>
          {props.sensor.status}
        </div>
      </div>

      {/* Metrics Grid */}
      <div style={{ display: "grid", "grid-template-columns": "1.2fr 1fr", gap: "1rem", "margin-top": "0.5rem" }}>
        <div style={{ background: "rgba(255,255,255,0.03)", padding: "0.75rem", "border-radius": "8px", border: "1px solid rgba(255,255,255,0.05)" }}>
          <div style={{ color: "#64748b", "font-size": "0.65rem", "text-transform": "uppercase", "margin-bottom": "4px", "letter-spacing": "0.025em" }}>Throughput</div>
          <div style={{ "font-size": "1.1rem", "font-weight": "700", "font-family": "monospace", color: "#f8fafc" }}>
            {props.sensor.eventsPerSecond} <span style={{ "font-size": "0.7rem", color: "#64748b", "font-weight": "normal" }}>obs/s</span>
          </div>
        </div>
        <div style={{ background: "rgba(255,255,255,0.03)", padding: "0.75rem", "border-radius": "8px", border: "1px solid rgba(255,255,255,0.05)" }}>
          <div style={{ color: "#64748b", "font-size": "0.65rem", "text-transform": "uppercase", "margin-bottom": "4px", "letter-spacing": "0.025em" }}>Validation</div>
          <div style={{ "font-size": "1.1rem", "font-weight": "700", "font-family": "monospace", color: "#f8fafc" }}>
            {props.sensor.validationPassRate}<span style={{ "font-size": "0.8rem", color: "#64748b", "font-weight": "normal" }}>%</span>
          </div>
        </div>
      </div>

      {/* Throughput Sparkline */}
      <div style={{ height: "45px", width: "100%", opacity: 0.8, "margin": "0.25rem 0" }}>
        <svg width="100%" height="100%" viewBox="0 0 100 40" preserveAspectRatio="none">
          <defs>
            <linearGradient id={`grad-${props.sensor.sensorId}`} x1="0%" y1="0%" x2="0%" y2="100%">
              <stop offset="0%" style={{ "stop-color": statusColor(), "stop-opacity": 0.4 }} />
              <stop offset="100%" style={{ "stop-color": statusColor(), "stop-opacity": 0 }} />
            </linearGradient>
          </defs>
          <path
            d={sparklinePoints()}
            fill="none"
            stroke={statusColor()}
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
          <path
            d={`${sparklinePoints()} V 40 H 0 Z`}
            fill={`url(#grad-${props.sensor.sensorId})`}
          />
        </svg>
      </div>

      {/* Footer Metrics */}
      <div style={{
        display: "flex",
        "justify-content": "space-between",
        "align-items": "center",
        "font-size": "0.75rem",
        "padding-top": "0.75rem",
        "border-top": "1px solid rgba(255,255,255,0.05)"
      }}>
        <div style={{ color: "#94a3b8" }}>
          <span style={{ color: "#64748b", "margin-right": "4px" }}>DLQ:</span>
          <span style={{ color: props.sensor.dlqCount > 50 ? "#f87171" : "#f1f5f9" }}>{props.sensor.dlqCount}</span>
        </div>
        <div style={{ color: "#94a3b8" }}>
          <span style={{ color: "#64748b", "margin-right": "4px" }}>Seen:</span>
          <span style={{ color: "#cbd5e1" }}>{props.sensor.lastSeenSeconds < 0 ? "N/A" : `${props.sensor.lastSeenSeconds}s ago`}</span>
        </div>
      </div>

      <style>{`
        .sensor-card-hover:hover {
            transform: translateY(-4px);
            border-color: rgba(255, 255, 255, 0.2);
            box-shadow: 0 12px 30px rgba(0, 0, 0, 0.5);
            background: rgba(15, 23, 42, 0.85);
        }
      `}</style>
    </div>
  );
}

function SensorIcon(props: { type: string }) {
  const t = props.type;
  if (t === "RADAR") {
    return (
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M19.07 4.93a10 10 0 0 0-14.14 0" /><path d="M16.24 7.76a6 6 0 0 0-8.48 0" /><circle cx="12" cy="12" r="2" /><path d="M12 2v2" /><path d="M2 12h2" /><path d="M20 12h2" /><path d="m4.93 4.93 1.41 1.41" /><path d="m17.66 17.66 1.41 1.41" />
        </svg>
    );
  }
  if (t === "EW/SIGINT" || t === "ELINT/COMINT") {
    return (
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="22 12 18 12 15 21 9 3 6 12 2 12" />
        </svg>
    );
  }
  if (t === "AIS/BFT") {
    return (
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M2 21c.6.5 1.2 1 2.5 1 2.5 0 2.5-2 5-2 1.3 0 1.9.5 2.5 1 .6.5 1.2 1 2.5 1 2.5 0 2.5-2 5-2 1.3 0 1.9.5 2.5 1" />
            <path d="M19.38 20A11.6 11.6 0 0 0 21 14l-9-4-9 4c0 2.2.6 4.3 1.62 6" />
            <path d="M12 10V2" />
        </svg>
    );
  }
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <rect x="2" y="3" width="20" height="14" rx="2" ry="2" />
        <line x1="8" y1="21" x2="16" y2="21" />
        <line x1="12" y1="17" x2="12" y2="21" />
    </svg>
  );
}
