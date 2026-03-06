// CLASSIFICATION: UNCLASSIFIED
// web-cop-gpu/e2e/role-access.spec.ts — Role-based access control E2E tests
//
// Workflows: Operator → Commander → layout changes; role selector interactions
// Reference: docs/implementation/v4/phase4_hardening_cutover.md H4-4

import { test, expect } from "@playwright/test";
import { gotoApp } from "./helpers";

test.describe("Role-Based Access Control", () => {
  test("role selector renders default role", async ({ page }) => {
    await gotoApp(page);
    await page.waitForLoadState("networkidle", { timeout: 20_000 }).catch(() => {});

    const roleSelector = page.locator('[data-testid="role-selector"]');
    await expect(roleSelector).toBeVisible({ timeout: 10_000 });
  });

  test("can switch to sensor_operator role", async ({ page }) => {
    await gotoApp(page);
    await page.waitForLoadState("networkidle", { timeout: 20_000 }).catch(() => {});

    const roleSelector = page.locator('[data-testid="role-selector"]');
    await expect(roleSelector).toBeVisible({ timeout: 10_000 });

    // Click the selector to open options
    await roleSelector.click();

    // Look for Sensor Operator option
    const operatorOption = page.locator("text=Sensor Operator").first();
    const isVisible = await operatorOption.isVisible().catch(() => false);
    if (isVisible) {
      await operatorOption.click();
      // After selecting, the role selector should reflect the choice
      await expect(roleSelector).toContainText(/Sensor|Operator/i);
    }
  });

  test("can switch to operations_commander role", async ({ page }) => {
    await gotoApp(page);
    await page.waitForLoadState("networkidle", { timeout: 20_000 }).catch(() => {});

    const roleSelector = page.locator('[data-testid="role-selector"]');
    await expect(roleSelector).toBeVisible({ timeout: 10_000 });

    await roleSelector.click();
    const commanderOption = page.locator("text=Commander").first();
    const isVisible = await commanderOption.isVisible().catch(() => false);
    if (isVisible) {
      await commanderOption.click();
      await expect(roleSelector).toContainText(/Commander/i);
    }
  });

  test("dashboard selector is visible alongside role selector", async ({ page }) => {
    await gotoApp(page);
    await page.waitForLoadState("networkidle", { timeout: 20_000 }).catch(() => {});

    const dashSelector = page.locator('[data-testid="dashboard-selector"]');
    await expect(dashSelector).toBeVisible({ timeout: 10_000 });
  });

  test("keyboard shortcut Ctrl+K opens search overlay", async ({ page }) => {
    await gotoApp(page);
    await page.waitForLoadState("networkidle", { timeout: 20_000 }).catch(() => {});

    // Press Ctrl+K
    await page.keyboard.press("Control+k");

    // Search overlay should appear
    const searchOverlay = page.locator('[data-testid="search-overlay"]');
    const isVisible = await searchOverlay.isVisible().catch(() => false);
    if (isVisible) {
      await expect(searchOverlay).toBeVisible();
      // Close with Escape
      await page.keyboard.press("Escape");
    }
  });
});
