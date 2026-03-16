// CLASSIFICATION: UNCLASSIFIED
// web-cop-gpu/e2e/quick-actions-assignment.spec.ts

import { expect, test, type Page } from "@playwright/test";
import { gotoApp } from "./helpers";

async function seedCommanderQuickActionScenario(page: Page): Promise<void> {
  await page.evaluate(async () => {
    const maybeGlobal = window as Window & {
      __RTSA_E2E_MOCKS__?: {
        submitFeedback?: (request: {
          trackId: string;
          operatorId: string;
          feedbackType: number;
          justification: string;
          alertId?: string;
          operatorClearance: number;
        }) => Promise<{
          feedbackId: string;
          trustScore: number;
          validated: boolean;
        }>;
        assignAlert?: (request: {
          alertId: string;
          assignerOperatorId: string;
          assigneeOperatorId: string;
          comment: string;
        }) => Promise<{
          success: boolean;
          assignedAt: { seconds: number };
        }>;
      };
    };

    maybeGlobal.__RTSA_E2E_MOCKS__ = {
      submitFeedback: async () => ({
        feedbackId: "fb-e2e",
        trustScore: 0.92,
        validated: true,
      }),
      assignAlert: async () => ({
        success: true,
        assignedAt: { seconds: 1700000000 },
      }),
    };

    const viewport = await import("/src/signals/viewport.ts");
    viewport.setRole("operations_commander");
    viewport.setDashboard("analytics");

    const alertsModule = await import("/src/signals/alerts.ts");
    alertsModule.updateAlerts([
      {
        alertId: "alert-e2e-001",
        trackId: "track-e2e-001",
        severity: "CRITICAL",
        description: "Synthetic quick-action test alert",
        detectedAtMs: Date.now(),
        acknowledged: false,
      },
    ]);
  });
}

test.describe("commander quick-action and assignment flows", () => {
  test("quick-action affordances render for seeded alert", async ({ page }) => {
    await gotoApp(page);
    await page
      .waitForLoadState("networkidle", { timeout: 20_000 })
      .catch(() => {});
    await seedCommanderQuickActionScenario(page);

    const sidebar = page.locator('[data-testid="alert-sidebar"]');
    await expect(sidebar).toBeVisible();

    await expect(
      sidebar.getByRole("button", { name: "Inspect alert" }),
    ).toBeVisible();
    await expect(
      sidebar.getByRole("button", { name: "Confirm alert" }),
    ).toBeVisible();
    await expect(
      sidebar.getByRole("button", { name: "Reject alert" }),
    ).toBeVisible();
    await expect(
      sidebar.getByRole("button", { name: "Assign alert" }),
    ).toBeVisible();
  });

  test("confirm and reject quick-action transitions update alert state", async ({
    page,
  }) => {
    await gotoApp(page);
    await page
      .waitForLoadState("networkidle", { timeout: 20_000 })
      .catch(() => {});
    await seedCommanderQuickActionScenario(page);

    const sidebar = page.locator('[data-testid="alert-sidebar"]');

    await sidebar.getByRole("button", { name: "Confirm alert" }).click();
    await expect(sidebar.getByText("Confirmed")).toBeVisible();

    await sidebar.getByRole("button", { name: "Reject alert" }).click();
    await expect(sidebar.getByText("Rejected")).toBeVisible();
  });

  test("assign quick-action captures assignee and shows assigned state", async ({
    page,
  }) => {
    await gotoApp(page);
    await page
      .waitForLoadState("networkidle", { timeout: 20_000 })
      .catch(() => {});
    await seedCommanderQuickActionScenario(page);

    const sidebar = page.locator('[data-testid="alert-sidebar"]');
    await sidebar.getByRole("button", { name: "Assign alert" }).click();

    await page.locator("#assign-assignee").fill("op-bravo");
    await page
      .locator("#assign-comment")
      .fill("Please investigate immediately");
    await page.getByRole("button", { name: "Assign" }).last().click();

    await expect(sidebar.getByText("Assigned: op-bravo")).toBeVisible();
  });

  test("inspect quick-action focuses track detail context", async ({
    page,
  }) => {
    await gotoApp(page);
    await page
      .waitForLoadState("networkidle", { timeout: 20_000 })
      .catch(() => {});
    await seedCommanderQuickActionScenario(page);

    const sidebar = page.locator('[data-testid="alert-sidebar"]');
    await sidebar.getByRole("button", { name: "Inspect alert" }).click();

    await expect(
      page.locator("[aria-label='Track detail panel']"),
    ).toBeVisible();
    await expect(page.getByText("Focused From Alert")).toBeVisible();
    await expect(page.getByText("alert-e2e-001")).toBeVisible();
  });
});
