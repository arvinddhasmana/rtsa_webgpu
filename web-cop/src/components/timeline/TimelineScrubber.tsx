// CLASSIFICATION: UNCLASSIFIED
// src/components/timeline/TimelineScrubber.tsx
// Historical map playback scrubber with speed control.

import React, { useEffect, useRef } from "react";

interface TimelineScrubberProps {
  /** Earliest time boundary (Unix timestamp ms) */
  startMs: number;
  /** Latest time boundary (Unix timestamp ms) */
  endMs: number;
  /** Currently selected playback time */
  currentMs: number;
  /** Whether the scrubber is actively playing */
  isPlaying: boolean;
  /** Playback speed multiplier */
  speed: number;
  onSeek: (ms: number) => void;
  onPlay: () => void;
  onPause: () => void;
  onSpeedChange: (speed: number) => void;
  onClose?: () => void;
}

const SPEED_OPTIONS = [0.5, 1, 2, 5, 10];

function formatTime(ms: number): string {
  const d = new Date(ms);
  return d.toISOString().replace("T", " ").slice(0, 19) + "Z";
}

/**
 * TimelineScrubber — horizontal slider for historical map replay.
 *
 * Displays current replay time, play/pause, speed selector, and a
 * labelled range slider. The parent drives actual query/animation logic.
 */
export const TimelineScrubber: React.FC<TimelineScrubberProps> = ({
  startMs,
  endMs,
  currentMs,
  isPlaying,
  speed,
  onSeek,
  onPlay,
  onPause,
  onSpeedChange,
  onClose,
}) => {
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  // Auto-advance when playing
  useEffect(() => {
    if (!isPlaying) {
      if (intervalRef.current) clearInterval(intervalRef.current);
      return;
    }
    intervalRef.current = setInterval(() => {
      const next = currentMs + speed * 10_000; // advance 10s * speed per tick
      if (next >= endMs) {
        onPause();
        onSeek(endMs);
      } else {
        onSeek(next);
      }
    }, 200);

    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current);
    };
  }, [isPlaying, speed, currentMs, endMs, onSeek, onPause]);

  const pct = endMs > startMs ? ((currentMs - startMs) / (endMs - startMs)) * 100 : 0;

  const handleSlider = (e: React.ChangeEvent<HTMLInputElement>) => {
    const val = Number(e.target.value);
    onSeek(startMs + (val / 100) * (endMs - startMs));
  };

  return (
    <div
      data-testid="timeline-scrubber"
      style={{
        position: "absolute",
        bottom: "56px",
        left: "50%",
        transform: "translateX(-50%)",
        width: "min(900px, calc(100vw - 64px))",
        backgroundColor: "rgba(15, 23, 42, 0.92)",
        backdropFilter: "blur(10px)",
        border: "1px solid #334155",
        borderRadius: "10px",
        padding: "12px 20px",
        zIndex: 20,
        boxShadow: "0 4px 24px rgba(0,0,0,0.5)",
      }}
    >
      <div style={{ display: "flex", alignItems: "center", gap: "14px" }}>
        {/* Play/Pause */}
        <button
          data-testid="scrubber-play-pause"
          onClick={isPlaying ? onPause : onPlay}
          style={{
            background: "none",
            border: "1px solid #475569",
            borderRadius: "50%",
            width: "32px",
            height: "32px",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            cursor: "pointer",
            color: "#F1F5F9",
            fontSize: "1rem",
            flexShrink: 0,
          }}
          title={isPlaying ? "Pause" : "Play"}
        >
          {isPlaying ? "⏸" : "▶"}
        </button>

        {/* Time label */}
        <div
          data-testid="scrubber-time-label"
          style={{
            fontFamily: "monospace",
            fontSize: "0.7rem",
            color: "#94A3B8",
            whiteSpace: "nowrap",
            minWidth: "160px",
          }}
        >
          {formatTime(currentMs)}
        </div>

        {/* Slider */}
        <div style={{ flex: 1, display: "flex", flexDirection: "column", gap: "2px" }}>
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              fontSize: "0.6rem",
              color: "#475569",
              fontFamily: "monospace",
            }}
          >
            <span>{formatTime(startMs)}</span>
            <span>{formatTime(endMs)}</span>
          </div>
          <div style={{ position: "relative", height: "6px" }}>
            {/* Progress fill */}
            <div
              style={{
                position: "absolute",
                left: 0,
                top: "50%",
                transform: "translateY(-50%)",
                width: `${pct}%`,
                height: "4px",
                backgroundColor: "#3B82F6",
                borderRadius: "2px",
                pointerEvents: "none",
              }}
            />
            <input
              type="range"
              min={0}
              max={100}
              step={0.01}
              value={pct}
              onChange={handleSlider}
              data-testid="scrubber-slider"
              style={{
                position: "absolute",
                inset: 0,
                width: "100%",
                opacity: 0,
                cursor: "pointer",
                height: "100%",
              }}
            />
            {/* Track background */}
            <div
              style={{
                position: "absolute",
                inset: 0,
                height: "4px",
                top: "50%",
                transform: "translateY(-50%)",
                backgroundColor: "#334155",
                borderRadius: "2px",
                pointerEvents: "none",
              }}
            />
          </div>
        </div>

        {/* Speed selector */}
        <select
          data-testid="scrubber-speed"
          value={speed}
          onChange={(e) => onSpeedChange(Number(e.target.value))}
          style={{
            backgroundColor: "#1E293B",
            color: "#F1F5F9",
            border: "1px solid #475569",
            borderRadius: "4px",
            padding: "4px 6px",
            fontSize: "0.7rem",
            cursor: "pointer",
            flexShrink: 0,
          }}
          title="Playback speed"
        >
          {SPEED_OPTIONS.map((s) => (
            <option key={s} value={s}>
              {s}×
            </option>
          ))}
        </select>

        {/* Live mode badge */}
        <div
          style={{
            fontSize: "0.65rem",
            color: "#10B981",
            backgroundColor: "rgba(16, 185, 129, 0.1)",
            border: "1px solid rgba(16,185,129,0.3)",
            borderRadius: "4px",
            padding: "3px 8px",
            whiteSpace: "nowrap",
            cursor: "pointer",
            flexShrink: 0,
          }}
          title="Jump to live"
          onClick={() => {
            onSeek(endMs);
            onPause();
          }}
        >
          ● LIVE
        </div>

        {/* Close */}
        {onClose && (
          <button
            data-testid="scrubber-close"
            onClick={onClose}
            style={{
              background: "none",
              border: "none",
              color: "#64748B",
              cursor: "pointer",
              fontSize: "1rem",
              flexShrink: 0,
            }}
          >
            ✕
          </button>
        )}
      </div>
    </div>
  );
};
