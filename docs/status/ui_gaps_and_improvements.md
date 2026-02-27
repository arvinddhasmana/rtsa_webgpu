# UI Gaps and Improvement Suggestions

**Status:** COMPLETED
**Date:** 2026-02-27

## Overview
This document analyzes the current implementation of the Common Operating Picture (COP) Web Application (`web-cop`) against the documented user-facing features described in `docs/user_guide/shared/04_ui_navigation.md`. The analysis is based on a direct review of the React component hierarchy and Playwright E2E browser test coverage executed against the Dockerized application setup.

---

## 1. Identified Functional Gaps (Missing Features)

Based on the UI Navigation documentation and verified through code analysis and E2E test execution, the following critical features are **missing or broken** in the current Web application:

### Main Toolbar (Navigation)
*   **Gap:** The top toolbar only contains a title, a connection indicator, a "FORENSICS" button, and a Theme dropdown.
*   **Missing Features:** The documented buttons for `Map`, `Settings` (Display modes like NVG High-Contrast), `Profile` (identity/clearance info), `Sensors`, `NATO`, and `Audit` do not exist in the layout (`MainLayout.tsx`).
*   **E2E Coverage:** There are no tests verifying the existence or functionality of these toolbar actions.

### Map Layer Controls
*   **Gap:** The Map UI (`MapView.tsx`) strictly renders pre-defined layers (Tracks, Halos, Geo-fences) but lacks the operator-facing **"Layers" toggle button** defined in the user guide.
*   **Impact:** Operators cannot currently toggle track labels, fade trails, grid overlays (MGRS), or sensor coverage on/off to declutter the map.
*   **E2E Coverage:** The test `map.spec.ts` verifies that the map renders features programmatically, but completely lacks coverage for user interactions relating to map layer toggles.

### Alert Quick Actions
*   **Gap:** The `AlertCard.tsx` component only supports a single click-to-acknowledge action.
*   **Missing Features:** The documented `[Inspect]`, `[Confirm]`, `[Reject]`, and `[Assign]` quick action buttons are not implemented inside the Alert panel.
*   **E2E Coverage:** `alerts.spec.ts` only serves as a smoke test verifying the panel exists and has severity buttons. There is no automated assessment of workflow actions or feedback submission initiated from an alert.

### Search and Entity Lookup
*   **Gap:** The global entity search function (invoked via `Ctrl + F`) is entirely missing.
*   **Impact:** Operators cannot look up tracks by Track ID (e.g., TRK-4501), Callsign, MMSI, or Geographic area/MGRS grid.
*   **E2E Coverage:** No search accessibility or lookup functionality is covered in the test suite.

### Keyboard Navigation & Accessibility
*   **Gap:** The app implies tactical-grade keyboard-only support (`M`, `A`, `H`, `F`, `Tab` navigation), but the React application lacks active focus trapping, explicit ARIA bindings, and defined `keydown` event listeners for these exact shortcuts.

---

## 2. Architecture & UI Improvement Suggestions

### Role-Based Dashboards & Navigation
*   **Issue:** Currently, `MainLayout.tsx` forces all panels (Map, Alerts, Details, Forensics) into a single unified grid view for all users, regardless of their role or intent.
*   **Improvement:** Introduce a **Role-Based Shell architecture (Sidebar/Top Navigation)**.
    *   **Operations Commander:** Defaults to the dense Map + Active Alerts grid.
    *   **Security Officer:** Defaults to an Audit / Feedback queue table view, taking up the full screen rather than cramming it alongside the tactical map.
    *   **Intelligence Analyst:** Defaults to the Forensics Map Replay and historical query builder.
    *   These views should be independently routeable rather than being toggled via simple global `zustand` state toggles overlaying the main map.
    *   The role-based shell should be implemented as a separate component that can be imported and used in the main layout.
    *   For demo purpose provide selection of Role on UI. On selection of Role, show role specific UI layout and functionality.

### E2E Test Suite Modernization
*   **Issue:** The existing E2E tests are superficial smoke tests ensuring DOM elements render without crashing.
*   **Improvement:** Expand the Playwright suite to mock gRPC-Web responses and assert full interaction flows:
    *   e.g. `User selects 'CRITICAL' filter -> Clicks Alert -> Clicks [Confirm] -> Feedback Request Dispatched`.

## 3. Additional Functional Gaps & Suggestions

### Classification Badges on Tracks
* **Gap:** Tracks displayed on the map do not show classification badges indicating NATO standard assessment status (e.g., UNCLASSIFIED, CONFIDENTIAL, SECRET). The UI lacks visual cues for track classification.
* **Impact:** Operators cannot quickly assess the security level of each track, which is critical for situational awareness and compliance.
* **Suggested Fix:** Extend the `MapView` track rendering to include a small badge overlay (e.g., colored dot or label) based on the track's classification property. Use NATO colour coding (UNCLASSIFIED – green, CONFIDENTIAL – blue, SECRET – red, etc.).

### Play Button on Forensics Tab
* **Gap:** The purpose of the **Play** button in the Forensics panel is undocumented and its functionality is unclear. Currently it does nothing when clicked.
* **Impact:** Users cannot replay historical track data or replay scenarios, limiting forensic analysis capabilities.
* **Suggested Fix:** Implement the Play button to control playback of recorded track data (start, pause, seek). Provide tooltip explaining its function.

### Dismiss Detail Panel
* **Gap:** Pressing **Escape** or clicking outside the detail panel does not dismiss it.
* **Impact:** Users cannot easily close the panel, leading to UI clutter and reduced usability, especially in tactical environments where quick view changes are needed.
* **Suggested Fix:** Add a global keydown listener for `Escape` and a click‑outside handler to collapse the panel. Update the `DetailPanel` component to respect these interactions.

### Entity Detail Panel – Incomplete Functionality
* **Gap:** Clicking a track or alert opens the detail panel, but many fields (e.g., identity, source attribution, history) are either empty or not interactive.
* **Impact:** Operators lack critical information about the selected entity, hindering decision‑making.
* **Suggested Fix:** Populate the panel from the track/alert store, ensure all sections render data, and add actions such as “Zoom to Track”, “Export Details”, and “Add Note”.

### Role Selection UI (Demo Purpose)
* **Gap:** The UI does not provide a mechanism to select a user role for demonstration purposes.
* **Impact:** Stakeholders cannot preview role‑specific dashboards without modifying code.
* **Suggested Fix:** Add a role selector dropdown in the toolbar that switches the layout to the appropriate role‑based view (as described in the Role‑Based Dashboards section).
