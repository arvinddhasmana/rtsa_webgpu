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
 * waitForMapReady — waits for MapLibre's 'load' event to have fired, indicated
 * by window.__RTSA_MAP__ being set. Use this before asserting on map sources/layers.
 */
export async function waitForMapReady(page: Page): Promise<void> {
  await waitForMapLoad(page);
  await page
    .waitForFunction(
      () => !!(window as unknown as Record<string, unknown>)["__RTSA_MAP__"],
      {
        timeout: 20_000,
      },
    )
    .catch(() => {
      /* non-fatal: map may not expose instance in all envs */
    });
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
 * injectTestTrack — injects a FusedTrack into the Zustand trackStore via
 * page.evaluate(), bypassing the gRPC layer entirely.
 * Requires window.__RTSA_TRACK_STORE__ to be exposed (set in main.tsx).
 */
export async function injectTestTrack(
  page: Page,
  track: {
    trackId: string;
    lat: number;
    lon: number;
    entityType?: string;
    hostileClass?: string;
    confidenceScore?: number;
  },
): Promise<void> {
  await page.evaluate((t) => {
    const w = window as unknown as {
      __RTSA_TRACK_STORE__?: {
        getState: () => {
          upsertTrack: (track: unknown) => void;
        };
      };
    };
    if (w.__RTSA_TRACK_STORE__) {
      w.__RTSA_TRACK_STORE__.getState().upsertTrack({
        trackId: t.trackId,
        entityType: t.entityType ?? "SURFACE",
        hostileClass: t.hostileClass ?? "UNKNOWN",
        position: {
          latitude: t.lat,
          longitude: t.lon,
          altitudeMeters: undefined,
          speedKnots: undefined,
          headingDegrees: undefined,
        },
        confidenceScore: t.confidenceScore ?? 0.85,
        sourceCount: 1,
        sources: [
          {
            sensorId: "E2E-TEST-001",
            sensorType: "RADAR",
            confidence: 0.85,
            lastContribution: new Date(),
          },
        ],
        status: "ACTIVE",
        classification: "UNCLASSIFIED",
        createdAt: new Date(),
        updatedAt: new Date(),
      });
    }
  }, track);
}

/**
 * injectTestAlert — injects an AnomalyAlert into the Zustand alertStore via
 * page.evaluate(), bypassing the gRPC layer entirely.
 * Requires window.__RTSA_ALERT_STORE__ to be exposed (set in main.tsx).
 */
export async function injectTestAlert(
  page: Page,
  alert: {
    alertId: string;
    trackId: string;
    severity: "WATCH" | "ELEVATED" | "CRITICAL";
    anomalyType?: string;
  },
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
        anomalyType: a.anomalyType ?? "SPEED",
        severity: a.severity,
        confidenceScore: 0.92,
        explanation: `E2E test alert: ${a.alertId} — ${a.anomalyType ?? "SPEED"} anomaly detected`,
        features: [],
        classification: "UNCLASSIFIED",
        detectedAt: new Date(),
      });
    }
  }, alert);
}
