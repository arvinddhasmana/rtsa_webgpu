// CLASSIFICATION: UNCLASSIFIED
// src/components/forensics/MapReplay.tsx

import React, { useEffect, useMemo, useRef, useState } from "react";
import { FusedTrack } from "../../types/track";
import { formatZuluTime } from "../../utils/time";

interface MapReplayProps {
  tracks: FusedTrack[];
}

type ReplaySpeed = 1 | 2 | 4 | 8;

type MapInstance = {
  getSource?: (id: string) => {
    setData?: (data: GeoJSON.FeatureCollection) => void;
  } | null;
};

/**
 * MapReplay — animate historical track positions on the live map.
 *
 * HOW IT WORKS:
 *   1. Tracks are sorted by updatedAt timestamp.
 *   2. Each frame represents one time step in that sorted sequence.
 *   3. At frame N, ALL tracks whose updatedAt ≤ T(N) are shown — giving a
 *      progressive appearance of tracks over time.
 *   4. Positions are written to the "replay-tracks" GeoJSON source on the map
 *      (added by MapView, separate from live tracks). Replay circles have a
 *      yellow stroke to distinguish them from live data.
 *   5. On unmount the replay source is cleared.
 *
 * CONTROLS: ▶ PLAY / ⏸ PAUSE, speed ×1/2/4/8, range scrubber, time display.
 */
export const MapReplay: React.FC<MapReplayProps> = ({ tracks }) => {
  const [isPlaying, setIsPlaying] = useState(false);
  const [speed, setSpeed] = useState<ReplaySpeed>(1);
  const [currentIndex, setCurrentIndex] = useState(0);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  // Sort tracks chronologically once; re-sort only when tracks array changes.
  const sortedTracks = useMemo(
    () =>
      [...tracks].sort((a, b) => a.updatedAt.getTime() - b.updatedAt.getTime()),
    [tracks],
  );

  const totalFrames = Math.max(sortedTracks.length, 1);

  // ── Playback interval ────────────────────────────────────────────────────
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

  // ── Map update — write current frame's tracks to the replay GeoJSON source ──
  useEffect(() => {
    const map = (window as unknown as Record<string, MapInstance>)[
      "__RTSA_MAP__"
    ];
    if (!map?.getSource) return;

    const replaySource = map.getSource("replay-tracks");
    if (!replaySource?.setData) return;

    const currentTime = sortedTracks[currentIndex]?.updatedAt;
    if (!currentTime) return;

    // Show all tracks at/before the current frame's timestamp
    const visible = sortedTracks.filter(
      (t) => t.updatedAt.getTime() <= currentTime.getTime(),
    );

    replaySource.setData({
      type: "FeatureCollection",
      features: visible.map((t) => ({
        type: "Feature",
        geometry: {
          type: "Point",
          coordinates: [t.position.longitude, t.position.latitude],
        },
        properties: {
          trackId: t.trackId,
          hostileClass: t.hostileClass,
          entityType: t.entityType,
          confidence: t.confidenceScore,
        },
      })),
    });
  }, [currentIndex, sortedTracks]);

  // ── Cleanup: clear replay source on unmount ──────────────────────────────
  useEffect(() => {
    return () => {
      const map = (window as unknown as Record<string, MapInstance>)[
        "__RTSA_MAP__"
      ];
      const replaySource = map?.getSource?.("replay-tracks");
      replaySource?.setData?.({
        type: "FeatureCollection",
        features: [],
      });
    };
  }, []);

  const currentTime = sortedTracks[currentIndex]?.updatedAt;

  const handleScrub = (e: React.ChangeEvent<HTMLInputElement>) => {
    setCurrentIndex(parseInt(e.target.value, 10));
    setIsPlaying(false);
  };

  return (
    <div
      data-testid="map-replay"
      style={{
        padding: "6px 8px",
        display: "flex",
        alignItems: "center",
        gap: "8px",
        backgroundColor: "#0F172A",
        borderTop: "1px solid #334155",
        flexShrink: 0,
      }}
    >
      {/* Play/Pause */}
      <button
        data-testid="replay-play-pause"
        title="Replay historical track positions on the map. Records play back chronologically at the selected speed."
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
          flexShrink: 0,
        }}
      >
        {isPlaying ? "⏸ PAUSE" : "▶ PLAY"}
      </button>

      {/* Time scrubber */}
      <input
        data-testid="replay-scrubber"
        type="range"
        min={0}
        max={totalFrames - 1}
        value={currentIndex}
        onChange={handleScrub}
        style={{ flex: 1, minWidth: 0 }}
      />

      {/* Frame counter */}
      <div
        style={{
          fontSize: "0.65rem",
          color: "#9CA3AF",
          minWidth: "52px",
          textAlign: "center",
        }}
      >
        {currentIndex + 1}/{totalFrames}
      </div>

      {/* Current timestamp */}
      {currentTime && (
        <div
          style={{
            fontSize: "0.65rem",
            color: "#60A5FA",
            fontFamily: "monospace",
            whiteSpace: "nowrap",
          }}
        >
          {formatZuluTime(currentTime)}
        </div>
      )}

      {/* Speed selector */}
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
          flexShrink: 0,
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
