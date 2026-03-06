// CLASSIFICATION: UNCLASSIFIED
// src/components/layout/NatoExchangeDashboard.tsx
//
// NATO Liaison — NATO Exchange Dashboard.
//
// Layout:
//   TOP:    LinkStatusHeader (Link 16, NFFI, MIP, STANAG 5516 indicators)
//   LEFT:   NominationQueue (outbound tracks for N-COP)
//   MAP:    Live map with NATO-shared tracks overlaid
//   RIGHT:  InboundTracksPanel (allied tracks)

import React from "react";
import { useTrackStore } from "../../stores/trackStore";
import { InboundTracksPanel } from "../nato/InboundTracksPanel";
import { LinkStatusHeader } from "../nato/LinkStatusHeader";
import { NominationQueue } from "../nato/NominationQueue";
import { MapView } from "../map/MapView";
import { CollapsiblePane } from "./CollapsiblePane";

/**
 * NatoExchangeDashboard — NATO Liaison default view.
 *
 * Wraps all four NATO sub-panels around the live map.
 */
export const NatoExchangeDashboard: React.FC = () => {
  const currentTracksMap = useTrackStore((s) => s.tracks);
  const tracks = Array.from(currentTracksMap.values());

  const leftCollapsed = false;
  const rightCollapsed = false;

  return (
    <div
      data-testid="nato-exchange-dashboard"
      style={{
        flex: 1,
        display: "flex",
        flexDirection: "column",
        overflow: "hidden",
      }}
    >
      {/* Link Status Header */}
      <LinkStatusHeader />

      {/* Main 3-column body */}
      <div style={{ flex: 1, display: "flex", overflow: "hidden" }}>
        {/* Left: Nomination Queue */}
        <CollapsiblePane
          title="Track Nomination"
          width={leftCollapsed ? "32px" : "300px"}
          height="100%"
          direction="horizontal"
        >
          {!leftCollapsed && <NominationQueue tracks={tracks} />}
        </CollapsiblePane>

        {/* Centre: Map */}
        <div
          style={{ flex: 1, overflow: "hidden", position: "relative" }}
          aria-label="Map View"
          role="region"
          tabIndex={0}
        >
          <MapView />

          {/* Exchange log overlay (bottom of map) */}
          <div
            style={{
              position: "absolute",
              bottom: 0,
              left: 0,
              right: 0,
              backgroundColor: "rgba(15, 23, 42, 0.88)",
              backdropFilter: "blur(8px)",
              borderTop: "1px solid #334155",
              padding: "6px 14px",
              maxHeight: "80px",
              overflowY: "auto",
            }}
          >
            <div
              style={{
                fontSize: "0.6rem",
                fontFamily: "monospace",
                lineHeight: "1.6",
                color: "#64748B",
              }}
            >
              <span style={{ color: "#10B981" }}>
                [TX] J3.2 Air Track (Trk 1045) → NATO N-COP Gateway
              </span>
              <br />
              [RX] ACK from N-COP Gateway — track accepted
              <br />
              <span style={{ color: "#10B981" }}>
                [RX] J3.1 Maritime Track (Trk 0922) ← GBR
              </span>
              <br />
              <span style={{ color: "#EF4444" }}>
                [TX DROP] Track 1048 — Classification mismatch (PROTECTED_B
                &gt; sharing ceiling)
              </span>
            </div>
          </div>
        </div>

        {/* Right: Inbound Allied Tracks */}
        <CollapsiblePane
          title="Inbound Tracks"
          width={rightCollapsed ? "32px" : "300px"}
          height="100%"
          direction="horizontal"
        >
          {!rightCollapsed && <InboundTracksPanel />}
        </CollapsiblePane>
      </div>
    </div>
  );
};
