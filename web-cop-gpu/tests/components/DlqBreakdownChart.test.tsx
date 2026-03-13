// CLASSIFICATION: UNCLASSIFIED
// tests/components/DlqBreakdownChart.test.tsx

import { render, screen } from "@solidjs/testing-library";
import { describe, expect, it } from "vitest";
import { DlqBreakdownChart } from "../../src/components/dashboard/DlqBreakdownChart";
import type { DlqReasonBar } from "../../src/components/dashboard/DlqBreakdownChart";

const reasons: DlqReasonBar[] = [
  { reason: "Schema Mismatch", percentage: 45.0, count: 90 },
  { reason: "CRC Error", percentage: 30.0, count: 60 },
  { reason: "Rate Limit", percentage: 15.0, count: 30 },
  { reason: "Unknown", percentage: 10.0, count: 20 },
];

describe("DlqBreakdownChart", () => {
  it("renders container", () => {
    render(() => <DlqBreakdownChart reasons={reasons} />);
    expect(screen.getByTestId("dlq-breakdown-chart")).toBeDefined();
  });

  it("renders correct labels", () => {
    render(() => <DlqBreakdownChart reasons={reasons} />);
    expect(screen.getByText("Schema Mismatch")).toBeDefined();
    expect(screen.getByText("CRC Error")).toBeDefined();
    expect(screen.getByText("Rate Limit")).toBeDefined();
    expect(screen.getByText("Unknown")).toBeDefined();
  });

  it("renders percentage values", () => {
    render(() => <DlqBreakdownChart reasons={reasons} />);
    expect(screen.getByText("45.0%")).toBeDefined();
    expect(screen.getByText("30.0%")).toBeDefined();
  });

  it("renders bar rows with testid for each reason", () => {
    render(() => <DlqBreakdownChart reasons={reasons} />);
    expect(screen.getByTestId("dlq-bar-schema-mismatch")).toBeDefined();
    expect(screen.getByTestId("dlq-bar-crc-error")).toBeDefined();
    expect(screen.getByTestId("dlq-bar-rate-limit")).toBeDefined();
    expect(screen.getByTestId("dlq-bar-unknown")).toBeDefined();
  });

  it("renders 'No DLQ rejections' when empty", () => {
    render(() => <DlqBreakdownChart reasons={[]} />);
    expect(screen.getByText("No DLQ rejections")).toBeDefined();
  });
});
