// CLASSIFICATION: UNCLASSIFIED
// src/utils/time.ts

/**
 * Formats a Date as a Zulu (UTC) time string: "2026-02-25T04:30:00Z"
 */
export function formatZulu(date: Date): string {
  return date.toISOString().replace(".000Z", "Z").replace(/\.\d{3}Z$/, "Z");
}

/**
 * Formats a Date as a short Zulu display string: "04:30:00Z"
 */
export function formatZuluTime(date: Date): string {
  const h = String(date.getUTCHours()).padStart(2, "0");
  const m = String(date.getUTCMinutes()).padStart(2, "0");
  const s = String(date.getUTCSeconds()).padStart(2, "0");
  return `${h}:${m}:${s}Z`;
}

/**
 * Returns a human-readable relative time string (e.g. "2 min ago").
 */
export function relativeTime(date: Date): string {
  const diffMs = Date.now() - date.getTime();
  const diffSec = Math.floor(diffMs / 1000);
  if (diffSec < 60) return `${diffSec}s ago`;
  const diffMin = Math.floor(diffSec / 60);
  if (diffMin < 60) return `${diffMin} min ago`;
  const diffHr = Math.floor(diffMin / 60);
  return `${diffHr} hr ago`;
}
