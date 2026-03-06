<!-- CLASSIFICATION: UNCLASSIFIED -->
# UI Gaps and Improvement Suggestions — Lead Architect Review (Post Phase-4)

**Status:** COMPLETED | **Date:** 2026-03-01 | **Reviewer:** Lead Architect & Web UI Designer

---

## Overview

This analysis evaluates the current implementation of the `web-cop` React application against the required designs specified in `docs/user_guide` and `docs/implementation/v2`.

Following the completion of Phase 3 and Phase 4, the core Role-Based Access Control (RBAC) shell, all 5 distinct role dashboards, and the component architectures are successfully implemented. The application state is significantly more advanced than previous iterations. However, several critical UI polish items, keyboard shortcuts, and aesthetic details remain incomplete.

---

## 1. Functional Implementation Gaps

### 1.1 Keyboard Shortcuts (Phase 4.3)
The `MainLayout.tsx` event listener implements placeholder logic rather than the correct documented behaviors:
* **F (Full-Screen)**: Currently changes the dashboard view to "multi-domain". **Fix**: Must call browser `document.documentElement.requestFullscreen()` and update `isFullscreen` boolean in the UI store.
* **Tab (Panel Cycling)**: Currently cycles between Level-2 dashboard tabs. **Fix**: Should explicitly cycle DOM focus (`tabIndex`) between Map, Alert Panel, Detail Panel, and Forensics Panel to ensure tactical keyboard accessibility.
* **Ctrl+Z (Undo Filter)**: Currently just acts as a toggle for the Alert Panel. **Fix**: Requires implementing a filter history stack in `uiStore` to revert to previous filter states upon undo.

### 1.2 Map Micro-Interactions (Phase 4.4)
The `MapView.tsx` WebGL map lacks the required interactive polish for entity markers:
* **Hover Tooltips**: Missing required tooltips rendering `track_id`, `entity_type`, `hostile_class`, and `confidence_score` upon hovering a track marker.
* **Hover Effects**: Tracks do not scale up by 20% or show the mandatory `box-shadow: var(--shadow-glow-blue)` glow effect on hover. Currently, it only alters the pointer cursor on mouse leave.

### 1.3 Operator Dashboard Layout Integration
In `OperatorDashboard.tsx`, the `DetailPanel` does not render. An empty `div` slot (`<div style={{ flex: 2... }}>`) was left as a placeholder. Clicking `[Inspect]` on an alert triggers state changes in the store to open the panel, but the panel is invisible to the Operator role.

---

## 2. Design & Polish Improvements (UI/UX)

As Web UI Designer, the following aesthetic and layout improvements are highly recommended to elevate the application from fully functional to strictly premium:

### 2.1 Glassmorphism Consistency
While `.glass-panel` exists in `index.css`, major floating elements like `SearchOverlay.tsx` and the action bars in `DetailPanel.tsx` use hard-coded solid background colors (e.g. `backgroundColor: "#1E293B"`) instead of the standard `var(--glass-bg)` token with `backdrop-filter: var(--glass-blur)`.
* **Action**: Audit all absolute/floating panels to enforce the standard transparent glassmorphism design system uniformly.

### 2.2 Animation Deficiencies
* **Breathing Connection**: The `ConnectionIndicator.tsx` is static and lacks the documented `@keyframes breathe` continuous animation to visually indicate a live, active data stream.
* **Alert Card Action States**: Clicking `[Confirm]` or `[Reject]` on an alert card replaces text abruptly. This should feature a micro-fade transition and a color-flash to positively reinforce operator action.

### 2.3 Visual Hierarchy & Typography
* The UI currently relies on scattered, hard-coded inline font sizes (e.g., `0.75rem`, `0.65rem`) across components rather than cohesive typography tokens in the CSS.
* The `DetailPanel` tab selection features a very basic border switch. Enhancing the active tab indicator with a smooth sliding underline animation would significantly increase the perceived premium quality of the COP interface.
