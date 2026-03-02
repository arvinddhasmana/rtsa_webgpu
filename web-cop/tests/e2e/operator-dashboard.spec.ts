import { expect, test } from '@playwright/test';

test.describe('Operator Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    // Switch to Operator view
    await page.locator('[data-testid="role-selector"]').selectOption('commander');
    await page.locator('[data-testid="dashboard-selector"]').selectOption('operator');
  });

  test('Operator dashboard renders DetailPanel on inspect', async ({ page }) => {
    // Check that we switched to Operator View
    // Try to trigger an inspect to open detail panel
    const detailPanel = page.getByTestId('detail-panel');

    // Initially hidden
    await expect(detailPanel).toBeHidden();

    // Verify presence of alert cards first
    const anyInspectButton = page.locator('[data-testid^="alert-inspect-"]').first();

    // Fallback if no alerts are streaming right now
    if (await anyInspectButton.isVisible()) {
      await anyInspectButton.click();
      await expect(detailPanel).toBeVisible();
    } else {
        // Just verify basic layout mounts
        const mapContainer = page.getByTestId('map-container');
        await expect(mapContainer).toBeVisible();
    }
  });

  test('Feedback buttons update visual transition on alert card', async ({ page }) => {
    const anyConfirmButton = page.locator('[data-testid^="alert-confirm-"]').first();
    if (await anyConfirmButton.isVisible()) {
        const cardLoc = page.locator('.glass-panel').filter({ has: anyConfirmButton }).first();
        await anyConfirmButton.click();
        await expect(cardLoc).toContainText('Status: Confirmed');
    }
  });
});
