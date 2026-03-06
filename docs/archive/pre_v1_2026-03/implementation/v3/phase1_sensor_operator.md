<!-- CLASSIFICATION: UNCLASSIFIED -->
# Phase 1: Sensor Operator — Detail Implementation Plan

**Status**: ✅ Complete
**Role**: Sensor Operator
**Use Cases**: UC001, UC017
**Dashboard**: Sensor Health (Variant C — Split-Pane Table + Map)

---

## Deliverables Summary

| Component | File | Status |
|---|---|---|
| CSS Design System | `src/styles/design-system.css` | ✅ |
| TypeScript Types | `src/types/sensor.ts` | ✅ |
| Zustand Store | `src/stores/sensorHealthStore.ts` | ✅ |
| gRPC Polling Hook | `src/hooks/useSensorHealth.ts` | ✅ |
| Dashboard | `src/components/layout/SensorHealthDashboard.tsx` | ✅ |
| Status Bar | `src/components/layout/StatusBar.tsx` | ✅ |
| Sparkline | `src/components/sensor/Sparkline.tsx` | ✅ |
| DLQ Popup | `src/components/sensor/DLQPopup.tsx` | ✅ |
| DLQ Viewer | `src/components/sensor/DLQViewer.tsx` | ✅ |
| Store Tests | `src/__tests__/stores/sensorHealthStore.test.ts` | ✅ (27 tests) |
| Dashboard Tests | `src/__tests__/components/SensorHealthDashboard.test.tsx` | ✅ (18 tests) |
| StatusBar Tests | `src/__tests__/components/StatusBar.test.tsx` | ✅ (8 tests) |
| Sparkline Tests | `src/__tests__/components/Sparkline.test.tsx` | ✅ (7 tests) |
| E2E Browser Tests | `e2e/sensor-health.spec.ts` | 🔄 Pending |

---

## Design Decisions

### Variant C — Split-Pane Table + Map

**Left pane** (resizable, 40% default):
- Tabs: Sensor Grid | Dead Letter Queue
- 4 KPI tiles: Active, Degraded, EPS throughput, Avg latency
- Sortable data table: Status dot, Sensor ID, Type, Rate (with sparkline), DLQ count (with popup icon), Latency, Last Seen
- Inline row expansion with drill-down (Identity, Metrics, Coverage)

**Right pane** (flex):
- MapView with sensor coverage overlays

**Interactions**:
- Drag resize handle between panes (min 20%, max 70%)
- Click DLQ icon (📊) on any sensor row → floating DLQ popup with pie chart
- Click row → inline expansion with 3-column drill-down
- Escape key → reset pane sizes and deselect

### Data Flow

```
IngestionService.ListSensorStatuses (gRPC, 3s poll)
        ↓
  useSensorHealth hook (maps proto → SensorStatus)
        ↓
  sensorHealthStore (Zustand)
        ↓
  SensorHealthDashboard (React component tree)
```

Fallback: When backend is unavailable, the hook generates synthetic demo data with realistic sensor types, rates, and DLQ events.

---

## Integration Changes

| File | Change |
|---|---|
| `MainLayout.tsx` | Replaced `SensorHealthPanel` with `StatusBar` |
| `DashboardSelector.tsx` | Analyst second view: `operator` → `intel-search` |
| `uiStore.ts` | Added `intel-search` to `DashboardView` union |
| `index.css` | Added JetBrains Mono font + `design-system.css` import |

---

## Test Results

- **226 tests passing** across 28 test files
- **Zero TypeScript compilation errors**
- Key coverage: `sensorHealthStore` 88%, `StatusBar` 100%, `Sparkline` 100%
