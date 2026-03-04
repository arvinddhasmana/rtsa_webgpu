// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/components/DomainMetricsOverlay.test.tsx

import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it } from "vitest";
import { DomainMetricsOverlay } from "../../components/dashboard/DomainMetricsOverlay";
import { useSensorHealthStore } from "../../stores/sensorHealthStore";
import { useTrackStore } from "../../stores/trackStore";
import { SensorStatus } from "../../types/sensor";
import { FusedTrack } from "../../types/track";

const makeTrack = (overrides: Partial<FusedTrack> = {}): FusedTrack => ({
  trackId: "TRK-001",
  entityType: "AIR",
  hostileClass: "UNKNOWN",
  confidenceScore: 0.8,
  position: { latitude: 0, longitude: 0 },
  status: "ACTIVE",
  classification: "UNCLASSIFIED",
  createdAt: new Date(),
  updatedAt: new Date(),
  sourceCount: 1,
  sources: [],
  ...overrides,
});

const makeSensor = (overrides: Partial<SensorStatus> = {}): SensorStatus => ({
  sensorId: "RADAR-001",
  sensorType: "RADAR",
  connected: true,
  connectionStatus: "connected",
  totalReceived: 100,
  totalAccepted: 100,
  totalRejected: 0,
  lastObservationTime: new Date(),
  eventsPerSecond: 10,
  acceptanceRate: 100,
  rateHistory: [],
  latencyMs: 10,
  ...overrides,
});

describe("DomainMetricsOverlay", () => {
  beforeEach(() => {
    useTrackStore.getState().clearAll();
    useSensorHealthStore.getState().clearAll();
  });

  it("renders without crashing and shows empty stats", () => {
    render(<DomainMetricsOverlay />);
    expect(screen.getByText("MULTI-DOMAIN")).toBeTruthy();
    expect(screen.getByText("TRACKS")).toBeTruthy();
    expect(screen.getAllByText("0").length).toBeGreaterThan(0); // For total tracks
  });

  it("aggregates tracks by domain and hostile status correctly", () => {
    useTrackStore.getState().upsertTrack(makeTrack({ trackId: "T1", entityType: "AIR", hostileClass: "HOSTILE" }));
    useTrackStore.getState().upsertTrack(makeTrack({ trackId: "T2", entityType: "AIR", hostileClass: "FRIENDLY" }));
    useTrackStore.getState().upsertTrack(makeTrack({ trackId: "T3", entityType: "SURFACE", hostileClass: "UNKNOWN" }));

    render(<DomainMetricsOverlay />);

    expect(screen.getByText("AIR")).toBeTruthy();
    expect(screen.getByText("SURFACE")).toBeTruthy();

    // There are 2 AIR tracks
    expect(screen.getAllByText("2").length).toBeGreaterThan(0);
    // 1 SURFACE track (and 1 HOSTILE track total, so there might be multiple '1's)
    expect(screen.getAllByText("1").length).toBeGreaterThan(0);
    // 1 Hostile AIR track
    expect(screen.getByText("1 HST")).toBeTruthy();
  });

  it("aggregates sensor eventsPerSecond into the correct domains", () => {
    useSensorHealthStore.getState().upsertSensors([
      makeSensor({ sensorId: "S1", sensorType: "RADAR", eventsPerSecond: 50 }),
      makeSensor({ sensorId: "S2", sensorType: "AIS", eventsPerSecond: 20 }),
    ]);

    render(<DomainMetricsOverlay />);

    // RADAR maps to AIR and SURFACE
    // AIS maps to SURFACE
    // AIR = 50 OBS/S
    expect(screen.getByText("50")).toBeTruthy();
    // SURFACE = 50 (from RADAR) + 20 (from AIS) = 70 OBS/S
    expect(screen.getAllByText("70").length).toBeGreaterThan(0);
  });

  it("can be collapsed and expanded", () => {
    useTrackStore.getState().upsertTrack(makeTrack({ trackId: "T1", entityType: "AIR" }));
    render(<DomainMetricsOverlay />);

    // Initially expanded
    expect(screen.getByText("AIR")).toBeTruthy();

    // Click to collapse
    fireEvent.click(screen.getByText("MULTI-DOMAIN"));
    expect(screen.queryByText("AIR")).toBeNull();

    // Click to expand
    fireEvent.click(screen.getByText("MULTI-DOMAIN"));
    expect(screen.getByText("AIR")).toBeTruthy();
  });
});
