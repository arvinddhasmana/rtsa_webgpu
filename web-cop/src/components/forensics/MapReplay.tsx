// CLASSIFICATION: UNCLASSIFIED
// src/components/forensics/MapReplay.tsx

import React, { useState, useEffect, useRef } from "react";
import { FusedTrack } from "../../types/track";

interface MapReplayProps {
  tracks: FusedTrack[];
}

type ReplaySpeed = 1 | 2 | 4 | 8;

/**
 * MapReplay — animate historical track positions on map.
 * Controls: play/pause/speed + time scrubber.
 */
export const MapReplay: React.FC<MapReplayProps> = ({ tracks }) => {
  const [isPlaying, setIsPlaying] = useState(false);
  const [speed, setSpeed] = useState<ReplaySpeed>(1);
  const [currentIndex, setCurrentIndex] = useState(0);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const totalFrames = Math.max(tracks.length, 1);

  useEffect(() => {
    if (isPlaying) {
      intervalRef.current = setInterval(() => {
        setCurrentIndex((idx) => {
          if (idx >= totalFrames - 1) {
            setIsPlaying(false);
            return idx;
          }
          return idx + 1;
        });
      }, 1000 / speed);
    } else {
      if (intervalRef.current !== null) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
    }

    return () => {
      if (intervalRef.current !== null) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
    };
  }, [isPlaying, speed, totalFrames]);

  const handleScrub = (e: React.ChangeEvent<HTMLInputElement>) => {
    setCurrentIndex(parseInt(e.target.value, 10));
    setIsPlaying(false);
  };

  return (
    <div
      data-testid="map-replay"
      style={{
        padding: "8px",
        display: "flex",
        alignItems: "center",
        gap: "8px",
        backgroundColor: "#0F172A",
        borderTop: "1px solid #334155",
      }}
    >
      <button
        data-testid="replay-play-pause"
        onClick={() => setIsPlaying((p) => !p)}
        style={{
          padding: "4px 12px",
          backgroundColor: isPlaying ? "#CA8A04" : "#16A34A",
          color: "white",
          border: "none",
          borderRadius: "4px",
          cursor: "pointer",
          fontSize: "0.75rem",
          fontWeight: "bold",
        }}
      >
        {isPlaying ? "⏸ PAUSE" : "▶ PLAY"}
      </button>

      <input
        data-testid="replay-scrubber"
        type="range"
        min={0}
        max={totalFrames - 1}
        value={currentIndex}
        onChange={handleScrub}
        style={{ flex: 1 }}
      />

      <div style={{ fontSize: "0.65rem", color: "#9CA3AF", minWidth: "60px" }}>
        {currentIndex + 1} / {totalFrames}
      </div>

      <select
        data-testid="replay-speed"
        value={speed}
        onChange={(e) => setSpeed(parseInt(e.target.value, 10) as ReplaySpeed)}
        style={{
          padding: "2px 4px",
          backgroundColor: "#374151",
          color: "#F1F5F9",
          border: "1px solid #475569",
          borderRadius: "4px",
          fontSize: "0.7rem",
        }}
      >
        <option value={1}>1×</option>
        <option value={2}>2×</option>
        <option value={4}>4×</option>
        <option value={8}>8×</option>
      </select>
    </div>
  );
};
