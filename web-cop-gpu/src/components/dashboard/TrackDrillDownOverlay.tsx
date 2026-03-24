// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/TrackDrillDownOverlay.tsx

import { Component, For, Show, createMemo } from "solid-js";
import { alerts } from "../../signals/alerts";
import { TrackDetail, setTrackDetail } from "../../signals/track";
import { setFeedbackOpen } from "../../signals/viewport";
import { EventTimeline } from "../timeline/EventTimeline";

interface TrackDrillDownOverlayProps {
  track: TrackDetail;
  onClose: () => void;
}

export const TrackDrillDownOverlay: Component<TrackDrillDownOverlayProps> = (props) => {
  const t = () => props.track;
  const hasActiveAlert = createMemo(() =>
    alerts().some(a => a.trackId === t().trackId && !a.acknowledged)
  );

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
        padding: "0.15rem 0.25rem",
        gap: "0.5rem",
        background: "transparent",
      }}
    >
        {/* Alert Indicator */}
        <Show when={hasActiveAlert()}>
            <div style={{
                background: "#ef4444",
                color: "#fff",
                padding: "4px 10px",
                "border-radius": "4px",
                "font-size": "0.65rem",
                "font-weight": "900",
                "text-transform": "uppercase",
                "letter-spacing": "0.15em",
                display: "flex",
                "align-items": "center",
                gap: "8px",
                border: "1px solid rgba(255, 255, 255, 0.2)",
                "box-shadow": "0 0 15px rgba(239, 68, 68, 0.5)"
            }}>
                <span style={{ "font-size": "0.9rem", "animation": "flash 1.5s infinite" }}>⚠</span> ALERT ACTIVE
                <style>{`
                    @keyframes flash {
                        0% { opacity: 1; }
                        50% { opacity: 0.4; }
                        100% { opacity: 1; }
                    }
                `}</style>
            </div>
        </Show>

        {/* Top Row: Icon + Confidence + Actions */}
        <div style={{ display: "flex", "justify-content": "space-between", "align-items": "center" }}>
             <div style={{ display: "flex", "align-items": "center", gap: "0.75rem" }}>
                 <span style={{ "font-size": "1.4rem" }} title={t().entityType.toUpperCase()}>{trackIcon(t().entityType)}</span>
                 <div style={{ display: "flex", "flex-direction": "column", width: "120px" }}>
                     <div style={{ "font-size": "0.6rem", color: "#94a3b8", "margin-bottom": "0.25rem", "font-weight": 800 }}>CONFIDENCE: <span style={{ color: "#38bdf8" }}>{(t().confidenceScore * 100).toFixed(0)}%</span></div>
                     <div style={{ height: "12px", background: "rgba(255,255,255,0.05)", "border-radius": "2px", overflow: "hidden", border: "1px solid rgba(255,255,255,0.1)" }}>
                         <div style={{ width: `${t().confidenceScore * 100}%`, height: "100%", background: "linear-gradient(90deg, #0ea5e9, #38bdf8)", "box-shadow": "0 0 8px rgba(56, 189, 248, 0.4)" }} />
                     </div>
                 </div>
             </div>

             <div style={{ display: "flex", gap: "6px" }}>
                  <button onClick={() => { setTrackDetail(t()); setFeedbackOpen(true); }} style={{ background: "rgba(59, 130, 246, 0.15)", border: "1px solid #3b82f6", color: "#60a5fa", padding: "4px 10px", "border-radius": "4px", "font-weight": "800", cursor: "pointer", "text-transform": "uppercase", "letter-spacing": "0.05em", "font-size": "0.6rem", transition: "all 0.2s" }}>FEEDBACK</button>
                  <button onClick={() => alert(`Sharing Track ${t().trackId} data to external liaison...`)} style={{ background: "rgba(255, 255, 255, 0.05)", border: "1px solid rgba(255,255,255,0.2)", color: "#e2e8f0", padding: "4px 10px", "border-radius": "4px", "font-weight": "700", cursor: "pointer", "font-size": "0.6rem" }}>SHARE</button>
             </div>
        </div>

        {/* Kinematics Row (Tactical HUD) */}
        <div style={{
            background: "rgba(255,255,255,0.03)",
            padding: "0.4rem 0.6rem",
            "border-radius": "6px",
            border: "1px solid rgba(255,255,255,0.05)",
            display: "flex",
            "justify-content": "space-between",
            "align-items": "center",
        }}>
            <div style={{ display: "flex", gap: "1.5rem" }}>
                <div style={{ display: "flex", "flex-direction": "column" }}>
                    <div style={{ "font-size": "0.5rem", color: "#64748b", "letter-spacing": "0.05em" }}>SPD</div>
                    <div style={{ "font-size": "0.85rem", "font-weight": "800", color: "#f1f5f9", "font-family": "monospace" }}>{t().speedKnots.toFixed(0)}<span style={{ "font-size": "0.6rem", color: "#64748b", "margin-left": "2px" }}>KTS</span></div>
                </div>
                <div style={{ display: "flex", "flex-direction": "column" }}>
                    <div style={{ "font-size": "0.5rem", color: "#64748b", "letter-spacing": "0.05em" }}>ALT</div>
                    <div style={{ "font-size": "0.85rem", "font-weight": "800", color: "#f1f5f9", "font-family": "monospace" }}>{t().altitudeMeters.toFixed(0)}<span style={{ "font-size": "0.6rem", color: "#64748b", "margin-left": "2px" }}>FT</span></div>
                </div>
            </div>

            <div style={{ display: "flex", "align-items": "center", gap: "8px" }}>
                <div style={{ "text-align": "right" }}>
                    <div style={{ "font-size": "0.5rem", color: "#64748b" }}>HEADING</div>
                    <div style={{ "font-size": "0.75rem", "font-weight": "700", color: "#f8fafc" }}>{t().headingDeg.toFixed(0)}°</div>
                </div>
                <div style={{ width: "32px", height: "32px", border: "1px solid rgba(255,255,255,0.1)", "border-radius": "50%", position: "relative", display: "flex", "align-items": "center", "justify-content": "center", background: "rgba(0,0,0,0.2)" }}>
                    <div style={{ position: "absolute", top: "2px", "font-size": "5px", color: "#f87171", "font-weight": "900" }}>N</div>
                    {/* Directional Arrow */}
                    <div style={{
                        position: "absolute",
                        width: "0",
                        height: "0",
                        "border-left": "4px solid transparent",
                        "border-right": "4px solid transparent",
                        "border-bottom": "12px solid #38bdf8",
                        transform: `rotate(${t().headingDeg}deg)`,
                        "transform-origin": "center 12px",
                        top: "4px"
                    }} />
                </div>
            </div>
        </div>

        {/* Source Pedigree */}
        <div style={{ display: "flex", "flex-direction": "column", gap: "0.25rem" }}>
            <div style={{ "font-size": "0.6rem", "font-weight": "800", color: "#94a3b8", "border-bottom": "1px solid rgba(255, 255, 255, 0.05)", "padding-bottom": "2px", "margin-bottom": "0.25rem", display: "flex", "justify-content": "space-between", "text-transform": "uppercase", "letter-spacing": "0.05em" }}>
                <span>SOURCE PEDIGREE</span>
                <span style={{ color: "#38bdf8" }}>{t().sourceContributions.length} CONTRIBUTORS</span>
            </div>

            <div style={{ position: "relative", padding: "0" }}>
                <div style={{ display: "flex", gap: "1rem", position: "relative", "overflow-x": "auto", "padding-bottom": "0.25rem", "scrollbar-width": "none" }}>
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
            <div style={{ "margin-top": "0.25rem" }}>
                <EventTimeline />
            </div>
        </div>
    </div>
  );
};

const PedigreeNode = (props: { time: string; label: string; sub: string; active: boolean; color: string; pulse?: boolean }) => (
    <div style={{ display: "flex", "flex-direction": "column", "align-items": "center", gap: "0.25rem", "z-index": "10", "min-width": "60px" }}>
        <div style={{ "font-size": "0.5rem", color: "#64748b" }}>{props.time}</div>
        <div style={{
            width: "12px",
            height: "12px",
            "border-radius": "50%",
            background: "#0f172a",
            border: `1.5px solid ${props.color}`,
            display: "flex",
            "align-items": "center",
            "justify-content": "center",
            "box-shadow": props.pulse ? `0 0 8px ${props.color}` : "none",
            animation: props.pulse ? "pulse-node 2s infinite" : "none"
        }}>
           <div style={{ width: "4px", height: "4px", background: props.color, "border-radius": "50%" }} />
        </div>
        <div style={{ "text-align": "center" }}>
            <div style={{ "font-size": "0.55rem", "font-weight": "800", color: "#e2e8f0", "white-space": "nowrap" }}>{props.label}</div>
            <div style={{ "font-size": "0.5rem", color: "#64748b", "white-space": "nowrap", opacity: 0.8 }}>{props.sub}</div>
        </div>
        <style>{`
          @keyframes pulse-node {
            0% { box-shadow: 0 0 0 0 rgba(56, 189, 248, 0.4); }
            70% { box-shadow: 0 0 0 8px rgba(56, 189, 248, 0); }
            100% { box-shadow: 0 0 0 0 rgba(56, 189, 248, 0); }
          }
        `}</style>
    </div>
);
