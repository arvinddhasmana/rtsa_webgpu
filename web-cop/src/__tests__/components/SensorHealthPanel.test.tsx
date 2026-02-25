// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/components/SensorHealthPanel.test.tsx

import React from "react";
import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { SensorHealthPanel } from "../../components/layout/SensorHealthPanel";

describe("SensorHealthPanel", () => {
  it("T13: renders all 6 sensor type cells", () => {
    render(<SensorHealthPanel />);
    expect(screen.getByTestId("sensor-cell-radar")).toBeTruthy();
    expect(screen.getByTestId("sensor-cell-ais")).toBeTruthy();
    expect(screen.getByTestId("sensor-cell-ew")).toBeTruthy();
    expect(screen.getByTestId("sensor-cell-elint")).toBeTruthy();
    expect(screen.getByTestId("sensor-cell-isr")).toBeTruthy();
    expect(screen.getByTestId("sensor-cell-cyber")).toBeTruthy();
  });

  it("T13: renders the panel header", () => {
    render(<SensorHealthPanel />);
    expect(screen.getByText("SENSOR HEALTH")).toBeTruthy();
  });

  it("T13: sensor cells are hidden when panel is collapsed", async () => {
    render(<SensorHealthPanel />);
    const header = screen.getByText("SENSOR HEALTH");
    await userEvent.click(header);
    expect(screen.queryByTestId("sensor-cell-radar")).toBeNull();
  });

  it("T13: sensor cells are shown after expanding collapsed panel", async () => {
    render(<SensorHealthPanel />);
    const header = screen.getByText("SENSOR HEALTH");
    // Collapse
    await userEvent.click(header);
    // Expand
    await userEvent.click(header);
    expect(screen.getByTestId("sensor-cell-radar")).toBeTruthy();
  });
});
