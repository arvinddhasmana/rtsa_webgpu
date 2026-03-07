// CLASSIFICATION: UNCLASSIFIED
// tests/components/AppShell.test.tsx

import { describe, it, expect } from "vitest";
import { render, screen } from "@solidjs/testing-library";
import { AppShell } from "../../src/components/shell/AppShell";

describe("AppShell", () => {
  it("renders toolbar and has correct data-testid", () => {
    render(() => (
      <AppShell
        toolbar={<div data-testid="mock-toolbar">Toolbar</div>}
        canvas={<div>Canvas</div>}
        rightPanel={<div>Right</div>}
        bottomPanel={<div>Bottom</div>}
      />
    ));

    expect(screen.getByTestId("app-toolbar")).toBeInTheDocument();
    expect(screen.getByTestId("mock-toolbar")).toBeInTheDocument();
  });
});
