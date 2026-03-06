// CLASSIFICATION: UNCLASSIFIED
// tests/components/SearchOverlay.test.tsx

import { describe, it, expect, afterEach, vi } from "vitest";
import { render, screen, fireEvent } from "@solidjs/testing-library";
import { SearchOverlay } from "../../src/components/search/SearchOverlay";
import { searchOpen, setSearchOpen } from "../../src/signals/viewport";

// Mock the query service
vi.mock("../../src/services/query", () => ({
  searchTracks: vi.fn().mockResolvedValue([]),
  fetchTrackDetail: vi.fn().mockResolvedValue(null),
  fetchTimeline: vi.fn().mockResolvedValue({ events: [] }),
}));

afterEach(() => {
  setSearchOpen(false);
});

describe("SearchOverlay", () => {
  it("renders nothing when searchOpen is false", () => {
    render(() => <SearchOverlay />);
    // The SearchOverlay registers keyboard listeners but renders nothing visually
    expect(screen.queryByRole("dialog")).toBeNull();
  });

  it("renders search dialog when searchOpen is true", () => {
    setSearchOpen(true);
    render(() => <SearchOverlay />);
    expect(screen.getByRole("dialog")).toBeDefined();
  });

  it("shows search input", () => {
    setSearchOpen(true);
    render(() => <SearchOverlay />);
    expect(screen.getByLabelText("Track search input")).toBeDefined();
  });

  it("shows Esc hint", () => {
    setSearchOpen(true);
    render(() => <SearchOverlay />);
    expect(screen.getByText("Press Esc to close")).toBeDefined();
  });

  it("closes on Escape key press", () => {
    setSearchOpen(true);
    render(() => <SearchOverlay />);
    fireEvent.keyDown(window, { key: "Escape" });
    expect(searchOpen()).toBe(false);
  });

  it("opens on Ctrl+K", () => {
    render(() => <SearchOverlay />);
    fireEvent.keyDown(window, { key: "k", ctrlKey: true });
    expect(searchOpen()).toBe(true);
  });
});
