// CLASSIFICATION: UNCLASSIFIED
// tests/signals/alerts.test.ts

import { describe, it, expect, afterEach } from "vitest";
import {
  alerts,
  setAlerts,
  updateAlerts,
  acknowledgeAlertLocally,
} from "../../src/signals/alerts";
import type { AlertPayload } from "../../src/workers/shared-protocol";

const makeAlert = (id: string, severity: AlertPayload["severity"] = "WATCH"): AlertPayload => ({
  alertId: id,
  trackId: `track-${id}`,
  severity,
  description: `Alert ${id}`,
  detectedAtMs: Date.now(),
  acknowledged: false,
});

afterEach(() => {
  setAlerts([]);
});

describe("alerts signal", () => {
  it("starts with empty list", () => {
    expect(alerts()).toHaveLength(0);
  });

  it("updateAlerts replaces the list", () => {
    const incoming = [makeAlert("1"), makeAlert("2")];
    updateAlerts(incoming);
    expect(alerts()).toHaveLength(2);
  });

  it("updateAlerts sorts CRITICAL before WATCH", () => {
    const w = makeAlert("w", "WATCH");
    const c = makeAlert("c", "CRITICAL");
    updateAlerts([w, c]);
    expect(alerts()[0].severity).toBe("CRITICAL");
    expect(alerts()[1].severity).toBe("WATCH");
  });

  it("acknowledgeAlertLocally marks alert acknowledged", () => {
    const a = makeAlert("a1");
    setAlerts([a]);
    acknowledgeAlertLocally("a1");
    expect(alerts()[0].acknowledged).toBe(true);
  });

  it("acknowledgeAlertLocally leaves other alerts unchanged", () => {
    setAlerts([makeAlert("a1"), makeAlert("a2")]);
    acknowledgeAlertLocally("a1");
    expect(alerts()[1].acknowledged).toBe(false);
  });
});
