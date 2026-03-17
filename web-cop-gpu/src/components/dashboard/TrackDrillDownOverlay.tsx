// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/TrackDrillDownOverlay.tsx

import { Component, Show, createEffect } from "solid-js";
import { TrackDetail } from "../../signals/track";

interface TrackDrillDownOverlayProps {
  track: TrackDetail | null;
  onClose: () => void;
}

export const TrackDrillDownOverlay: Component<TrackDrillDownOverlayProps> = (props) => {
  createEffect(() => console.log("[TrackDrillDownOverlay] Track changed:", props.track));
  return (
    <Show when={props.track}>
      {(t) => (
        <div
          style={{
            position: "absolute",
            bottom: "2rem",
            left: "50%",
            transform: "translateX(-50%)",
            width: "600px",
            background: "rgba(15, 23, 42, 0.7)",
            "backdrop-filter": "blur(16px)",
            border: "1px solid rgba(255, 255, 255, 0.1)",
            "border-radius": "12px",
            "box-shadow": "0 20px 50px rgba(0, 0, 0, 0.8)",
            padding: "24px",
            color: "#f8fafc",
            "font-family": "Inter, sans-serif",
            "z-index": "100",
          }}
        >
          {/* Header */}
          <div style={{ display: "flex", "justify-content": "space-between", "align-items": "start", "margin-bottom": "20px" }}>
            <div>
              <div style={{ "font-size": "0.75rem", "text-transform": "uppercase", "letter-spacing": "0.1em", color: "#94a3b8" }}>
                Fused Track Data
              </div>
              <h2 style={{ margin: "4px 0 0 0", "font-size": "1.5rem", "font-weight": "700" }}>
                TRACK ID: {t().trackId}
              </h2>
            </div>
            <button
              onClick={props.onClose}
              style={{
                background: "transparent",
                border: "none",
                color: "#94a3b8",
                cursor: "pointer",
                "font-size": "1.5rem",
              }}
            >
              ×
            </button>
          </div>

          {/* Main Content Grid */}
          <div style={{ display: "grid", "grid-template-columns": "1fr 1fr", gap: "24px" }}>
            {/* Left Column: Classification & Confidence */}
            <div>
              <div style={{ "margin-bottom": "16px" }}>
                <div style={{ "font-size": "0.75rem", color: "#94a3b8", "margin-bottom": "4px" }}>Classification</div>
                <div style={{ "font-size": "1.1rem", "font-weight": "600", color: "#60a5fa" }}>
                  {t().classification}
                </div>
              </div>

              <div style={{ "margin-bottom": "16px" }}>
                <div style={{ "font-size": "0.75rem", color: "#94a3b8", "margin-bottom": "4px" }}>Confidence</div>
                <div style={{ display: "flex", "align-items": "center", gap: "10px" }}>
                  <div style={{ flex: 1, height: "8px", background: "rgba(255,255,255,0.1)", "border-radius": "4px", overflow: "hidden" }}>
                    <div style={{
                      width: `${t().confidenceScore}%`,
                      height: "100%",
                      background: t().confidenceScore > 80 ? "#22c55e" : t().confidenceScore > 50 ? "#eab308" : "#ef4444",
                    }} />
                  </div>
                  <span style={{ "font-family": "monospace", "font-weight": "bold" }}>{t().confidenceScore}%</span>
                </div>
              </div>

              <div>
                <div style={{ "font-size": "0.75rem", color: "#94a3b8", "margin-bottom": "8px" }}>Source Pedigree</div>
                <div style={{ display: "flex", gap: "8px" }}>
                  {/* Mock Pedigree Nodes */}
                  <PedigreeNode label="RADAR" time="14:05" active />
                  <PedigreeLine />
                  <PedigreeNode label="AIS" time="14:18" active />
                  <PedigreeLine />
                  <PedigreeNode label="SIGINT" time="14:27" alert />
                </div>
              </div>
            </div>

            {/* Right Column: Kinematics */}
            <div style={{ background: "rgba(0,0,0,0.2)", padding: "16px", "border-radius": "8px", border: "1px solid rgba(255,255,255,0.05)" }}>
              <div style={{ "font-size": "0.85rem", "font-weight": "600", "margin-bottom": "12px", "border-bottom": "1px solid rgba(255,255,255,0.1)", "padding-bottom": "4px" }}>
                KINEMATICS
              </div>
              <div style={{ display: "flex", "flex-direction": "column", gap: "12px" }}>
                <KpiRow label="Velocity" value={`${t().speedKnots.toFixed(1)} KTS`} />
                <KpiRow label="Heading" value={`${t().headingDeg.toFixed(0)}°`} />
                <KpiRow label="Altitude" value={`${t().altitudeMeters.toFixed(0)} FT`} />
                <KpiRow label="Status" value="ACTIVE / TRACKING" color="#22c55e" />
              </div>
            </div>
          </div>

          {/* Footer Actions */}
          <div style={{ display: "flex", "justify-content": "flex-end", gap: "12px", "margin-top": "24px", "border-top": "1px solid rgba(255,255,255,0.1)", "padding-top": "16px" }}>
            <ActionButton label="Analyze" />
            <ActionButton label="Share" />
            <ActionButton label="Close" primary onClick={props.onClose} />
          </div>
        </div>
      )}
    </Show>
  );
};

const KpiRow = (props: { label: string; value: string; color?: string }) => (
  <div style={{ display: "flex", "justify-content": "space-between", "font-family": "monospace" }}>
    <span style={{ color: "#94a3b8" }}>{props.label}:</span>
    <span style={{ color: props.color || "#f8fafc" }}>{props.value}</span>
  </div>
);

const ActionButton = (props: { label: string; primary?: boolean; onClick?: () => void }) => (
  <button
    onClick={props.onClick}
    style={{
      padding: "8px 16px",
      "border-radius": "6px",
      border: props.primary ? "none" : "1px solid rgba(255,255,255,0.2)",
      background: props.primary ? "#2563eb" : "rgba(255,255,255,0.05)",
      color: "white",
      cursor: "pointer",
      "font-size": "0.85rem",
      "font-weight": "600",
      transition: "all 0.2s",
    }}
  >
    {props.label}
  </button>
);

const PedigreeNode = (props: { label: string; time: string; active?: boolean; alert?: boolean }) => (
  <div style={{ display: "flex", "flex-direction": "column", "align-items": "center", gap: "4px" }}>
    <div style={{
      width: "32px",
      height: "32px",
      "border-radius": "50%",
      background: props.alert ? "rgba(239, 68, 68, 0.2)" : props.active ? "rgba(37, 99, 235, 0.2)" : "rgba(255,255,255,0.05)",
      border: `1px solid ${props.alert ? "#ef4444" : props.active ? "#2563eb" : "rgba(255,255,255,0.2)"}`,
      display: "flex",
      "align-items": "center",
      "justify-content": "center",
      "font-size": "0.6rem",
      "font-weight": "bold",
      color: props.alert ? "#ef4444" : props.active ? "#60a5fa" : "#94a3b8",
    }}>
      {props.label[0]}
    </div>
    <span style={{ "font-size": "0.6rem", color: "#64748b" }}>{props.time}</span>
  </div>
);

const PedigreeLine = () => (
  <div style={{ width: "12px", height: "1px", background: "rgba(255,255,255,0.1)", "margin-top": "16px" }} />
);
