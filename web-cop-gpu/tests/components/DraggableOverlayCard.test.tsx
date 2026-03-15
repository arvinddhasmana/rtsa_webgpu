// CLASSIFICATION: UNCLASSIFIED
// tests/components/DraggableOverlayCard.test.tsx

import { fireEvent, render, screen } from "@solidjs/testing-library";
import { describe, expect, it } from "vitest";
import { DraggableOverlayCard } from "../../src/components/dashboard/DraggableOverlayCard";

describe("DraggableOverlayCard", () => {
  it("renders title", () => {
    render(() => (
      <DraggableOverlayCard title="Sensor Fleet">
        <div>content</div>
      </DraggableOverlayCard>
    ));
    expect(screen.getByText("Sensor Fleet")).toBeDefined();
  });

  it("has correct data-testid", () => {
    render(() => (
      <DraggableOverlayCard title="Fleet List">
        <div>content</div>
      </DraggableOverlayCard>
    ));
    expect(screen.getByTestId("overlay-card-fleet-list")).toBeDefined();
  });

  it("renders children when not minimized", () => {
    render(() => (
      <DraggableOverlayCard title="Test">
        <div data-testid="panel-content">panel</div>
      </DraggableOverlayCard>
    ));
    expect(screen.getByTestId("panel-content")).toBeDefined();
  });

  it("hides children when minimized via minimize button", () => {
    render(() => (
      <DraggableOverlayCard title="Test Panel" scrollKey="test">
        <div data-testid="panel-content">panel</div>
      </DraggableOverlayCard>
    ));
    // Initially visible
    expect(screen.getByTestId("panel-content")).toBeDefined();
    // Click minimize button
    const btn = screen.getByTestId("overlay-minimize-test-panel");
    fireEvent.click(btn);
    // Now hidden
    expect(screen.queryByTestId("panel-content")).toBeNull();
  });

  it("restores children after minimize then restore", () => {
    render(() => (
      <DraggableOverlayCard title="Test Panel" scrollKey="test2">
        <div data-testid="panel-content2">panel</div>
      </DraggableOverlayCard>
    ));
    const btn = screen.getByTestId("overlay-minimize-test-panel");
    // minimize
    fireEvent.click(btn);
    expect(screen.queryByTestId("panel-content2")).toBeNull();
    // restore
    fireEvent.click(btn);
    expect(screen.getByTestId("panel-content2")).toBeDefined();
  });

  it("renders minimize button with correct title", () => {
    render(() => (
      <DraggableOverlayCard title="My Panel" scrollKey="mp">
        <div>content</div>
      </DraggableOverlayCard>
    ));
    const btn = screen.getByTitle("Minimize panel");
    expect(btn).toBeDefined();
  });
});
