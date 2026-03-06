// CLASSIFICATION: UNCLASSIFIED
// src/components/nato/LinkStatusHeader.tsx
//
// Link 16 / NFFI / MIP connectivity status bar for NATO Liaison view.

import React from "react";

interface LinkStatus {
  name: string;
  status: "ACTIVE" | "DEGRADED" | "OFFLINE" | "SYNCED";
  detail?: string;
}

interface LinkStatusHeaderProps {
  links?: LinkStatus[];
}

const STATUS_COLOR: Record<string, string> = {
  ACTIVE:   "#10B981",
  SYNCED:   "#10B981",
  DEGRADED: "#F59E0B",
  OFFLINE:  "#EF4444",
};

const STATUS_DOT: Record<string, string> = {
  ACTIVE:   "●",
  SYNCED:   "●",
  DEGRADED: "◐",
  OFFLINE:  "○",
};

const DEFAULT_LINKS: LinkStatus[] = [
  { name: "Link 16", status: "ACTIVE",   detail: "Latency 42ms" },
  { name: "NFFI",    status: "SYNCED",   detail: "Last sync 12s ago" },
  { name: "MIP",     status: "DEGRADED", detail: "Partial connectivity" },
  { name: "STANAG 5516", status: "ACTIVE", detail: "GMTI active" },
];

/**
 * LinkStatusHeader — horizontal bar showing NATO link connectivity status.
 * Displays Link 16, NFFI, MIP, and STANAG 5516 indicators.
 */
export const LinkStatusHeader: React.FC<LinkStatusHeaderProps> = ({
  links = DEFAULT_LINKS,
}) => {
  return (
    <div
      data-testid="link-status-header"
      style={{
        display: "flex",
        alignItems: "center",
        gap: "16px",
        padding: "6px 16px",
        backgroundColor: "rgba(15, 23, 42, 0.95)",
        borderBottom: "1px solid #334155",
        flexShrink: 0,
      }}
    >
      <span
        style={{
          fontSize: "0.65rem",
          fontWeight: "bold",
          color: "#A855F7",
          letterSpacing: "0.06em",
          marginRight: "4px",
        }}
      >
        NATO LINKS
      </span>

      {links.map((link) => {
        const color = STATUS_COLOR[link.status] ?? "#64748B";
        const dot = STATUS_DOT[link.status] ?? "○";
        return (
          <div
            key={link.name}
            data-testid={`link-status-${link.name.replace(/\s+/g, "-")}`}
            title={link.detail}
            style={{
              display: "flex",
              alignItems: "center",
              gap: "5px",
              fontSize: "0.7rem",
              color: "#94A3B8",
            }}
          >
            <span style={{ color, fontSize: "0.8rem" }}>{dot}</span>
            <span style={{ fontWeight: "bold", color }}>{link.name}</span>
            <span style={{ color: "#475569", fontSize: "0.6rem" }}>
              {link.status}
            </span>
          </div>
        );
      })}
    </div>
  );
};
