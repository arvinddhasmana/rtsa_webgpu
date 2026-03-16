// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/OperatorUiCommanderDashboard.tsx

import type { JSX } from "solid-js";

interface OperatorUiCommanderDashboardProps {
  alertColumnContent: JSX.Element;
  detailPaneContent: JSX.Element;
  timelinePaneContent: JSX.Element;
}

/** Operations Commander Operator UI dashboard skeleton (three-pane). */
export function OperatorUiCommanderDashboard(
  props: OperatorUiCommanderDashboardProps,
) {
  return (
    <div
      data-testid="commander-operator-ui-dashboard"
      style={{
        display: "grid",
        "grid-template-columns": "20rem 1fr",
        "grid-template-rows": "1fr 14rem",
        "grid-template-areas": '"alerts detail" "alerts timeline"',
        width: "100%",
        height: "100%",
        background: "#0a0f1a",
        overflow: "hidden",
      }}
    >
      <section
        data-testid="commander-operator-alert-column"
        aria-label="Alert column"
        style={{
          "grid-area": "alerts",
          "border-right": "1px solid #1e2a3a",
          background: "#0d1424",
          overflow: "hidden",
        }}
      >
        {props.alertColumnContent}
      </section>

      <section
        data-testid="commander-operator-detail-pane"
        aria-label="Detail pane"
        style={{
          "grid-area": "detail",
          overflow: "hidden auto",
          "min-width": "0",
        }}
      >
        {props.detailPaneContent}
      </section>

      <section
        data-testid="commander-operator-timeline-pane"
        aria-label="Timeline pane"
        style={{
          "grid-area": "timeline",
          "border-top": "1px solid #1e2a3a",
          background: "#0d1424",
          overflow: "hidden auto",
        }}
      >
        {props.timelinePaneContent}
      </section>
    </div>
  );
}
