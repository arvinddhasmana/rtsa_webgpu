// CLASSIFICATION: UNCLASSIFIED
// tests/components/AppShell.test.tsx

import { describe, it, expect } from "vitest";
import { render, screen } from "@solidjs/testing-library";
import { AppShell } from "../../src/components/shell/AppShell";

describe("AppShell", () => {
  it("renders the app-toolbar with data-testid", () => {
    render(() => (
      <AppShell
        toolbar={<div>toolbar</div>}
        canvas={<div>canvas</div>}
        rightPanel={<div>right</div>}
        bottomPanel={<div>bottom</div>}
      />
    ));
    expect(screen.getByTestId("app-toolbar")).toBeDefined();
  });

  it("renders toolbar slot content inside app-toolbar", () => {
    render(() => (
      <AppShell
        toolbar={<div>My Toolbar</div>}
        canvas={<div>canvas</div>}
        rightPanel={<div>right</div>}
        bottomPanel={<div>bottom</div>}
      />
    ));
    expect(screen.getByText("My Toolbar")).toBeDefined();
  });
});
