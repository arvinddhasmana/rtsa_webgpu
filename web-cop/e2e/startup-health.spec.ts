// CLASSIFICATION: UNCLASSIFIED
// web-cop/e2e/startup-health.spec.ts
//
// Startup health tests — detect the class of bugs that manifested during the
// Feb 2026 demo: silent init failures, unexpected 404s, broken console output.
//
// Each test in this file is deliberately STRICT (no "null = headless" escapes)
// and runs against the live containerised app at http://localhost:5173.
//
// Bugs caught by this suite:
//   BUG-1  maplibre-gl dynamic import returned CJS wrapper, not the maplibregl
//          object → new maplibregl.Map() threw TypeError silently in try/catch
//          → map never initialized, tracks never rendered.
//   BUG-2  glyphs URL referenced protomaps / demotiles font server which 404s
//          for the font stacks used (Open Sans Semibold, etc.)
//   BUG-3  useSensorCoverage polled IngestionService.ListSensorStatuses every
//          30 s → HTTP 404 because no Envoy route exists for that RPC.
//
// Requirements: CR-UI-001, CR-UI-002
// UC: UC006 Situational Awareness UI, UC012

import { expect, Page, test } from "@playwright/test";
import { injectTestTrack } from "./helpers";

// ─────────────────────────────────────────────────────────────────────────────
// Shared setup: capture console errors & network 404s before page loads
// ─────────────────────────────────────────────────────────────────────────────

interface PageDiagnostics {
  consoleErrors: string[];
  networkFailures: { url: string; status: number }[];
}

async function openWithDiagnostics(
  page: Page,
  url = "/",
): Promise<PageDiagnostics> {
  const diag: PageDiagnostics = { consoleErrors: [], networkFailures: [] };

  // Capture console.error / unhandled errors
  page.on("console", (msg) => {
    if (msg.type() === "error") {
      // Ignore browser-extension messages (Edge Copilot, etc.)
      const text = msg.text();
      if (text.includes("message channel closed")) return;
      if (text.includes("Extension")) return;
      diag.consoleErrors.push(text);
    }
  });

  page.on("pageerror", (err) => {
    diag.consoleErrors.push(`pageerror: ${err.message}`);
  });

  // Intercept network responses to catch unexpected 4xx/5xx
  page.on("response", (res) => {
    const status = res.status();
    const url = res.url();
    if (status >= 400 && status < 600) {
      // Ignore OSM tile 404s (tiles may not be cached) and favicon
      if (url.includes("tile.openstreetmap.org")) return;
      if (url.includes("favicon.ico")) return;
      // Font PBF requests — BUG-2 would appear here
      if (url.includes(".pbf")) {
        diag.networkFailures.push({ url, status });
        return;
      }
      // gRPC-Web RPC calls — BUG-3 would appear here
      if (url.includes("/rtsa.")) {
        diag.networkFailures.push({ url, status });
      }
    }
  });

  await page.goto(url);
  await page
    .waitForLoadState("networkidle", { timeout: 20_000 })
    .catch(() => {});
  // Extra wait for async initialisation (dynamic imports, gRPC connection)
  await page.waitForTimeout(3_000);

  return diag;
}

// ─────────────────────────────────────────────────────────────────────────────
// 1. No unexpected console.error output on startup (catches BUG-1)
// ─────────────────────────────────────────────────────────────────────────────
test.describe("Startup — no console errors (BUG-1 regression)", () => {
  test("no console errors on page load [CR-UI-001]", async ({ page }) => {
    const diag = await openWithDiagnostics(page);

    // Filter to truly unexpected errors (not SW registration notices, not dev
    // environment cert errors from the Envoy self-signed cert, not browser
    // resource-load failures which are surfaced at the network layer separately)
    const unexpected = diag.consoleErrors.filter(
      (e) =>
        !e.includes("SW registered") &&
        !e.includes("serviceWorker") &&
        !e.toLowerCase().includes("favicon") &&
        // Dev-env: Nginx proxy uses self-signed cert for Envoy — Chromium logs this
        !e.includes("ERR_CERT_AUTHORITY_INVALID") &&
        !e.includes("ERR_FAILED") &&
        !e.includes("Failed to load resource") &&
        // OSM tile server quota / rate limit errors
        !e.includes("tile.openstreetmap") &&
        // Browser extension messages
        !e.includes("message channel closed"),
    );

    if (unexpected.length > 0) {
      console.error("Unexpected console errors:", unexpected);
    }
    expect(
      unexpected,
      `Unexpected console errors: ${unexpected.join(" | ")}`,
    ).toHaveLength(0);
  });

  test("no TypeError from maplibre-gl initialisation [BUG-1]", async ({
    page,
  }) => {
    const typeErrors: string[] = [];
    page.on("console", (msg) => {
      if (msg.type() === "error" && msg.text().includes("TypeError")) {
        typeErrors.push(msg.text());
      }
    });
    page.on("pageerror", (err) => {
      if (err.message.includes("TypeError")) {
        typeErrors.push(err.message);
      }
    });

    await page.goto("/");
    await page
      .waitForLoadState("networkidle", { timeout: 20_000 })
      .catch(() => {});
    await page.waitForTimeout(3_000);

    expect(
      typeErrors,
      `TypeError(s) during init (likely broken dynamic import): ${typeErrors.join("; ")}`,
    ).toHaveLength(0);
  });
});

// ─────────────────────────────────────────────────────────────────────────────
// 2. No unexpected HTTP 404s (catches BUG-2 and BUG-3)
// ─────────────────────────────────────────────────────────────────────────────
test.describe("Startup — no unexpected HTTP 404s (BUG-2 & BUG-3 regression)", () => {
  test("no .pbf font file 404s from MapLibre [BUG-2]", async ({ page }) => {
    const diag = await openWithDiagnostics(page);
    const fontFailures = diag.networkFailures.filter((f) =>
      f.url.includes(".pbf"),
    );

    if (fontFailures.length > 0) {
      console.error("Font 404s:", fontFailures.map((f) => f.url).join("\n"));
    }
    expect(
      fontFailures,
      `Font PBF 404s detected — remove glyphs from MapLibre style or use a valid font server:\n${fontFailures.map((f) => f.url).join("\n")}`,
    ).toHaveLength(0);
  });

  test("no gRPC 404s on unrouted RPCs [BUG-3]", async ({ page }) => {
    const diag = await openWithDiagnostics(page);
    const grpcFailures = diag.networkFailures.filter(
      (f) => f.url.includes("/rtsa.") && f.status === 404,
    );

    if (grpcFailures.length > 0) {
      console.error(
        "Unrouted gRPC RPCs returning 404:",
        grpcFailures.map((f) => `${f.url} → ${f.status}`).join("\n"),
      );
    }
    expect(
      grpcFailures,
      `Unrouted gRPC RPC(s) — add Envoy route or stub the client hook:\n${grpcFailures.map((f) => f.url).join("\n")}`,
    ).toHaveLength(0);
  });

  test("ListSensorStatuses is not called on startup [BUG-3]", async ({
    page,
  }) => {
    const listSensorCalls: string[] = [];

    page.on("request", (req) => {
      if (req.url().includes("ListSensorStatuses")) {
        listSensorCalls.push(req.url());
      }
    });

    await page.goto("/");
    // Wait long enough to catch 30s polling interval trigger on first call
    await page
      .waitForLoadState("networkidle", { timeout: 20_000 })
      .catch(() => {});
    await page.waitForTimeout(3_000);

    expect(
      listSensorCalls,
      `ListSensorStatuses was called but is not routed through Envoy — stub the hook or add the Envoy route`,
    ).toHaveLength(0);
  });
});

// ─────────────────────────────────────────────────────────────────────────────
// 3. MapLibre strict initialization (catches BUG-1 definitively)
// ─────────────────────────────────────────────────────────────────────────────
test.describe("MapLibre initialisation — strict (BUG-1 regression)", () => {
  test("window.__RTSA_MAP__ is set after load [BUG-1]", async ({ page }) => {
    await page.goto("/");
    // Wait for the map 'load' event which sets __RTSA_MAP__
    const mapReady = await page
      .waitForFunction(
        () => !!(window as unknown as Record<string, unknown>)["__RTSA_MAP__"],
        { timeout: 30_000 },
      )
      .then(() => true)
      .catch(() => false);

    expect(
      mapReady,
      "__RTSA_MAP__ was never set — MapLibre.Map() failed to initialize. " +
        "Check that the dynamic import resolves to the actual maplibregl object " +
        "(UMD interop: use `(await import('maplibre-gl')).default`)",
    ).toBe(true);
  });

  test("map canvas element is present and has non-zero size [BUG-1]", async ({
    page,
  }) => {
    await page.goto("/");
    const canvas = page.locator("canvas").first();
    await expect(canvas).toBeVisible({ timeout: 30_000 });
    const box = await canvas.boundingBox();
    expect(
      box,
      "Canvas bounding box is null — map container has zero size",
    ).not.toBeNull();
    expect(box!.width, "Canvas width is 0").toBeGreaterThan(100);
    expect(box!.height, "Canvas height is 0").toBeGreaterThan(100);
  });

  test("tracks-circle layer exists on the map [BUG-1]", async ({ page }) => {
    await page.goto("/");
    await page
      .waitForFunction(
        () => !!(window as unknown as Record<string, unknown>)["__RTSA_MAP__"],
        { timeout: 30_000 },
      )
      .catch(() => {});

    const hasTracksLayer = await page.evaluate(() => {
      const map = (window as unknown as Record<string, unknown>)[
        "__RTSA_MAP__"
      ] as { getLayer?: (id: string) => unknown } | undefined;
      return map?.getLayer?.("tracks-circle") != null;
    });

    expect(
      hasTracksLayer,
      "tracks-circle layer not found — map did not initialize correctly",
    ).toBe(true);
  });

  test("injected track appears in tracks GeoJSON source (strict) [BUG-1]", async ({
    page,
  }) => {
    await page.goto("/");
    await page
      .waitForFunction(
        () => !!(window as unknown as Record<string, unknown>)["__RTSA_MAP__"],
        { timeout: 30_000 },
      )
      .catch(() => {});

    await injectTestTrack(page, {
      trackId: "health-check-001",
      lat: 45.0,
      lon: -62.0,
      hostileClass: "UNKNOWN",
    });

    await page.waitForTimeout(500);

    const trackInSource = await page.evaluate(() => {
      const map = (window as unknown as Record<string, unknown>)[
        "__RTSA_MAP__"
      ] as
        | {
            getSource?: (id: string) => {
              _data?: {
                features?: Array<{
                  properties?: { trackId?: string };
                }>;
              };
            };
          }
        | undefined;
      if (!map?.getSource) return "MAP_NOT_READY";
      const features = map.getSource("tracks")?._data?.features ?? [];
      return features.some((f) => f.properties?.trackId === "health-check-001")
        ? "FOUND"
        : "NOT_FOUND";
    });

    expect(
      trackInSource,
      `Track not found in GeoJSON source (got: ${trackInSource}). ` +
        "Either map init failed or updateMapData() did not run.",
    ).toBe("FOUND");
  });
});

// ─────────────────────────────────────────────────────────────────────────────
// 4. Component smoke tests — key UI regions are mounted
// ─────────────────────────────────────────────────────────────────────────────
test.describe("Component smoke tests", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/");
    await page
      .waitForLoadState("networkidle", { timeout: 20_000 })
      .catch(() => {});
  });

  test("UNCLASSIFIED classification banner visible at top and bottom", async ({
    page,
  }) => {
    // ClassificationBanner uses data-testid="classification-banner-{position}"
    await expect(
      page.locator('[data-testid="classification-banner-top"]'),
    ).toBeVisible({ timeout: 10_000 });
    await expect(
      page.locator('[data-testid="classification-banner-bottom"]'),
    ).toBeVisible({ timeout: 10_000 });
  });

  test("RTSA COP toolbar is visible with correct labels", async ({ page }) => {
    await expect(page.getByTestId("toolbar-map")).toBeVisible({
      timeout: 10_000,
    });
    await expect(page.getByTestId("toolbar-alerts")).toBeVisible();
    await expect(page.getByTestId("toolbar-history")).toBeVisible();
  });

  test("connection indicator is present", async ({ page }) => {
    const indicator = page.locator('[data-testid="connection-indicator"]');
    await expect(indicator).toBeVisible({ timeout: 10_000 });
  });

  test("Active Tracks counter is present in FusionSidePanel", async ({
    page,
  }) => {
    // FusionSidePanel shows "Active Tracks" metric label
    await expect(page.getByText("Active Tracks")).toBeVisible({
      timeout: 10_000,
    });
  });
});
