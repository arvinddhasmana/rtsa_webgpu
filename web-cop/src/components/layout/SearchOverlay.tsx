// CLASSIFICATION: UNCLASSIFIED
// src/components/layout/SearchOverlay.tsx

import React, { useEffect, useRef } from "react";
import { useTrackStore } from "../../stores/trackStore";
import { useUIStore } from "../../stores/uiStore";

/**
 * SearchOverlay — full-width overlay below toolbar, triggered by Ctrl+F.
 */
export const SearchOverlay: React.FC = () => {
  const searchOpen = useUIStore((s) => s.searchOpen);
  const searchQuery = useUIStore((s) => s.searchQuery);
  const closeSearch = useUIStore((s) => s.closeSearch);
  const setSearchQuery = useUIStore((s) => s.setSearchQuery);

  const tracks = useTrackStore((s) => s.tracks);
  const selectTrack = useTrackStore((s) => s.selectTrack);

  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (searchOpen && inputRef.current) {
      inputRef.current.focus();
    }
  }, [searchOpen]);

  if (!searchOpen) return null;

  const handleResultClick = (trackId: string) => {
    selectTrack(trackId);
    useUIStore.getState().toggleDetailPanel(); // Opening detail panel implicitly opens it if toggle is called while closed. Wait, let's just make it toggle or explicitly open? Oh wait, toggle is toggle. We should really have an open. But using toggle if closed is fine.
    // Actually, I'll check if it's open:
    if (!useUIStore.getState().detailPanelOpen) {
      useUIStore.getState().toggleDetailPanel();
    }
    closeSearch();
    setSearchQuery("");
  };

  const query = searchQuery.toLowerCase();
  const results = Array.from(tracks.values())
    .filter(
      (t) =>
        t.trackId.toLowerCase().includes(query) ||
        t.entityType.toLowerCase().includes(query) ||
        t.hostileClass.toLowerCase().includes(query)
    )
    .slice(0, 10);

  return (
    <div
      data-testid="search-overlay"
      style={{
        position: "absolute",
        top: "76px", // Below toolbar
        left: "50%",
        transform: "translateX(-50%)",
        width: "600px",
        minHeight: "100px",
        backgroundColor: "#1E293B",
        border: "1px solid #334155",
        borderRadius: "4px",
        padding: "16px",
        zIndex: 9999, // Ensure it's above map controls
        boxShadow: "0 10px 15px -3px rgba(0, 0, 0, 0.5)",
      }}
    >
      <div style={{ display: "flex", justifyContent: "space-between", marginBottom: "8px" }}>
        <span style={{ fontSize: "0.75rem", fontWeight: "bold", color: "#9CA3AF" }}>GLOBAL SEARCH</span>
        <button
          onClick={closeSearch}
          style={{ background: "none", border: "none", color: "#9CA3AF", cursor: "pointer", fontSize: "0.8rem" }}
        >
          ✕
        </button>
      </div>

      <input
        ref={inputRef}
        data-testid="search-input"
        type="text"
        placeholder="Search by Track ID, MMSI, Callsign, or MGRS…"
        value={searchQuery}
        onChange={(e) => setSearchQuery(e.target.value)}
        onKeyDown={(e) => {
           if (e.key === "Escape") closeSearch();
        }}
        style={{
          width: "100%",
          padding: "8px",
          backgroundColor: "#0F172A",
          color: "#F1F5F9",
          border: "1px solid #4B5563",
          borderRadius: "4px",
          outline: "none",
          boxSizing: "border-box",
        }}
      />

      {query.length > 0 && (
        <div style={{ marginTop: "12px", display: "flex", flexDirection: "column", gap: "4px" }}>
          {results.length === 0 ? (
            <div style={{ color: "#6B7280", fontSize: "0.75rem", padding: "8px" }}>No results found.</div>
          ) : (
            results.map((t) => (
              <div
                key={t.trackId}
                data-testid={`search-result-${t.trackId}`}
                onClick={() => handleResultClick(t.trackId)}
                style={{
                  padding: "8px",
                  backgroundColor: "#0F172A",
                  border: "1px solid #334155",
                  borderRadius: "4px",
                  cursor: "pointer",
                  display: "flex",
                  justifyContent: "space-between",
                  fontSize: "0.75rem",
                }}
              >
                <span style={{ fontWeight: "bold", color: "#60A5FA" }}>{t.trackId}</span>
                <span style={{ color: "#9CA3AF" }}>{t.entityType} | {t.hostileClass.replace("_", " ")}</span>
              </div>
            ))
          )}
        </div>
      )}
    </div>
  );
};
