// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/CommanderDashboards.test.tsx

import { render, screen } from "@solidjs/testing-library";
import { describe, expect, it } from "vitest";
import { FusionCommanderDashboard } from "./FusionCommanderDashboard";
import { MultiDomainCommanderDashboard } from "./MultiDomainCommanderDashboard";
import { OperatorUiCommanderDashboard } from "./OperatorUiCommanderDashboard";

describe("Commander dashboard skeletons", () => {
  it("renders Fusion dashboard required containers and layer mounts", () => {
    render(() => <FusionCommanderDashboard mapContent={<div>map</div>} />);
    expect(screen.getByTestId("commander-fusion-map-container")).toBeTruthy();
    expect(screen.getByTestId("commander-fusion-side-panel")).toBeTruthy();
    expect(
      screen.getByTestId("commander-observation-layer-mount"),
    ).toBeTruthy();
    expect(screen.getByTestId("commander-fused-layer-mount")).toBeTruthy();
  });

  it("renders Multi-Domain dashboard KPI overlay and layer toggles", () => {
    render(() => <MultiDomainCommanderDashboard mapContent={<div>map</div>} />);
    expect(
      screen.getByTestId("commander-multidomain-kpi-overlay"),
    ).toBeTruthy();
    expect(
      screen.getByTestId("commander-multidomain-layer-toggles"),
    ).toBeTruthy();
    expect(
      screen.getByTestId("commander-multidomain-map-container"),
    ).toBeTruthy();
  });

  it("renders Operator UI dashboard alert/detail/timeline panes", () => {
    render(() => (
      <OperatorUiCommanderDashboard
        alertColumnContent={<div>alerts</div>}
        detailPaneContent={<div>detail</div>}
        timelinePaneContent={<div>timeline</div>}
      />
    ));
    expect(screen.getByTestId("commander-operator-alert-column")).toBeTruthy();
    expect(screen.getByTestId("commander-operator-detail-pane")).toBeTruthy();
    expect(screen.getByTestId("commander-operator-timeline-pane")).toBeTruthy();
  });
});
