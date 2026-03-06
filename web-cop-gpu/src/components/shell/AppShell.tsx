// CLASSIFICATION: UNCLASSIFIED
// src/components/shell/AppShell.tsx
//
// Root layout grid: classification banner (top) → toolbar (left) → canvas
// (centre) → panels (right/bottom) → classification banner (bottom).
// Reference: docs/implementation/v4/phase3_ui_interaction.md §3 U3-1

import type { JSX } from "solid-js";
import { ClassificationBanner } from "./ClassificationBanner";

interface AppShellProps {
  toolbar: JSX.Element;
  canvas: JSX.Element;
  rightPanel: JSX.Element;
  bottomPanel: JSX.Element;
  overlay?: JSX.Element;
}

/** Root CSS-grid application shell. Never destructure props. */
export function AppShell(props: AppShellProps) {
  return (
    <div
      style={{
        display: "flex",
        "flex-direction": "column",
        width: "100vw",
        height: "100vh",
        background: "#0a0f1a",
        color: "#e2e8f0",
        overflow: "hidden",
        "font-family": "monospace",
      }}
    >
      {/* Top classification banner */}
      <ClassificationBanner />

      {/* Main content row */}
      <div
        style={{
          display: "flex",
          flex: "1",
          overflow: "hidden",
          position: "relative",
        }}
      >
        {/* Left toolbar */}
        <div
          style={{
            width: "14rem",
            "flex-shrink": "0",
            background: "#0d1424",
            "border-right": "1px solid #1e2a3a",
            display: "flex",
            "flex-direction": "column",
            overflow: "hidden auto",
          }}
        >
          {props.toolbar}
        </div>

        {/* Centre: canvas + bottom status */}
        <div
          style={{
            flex: "1",
            display: "flex",
            "flex-direction": "column",
            overflow: "hidden",
            position: "relative",
          }}
        >
          <div style={{ flex: "1", position: "relative", overflow: "hidden" }}>
            {props.canvas}
          </div>
          <div
            style={{
              "flex-shrink": "0",
              background: "#0d1424",
              "border-top": "1px solid #1e2a3a",
            }}
          >
            {props.bottomPanel}
          </div>
        </div>

        {/* Right panels */}
        <div
          style={{
            width: "22rem",
            "flex-shrink": "0",
            background: "#0d1424",
            "border-left": "1px solid #1e2a3a",
            display: "flex",
            "flex-direction": "column",
            overflow: "hidden auto",
          }}
        >
          {props.rightPanel}
        </div>

        {/* Overlay (search, modals) */}
        {props.overlay}
      </div>

      {/* Bottom classification banner */}
      <ClassificationBanner />
    </div>
  );
}
