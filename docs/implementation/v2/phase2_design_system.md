<!-- CLASSIFICATION: UNCLASSIFIED -->
# Phase 2 — Design System & Collapsible Pane Architecture

> **Phase**: 2 of 4 | **Depends on**: Nothing (parallel with Phase 1) | **Blocks**: Phase 3
> **Scope**: CSS/Tailwind design tokens, uiStore state, RoleSelector, DashboardSelector, MainLayout refactor

---

## Purpose

Establish a premium, mission-critical visual design system and refactor the layout architecture to support the Two-Level Role-Based shell. This phase touches NO dashboard-specific components — it only builds the infrastructure that Phase 3 will compose into specific views.

---

## Step 1: Typography & Font Setup

### 1.1 Add Google Font

Open `web-cop/index.html` (the root HTML file served by Vite). Add the Inter font import inside the `<head>` tag:

```html
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap" rel="stylesheet">
```

### 1.2 Update base CSS

Open `web-cop/src/index.css`. Replace the current `body` font-family rule:

Replace:
```css
font-family: 'Courier New', Courier, monospace;
```

With:
```css
font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
```

---

## Step 2: Define Premium Design Tokens in CSS

In the same `web-cop/src/index.css` file, add CSS custom properties at the top (inside `:root` or on the `body` selector). These will be the foundation for all components:

```css
:root {
  /* === Background Layers === */
  --bg-base: #0B1120;
  --bg-surface: #111827;
  --bg-panel: rgba(17, 24, 39, 0.75);
  --bg-panel-hover: rgba(30, 41, 59, 0.85);

  /* === Glassmorphism === */
  --glass-bg: rgba(15, 23, 42, 0.60);
  --glass-border: rgba(51, 65, 85, 0.50);
  --glass-blur: 12px;

  /* === Text === */
  --text-primary: #F1F5F9;
  --text-secondary: #94A3B8;
  --text-muted: #64748B;

  /* === Accent - Blue (Primary) === */
  --accent-blue: #3B82F6;
  --accent-blue-dim: rgba(59, 130, 246, 0.15);
  --accent-blue-glow: rgba(59, 130, 246, 0.40);

  /* === Accent - Amber (Warning/Highlight) === */
  --accent-amber: #F59E0B;
  --accent-amber-dim: rgba(245, 158, 11, 0.15);

  /* === Accent - Red (Critical/Hostile) === */
  --accent-red: #EF4444;
  --accent-red-dim: rgba(239, 68, 68, 0.15);
  --accent-red-glow: rgba(220, 38, 38, 0.50);

  /* === Accent - Green (Friendly/Connected) === */
  --accent-green: #22C55E;
  --accent-green-dim: rgba(34, 197, 94, 0.15);

  /* === Borders === */
  --border-subtle: rgba(51, 65, 85, 0.40);
  --border-active: rgba(59, 130, 246, 0.60);

  /* === Shadows === */
  --shadow-panel: 0 4px 24px rgba(0, 0, 0, 0.40);
  --shadow-glow-blue: 0 0 20px rgba(59, 130, 246, 0.25);

  /* === Transitions === */
  --transition-fast: 150ms ease;
  --transition-normal: 250ms ease;
  --transition-slow: 400ms ease;

  /* === Spacing === */
  --toolbar-height: 44px;
  --banner-height: 28px;
}
```

### 2.1 Add NVG theme variables

Below the `:root` block, add:

```css
[data-theme="nvg"] {
  --bg-base: #000000;
  --bg-surface: #0A1A0A;
  --bg-panel: rgba(10, 26, 10, 0.75);
  --text-primary: #33FF33;
  --text-secondary: #228B22;
  --accent-blue: #00FF00;
  --accent-amber: #33FF33;
  --accent-red: #FF3333;
  --accent-green: #00FF00;
  --border-subtle: rgba(34, 139, 34, 0.40);
}
```

### 2.2 Add glassmorphism utility class

```css
.glass-panel {
  background: var(--glass-bg);
  backdrop-filter: blur(var(--glass-blur));
  -webkit-backdrop-filter: blur(var(--glass-blur));
  border: 1px solid var(--glass-border);
  border-radius: 8px;
  box-shadow: var(--shadow-panel);
}
```

### 2.3 Add collapsible panel transition utility

```css
.panel-collapsible {
  transition: width var(--transition-normal),
              height var(--transition-normal),
              opacity var(--transition-fast);
  overflow: hidden;
}

.panel-collapsible.collapsed {
  width: 0 !important;
  min-width: 0 !important;
  opacity: 0;
  padding: 0;
  border: none;
}

.panel-collapsible-vertical.collapsed {
  height: 0 !important;
  min-height: 0 !important;
  opacity: 0;
  padding: 0;
  border: none;
}
```

---

## Step 3: Update `uiStore.ts` with Level-2 Dashboard State and All 5 Roles

Open `web-cop/src/stores/uiStore.ts`.

### 3.1 Extend the `ActiveRole` type

Replace:
```typescript
export type ActiveRole = "commander" | "security" | "analyst";
```

With:
```typescript
export type ActiveRole = "commander" | "security" | "analyst" | "sensor_operator" | "nato_liaison";
```

### 3.2 Add the `DashboardView` type

Add the following type definition:

```typescript
export type DashboardView =
  | "fusion"          // Commander default
  | "multi-domain"    // Commander option
  | "operator"        // Commander option
  | "forensics"       // Analyst default
  | "intel-search"    // Analyst option
  | "audit"           // Security default
  | "sensor-health"   // Sensor Operator default
  | "nato-exchange";  // NATO Liaison default
```

### 3.3 Add state and actions to the store interface

Add these fields to the `UIState` interface:

```typescript
activeDashboardView: DashboardView;
setDashboardView: (view: DashboardView) => void;
```

### 3.4 Add default state and action implementation

In the `create<UIState>` call, add:

```typescript
activeDashboardView: "fusion",
setDashboardView: (view) => set({ activeDashboardView: view }),
```

### 3.5 Update `setActiveRole` to auto-set the default dashboard view

Modify the existing `setActiveRole` implementation to also reset the dashboard view to the role's default:

```typescript
setActiveRole: (role) => {
  const defaultViews: Record<ActiveRole, DashboardView> = {
    commander: "fusion",
    analyst: "forensics",
    security: "audit",
    sensor_operator: "sensor-health",
    nato_liaison: "nato-exchange",
  };
  set({ activeRole: role, activeDashboardView: defaultViews[role] });
},
```

---

## Step 4: Update `RoleSelector.tsx` with All 5 Roles

Open `web-cop/src/components/layout/RoleSelector.tsx`.

Replace the `ROLES` array:

```typescript
const ROLES: { value: ActiveRole; label: string }[] = [
  { value: "commander", label: "Operations Commander" },
  { value: "analyst", label: "Intelligence Analyst" },
  { value: "security", label: "Security Officer" },
  { value: "sensor_operator", label: "Sensor Operator" },
  { value: "nato_liaison", label: "NATO Liaison" },
];
```

Apply the glassmorphism styling to the `<select>` element. Replace the existing inline `style` object with one that uses the CSS custom properties:

```typescript
style={{
  padding: "6px 12px",
  background: "var(--glass-bg)",
  backdropFilter: "blur(8px)",
  color: "var(--text-primary)",
  border: "1px solid var(--glass-border)",
  borderRadius: "6px",
  cursor: "pointer",
  fontSize: "0.8rem",
  fontWeight: 500,
}}
```

---

## Step 5: Create `DashboardSelector.tsx` (Level 2)

Create a new file: `web-cop/src/components/layout/DashboardSelector.tsx`

This component should:

1. Read `activeRole` and `activeDashboardView` from `useUIStore`.
2. Based on the `activeRole`, render a set of tab-style buttons for the available dashboard views.
3. Clicking a button calls `setDashboardView(view)`.
4. The active view button should be highlighted with the blue accent.

The role-to-views mapping:

```typescript
const ROLE_VIEWS: Record<ActiveRole, { value: DashboardView; label: string; icon: string }[]> = {
  commander: [
    { value: "fusion", label: "Fusion", icon: "🔗" },
    { value: "multi-domain", label: "Multi-Domain", icon: "🌐" },
    { value: "operator", label: "Operator", icon: "🎯" },
  ],
  analyst: [
    { value: "forensics", label: "Forensics", icon: "🔍" },
    { value: "intel-search", label: "Intel Search", icon: "📊" },
  ],
  security: [
    { value: "audit", label: "Audit & Feedback", icon: "🔒" },
  ],
  sensor_operator: [
    { value: "sensor-health", label: "Sensor Health", icon: "📡" },
  ],
  nato_liaison: [
    { value: "nato-exchange", label: "NATO Exchange", icon: "🌐" },
  ],
};
```

Style each button using the glassmorphism design tokens. The active button should have `background: var(--accent-blue-dim)` and `border-color: var(--accent-blue)`. Inactive buttons should have `background: transparent`.

---

## Step 6: Refactor `MainLayout.tsx`

Open `web-cop/src/components/layout/MainLayout.tsx`.

### 6.1 Import the new components

Add imports for `DashboardSelector`. Also prepare for dashboard-specific layout components that Phase 3 will create — for now, render placeholder divs.

### 6.2 Add DashboardSelector to the toolbar

In the toolbar JSX, add `<DashboardSelector />` after `<RoleSelector />` and before the spacer `<div style={{ flex: 1 }} />`.

### 6.3 Apply design tokens to the toolbar

Replace the existing inline `backgroundColor: "#1E293B"` style on the toolbar div with:

```typescript
style={{
  display: "flex",
  alignItems: "center",
  padding: "6px 16px",
  background: "var(--bg-surface)",
  borderBottom: "1px solid var(--border-subtle)",
  gap: "12px",
  height: "var(--toolbar-height)",
}}
```

### 6.4 Apply design tokens to the outer container

Replace the root `<div>` style to use the new tokens:

```typescript
style={{
  display: "flex",
  flexDirection: "column",
  height: "100vh",
  backgroundColor: "var(--bg-base)",
  color: "var(--text-primary)",
  paddingTop: "var(--banner-height)",
  paddingBottom: "var(--banner-height)",
  boxSizing: "border-box",
}}
```

### 6.5 Refactor the main content area to use a view router

Replace the existing role-based conditional rendering (`activeRole === "security" ? ...`) with a single switch based on `activeDashboardView`. For now, render simple placeholder panels that Phase 3 will replace:

```tsx
{/* Main content — routed by activeDashboardView */}
<div style={{ display: "flex", flex: 1, overflow: "hidden" }}>
  {activeDashboardView === "fusion" && <div>Fusion Dashboard (Phase 3)</div>}
  {activeDashboardView === "multi-domain" && <div>Multi-Domain Dashboard (Phase 3)</div>}
  {activeDashboardView === "operator" && <div>Operator UI (Phase 3)</div>}
  {activeDashboardView === "forensics" && (
    <>
      <div style={{ flex: 1, overflow: "hidden" }}><MapView /></div>
      <div style={{ width: "40%", overflow: "auto" }}><ForensicsPanel /></div>
    </>
  )}
  {activeDashboardView === "audit" && (
    <>
      <div style={{ width: "30%", overflow: "hidden" }}><AlertPanel /></div>
      <div style={{ flex: 1, display: "flex", alignItems: "center", justifyContent: "center", color: "var(--text-muted)" }}>
        [Security Officer] Audit & Feedback Queue View
      </div>
    </>
  )}
  {activeDashboardView === "sensor-health" && <div>Sensor Health Dashboard (Phase 4)</div>}
  {activeDashboardView === "nato-exchange" && <div>NATO Exchange Dashboard (Phase 4)</div>}
</div>
```

### 6.6 Add `data-theme` attribute for NVG support

In the root `<div>`, add `data-theme={theme}` so that the CSS custom properties from the NVG overrides apply:

```tsx
<div data-theme={theme} style={{ ... }}>
```

---

## Step 7: Create Collapsible Panel Wrapper Component

Create a new file: `web-cop/src/components/layout/CollapsiblePanel.tsx`

This is a reusable wrapper that Phase 3 dashboard components will use. It should:

1. Accept props: `title: string`, `collapsed: boolean`, `onToggle: () => void`, `direction: 'horizontal' | 'vertical'`, `defaultSize: string`, `children: React.ReactNode`
2. Render a header bar with the `title` and a chevron icon (`▸` when collapsed, `▾` when expanded).
3. When expanded, render `children` with the specified `defaultSize` as width (horizontal) or height (vertical).
4. When collapsed, apply the `.panel-collapsible.collapsed` or `.panel-collapsible-vertical.collapsed` CSS class from the utilities defined in Step 2.
5. The header bar should always be visible (even when collapsed) so the user can re-expand.
6. Apply the `glass-panel` CSS class to the outer wrapper.

---

## Step 8: Update Toolbar Buttons with Design Tokens

Go through each toolbar button in `MainLayout.tsx` and replace the hardcoded `toolbarButtonStyle` with a new style object using CSS custom properties:

```typescript
const toolbarButtonStyle: React.CSSProperties = {
  padding: "5px 14px",
  background: "var(--bg-panel)",
  color: "var(--text-primary)",
  border: "1px solid var(--border-subtle)",
  borderRadius: "6px",
  cursor: "pointer",
  fontSize: "0.75rem",
  fontWeight: 500,
  transition: "var(--transition-fast)",
};
```

Add hover effect using onMouseEnter/onMouseLeave state or a CSS class with `:hover` pseudo-class in `index.css`:

```css
.toolbar-btn:hover {
  background: var(--bg-panel-hover);
  border-color: var(--border-active);
  box-shadow: var(--shadow-glow-blue);
}
```

---

## Verification

### Automated

1. Build the web-cop:
   ```bash
   cd web-cop && npm run build
   ```
   Must complete with zero TypeScript errors.

2. Run existing unit tests:
   ```bash
   cd web-cop && npm run test
   ```

3. Run existing E2E tests:
   ```bash
   cd web-cop && npm run test:e2e
   ```
   Note: Some existing tests that relied on the old 3-role layout may need minor locator updates. Fix those as you go.

### Visual Verification

1. Start dev server: `cd web-cop && npm run dev`
2. Open `http://localhost:5173` in the browser
3. Verify:
   - [ ] Inter font is rendering (no Courier New)
   - [ ] Dark theme uses the new CSS custom properties (deeper blacks, subtle borders)
   - [ ] NVG theme produces green-on-black when selected
   - [ ] All 5 roles appear in the Role Selector dropdown
   - [ ] Switching roles shows the correct Level-2 dashboard tabs
   - [ ] Commander role shows Fusion / Multi-Domain / Operator tabs
   - [ ] Clicking tabs updates the main content area (placeholder text is OK for now)
   - [ ] Toolbar buttons have glassmorphism styling and hover glow effect
