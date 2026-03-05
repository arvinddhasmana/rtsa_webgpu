// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/components/TimelineView.test.tsx

import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { TimelineView } from "../../components/timeline/TimelineView";
import * as eventTimelineHook from "../../hooks/useEventTimeline";
import { useTrackStore } from "../../stores/trackStore";

// Mock the hook
vi.mock("../../hooks/useEventTimeline", () => ({
  useEventTimeline: vi.fn(),
}));

describe("TimelineView", () => {
  const mockSetSelectedEventId = vi.fn();
  const mockHook = eventTimelineHook.useEventTimeline as any;

  beforeEach(() => {
    vi.clearAllMocks();
    mockHook.mockReturnValue({
      events: [],
      loading: false,
      error: null,
      isRefreshing: false,
    });
  });

  it("shows empty state when no track is selected", () => {
    render(<TimelineView />);
    expect(screen.getByText("Select a track from the Map or Alert Queue to view its historical event timeline.")).toBeTruthy();
  });

  it("shows loading state when fetching initial data", () => {
    // Need to mock the component behavior or store to simulate trackId present
    useTrackStore.setState({ selectedTrackId: "TRK-1" });
    mockHook.mockReturnValue({ events: [], loading: true, error: null, refreshing: false });
    render(<TimelineView />);
    expect(screen.getByText("Loading timeline events...")).toBeTruthy();
  });

  it("shows refreshing indicator when background polling", () => {
    useTrackStore.setState({ selectedTrackId: "TRK-1" });
    mockHook.mockReturnValue({
      events: [{ id: "E1", type: "track", eventTypeStr: "Track Created", eventTime: { seconds: Math.floor(Date.now() / 1000) }, sourceId: "SYS", summary: "Test Track Created", typeColor: "#FFF", icon: "📍" }],
      loading: false,
      error: null,
      refreshing: true
    });
    render(<TimelineView />);
    expect(screen.getByText("Refreshing...")).toBeTruthy();
  });

  it("renders filter chips", () => {
    useTrackStore.setState({ selectedTrackId: "TRK-1" });
    render(<TimelineView />);
    expect(screen.getByText("ALL")).toBeTruthy();
    expect(screen.getByText("TRACK")).toBeTruthy();
    expect(screen.getByText("ANOMALY")).toBeTruthy();
    expect(screen.getByText("FEEDBACK")).toBeTruthy();
    expect(screen.getByText("ALERT")).toBeTruthy();
  });

  it("clicking filter chip changes active filter", () => {
    useTrackStore.setState({ selectedTrackId: "TRK-1" });
    render(<TimelineView />);

    // ANOMALY filter
    const anomalyChip = screen.getByText("ANOMALY");
    fireEvent.click(anomalyChip);

    // The hook should be called with the new filter
    expect(mockHook).toHaveBeenLastCalledWith("TRK-1", "ANOMALY");
  });

  it("renders events", () => {
    useTrackStore.setState({ selectedTrackId: "TRK-1" });
    mockHook.mockReturnValue({
      events: [
        { id: "E1", type: "track", eventTypeStr: "Track Created", eventTime: { seconds: Math.floor(Date.now() / 1000) }, sourceId: "SYS", summary: "Test Track Created", typeColor: "#FFF", icon: "📍" },
        { id: "E2", type: "anomaly", eventTypeStr: "Speed Anomaly", eventTime: { seconds: Math.floor(Date.now() / 1000) }, sourceId: "AI", summary: "Test Speed Anomaly", typeColor: "#F00", icon: "⚠" }
      ],
      loading: false,
      error: null,
      refreshing: false
    });
    render(<TimelineView />);

    expect(screen.getByText("Track Created")).toBeTruthy();
    expect(screen.getByText("Speed Anomaly")).toBeTruthy();
  });
});
