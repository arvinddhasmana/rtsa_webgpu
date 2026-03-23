// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/TrackDrillDownOverlay.tsx

import { Component, For } from "solid-js";
import { TrackDetail } from "../../signals/track";
import { setFeedbackOpen } from "../../signals/viewport";
import { EventTimeline } from "../timeline/EventTimeline";

interface TrackDrillDownOverlayProps {
  track: TrackDetail;
  onClose: () => void;
}

export const TrackDrillDownOverlay: Component<TrackDrillDownOverlayProps> = (props) => {
  const t = () => props.track;

  const trackIcon = (type: string) => {
    switch(type.toUpperCase()) {
      case "SURFACE": return "🚢";
      case "SUBSURFACE": return "🐠";
      case "LAND": return "⛟";
      case "AIR": return "✈";
      case "CYBER": return "💻";
      default: return "❓";
    }
  };

  return (
    <div
      style={{
        display: "flex",
        "flex-direction": "column",
        width: "550px",
        padding: "0.5rem 1rem",
        gap: "1rem",
        background: "transparent",
      }}
    >
        {/* Top Actions Row */}
        <div style={{ display: "flex", "justify-content": "flex-end", gap: "8px" }}>
              <button onClick={() => setFeedbackOpen(true)} style={{ background: "linear-gradient(135deg, #3b82f6 0%, #2563eb 100%)", border: "none", color: "#fff", padding: "4px 8px", "border-radius": "4px", "font-weight": "800", cursor: "pointer", "text-transform": "uppercase", "letter-spacing": "0.05em", "font-size": "0.6rem" }}>FEEDBACK</button>
              <button onClick={() => alert(`Sharing Track ${t().trackId} data to external liaison...`)} style={{ background: "rgba(56, 189, 248, 0.1)", border: "1px solid #38bdf8", color: "#38bdf8", padding: "4px 8px", "border-radius": "4px", "font-weight": "700", cursor: "pointer", "font-size": "0.6rem" }}>SHARE</button>
        </div>

        {/* Kinematics Row (Horizontal) */}
        <div style={{
            background: "rgba(0,0,0,0.2)",
            padding: "0.75rem",
            "border-radius": "8px",
            display: "flex",
            "justify-content": "space-around",
            "align-items": "center",
            border: "1px solid rgba(255,255,255,0.05)",
        }}>
            <div style={{ "font-size": "0.65rem", "font-weight": "800", color: "#94a3b8" }}>KINEMATICS</div>
            <div style={{ display: "flex", "flex-direction": "column", "align-items": "center" }}>
                <div style={{ "font-size": "0.55rem", color: "#64748b", "margin-bottom": "0.25rem" }}>VELOCITY (KTS): {t().speedKnots.toFixed(1)}</div>
                <div style={{ width: "60px", height: "30px", position: "relative", overflow: "hidden" }}>
                    <div style={{ width: "60px", height: "60px", border: "4px solid rgba(255,255,255,0.05)", "border-radius": "50%", position: "absolute", top: "0" }} />
                    <div style={{ width: "60px", height: "60px", border: "4px solid transparent", "border-left-color": "#38bdf8", "border-top-color": "#38bdf8", "border-radius": "50%", position: "absolute", top: "0", transform: `rotate(${t().speedKnots/10}deg)` }} />
                </div>
            </div>

            <div style={{ display: "flex", "flex-direction": "column", "align-items": "center" }}>
                <div style={{ "font-size": "0.55rem", color: "#64748b", "margin-bottom": "0.25rem" }}>HEADING: {t().headingDeg.toFixed(0)}°</div>
                <div style={{ width: "40px", height: "40px", border: "1px solid rgba(255,255,255,0.1)", "border-radius": "50%", position: "relative", display: "flex", "align-items": "center", "justify-content": "center" }}>
                    <div style={{ position: "absolute", top: "2px", "font-size": "6px", color: "#f87171" }}>N</div>
                    <div style={{ width: "2px", height: "16px", background: "#f8fafc", transform: `rotate(${t().headingDeg}deg)`, "transform-origin": "bottom center", position: "absolute", top: "4px" }} />
                </div>
            </div>

            <div style={{ display: "flex", "flex-direction": "column", "align-items": "center" }}>
                <div style={{ "font-size": "0.55rem", color: "#64748b", "margin-bottom": "0.25rem" }}>ALTITUDE (FT): {t().altitudeMeters.toFixed(0)}</div>
                <div style={{ display: "flex", "align-items": "flex-end", height: "30px" }}>
                    <div style={{ height: "100%", width: "12px", border: "1px solid rgba(255,255,255,0.1)", position: "relative" }}>
                         <div style={{ position: "absolute", bottom: `${Math.min(100, (t().altitudeMeters/100))}%`, left: "0", right: "0", height: "20%", background: "#38bdf8", "box-shadow": "0 0 4px #38bdf8" }} />
                    </div>
                </div>
            </div>
        </div>

        {/* Confidence & Classification */}
        <div style={{ display: "grid", "grid-template-columns": "1fr 1fr", gap: "1rem", "align-items": "center" }}>
            <div style={{ background: "rgba(0,0,0,0.1)", padding: "10px", "border-radius": "6px", border: "1px solid rgba(255,255,255,0.05)" }}>
                <div style={{ "font-size": "0.6rem", color: "#64748b", "margin-bottom": "0.5rem" }}>TRACK CONFIDENCE: <span style={{ color: "#38bdf8" }}>{t().confidenceScore.toFixed(0)}%</span></div>
                <div style={{ height: "4px", background: "rgba(255,255,255,0.05)", "border-radius": "2px", overflow: "hidden" }}>
                    <div style={{ width: `${t().confidenceScore}%`, height: "100%", background: "#38bdf8" }} />
                </div>
            </div>
            <div style={{ background: "rgba(0,0,0,0.1)", padding: "10px", "border-radius": "6px", border: "1px solid rgba(255,255,255,0.05)" }}>
                <div style={{ "font-size": "0.6rem", color: "#64748b", "margin-bottom": "0.5rem" }}>CLASSIFICATION:</div>
                <div style={{ display: "flex", "align-items": "center", gap: "0.5rem" }}>
                   <span style={{ "font-size": "1.2rem" }}>{trackIcon(t().entityType)}</span>
                   <span style={{ "font-size": "0.85rem", "font-weight": "700", "letter-spacing": "0.05em" }}>{t().entityType.toUpperCase()}</span>
                </div>
            </div>
        </div>

        {/* Source Pedigree */}
        <div style={{ display: "flex", "flex-direction": "column" }}>
            <div style={{ "font-size": "0.65rem", "font-weight": "700", color: "#f8fafc", "border-bottom": "1px solid rgba(56, 189, 248, 0.3)", "padding-bottom": "0.25rem", "margin-bottom": "0.5rem", display: "flex", "justify-content": "space-between" }}>
                <span>SOURCE PEDIGREE</span>
                <span style={{ color: "#38bdf8", "font-size": "0.55rem" }}>{t().sourceContributions.length} CONTRIBUTORS</span>
            </div>

            <div style={{ position: "relative", "padding-top": "0.5rem", "margin-bottom": "0.5rem" }}>
                {/* Connecting Line */}
                <div style={{ position: "absolute", top: "1.25rem", left: "20px", right: "20px", height: "2px", background: "linear-gradient(90deg, #38bdf8 0%, #fbbf24 100%)", opacity: "0.2" }} />

                <div style={{ display: "flex", gap: "1rem", position: "relative", "overflow-x": "auto", "padding-bottom": "0.5rem" }}>
                     <For each={t().sourceContributions}>
                        {(source, index) => (
                            <PedigreeNode
                                time={new Date(source.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                                label={source.sourceName}
                                sub={source.data}
                                active
                                color={source.signalStrength > 0.8 ? "#10b981" : "#38bdf8"}
                                pulse={index() === t().sourceContributions.length - 1}
                            />
                        )}
                     </For>
                </div>
            </div>

            {/* Event Timeline */}
            <div style={{ border: "1px solid rgba(56, 189, 248, 0.1)", "border-radius": "8px", background: "rgba(0,0,0,0.2)", padding: "0.5rem" }}>
                <EventTimeline />
            </div>
        </div>
    </div>
  );
};

const PedigreeNode = (props: { time: string; label: string; sub: string; active: boolean; color: string; pulse?: boolean }) => (
    <div style={{ display: "flex", "flex-direction": "column", "align-items": "center", gap: "1rem", "z-index": "10", width: "80px" }}>
        <div style={{ "font-size": "0.6rem", color: "#64748b" }}>{props.time}</div>
        <div style={{
            width: "32px",
            height: "32px",
            "border-radius": "50%",
            background: "#0f172a",
            border: `2px solid ${props.color}`,
            display: "flex",
            "align-items": "center",
            "justify-content": "center",
            "box-shadow": props.pulse ? `0 0 15px ${props.color}` : "none",
            animation: props.pulse ? "pulse-node 2s infinite" : "none"
        }}>
           <div style={{ width: "8px", height: "8px", background: props.color, "border-radius": "50%" }} />
        </div>
        <div style={{ "text-align": "center" }}>
            <div style={{ "font-size": "0.6rem", "font-weight": "700", color: "#e2e8f0" }}>{props.label}</div>
            <div style={{ "font-size": "0.55rem", color: "#64748b" }}>{props.sub}</div>
        </div>
        <style>{`
          @keyframes pulse-node {
            0% { box-shadow: 0 0 0 0 rgba(251, 191, 36, 0.4); }
            70% { box-shadow: 0 0 0 15px rgba(251, 191, 36, 0); }
            100% { box-shadow: 0 0 0 0 rgba(251, 191, 36, 0); }
          }
        `}</style>
    </div>
);
