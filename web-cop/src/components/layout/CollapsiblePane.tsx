// CLASSIFICATION: UNCLASSIFIED
// src/components/layout/CollapsiblePane.tsx

import React, { useState } from "react";

interface CollapsiblePaneProps {
  title: string;
  defaultOpen?: boolean;
  children: React.ReactNode;
  width?: string | number;
  height?: string | number;
  direction?: "vertical" | "horizontal";
}

export const CollapsiblePane: React.FC<CollapsiblePaneProps> = ({
  title,
  defaultOpen = true,
  children,
  width,
  height,
  direction = "vertical",
}) => {
  const [isOpen, setIsOpen] = useState(defaultOpen);

  // When horizontal collapse, width becomes strict min-content.
  // When vertical collapse, height becomes fixed 36px header size.
  const collapsedWidth = direction === "horizontal" ? "36px" : width;
  const collapsedHeight = direction === "vertical" ? "36px" : height;

  return (
    <div
      style={{
        width: isOpen ? width : collapsedWidth,
        height: isOpen ? height : collapsedHeight,
        minWidth: isOpen && direction === "horizontal" ? width : "36px",
        minHeight: isOpen && direction === "vertical" ? height : "36px",
        backgroundColor: "var(--glass-bg, #1E293B)",
        backdropFilter: "var(--glass-blur, blur(12px))",
        border: "1px solid var(--glass-border, #334155)",
        display: "flex",
        flexDirection: "column",
        transition: "all 0.3s ease-in-out",
        overflow: "hidden",
      }}
    >
      <div
        onClick={() => setIsOpen(!isOpen)}
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          padding: "8px 12px",
          backgroundColor: "rgba(0, 0, 0, 0.2)",
          cursor: "pointer",
          fontWeight: "600",
          fontSize: "0.85rem",
          userSelect: "none",
          whiteSpace: "nowrap",
          height: "36px",
          borderBottom: isOpen ? "1px solid var(--glass-border, #334155)" : "none",
        }}
        aria-expanded={isOpen}
      >
        {isOpen || direction === "vertical" ? (
          <span>{title}</span>
        ) : (
          <span style={{ transform: "rotate(-90deg)", transformOrigin: "center left", width: "100%", textAlign: "right" }}>{title}</span>
        )}
        <span style={{ color: "var(--color-accent-amber, #f59e0b)" }}>
          {isOpen
            ? direction === "horizontal"
              ? "▶"
              : "▼"
            : direction === "horizontal"
            ? "◀"
            : "▲"}
        </span>
      </div>
      {isOpen && (
        <div style={{ flex: 1, overflow: "auto", display: "flex", flexDirection: "column" }}>
          {children}
        </div>
      )}
    </div>
  );
};
