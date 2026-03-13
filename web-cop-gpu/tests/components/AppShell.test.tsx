// CLASSIFICATION: UNCLASSIFIED
// tests/components/AppShell.test.tsx

import { render, screen } from "@solidjs/testing-library";
import { describe, expect, it } from "vitest";
import { AppShell } from "../../src/components/shell/AppShell";

describe("AppShell", () => {
  it("renders app-header with data-testid when headerBar prop is provided", () => {
    render(() => (
      <AppShell
        headerBar={<div>header content</div>}
        canvas={<div>canvas</div>}
        rightPanel={<div>right</div>}
        bottomPanel={<div>bottom</div>}
      />
    ));
    expect(screen.getByTestId("app-header")).toBeDefined();
  });

  it("renders headerBar slot content inside app-header", () => {
    render(() => (
      <AppShell
        headerBar={<div>My Header Bar</div>}
        canvas={<div>canvas</div>}
        rightPanel={<div>right</div>}
        bottomPanel={<div>bottom</div>}
      />
    ));
    const header = screen.getByTestId("app-header");
    expect(header.textContent).toContain("My Header Bar");
  });

  it("does not render app-header when headerBar prop is omitted", () => {
    render(() => (
      <AppShell canvas={<div>canvas</div>} bottomPanel={<div>bottom</div>} />
    ));
    expect(screen.queryByTestId("app-header")).toBeNull();
  });

  it("does not render app-toolbar", () => {
    render(() => (
      <AppShell
        headerBar={<div>header</div>}
        canvas={<div>canvas</div>}
        bottomPanel={<div>bottom</div>}
      />
    ));
    expect(screen.queryByTestId("app-toolbar")).toBeNull();
  });
});
