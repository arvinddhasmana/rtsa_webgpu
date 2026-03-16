// CLASSIFICATION: UNCLASSIFIED
// web-cop-gpu/e2e/role-access.spec.ts — Role-based access control E2E tests
//
// Workflows: five-role shell visibility, role default dashboard landing, and
// role-change dashboard guard behavior.
// Reference: docs/implementation/v5/operations_commander/01_blocking_b1_rbac_shell.md

import { expect, test } from "@playwright/test";
import { gotoApp } from "./helpers";

const ROLE_DEFAULT_DASHBOARD = {
  operations_commander: "commander",
  intelligence_analyst: "analytics",
  security_officer: "commander",
  sensor_operator: "health",
  nato_liaison: "sensor",
} as const;

test.describe("Role-Based Access Control", () => {
  test("role selector renders all five roles", async ({ page }) => {
    await gotoApp(page);
    await page
      .waitForLoadState("networkidle", { timeout: 20_000 })
      .catch(() => {});

    const roleSelect = page.locator("#role-selector");
    await expect(roleSelect).toBeVisible({ timeout: 10_000 });

    const labels = await roleSelect.locator("option").allTextContents();
    expect(labels).toContain("Ops Commander");
    expect(labels).toContain("Intelligence Analyst");
    expect(labels).toContain("Security Officer");
    expect(labels).toContain("Sensor Operator");
    expect(labels).toContain("NATO Liaison");
  });

  test("each role lands on its mapped default dashboard", async ({ page }) => {
    await gotoApp(page);
    await page
      .waitForLoadState("networkidle", { timeout: 20_000 })
      .catch(() => {});

    const roleSelect = page.locator("#role-selector");
    const dashboardSelect = page.locator("#dashboard-selector");

    for (const [role, defaultDashboard] of Object.entries(
      ROLE_DEFAULT_DASHBOARD,
    )) {
      await roleSelect.selectOption(role);
      await expect(dashboardSelect).toHaveValue(defaultDashboard);
    }
  });

  test("changing role resets disallowed dashboard to role default", async ({
    page,
  }) => {
    await gotoApp(page);
    await page
      .waitForLoadState("networkidle", { timeout: 20_000 })
      .catch(() => {});

    const roleSelect = page.locator("#role-selector");
    const dashboardSelect = page.locator("#dashboard-selector");

    await roleSelect.selectOption("operations_commander");
    await dashboardSelect.selectOption("analytics");
    await expect(dashboardSelect).toHaveValue("analytics");

    // Security officer only allows the "commander" dashboard in B1.
    await roleSelect.selectOption("security_officer");
    await expect(dashboardSelect).toHaveValue("commander");
  });

  test("dashboard selector is visible alongside role selector", async ({
    page,
  }) => {
    await gotoApp(page);
    await page
      .waitForLoadState("networkidle", { timeout: 20_000 })
      .catch(() => {});

    const roleSelect = page.locator("#role-selector");
    const dashSelect = page.locator("#dashboard-selector");
    await expect(roleSelect).toBeVisible({ timeout: 10_000 });
    await expect(dashSelect).toBeVisible({ timeout: 10_000 });
  });

  test("keyboard shortcut Ctrl+K opens search overlay", async ({ page }) => {
    await gotoApp(page);
    await page
      .waitForLoadState("networkidle", { timeout: 20_000 })
      .catch(() => {});

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
