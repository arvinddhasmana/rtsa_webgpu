// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/hooks/useClassification.test.ts

import { describe, it, expect, beforeEach } from "vitest";
import { renderHook } from "@testing-library/react";
import { useClassification } from "../../hooks/useClassification";
import { useTrackStore } from "../../stores/trackStore";
import { useAlertStore } from "../../stores/alertStore";
import { FusedTrack } from "../../types/track";
import { AnomalyAlert } from "../../types/alert";

const makeTrack = (classification: FusedTrack["classification"]): FusedTrack => ({
  trackId: "TRK-001",
  entityType: "SURFACE",
  hostileClass: "UNKNOWN",
  position: { latitude: 45, longitude: -60 },
  confidenceScore: 0.9,
  sourceCount: 1,
  sources: [],
  status: "ACTIVE",
  classification,
  createdAt: new Date(),
  updatedAt: new Date(),
});

const makeAlert = (classification: AnomalyAlert["classification"]): AnomalyAlert => ({
  alertId: "ALT-001",
  trackId: "TRK-001",
  anomalyType: "SPEED",
  severity: "WATCH",
  confidenceScore: 0.8,
  explanation: "Test",
  features: [],
  classification,
  detectedAt: new Date(),
});

describe("useClassification", () => {
  beforeEach(() => {
    useTrackStore.getState().clearAll();
    useAlertStore.getState().clearAll();
  });

  it("returns UNCLASSIFIED when no data", () => {
    const { result } = renderHook(() => useClassification());
    expect(result.current.effectiveClassification).toBe("UNCLASSIFIED");
  });

  it("returns highest classification from tracks", () => {
    useTrackStore.getState().upsertTrack(makeTrack("UNCLASSIFIED"));
    useTrackStore.getState().upsertTrack({ ...makeTrack("PROTECTED_B"), trackId: "TRK-002" });
    const { result } = renderHook(() => useClassification());
    expect(result.current.effectiveClassification).toBe("PROTECTED_B");
  });

  it("returns highest classification from alerts", () => {
    useAlertStore.getState().addAlert(makeAlert("PROTECTED_A"));
    useAlertStore.getState().addAlert({ ...makeAlert("SECRET"), alertId: "ALT-002" });
    const { result } = renderHook(() => useClassification());
    expect(result.current.effectiveClassification).toBe("SECRET");
  });

  it("returns highest across tracks and alerts", () => {
    useTrackStore.getState().upsertTrack(makeTrack("PROTECTED_C"));
    useAlertStore.getState().addAlert(makeAlert("PROTECTED_A"));
    const { result } = renderHook(() => useClassification());
    expect(result.current.effectiveClassification).toBe("PROTECTED_C");
  });
});
