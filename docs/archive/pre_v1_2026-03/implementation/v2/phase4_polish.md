<!-- CLASSIFICATION: UNCLASSIFIED -->
# Phase 4 — Polish & Remaining Roles

> **Phase**: 4 of 4 | **Depends on**: Phase 3 | **Blocks**: Nothing
> **Scope**: Sensor Operator dashboard, NATO Liaison dashboard, keyboard shortcuts, micro-animations, E2E test suite

---

## Purpose

Complete the remaining two role-specific dashboards (Sensor Operator and NATO Liaison), implement all documented keyboard shortcuts, add premium micro-animations, and build out the Playwright E2E test suite to cover the new features.

---

## Step 1: Sensor Operator Dashboard

### 1.1 Create the hook `useSensorStatuses.ts`

Create `web-cop/src/hooks/useSensorStatuses.ts`. This hook should:

1. Import the gRPC-Web client for `IngestionService.ListSensorStatuses`.
2. Use `@tanstack/react-query` with a `refetchInterval` of 5000ms (5 seconds).
3. Call `ListSensorStatuses({ sensorTypes: [], activeWithinSeconds: 300 })` — return all sensors active in the last 5 minutes.
4. Return `{ sensors: SensorStatusResponse[], isLoading: boolean }`.

### 1.2 Create `SensorHealthDashboard.tsx`

Create `web-cop/src/components/layout/SensorHealthDashboard.tsx`. Layout:

```
┌──────────────────────┬──────────────────────────────────────────┐
│                      │                                          │
│  Sensor Status List  │  MapView with SensorCoverageLayer        │
│  (collapsible, 30%)  │  (sensor positions marked)               │
│                      │                                          │
│  Cards per sensor:   │                                          │
│  - sensor_id         │                                          │
│  - sensor_type icon  │                                          │
│  - connected (●/○)   │                                          │
│  - events/sec        │                                          │
│  - total received    │                                          │
│  - last observation  │                                          │
│  - acceptance rate   │                                          │
│                      │                                          │
└──────────────────────┴──────────────────────────────────────────┘
```

The left panel should:
1. Render a list of sensor cards from `useSensorStatuses`.
2. Each card should use `glass-panel` styling.
3. Connected sensors show a green dot (`var(--accent-green)`); disconnected sensors show a red dot (`var(--accent-red)`).
4. Events/sec should be rendered as a small sparkline or just a number.
5. Acceptance rate = `total_accepted / total_received * 100`, displayed as a percentage bar.
6. Clicking a sensor card centres the map on that sensor's position (from `coverage.sensor_position`).

The right panel (map) should:
1. Render `MapView` with `SensorCoverageLayer` enabled by default.
2. Show sensor markers at each sensor position with the sensor_id label.
3. When a sensor is clicked on the map, its card in the left panel should scroll into view and highlight.

### 1.3 Wire into `MainLayout.tsx`

Import `SensorHealthDashboard` and replace the Phase 2 placeholder:

```tsx
{activeDashboardView === "sensor-health" && <SensorHealthDashboard />}
```

---

## Step 2: NATO Liaison Dashboard

### 2.1 Create `NATOExchangeDashboard.tsx`

Create `web-cop/src/components/layout/NATOExchangeDashboard.tsx`. Layout:

```
┌────────────────────────────────────────────────────────────────────┐
│  NATO Exchange Status Header                                       │
│  Link 16: ● Connected | NFFI: ● Connected | Last Exchange: 14:32  │
├────────────────────────────┬───────────────────────────────────────┤
│                            │                                       │
│  Track Nomination Queue    │   MapView showing NATO-shared tracks  │
│  (collapsible, 35%)       │   with REL TO markings                │
│                            │                                       │
│  - Tracks pending review   │   Tracks marked for NATO sharing     │
│  - [Nominate] button      │   shown with distinct NATO icon       │
│  - [Revoke] button        │                                       │
│  - Filters: status, type  │                                       │
│                            │                                       │
└────────────────────────────┴───────────────────────────────────────┘
```

Since the backend `NATOService` proto (`nato_service.proto`) already exists, use the existing RPCs.

The Status Header should:
1. Display connectivity status for Link 16 and NFFI gateways — you can derive this from the connection indicator pattern or from a health-check call.
2. Show the timestamp of the last successful data exchange.

The Track Nomination Queue should:
1. Query tracks that have been nominated for NATO sharing (use a filter on `tracks.fused.*` where classification allows sharing).
2. Each track card shows: track_id, entity_type, hostile_class, classification, and sharing status.
3. The `[Nominate]` button should — for demo purposes — log the action and mark the track with a visual badge.
4. The `[Revoke]` button should reverse the nomination.

The MapView should:
1. Render normally but highlight tracks that have been nominated for NATO sharing with a distinct icon (e.g., a NATO compass star overlay).

### 2.2 Wire into `MainLayout.tsx`

```tsx
{activeDashboardView === "nato-exchange" && <NATOExchangeDashboard />}
```

---

## Step 3: Complete All Keyboard Shortcuts

Open `web-cop/src/components/layout/MainLayout.tsx`. Update the `handleKeyDown` function.

### 3.1 Map Focus (`M` key)

When `M` is pressed:
1. Collapse all side panels (alert, detail, forensics).
2. Set `activeDashboardView` to the current role's default (which always includes a map).
3. This effectively gives the user a "full map focus" mode.

### 3.2 Full-Screen Map (`F` key)

When `F` is pressed:
1. Call `document.documentElement.requestFullscreen()` (or `exitFullscreen()` if already in fullscreen).
2. Toggle a boolean `isFullscreen` in `uiStore`.

### 3.3 Tab Panel Cycling (`Tab` key)

When `Tab` is pressed (and no input is focused):
1. Cycle focus between available panels: Map → Alert Panel → Detail Panel → Forensics Panel → Map.
2. Use `focus()` on the panel's `tabIndex` container element.

### 3.4 Undo Filter (`Ctrl+Z`)

This requires storing a filter history stack in `uiStore`:
1. Add `filterHistory: Array<{ entityTypeFilter, hostileClassFilter }>` to `uiStore`.
2. When any filter is set, push the previous filter state to the stack.
3. On `Ctrl+Z`, pop the stack and restore the previous filter state.

---

## Step 4: Micro-Animations & Premium Polish

### 4.1 Panel Collapse/Expand Transitions

In `web-cop/src/index.css`, the transition classes from Phase 2 should already handle this. Verify that all `CollapsiblePanel` instances actually apply the `panel-collapsible` class and that the animation is smooth.

Add a subtle slide-in effect for the detail panel:

```css
@keyframes slideUp {
  from { transform: translateY(100%); opacity: 0; }
  to { transform: translateY(0); opacity: 1; }
}

.detail-panel-enter {
  animation: slideUp var(--transition-normal) ease-out;
}
```

### 4.2 Track Icon Hover Effects

In the MapView map component, when the user hovers over a track marker:
1. Increase the marker size by 20%.
2. Show a tooltip with: track_id, entity_type, hostile_class, confidence_score.
3. Apply a subtle glow effect using `box-shadow: var(--shadow-glow-blue)`.

### 4.3 Alert Pulse Animation Enhancement

Update the existing `@keyframes pulse` in `index.css` to be more visually striking:

```css
@keyframes pulse-critical {
  0%, 100% {
    box-shadow: 0 0 0 0 var(--accent-red-glow);
  }
  50% {
    box-shadow: 0 0 12px 4px var(--accent-red-glow);
  }
}

@keyframes pulse-elevated {
  0%, 100% {
    box-shadow: 0 0 0 0 var(--accent-amber-dim);
  }
  50% {
    box-shadow: 0 0 8px 3px var(--accent-amber-dim);
  }
}
```

### 4.4 Connection Indicator Animation

Add a breathing animation to the connection indicator:

```css
@keyframes breathe {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.connection-live {
  animation: breathe 2s ease-in-out infinite;
  color: var(--accent-green);
}
```

### 4.5 Toolbar Button Hover Effects

Ensure the `.toolbar-btn:hover` CSS from Phase 2 is applied. Add an active/pressed state:

```css
.toolbar-btn:active {
  transform: scale(0.96);
  transition: transform 100ms ease;
}
```

---

## Step 5: E2E Test Suite Expansion

### 5.1 Create `role-switching.spec.ts`

Create `web-cop/tests/e2e/role-switching.spec.ts`:

```typescript
test('Commander role shows Fusion/Multi-Domain/Operator tabs', async ({ page }) => {
  await page.goto('/');
  await page.selectOption('[data-testid="role-selector"]', 'commander');
  await expect(page.locator('text=Fusion')).toBeVisible();
  await expect(page.locator('text=Multi-Domain')).toBeVisible();
  await expect(page.locator('text=Operator')).toBeVisible();
});

test('Analyst role shows Forensics/Intel Search tabs', async ({ page }) => {
  await page.goto('/');
  await page.selectOption('[data-testid="role-selector"]', 'analyst');
  await expect(page.locator('text=Forensics')).toBeVisible();
  await expect(page.locator('text=Intel Search')).toBeVisible();
});

test('Switching roles resets to default dashboard view', async ({ page }) => {
  await page.goto('/');
  await page.selectOption('[data-testid="role-selector"]', 'commander');
  await page.click('text=Multi-Domain');
  await page.selectOption('[data-testid="role-selector"]', 'analyst');
  await expect(page.locator('text=Forensics')).toHaveAttribute('aria-selected', 'true');
});
```

### 5.2 Create `fusion-dashboard.spec.ts`

Test that:
- Fusion side panel renders with detection metrics
- Track list shows active tracks
- Clicking a track opens the detail panel
- Side panel can be collapsed and expanded

### 5.3 Create `operator-dashboard.spec.ts`

Test that:
- Map has blur effect applied (check CSS filter property)
- Alert cards have Inspect/Confirm/Reject/Assign buttons
- Clicking Confirm shows a feedback success indication
- Timeline view renders with at least one event

### 5.4 Create `multi-domain-dashboard.spec.ts`

Test that:
- All panels are collapsed by default
- Domain metrics overlay is visible
- The overlay shows counts for at least one domain

### 5.5 Create `keyboard-shortcuts.spec.ts`

Test that:
- `Escape` closes the detail panel
- `A` toggles the alert panel
- `H` opens the forensics panel (for analyst role)
- `Ctrl+F` opens the search overlay

### 5.6 Update existing tests

Review the existing E2E tests (`toolbar.spec.ts`, `role-layout.spec.ts`, etc.) and update any locators that have changed due to the refactored layout. Ensure all existing tests pass.

---

## Verification

### Automated

```bash
# Full build
cd web-cop && npm run build

# Unit tests
cd web-cop && npm run test

# E2E tests
cd web-cop && npm run test:e2e

# E2E with visual report
cd web-cop && npm run test:e2e:report
```

All must pass with zero failures.

### Visual Verification Checklist

- [ ] **All 5 roles** accessible and each shows correct Level-2 views
- [ ] **Sensor Operator**: Sensor health cards render, coverage overlays on map, connected/disconnected status visible
- [ ] **NATO Liaison**: Exchange status header, track nomination queue functional
- [ ] **Keyboard shortcuts**: M (map focus), F (fullscreen), A (alerts), H (history), Escape (dismiss), Tab (cycle), Ctrl+F (search)
- [ ] **Micro-animations**: Panel slide-in, track hover glow, alert pulse, toolbar button hover, breathing connection indicator
- [ ] **Glassmorphism**: All panels have translucent backgrounds with blur effect
- [ ] **NVG theme**: Green-on-black rendering when NVG mode selected
- [ ] **Responsive**: Panels collapse gracefully, map maximises when panels are hidden
- [ ] **No regressions**: All pre-existing features still function correctly
