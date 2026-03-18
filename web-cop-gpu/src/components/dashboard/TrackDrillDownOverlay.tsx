// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/TrackDrillDownOverlay.tsx

import { Component, createEffect, For, Show } from "solid-js";
import { TrackDetail } from "../../signals/track";
import { setFeedbackOpen } from "../../signals/viewport";
import { EventTimeline } from "../timeline/EventTimeline";

interface TrackDrillDownOverlayProps {
  track: TrackDetail | null;
  onClose: () => void;
}

export const TrackDrillDownOverlay: Component<TrackDrillDownOverlayProps> = (props) => {
  createEffect(() => {
    if (props.track) {
        console.log("[TrackHUD] Selected Track:", props.track.trackId);
    }
  });

  return (
    <Show when={props.track}>
      {(t) => (
        <div
          style={{
            position: "absolute",
            top: "50%",
            left: "50%",
            transform: "translate(-50%, -50%)",
            width: "800px",
            height: "600px",
            background: "rgba(10, 20, 30, 0.7)",
            "backdrop-filter": "blur(24px)",
            border: "2px solid rgba(56, 189, 248, 0.2)",
            "border-radius": "16px",
            display: "flex",
            "flex-direction": "column",
            overflow: "hidden",
            "box-shadow": "0 0 100px rgba(0,0,0,0.8), inset 0 0 40px rgba(56, 189, 248, 0.05)",
            "z-index": "1000",
            animation: "hud-appear 0.4s cubic-bezier(0.16, 1, 0.3, 1)",
          }}
        >
          {/* Header Bar */}
          <div style={{
            padding: "0.75rem 1.5rem",
            background: "linear-gradient(90deg, rgba(56, 189, 248, 0.1) 0%, transparent 100%)",
            "border-bottom": "1px solid rgba(255, 255, 255, 0.1)",
            display: "flex",
            "justify-content": "space-between",
            "align-items": "center",
          }}>
            <div style={{ display: "flex", "align-items": "center", gap: "1rem" }}>
                <span style={{ "font-size": "0.7rem", "font-weight": "800", color: "#38bdf8", "letter-spacing": "0.15em" }}>FUSED TRACK DATA</span>
                <span style={{ color: "rgba(255,255,255,0.2)" }}>|</span>
                <span style={{ "font-size": "0.7rem", "font-weight": "700", color: "#f8fafc" }}>TRACK ID: {t().trackId}</span>
            </div>
            <button onClick={props.onClose} style={{ background: "transparent", border: "none", color: "#94a3b8", cursor: "pointer", "font-size": "1.2rem" }}>×</button>
          </div>

          <div style={{ flex: "1", padding: "0", display: "grid", "grid-template-columns": "1fr 240px", gap: "0", "min-height": "0", overflow: "hidden" }}>
             {/* Left Section: Details and Pedigree */}
             <div style={{
               display: "flex",
               "flex-direction": "column",
               gap: "1.5rem",
               padding: "1.5rem",
               "overflow-y": "auto",
               "border-right": "1px solid rgba(255,255,255,0.05)"
             }}>
                {/* Confidence & Classification Row */}
                <div style={{ display: "grid", "grid-template-columns": "200px 1fr", gap: "2rem" }}>
                    <div>
                        <div style={{ "font-size": "0.65rem", color: "#64748b", "margin-bottom": "0.5rem" }}>TRACK CONFIDENCE: <span style={{ color: "#38bdf8" }}>{t().confidenceScore.toFixed(0)}%</span></div>
                        <div style={{ height: "4px", background: "rgba(255,255,255,0.05)", "border-radius": "2px", overflow: "hidden" }}>
                            <div style={{ width: `${t().confidenceScore}%`, height: "100%", background: "#38bdf8" }} />
                        </div>
                    </div>
                    <div>
                        <div style={{ "font-size": "0.65rem", color: "#64748b", "margin-bottom": "0.5rem" }}>CLASSIFICATION:</div>
                        <div style={{ display: "flex", "align-items": "center", gap: "0.5rem" }}>
                           <span style={{ "font-size": "1.5rem" }}>{t().entityType === "Surface" ? "🚢" : "✈"}</span>
                           <span style={{ "font-size": "1.1rem", "font-weight": "700", "letter-spacing": "0.05em" }}>{t().entityType.toUpperCase()}</span>
                        </div>
                    </div>
                </div>

                {/* Source Pedigree & Timeline Section */}
                <div style={{ display: "flex", "flex-direction": "column" }}>
                    <div style={{ "font-size": "0.75rem", "font-weight": "700", color: "#f8fafc", "border-bottom": "1px solid rgba(56, 189, 248, 0.3)", "padding-bottom": "0.5rem", "margin-bottom": "1rem", display: "flex", "justify-content": "space-between" }}>
                        <span>SOURCE PEDIGREE</span>
                        <span style={{ color: "#38bdf8", "font-size": "0.6rem" }}>{t().sourceContributions.length} CONTRIBUTORS</span>
                    </div>

                    <div style={{ position: "relative", "padding-top": "1rem", "margin-bottom": "1.5rem" }}>
                        {/* Connecting Line */}
                        <div style={{ position: "absolute", top: "2.5rem", left: "20px", right: "20px", height: "2px", background: "linear-gradient(90deg, #38bdf8 0%, #fbbf24 100%)", opacity: "0.2" }} />

                        <div style={{ display: "flex", gap: "1rem", position: "relative", "overflow-x": "auto", "padding-bottom": "1rem" }}>
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

                    {/* Integrated Event Timeline */}
                    <div style={{ border: "1px solid rgba(56, 189, 248, 0.1)", "border-radius": "8px", background: "rgba(0,0,0,0.2)", padding: "0.5rem 1rem" }}>
                        <EventTimeline />
                    </div>
                </div>
             </div>

             {/* Right Section: HUD Gauges */}
             <div style={{
               background: "rgba(0,0,0,0.2)",
               padding: "1.25rem",
               display: "flex",
               "flex-direction": "column",
               gap: "1.5rem",
               "overflow-y": "auto"
             }}>
                <div style={{ "font-size": "0.7rem", "font-weight": "800", color: "#94a3b8" }}>KINEMATICS</div>

                {/* Velocity Gauge */}
                <div style={{ display: "flex", "flex-direction": "column", "align-items": "center" }}>
                    <div style={{ "font-size": "0.6rem", color: "#64748b", "margin-bottom": "0.5rem" }}>VELOCITY (KTS): {t().speedKnots.toFixed(1)}</div>
                    <div style={{ width: "120px", height: "60px", position: "relative", overflow: "hidden" }}>
                        <div style={{ width: "120px", height: "120px", border: "8px solid rgba(255,255,255,0.05)", "border-radius": "50%", position: "absolute", top: "0" }} />
                        <div style={{ width: "120px", height: "120px", border: "8px solid transparent", "border-left-color": "#38bdf8", "border-top-color": "#38bdf8", "border-radius": "50%", position: "absolute", top: "0", transform: `rotate(${t().speedKnots/10}deg)` }} />
                    </div>
                </div>

                <div style={{ height: "1px", background: "rgba(255,255,255,0.05)" }} />

                {/* Heading Compass */}
                <div style={{ display: "flex", "flex-direction": "column", "align-items": "center" }}>
                    <div style={{ "font-size": "0.6rem", color: "#64748b", "margin-bottom": "1rem" }}>HEADING: {t().headingDeg.toFixed(0)}°</div>
                    <div style={{ width: "100px", height: "100px", border: "2px solid rgba(255,255,255,0.1)", "border-radius": "50%", position: "relative", display: "flex", "align-items": "center", "justify-content": "center" }}>
                        <div style={{ position: "absolute", top: "5px", "font-size": "8px", color: "#f87171" }}>N</div>
                        <div style={{ width: "2px", height: "40px", background: "#f8fafc", transform: `rotate(${t().headingDeg}deg)`, "transform-origin": "bottom center", position: "absolute", top: "10px" }} />
                        <div style={{ "font-size": "10px", color: "#38bdf8" }}>{t().headingDeg.toFixed(0)}°</div>
                    </div>
                </div>

                <div style={{ height: "1px", background: "rgba(255,255,255,0.05)" }} />

                {/* Altitude */}
                <div style={{ display: "flex", "flex-direction": "column" }}>
                    <div style={{ "font-size": "0.6rem", color: "#64748b", "margin-bottom": "0.5rem" }}>ALTITUDE (FT):</div>
                    <div style={{ display: "flex", "align-items": "flex-end", gap: "1rem", height: "60px" }}>
                        <div style={{ flex: "1", height: "100%", border: "1px solid rgba(255,255,255,0.1)", position: "relative" }}>
                             <div style={{ position: "absolute", bottom: `${(t().altitudeMeters/100)}%`, left: "0", right: "0", height: "2px", background: "#38bdf8", "box-shadow": "0 0 8px #38bdf8" }} />
                        </div>
                        <div style={{ "font-size": "1.2rem", "font-weight": "700" }}>{t().altitudeMeters.toFixed(0)}</div>
                    </div>
                </div>
             </div>
          </div>

          {/* Action Footer (Fixed at the bottom) */}
          <div style={{
            padding: "1rem 1.5rem",
            "border-top": "1px solid rgba(255,255,255,0.05)",
            background: "rgba(0,0,0,0.2)",
            display: "flex",
            gap: "1rem"
          }}>
              <button onClick={props.onClose} style={{ flex: "1", background: "rgba(255, 255, 255, 0.05)", border: "1px solid rgba(255, 255, 255, 0.1)", color: "#94a3b8", padding: "0.75rem", "border-radius": "4px", "font-weight": "700", cursor: "pointer", transition: "all 0.2s" }}>DISMISS</button>
              <button onClick={() => setFeedbackOpen(true)} style={{ flex: "2", background: "linear-gradient(135deg, #3b82f6 0%, #2563eb 100%)", border: "none", color: "#fff", padding: "0.75rem", "border-radius": "4px", "font-weight": "800", cursor: "pointer", "box-shadow": "0 4px 15px rgba(59, 130, 246, 0.3)", "text-transform": "uppercase", "letter-spacing": "0.05em" }}>OPERATIONAL FEEDBACK</button>
              <button onClick={() => alert(`Sharing Track ${t().trackId} data to external liaison...`)} style={{ flex: "1", background: "rgba(56, 189, 248, 0.1)", border: "1px solid #38bdf8", color: "#38bdf8", padding: "0.75rem", "border-radius": "4px", "font-weight": "700", cursor: "pointer" }}>SHARE</button>
          </div>

          <style>{`
            @keyframes hud-appear {
              from { opacity: 0; transform: translate(-50%, -45%) scale(0.95); }
              to { opacity: 1; transform: translate(-50%, -50%) scale(1); }
            }
          `}</style>
        </div>
      )}
    </Show>
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
