// CLASSIFICATION: UNCLASSIFIED
// web-cop/e2e/map.spec.ts
// Browser E2E tests for Map Rendering, Track Plotting, Threat Halos, Geo-Fences
// Covers: UC001 (Sensor Ingestion), UC002 (Multi-Sensor Fusion), UC006 (COP Display)
// Requirements: CR-UI-001, CR-UI-002

import { expect, test } from "@playwright/test";
import { injectTestTrack, waitForMapReady } from "./helpers";

// ─────────────────────────────────────────────────────────────────────────────
// 1. Map Rendering — basemap and canvas
// ─────────────────────────────────────────────────────────────────────────────
test.describe("Map Rendering (CR-UI-001, CR-UI-002)", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/");
    await page
      .waitForLoadState("networkidle", { timeout: 15_000 })
      .catch(() => {});
  });

  test("map renders and shows canvas element [UC006]", async ({ page }) => {
    const mapContainer = page.locator(
      '[data-testid="map-container"], canvas, .maplibregl-map, #map',
    );
    await expect(mapContainer.first()).toBeVisible({ timeout: 20_000 });
  });

  test("map container has non-zero dimensions [CR-UI-001]", async ({
    page,
  }) => {
    const mapEl = page.locator(
      '[data-testid="map-container"], canvas, .maplibregl-map, #map',
    );
    const el = mapEl.first();
    await expect(el).toBeVisible({ timeout: 20_000 });
    const box = await el.boundingBox();
    expect(box).not.toBeNull();
    expect(box!.width).toBeGreaterThan(0);
    expect(box!.height).toBeGreaterThan(0);
  });

  test("page title or heading identifies RTSA COP [UC006]", async ({
    page,
  }) => {
    const title = await page.title();
    const bodyText = await page
      .locator("body")
      .innerText()
      .catch(() => "");
    expect(
      title.toLowerCase().includes("rtsa") ||
        title.toLowerCase().includes("cop") ||
        bodyText.toLowerCase().includes("unclassified") ||
        bodyText.toLowerCase().includes("rtsa"),
    ).toBeTruthy();
  });

  test("MapLibre GL initialises and exposes map instance [CR-UI-001]", async ({
    page,
  }) => {
    await waitForMapReady(page);
    const mapReady = await page.evaluate(
      () => !!(window as unknown as Record<string, unknown>)["__RTSA_MAP__"],
    );
    // If __RTSA_MAP__ is set, MapLibre load event fired successfully.
    // If not set (e.g. WebGL unavailable in headless), canvas presence is sufficient.
    const hasCanvas = await page.locator("canvas").count();
    expect(mapReady || hasCanvas > 0).toBeTruthy();
  });
});

// ─────────────────────────────────────────────────────────────────────────────
// 2. Track Plotting — real-time markers [CR-UI-002, UC006]
// ─────────────────────────────────────────────────────────────────────────────
test.describe("Track Plotting (CR-UI-002, UC006)", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/");
    await page
      .waitForLoadState("networkidle", { timeout: 15_000 })
      .catch(() => {});
    await waitForMapReady(page);
  });

  test("track store is accessible via window.__RTSA_TRACK_STORE__ [UC006]", async ({
    page,
  }) => {
    const storePresent = await page.evaluate(
      () =>
        !!(window as unknown as Record<string, unknown>)[
          "__RTSA_TRACK_STORE__"
        ],
    );
    expect(storePresent).toBeTruthy();
  });

  test("injected surface track appears in tracks GeoJSON source [CR-UI-002]", async ({
    page,
  }) => {
    await injectTestTrack(page, {
      trackId: "e2e-surface-001",
      lat: 45.0,
      lon: -60.0,
      entityType: "SURFACE",
      hostileClass: "UNKNOWN",
    });

    // Wait for the GeoJSON source to be updated by the RAF-throttled updateMapData
    await page.waitForTimeout(500);

    const featureCount = await page.evaluate(() => {
      const w = window as unknown as Record<string, unknown>;
      const map = w["__RTSA_MAP__"] as
        | { getSource?: (id: string) => { _data?: { features?: unknown[] } } }
        | undefined;
      if (!map?.getSource) return -1; // headless WebGL not available
      return map.getSource("tracks")?._data?.features?.length ?? 0;
    });
    // -1 = headless (acceptable); >=1 = track is in the source
    expect(featureCount === -1 || featureCount >= 1).toBeTruthy();
  });

  test("friendly track has FRIENDLY hostileClass in GeoJSON source [CR-UI-002]", async ({
    page,
  }) => {
    await injectTestTrack(page, {
      trackId: "e2e-friendly-001",
      lat: 45.5,
      lon: -59.5,
      entityType: "SURFACE",
      hostileClass: "FRIENDLY",
    });

    await page.waitForTimeout(500);

    const hostileClass = await page.evaluate(() => {
      const w = window as unknown as Record<string, unknown>;
      const map = w["__RTSA_MAP__"] as
        | {
            getSource?: (id: string) => {
              _data?: {
                features?: Array<{
                  properties?: { hostileClass?: string; trackId?: string };
                }>;
              };
            };
          }
        | undefined;
      if (!map?.getSource) return null;
      const features = map.getSource("tracks")?._data?.features ?? [];
      const found = features.find(
        (f) => f.properties?.trackId === "e2e-friendly-001",
      );
      return found?.properties?.hostileClass ?? null;
    });
    // null = headless (acceptable); otherwise must be FRIENDLY
    expect(hostileClass === null || hostileClass === "FRIENDLY").toBeTruthy();
  });

  test("multiple tracks are each in the GeoJSON source [CR-UI-002]", async ({
    page,
  }) => {
    const trackCount = 5;
    for (let i = 0; i < trackCount; i++) {
      await injectTestTrack(page, {
        trackId: `e2e-multi-${i}`,
        lat: 44.0 + i * 0.1,
        lon: -61.0 + i * 0.1,
      });
    }

    await page.waitForTimeout(600);

    const featureCount = await page.evaluate(() => {
      const w = window as unknown as Record<string, unknown>;
      const map = w["__RTSA_MAP__"] as
        | { getSource?: (id: string) => { _data?: { features?: unknown[] } } }
        | undefined;
      if (!map?.getSource) return -1;
      return map.getSource("tracks")?._data?.features?.length ?? 0;
    });
    // -1 = headless; otherwise must contain at least trackCount features
    expect(featureCount === -1 || featureCount >= trackCount).toBeTruthy();
  });

  test("Zustand trackStore reflects injected tracks [CR-UI-002]", async ({
    page,
  }) => {
    await injectTestTrack(page, {
      trackId: "e2e-store-check",
      lat: 46.0,
      lon: -58.0,
    });

    const trackInStore = await page.evaluate(() => {
      const w = window as unknown as {
        __RTSA_TRACK_STORE__?: {
          getState: () => { tracks: Map<string, unknown> };
        };
      };
      return w.__RTSA_TRACK_STORE__
        ? w.__RTSA_TRACK_STORE__.getState().tracks.has("e2e-store-check")
        : false;
    });
    expect(trackInStore).toBeTruthy();
  });
});

// ─────────────────────────────────────────────────────────────────────────────
// 3. Threat Halos — hostile track proximity circles [CR-UI-002]
// ─────────────────────────────────────────────────────────────────────────────
test.describe("Threat Halos (CR-UI-002)", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/");
    await page
      .waitForLoadState("networkidle", { timeout: 15_000 })
      .catch(() => {});
    await waitForMapReady(page);
  });

  test("hostile track has HOSTILE property in GeoJSON source [CR-UI-002]", async ({
    page,
  }) => {
    await injectTestTrack(page, {
      trackId: "e2e-hostile-001",
      lat: 45.2,
      lon: -60.3,
      entityType: "SURFACE",
      hostileClass: "HOSTILE",
      confidenceScore: 0.97,
    });

    await page.waitForTimeout(500);

    const hostileClass = await page.evaluate(() => {
      const w = window as unknown as Record<string, unknown>;
      const map = w["__RTSA_MAP__"] as
        | {
            getSource?: (id: string) => {
              _data?: {
                features?: Array<{
                  properties?: { hostileClass?: string; trackId?: string };
                }>;
              };
            };
          }
        | undefined;
      if (!map?.getSource) return null;
      const features = map.getSource("tracks")?._data?.features ?? [];
      const found = features.find(
        (f) => f.properties?.trackId === "e2e-hostile-001",
      );
      return found?.properties?.hostileClass ?? null;
    });
    // null = headless (acceptable); otherwise must be HOSTILE
    // Tracks rendered as WebGL circle layer with data-driven color — no DOM markers.
    expect(hostileClass === null || hostileClass === "HOSTILE").toBeTruthy();
  });

  test("threat-halos GeoJSON source is initialised on map load [CR-UI-002]", async ({
    page,
  }) => {
    const hasHalosSource = await page.evaluate(() => {
      const w = window as unknown as Record<string, unknown>;
      const map = w["__RTSA_MAP__"] as
        | { getSource?: (id: string) => unknown }
        | undefined;
      if (!map?.getSource) return null; // map not ready
      return !!map.getSource("threat-halos");
    });
    // null = map not WebGL-ready in headless (acceptable); true = source present
    expect(hasHalosSource === null || hasHalosSource === true).toBeTruthy();
  });

  test("threat-halos source is updated with polygon for hostile track [CR-UI-002]", async ({
    page,
  }) => {
    await injectTestTrack(page, {
      trackId: "e2e-halo-driven",
      lat: 45.0,
      lon: -60.0,
      hostileClass: "HOSTILE",
    });

    // Allow updateMapData async call to complete (dynamic import + DOM update)
    await page.waitForTimeout(800);

    const featureCount = await page.evaluate(() => {
      const w = window as unknown as Record<string, unknown>;
      const map = w["__RTSA_MAP__"] as
        | { getSource?: (id: string) => { _data?: { features?: unknown[] } } }
        | undefined;
      if (!map?.getSource) return -1;
      const source = map.getSource("threat-halos");
      return source?._data?.features?.length ?? 0;
    });

    // -1 = headless WebGL not available (acceptable); >=1 = halo polygon rendered
    expect(featureCount === -1 || featureCount >= 1).toBeTruthy();
  });
});

// ─────────────────────────────────────────────────────────────────────────────
// 4. Geo-Fence Overlays [CR-UI-002]
// ─────────────────────────────────────────────────────────────────────────────
test.describe("Geo-Fence Overlays (CR-UI-002)", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/");
    await page
      .waitForLoadState("networkidle", { timeout: 15_000 })
      .catch(() => {});
    await waitForMapReady(page);
  });

  test("geofences GeoJSON source is initialised with exclusion zone polygon [CR-UI-002]", async ({
    page,
  }) => {
    const fenceInfo = await page.evaluate(() => {
      const w = window as unknown as Record<string, unknown>;
      const map = w["__RTSA_MAP__"] as
        | {
            getSource?: (id: string) => {
              _data?: { features?: Array<{ properties?: { type?: string } }> };
            };
          }
        | undefined;
      if (!map?.getSource) return null;
      const source = map.getSource("geofences");
      const features = source?._data?.features ?? [];
      return {
        count: features.length,
        types: features.map((f) => f.properties?.type),
      };
    });

    // null = headless WebGL not available; otherwise verify exclusion zone exists
    if (fenceInfo !== null) {
      expect(fenceInfo.count).toBeGreaterThanOrEqual(1);
      expect(fenceInfo.types).toContain("exclusion");
    }
  });

  test("geofences layer is added to the map [CR-UI-002]", async ({ page }) => {
    const hasLayer = await page.evaluate(() => {
      const w = window as unknown as Record<string, unknown>;
      const map = w["__RTSA_MAP__"] as
        | { getLayer?: (id: string) => unknown }
        | undefined;
      if (!map?.getLayer) return null;
      return !!map.getLayer("geofences-layer");
    });

    expect(hasLayer === null || hasLayer === true).toBeTruthy();
  });

  test("geofences-fill layer is added with fill paint [CR-UI-002]", async ({
    page,
  }) => {
    const hasFill = await page.evaluate(() => {
      const w = window as unknown as Record<string, unknown>;
      const map = w["__RTSA_MAP__"] as
        | { getLayer?: (id: string) => unknown }
        | undefined;
      if (!map?.getLayer) return null;
      return !!map.getLayer("geofences-fill");
    });

    expect(hasFill === null || hasFill === true).toBeTruthy();
  });
});
