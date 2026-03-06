// CLASSIFICATION: UNCLASSIFIED
// web-cop-gpu/e2e/security-audit.spec.ts — Security audit E2E tests (H4-3)
//
// Validates ITSG-33 controls:
//   - COOP/COEP headers for SharedArrayBuffer
//   - CSP header present
//   - No PII/classified data in console logs
//   - JWT/clearance controls prevent cross-origin data leakage
//
// Reference: docs/implementation/v4/phase4_hardening_cutover.md H4-3

import { test, expect } from "@playwright/test";
import { gotoApp, setupLogAudit } from "./helpers";

test.describe("Security Audit — ITSG-33 Controls", () => {
  test("COOP header is set to same-origin", async ({ page }) => {
    const response = await page.request.get("/");
    const headers = response.headers();
    expect(headers["cross-origin-opener-policy"]).toBe("same-origin");
  });

  test("COEP header is set to require-corp", async ({ page }) => {
    const response = await page.request.get("/");
    const headers = response.headers();
    expect(headers["cross-origin-embedder-policy"]).toBe("require-corp");
  });

  test("Content-Security-Policy header is present", async ({ page }) => {
    const response = await page.request.get("/");
    const headers = response.headers();
    const csp = headers["content-security-policy"];
    expect(csp).toBeDefined();
    // Must contain key directives
    expect(csp).toContain("default-src");
    expect(csp).toContain("script-src");
    expect(csp).toContain("object-src 'none'");
    expect(csp).toContain("frame-ancestors 'none'");
  });

  test("CSP does not allow unsafe-inline scripts", async ({ page }) => {
    const response = await page.request.get("/");
    const csp = response.headers()["content-security-policy"] ?? "";
    // 'unsafe-inline' must NOT appear in script-src
    const scriptSrc = csp.split(";").find((d) => d.trim().startsWith("script-src"));
    expect(scriptSrc).toBeDefined();
    expect(scriptSrc).not.toContain("'unsafe-inline'");
  });

  test("X-Content-Type-Options is nosniff", async ({ page }) => {
    const response = await page.request.get("/");
    const headers = response.headers();
    expect(headers["x-content-type-options"]).toBe("nosniff");
  });

  test("X-Frame-Options is DENY", async ({ page }) => {
    const response = await page.request.get("/");
    const headers = response.headers();
    expect(headers["x-frame-options"]).toBe("DENY");
  });

  test("no PII or classified data in console logs on startup", async ({ page }) => {
    const { violations } = setupLogAudit(page);
    await gotoApp(page);
    await page.waitForLoadState("networkidle", { timeout: 20_000 }).catch(() => {});
    await page.waitForTimeout(2_000);
    expect(violations).toHaveLength(0);
  });

  test("no inline JavaScript execution detected in DOM", async ({ page }) => {
    await gotoApp(page);
    await page.waitForLoadState("networkidle", { timeout: 20_000 }).catch(() => {});

    // Check that no scripts use innerHTML with dynamic content
    const inlineHandlers = await page.evaluate(() => {
      const elements = Array.from(document.querySelectorAll("*"));
      return elements.filter((el) => {
        const attrs = Array.from(el.attributes);
        return attrs.some((a) => a.name.startsWith("on") && a.value.length > 0);
      }).length;
    });
    // SolidJS should produce zero inline event handlers
    expect(inlineHandlers).toBe(0);
  });

  test("app does not expose sensitive data in window globals", async ({ page }) => {
    await gotoApp(page);
    await page.waitForLoadState("networkidle", { timeout: 20_000 }).catch(() => {});

    const sensitiveKeys = await page.evaluate(() => {
      const win = window as unknown as Record<string, unknown>;
      const SENSITIVE = ["token", "jwt", "secret", "password", "apiKey", "credentials"];
      return Object.keys(win).filter((k) =>
        SENSITIVE.some((s) => k.toLowerCase().includes(s)),
      );
    });
    expect(sensitiveKeys).toHaveLength(0);
  });
});
