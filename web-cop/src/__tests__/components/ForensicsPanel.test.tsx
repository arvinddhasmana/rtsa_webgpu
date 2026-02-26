// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/components/ForensicsPanel.test.tsx

import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { queryClient } from "../../api/query-client";
import { ForensicsPanel } from "../../components/forensics/ForensicsPanel";

// Mock the queryClient
vi.mock("../../api/query-client", () => {
  return {
    queryClient: {
      queryHistory: vi.fn(),
    },
  };
});

// Mock MapReplay component since it uses maplibre-gl
vi.mock("../../components/forensics/MapReplay", () => ({
  MapReplay: () => <div data-testid="map-replay">Map Replay</div>,
}));

describe("ForensicsPanel", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders the panel header", () => {
    render(<ForensicsPanel />);
    expect(screen.getByText("FORENSICS")).toBeInTheDocument();
  });

  it("calls queryClient.queryHistory on form submission", async () => {
    // Mock successful response
    const mockResponse = {
      tracks: [
        {
          trackId: "T1",
          entityType: "SHIP",
          hostileClass: "FRIENDLY",
          latitude: 45,
          longitude: -60,
          status: "ACTIVE",
          confidence: 0.9,
          lastUpdateTime: new Date(),
          sourceSensors: ["RADAR-1"],
        },
      ],
      alerts: [
        {
          alertId: "A1",
          trackId: "T1",
          anomalyType: "KINEMATIC",
          severity: "CRITICAL",
          description: "Test alert",
          detectedAt: new Date("2024-01-01T12:00:00Z"),
          confidenceScore: 0.9,
        },
      ],
      totalCount: 1,
    };

    (queryClient.queryHistory as any).mockResolvedValue(mockResponse);

    render(<ForensicsPanel />);

    const runButton = screen.getByRole("button", { name: "Run Query" });
    fireEvent.click(runButton);

    await waitFor(() => {
      expect(queryClient.queryHistory).toHaveBeenCalled();
    });

    // ResultsView renders alerts, not tracks directly
    expect(await screen.findByText("T1")).toBeInTheDocument();
    expect(screen.getByTestId("map-replay")).toBeInTheDocument();
  });

  it("displays error message on query failure", async () => {
    (queryClient.queryHistory as any).mockRejectedValue(
      new Error("Query failed"),
    );

    render(<ForensicsPanel />);

    const runButton = screen.getByRole("button", { name: "Run Query" });
    fireEvent.click(runButton);

    expect(await screen.findByText("Error: Query failed")).toBeInTheDocument();
  });
});
