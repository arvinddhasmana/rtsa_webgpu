// CLASSIFICATION: UNCLASSIFIED
// web-cop-gpu/e2e/helpers.ts — Shared Playwright test utilities for WebGPU COP
//
// Reference: docs/implementation/v4/phase4_hardening_cutover.md H4-4

import { Page, expect } from "@playwright/test";

/**
 * Navigate to the app and wait for the initial load to settle.
 * In test environments, the app may show a degraded notice when WebGPU
 * is unavailable — this function handles both paths.
 */
export async function gotoApp(page: Page): Promise<void> {
  await page.goto("/");
  await page.waitForLoadState("domcontentloaded");
}

/**
 * Wait for the WebGPU canvas to be visible, indicating full render mode.
 */
export async function waitForGpuCanvas(page: Page): Promise<void> {
  await page.waitForSelector("#gpu-canvas", { state: "visible", timeout: 20_000 });
}

/**
 * Wait for the classification banner to be visible.
 * The banner must be shown at all times including degraded mode.
 */
export async function waitForClassificationBanner(page: Page): Promise<void> {
  await page.waitForSelector('[data-testid="classification-banner-top"]', {
    state: "visible",
    timeout: 10_000,
  });
}

/**
 * Wait for the WebTransport connection indicator to be visible.
 */
export async function waitForConnectionIndicator(page: Page): Promise<void> {
  await page.waitForSelector('[data-testid="connection-indicator"]', {
    state: "visible",
    timeout: 15_000,
  });
}

/**
 * Simulate app startup and wait for all core UI elements to be rendered.
 * Returns true if the app is in full WebGPU mode, false if degraded.
 */
export async function waitForAppReady(page: Page): Promise<boolean> {
  await gotoApp(page);
  // Wait for the page to finish loading
  await page.waitForLoadState("networkidle", { timeout: 20_000 }).catch(() => {});

  // Check whether we have the canvas (full mode) or the degraded notice
  const canvasLocator = page.locator("#gpu-canvas");
  const degradedLocator = page.locator('[role="alert"]');

  const [canvasVisible, degradedVisible] = await Promise.all([
    canvasLocator.isVisible().catch(() => false),
    degradedLocator.isVisible().catch(() => false),
  ]);

  return canvasVisible && !degradedVisible;
}

/**
 * Block WebTransport traffic to simulate a server restart/disconnect.
 */
export async function blockWebTransport(page: Page): Promise<void> {
  await page.route("**/*", (route) => {
    const url = route.request().url();
    if (url.includes(":4443") || url.includes("/wt")) {
      route.abort();
    } else {
      route.continue();
    }
  });
}

/**
 * Restore WebTransport traffic (unroute all blocked patterns).
 */
export async function restoreWebTransport(page: Page): Promise<void> {
  await page.unrouteAll();
}

/**
 * Verify that no PII or classified data appears in console logs.
 * Sets up a listener — call before page.goto().
 */
export function setupLogAudit(page: Page): { violations: string[] } {
  const violations: string[] = [];
  const PII_PATTERNS = [/\b\d{3}-\d{2}-\d{4}\b/, /operator_id:\s*\w+@/i];
  const CLASSIFIED_PATTERNS = [/SECRET/i, /TOP SECRET/i, /TS\/SCI/i];

  page.on("console", (msg) => {
    const text = msg.text();
    for (const pattern of [...PII_PATTERNS, ...CLASSIFIED_PATTERNS]) {
      if (pattern.test(text)) {
        violations.push(`[${msg.type()}] ${text}`);
      }
    }
  });

  return { violations };
}

/**
 * Assert that no security header violations occurred during page load.
 * Checks that COOP/COEP headers were sent (required for SharedArrayBuffer).
 */
export async function assertSecurityHeaders(page: Page): Promise<void> {
  const response = await page.request.get("/");
  const headers = response.headers();

  expect(headers["cross-origin-opener-policy"]).toBe("same-origin");
  expect(headers["cross-origin-embedder-policy"]).toBe("require-corp");
}
