// CLASSIFICATION: UNCLASSIFIED
// src/components/detail/DetailPanel.tsx

import React, { useState } from "react";
import { useAlertStore } from "../../stores/alertStore";
import { useAuthStore } from "../../stores/authStore";
import { useTrackStore } from "../../stores/trackStore";
import { useUIStore } from "../../stores/uiStore";
import { EntityTimeline } from "./EntityTimeline";
import { FeedbackForm } from "./FeedbackForm";
import { IdentitySection } from "./IdentitySection";
import { PositionSection } from "./PositionSection";
import { SourceAttributionSection } from "./SourceAttribution";

type ActiveTab = "identity" | "position" | "sources" | "timeline" | "feedback";

/**
 * DetailPanel — shows full details for the selected track entity.
 *
 * Sections (tabs):
 *   1. Identity: track_id, entity_type, hostile_class, confidence, status
 *   2. Position: lat/lon (DMS), altitude, speed, heading, last update
 *   3. Sources: contributing sensors with confidence per source
 *   4. Timeline: chronological history of track updates
 *   5. Feedback: operator feedback submission form
 *
 * Classification: greys out if above operator clearance.
 */
export const DetailPanel: React.FC = () => {
  const selectedTrackId = useTrackStore((s) => s.selectedTrackId);
  const getTrackById = useTrackStore((s) => s.getTrackById);

  const getAlertsByTrack = useAlertStore((s) => s.getAlertsByTrack);
  const acknowledgeAlert = useAlertStore((s) => s.acknowledgeAlert);
  const acknowledgedIds = useAlertStore((s) => s.acknowledgedIds);

  const canAccess = useAuthStore((s) => s.canAccess);
  const closeDetailPanel = useUIStore((s) => s.closeDetailPanel);
  const setMapView = useUIStore((s) => s.setMapView);
  const [activeTab, setActiveTab] = useState<ActiveTab>("identity");

  const [noteOpen, setNoteOpen] = useState(false);
  const [noteText, setNoteText] = useState("");

  if (!selectedTrackId) {
    return (
      <div
        data-testid="detail-panel-empty"
        style={{
          padding: "16px",
          color: "#9CA3AF",
          fontSize: "0.8rem",
          textAlign: "center",
        }}
      >
        Select a track to view details
      </div>
    );
  }

  const track = getTrackById(selectedTrackId);

  if (!track) {
    return (
      <div data-testid="detail-panel-not-found" style={{ padding: "16px", color: "#DC2626", fontSize: "0.8rem" }}>
        Track not found
      </div>
    );
  }

  if (!canAccess(track.classification)) {
    return (
      <div data-testid="detail-panel-restricted" style={{ padding: "16px", color: "#6B7280", fontSize: "0.8rem" }}>
        RESTRICTED — insufficient clearance to view this track
      </div>
    );
  }

  // Check for unacknowledged critical alerts for this track
  const trackAlerts = getAlertsByTrack(track.trackId);
  const unownedCriticalAlerts = trackAlerts.filter(
    (a) => a.severity === "CRITICAL" && !acknowledgedIds.has(a.alertId)
  );

  const TABS: { key: ActiveTab; label: string }[] = [
    { key: "identity", label: "Identity" },
    { key: "position", label: "Position" },
    { key: "sources", label: "Sources" },
    { key: "timeline", label: "Timeline" },
    { key: "feedback", label: "Feedback" },
  ];

  const handleZoom = () => {
    setMapView([track.position.longitude, track.position.latitude], 12);
  };

  const handleExport = () => {
    const blob = new Blob([JSON.stringify(track, null, 2)], { type: "application/json" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `track-${track.trackId}.json`;
    a.click();
    URL.revokeObjectURL(url);
  };

  const submitNote = () => {
    if (noteText.trim()) {
      console.log(`[NOTE] Track ${track.trackId}: ${noteText}`);
      setNoteText("");
      setNoteOpen(false);
    }
  };

  const handleAcknowledgeAll = () => {
    unownedCriticalAlerts.forEach((a) => acknowledgeAlert(a.alertId));
  };

  return (
    <div data-testid="detail-panel" style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      {/* Tab bar */}
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          borderBottom: "1px solid #334155",
          backgroundColor: "#0F172A",
        }}
      >
        <div style={{ display: "flex" }}>
          {TABS.map((tab) => (
            <button
              key={tab.key}
              data-testid={`tab-${tab.key}`}
              onClick={() => setActiveTab(tab.key)}
              style={{
                padding: "6px 12px",
                fontSize: "0.7rem",
                fontWeight: activeTab === tab.key ? "bold" : "normal",
                backgroundColor:
                  activeTab === tab.key ? "#1E293B" : "transparent",
                color: activeTab === tab.key ? "#F1F5F9" : "#9CA3AF",
                border: "none",
                borderBottom:
                  activeTab === tab.key ? "2px solid #3B82F6" : "2px solid transparent",
                cursor: "pointer",
                letterSpacing: "0.05em",
                transition: "all 0.2s ease-in-out",
              }}
            >
              {tab.label.toUpperCase()}
            </button>
          ))}
        </div>
        <button
          data-testid="detail-panel-close"
          onClick={closeDetailPanel}
          style={{
            background: "none",
            border: "none",
            color: "#9CA3AF",
            cursor: "pointer",
            padding: "0 12px",
            fontSize: "1rem"
          }}
        >
          ✕
        </button>
      </div>

      {unownedCriticalAlerts.length > 0 && (
         <div style={{
           padding: "8px 12px",
           backgroundColor: "rgba(220, 38, 38, 0.15)",
           borderBottom: "1px solid #DC2626",
           display: "flex",
           justifyContent: "space-between",
           alignItems: "center"
         }}>
           <span style={{ color: "#FCA5A5", fontSize: "0.75rem", fontWeight: "bold" }}>
             ⚠ {unownedCriticalAlerts.length} Unacknowledged CRITICAL Alert(s)
           </span>
           <button
             onClick={handleAcknowledgeAll}
             style={{
               backgroundColor: "#DC2626",
               color: "white",
               border: "none",
               padding: "4px 8px",
               borderRadius: "4px",
               fontSize: "0.65rem",
               fontWeight: "bold",
               cursor: "pointer"
             }}
           >
             ACKNOWLEDGE
           </button>
         </div>
      )}

      {/* Action Bar */}
      <div className="glass-panel" style={{ display: "flex", gap: "8px", padding: "8px", borderBottom: "1px solid var(--glass-border)", backgroundColor: "var(--glass-bg)" }}>
        <button data-testid="detail-zoom" onClick={handleZoom} style={actionButtonStyle}>Zoom</button>
        <button data-testid="detail-export" onClick={handleExport} style={actionButtonStyle}>Export</button>
        <button data-testid="detail-add-note" onClick={() => setNoteOpen(!noteOpen)} style={actionButtonStyle}>Note</button>
      </div>

      {noteOpen && (
        <div style={{ padding: "8px", borderBottom: "1px solid #334155", display: "flex", gap: "8px" }}>
          <input
            type="text"
            value={noteText}
            onChange={e => setNoteText(e.target.value)}
            placeholder="Type note..."
            style={{ flex: 1, padding: "4px", fontSize: "0.75rem", backgroundColor: "#0F172A", color: "#F1F5F9", border: "1px solid #334155" }}
          />
          <button onClick={submitNote} style={actionButtonStyle}>Save</button>
        </div>
      )}

      {/* Tab content */}
      <div style={{ flex: 1, overflowY: "auto" }}>
        {activeTab === "identity" && <IdentitySection track={track} />}
        {activeTab === "position" && <PositionSection track={track} />}
        {activeTab === "sources" && <SourceAttributionSection track={track} />}
        {activeTab === "timeline" && <EntityTimeline track={track} />}
        {activeTab === "feedback" && (
          <FeedbackForm trackId={track.trackId} />
        )}
      </div>
    </div>
  );
};

const actionButtonStyle: React.CSSProperties = {
  padding: "4px 8px",
  backgroundColor: "rgba(55, 65, 81, 0.5)",
  color: "#F1F5F9",
  border: "1px solid var(--glass-border)",
  borderRadius: "4px",
  cursor: "pointer",
  fontSize: "var(--text-xs)"
};
