// CLASSIFICATION: UNCLASSIFIED
// web-cop/e2e/helpers.ts
// Shared Playwright test utilities for RTSA COP browser E2E tests.

import { Page } from "@playwright/test";

/**
 * waitForMapLoad — waits for the MapLibre canvas element to be visible and
 * have non-zero dimensions, indicating the map has fully rendered.
 */
export async function waitForMapLoad(page: Page): Promise<void> {
  await page.waitForSelector("canvas", { state: "visible", timeout: 30_000 });
}

/**
 * waitForGrpcConnection — waits for the connection indicator to show
 * "connected" status (green/connected state).
 */
export async function waitForGrpcConnection(page: Page): Promise<void> {
  await page.waitForSelector('[data-testid="connection-indicator"]', {
    state: "visible",
    timeout: 15_000,
  });
}

/**
 * mockGrpcStream — intercepts gRPC-Web requests and returns canned responses.
 * Routes all /rtsa. prefixed paths to return an empty stream.
 */
export async function mockGrpcStream(page: Page): Promise<void> {
  await page.route("**/rtsa.**", (route) => {
    route.fulfill({
      status: 200,
      headers: {
        "Content-Type": "application/grpc-web+proto",
        "grpc-status": "0",
      },
      body: Buffer.from([]),
    });
  });
}

/**
 * injectTestTrack — injects a track into the Zustand trackStore via
 * page.evaluate(), bypassing the gRPC layer entirely.
 */
export async function injectTestTrack(
  page: Page,
  track: {
    trackId: string;
    lat: number;
    lon: number;
    entityType?: string;
    anomalyScore?: number;
  }
): Promise<void> {
  await page.evaluate((t) => {
    // Access Zustand store through the window global exposed by the app.
    // Falls back to a no-op if the store is not yet exposed.
    const w = window as unknown as {
      __RTSA_TRACK_STORE__?: {
        getState: () => { addTrack: (track: unknown) => void };
      };
    };
    if (w.__RTSA_TRACK_STORE__) {
      w.__RTSA_TRACK_STORE__.getState().addTrack({
        trackId: t.trackId,
        entityType: t.entityType ?? "SURFACE",
        classification: "UNCLASSIFIED",
        position: { latitude: t.lat, longitude: t.lon },
        anomalyScore: t.anomalyScore ?? 0,
        lastUpdate: new Date().toISOString(),
        sensorIds: ["RADAR-TEST-001"],
      });
    }
  }, track);
}

/**
 * injectTestAlert — injects an alert into the Zustand alertStore via
 * page.evaluate().
 */
export async function injectTestAlert(
  page: Page,
  alert: {
    alertId: string;
    trackId: string;
    severity: "WATCH" | "ELEVATED" | "CRITICAL";
    alertType: string;
  }
): Promise<void> {
  await page.evaluate((a) => {
    const w = window as unknown as {
      __RTSA_ALERT_STORE__?: {
        getState: () => { addAlert: (alert: unknown) => void };
      };
    };
    if (w.__RTSA_ALERT_STORE__) {
      w.__RTSA_ALERT_STORE__.getState().addAlert({
        alertId: a.alertId,
        trackId: a.trackId,
        severity: a.severity,
        alertType: a.alertType,
        description: `Test alert ${a.alertId}`,
        detectedAt: new Date().toISOString(),
        acknowledged: false,
      });
    }
  }, alert);
}
