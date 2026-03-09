<!-- CLASSIFICATION: UNCLASSIFIED -->

# B1 — Missing `data-testid` Attributes (E2E Tests Broken)

> **Batch**: B1 of 7
> **Theme**: E2E Testid Attributes
> **Priority**: BLOCKING — Every Playwright E2E test fails immediately
> **Agent Profile**: `greatest-ever-dev-forworkflow`
> **Source**: v4 Implementation Review R-001 through R-004

---

## Context

All Playwright E2E tests in `web-cop-gpu/e2e/` select elements exclusively via `data-testid` attributes. No component in `web-cop-gpu/src/components/` emits any `data-testid` attribute. This means **every single E2E test fails at the first selector** without even reaching the test logic.

Do not add `data-testid` to elements that already have them. Check each component file before editing.

---

## Issue R-001 — `ClassificationBanner` missing `data-testid`

**File**: `web-cop-gpu/src/components/shell/ClassificationBanner.tsx`

**Required change**: Add `data-testid="classification-banner-top"` to the root `<div role="banner">` element.

**Evidence**: E2E tests reference this selector in:

- `web-cop-gpu/e2e/cold-boot.spec.ts` (line ~26)
- `web-cop-gpu/e2e/helpers.ts` (line ~30)

---

## Issue R-002 — `StatusBar` missing `data-testid`

**File**: `web-cop-gpu/src/components/status/StatusBar.tsx`

**Required changes**:

- Add `data-testid="status-bar"` to the root `<div>`.
- Add `data-testid="fps-display"` to the element that renders the FPS value.
- Add `data-testid="latency-display"` to the element that renders the latency value.

**Evidence**: E2E tests reference these selectors in:

- `web-cop-gpu/e2e/cold-boot.spec.ts` (line ~64)
- `web-cop-gpu/e2e/visual-regression.spec.ts` (lines ~75–76)

---

## Issue R-003 — Toolbar components missing `data-testid`

**Files**:

| File                                                         | `data-testid` to add     | Element      |
| ------------------------------------------------------------ | ------------------------ | ------------ |
| `web-cop-gpu/src/components/toolbar/RoleSelector.tsx`        | `"role-selector"`        | Root `<div>` |
| `web-cop-gpu/src/components/toolbar/DashboardSelector.tsx`   | `"dashboard-selector"`   | Root `<div>` |
| `web-cop-gpu/src/components/toolbar/ConnectionIndicator.tsx` | `"connection-indicator"` | Root `<div>` |

**Evidence**: E2E tests reference these selectors in:

- `web-cop-gpu/e2e/cold-boot.spec.ts` (line ~72)
- `web-cop-gpu/e2e/role-access.spec.ts` (lines ~15, 23, 43, 59)
- `web-cop-gpu/e2e/helpers.ts` (line ~40)
- `web-cop-gpu/e2e/reconnection.spec.ts` (line ~16)

---

## Issue R-004 — Panel and shell components missing `data-testid`

**Files**:

| File                                                  | `data-testid` to add | Element                   |
| ----------------------------------------------------- | -------------------- | ------------------------- |
| `web-cop-gpu/src/components/panels/AlertSidebar.tsx`  | `"alert-sidebar"`    | Root `<div>`              |
| `web-cop-gpu/src/components/panels/FeedbackForm.tsx`  | `"feedback-form"`    | Modal container `<div>`   |
| `web-cop-gpu/src/components/search/SearchOverlay.tsx` | `"search-overlay"`   | Overlay container `<div>` |
| `web-cop-gpu/src/components/shell/AppShell.tsx`       | `"app-toolbar"`      | Left toolbar `<div>`      |

**Evidence**: E2E tests reference these selectors in:

- `web-cop-gpu/e2e/alerts-feedback.spec.ts` (lines ~18, 32, 54, 71)
- `web-cop-gpu/e2e/visual-regression.spec.ts` (line ~103)
- `web-cop-gpu/e2e/role-access.spec.ts` (line ~71)

---

## Implementation Rules

1. Read each component file before editing. Only add the `data-testid` attribute to the correct element — do not restructure or refactor any other code.
2. **Do not** add `data-testid` to child elements unless explicitly listed above.
3. **Do not** add TypeScript type annotations, docstrings, or comments to code you did not change.
4. Follow `docs/sdlc_guidelines/04_coding_standards/solidjs_standards.md` — do not destructure props.
5. All files must retain the `// CLASSIFICATION: UNCLASSIFIED` header.

---

## Unit & Integration Tests Required

- Confirm existing Vitest unit tests still pass after changes: `cd web-cop-gpu && pnpm test`
- At minimum, add or confirm that a test for each component renders without error and that the `data-testid` attribute is present in the rendered output.
- Example pattern (adapt to the project's existing test setup):

```typescript
it("renders classification banner with data-testid", () => {
  render(() => <ClassificationBanner />);
  expect(screen.getByTestId("classification-banner-top")).toBeInTheDocument();
});
```

---

## Verification Steps

1. `cd web-cop-gpu && pnpm test` — all unit tests must pass.
2. `cd web-cop-gpu && pnpm test:e2e` — E2E selectors must now resolve; cold-boot, role-access, and visual-regression tests must pass the element-visibility assertions.

---

## PR Instructions

- PR title: `fix(web-cop-gpu): add data-testid attributes to all E2E-targeted components (B1)`
- Label: `ai-orchestrator`
- After the PR is merged, move this file from `.github/instructions/todo/B1-E2E-Testid-Attributes.md` to `.github/instructions/done/B1-E2E-Testid-Attributes.md`.
