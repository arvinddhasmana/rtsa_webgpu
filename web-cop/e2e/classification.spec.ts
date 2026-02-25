// CLASSIFICATION: UNCLASSIFIED
// web-cop/e2e/classification.spec.ts
// Browser E2E tests for Classification Banner & Security — CR-SEC-001
// Covers: UC012 (Security Enforcement)

import { test, expect } from "@playwright/test";

test.describe("Classification Banner (CR-SEC-001)", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/");
    await page.waitForLoadState("networkidle", { timeout: 15_000 }).catch(() => {});
  });

  test("top classification banner is visible on page load [CR-SEC-001, UC012]", async ({ page }) => {
    const banner = page.locator('[data-testid="classification-banner-top"]');
    await expect(banner).toBeVisible({ timeout: 15_000 });
  });

  test("bottom classification banner is visible on page load [CR-SEC-001, UC012]", async ({ page }) => {
    const banner = page.locator('[data-testid="classification-banner-bottom"]');
    await expect(banner).toBeVisible({ timeout: 15_000 });
  });

  test("classification banner displays classification level text [CR-SEC-001]", async ({ page }) => {
    const banner = page.locator('[data-testid="classification-banner-top"]');
    await expect(banner).toBeVisible({ timeout: 15_000 });
    const text = await banner.innerText();
    expect(
      text.toUpperCase().includes("UNCLASSIFIED") ||
      text.toUpperCase().includes("PROTECTED") ||
      text.toUpperCase().includes("SECRET")
    ).toBeTruthy();
  });

  test("classification banner has non-transparent background colour [CR-SEC-001]", async ({ page }) => {
    const banner = page.locator('[data-testid="classification-banner-top"]');
    await expect(banner).toBeVisible({ timeout: 15_000 });
    const bgColor = await banner.evaluate(
      (el) => window.getComputedStyle(el).backgroundColor
    );
    expect(bgColor).not.toBe("rgba(0, 0, 0, 0)");
    expect(bgColor).not.toBe("transparent");
  });
});
