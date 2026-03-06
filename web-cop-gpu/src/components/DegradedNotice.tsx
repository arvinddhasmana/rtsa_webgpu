// CLASSIFICATION: UNCLASSIFIED
// src/components/DegradedNotice.tsx — Degraded mode notice

import { JSX } from "solid-js";
import { type Capabilities } from "../services/capabilities";

const CAPABILITY_LABELS: Record<keyof Capabilities, string> = {
  webgpu: "WebGPU",
  webtransport: "WebTransport",
  sharedArrayBuffer: "SharedArrayBuffer",
  offscreenCanvas: "OffscreenCanvas",
};

/**
 * Render a static degraded-mode notice listing missing capabilities.
 * Called when one or more required browser APIs are unavailable.
 */
export function renderDegradedNotice(caps: Capabilities): JSX.Element {
  const missing = (Object.keys(caps) as Array<keyof Capabilities>).filter(
    (k) => !caps[k],
  );

  return (
    <div
      role="alert"
      style={{
        padding: "2rem",
        background: "#1a1f2e",
        border: "1px solid #f59e0b",
        "border-radius": "8px",
        margin: "2rem",
        "max-width": "480px",
      }}
    >
      <h2
        style={{
          color: "#f59e0b",
          margin: "0 0 1rem 0",
          "font-size": "1.2rem",
        }}
      >
        Browser Not Supported
      </h2>
      <p style={{ color: "#e2e8f0", margin: "0 0 0.75rem 0" }}>
        The following required browser capabilities are unavailable:
      </p>
      <ul style={{ color: "#e2e8f0", margin: "0 0 1rem 0", "padding-left": "1.5rem" }}>
        {missing.map((k) => (
          <li>{CAPABILITY_LABELS[k]}</li>
        ))}
      </ul>
      <p style={{ color: "#94a3b8", margin: 0, "font-size": "0.875rem" }}>
        Please use a recent version of Google Chrome or Microsoft Edge with
        WebGPU enabled.
      </p>
    </div>
  );
}
