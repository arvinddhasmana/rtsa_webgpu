// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/OperatorUiCommanderDashboard.tsx

import type { JSX } from "solid-js";

interface OperatorUiCommanderDashboardProps {
  alertColumnContent: JSX.Element;
  detailPaneContent: JSX.Element;
}

/** Operations Commander Operator UI dashboard skeleton (two-pane). */
export function OperatorUiCommanderDashboard(
  props: OperatorUiCommanderDashboardProps,
) {
  return (
    <div
      data-testid="commander-operator-ui-dashboard"
      style={{
        display: "flex",
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
          width: "20rem",
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
          flex: 1,
          overflow: "hidden auto",
          "min-width": "0",
        }}
      >
        {props.detailPaneContent}
      </section>
    </div>
  );
}
