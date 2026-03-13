# Implementation Plan - Sensor Coverage Dashboard UI & E2E Upgrade

Upgrade the RTSA UI to match high-fidelity mockups using SolidJS and WebGPU, and extend the backend pipeline to propagate real-time coverage geometry.

## User Review Required

> [!IMPORTANT]
> This upgrade requires a **full-stack deployment**: Simulator, Ingestion Services, and Web UI must be updated in sync to transmit and render coverage geometry.
>
> [!WARNING]
> Synthetic data in the simulator must be calibrated to provide realistic bearing sectors (for Radar) and polygons (for ISR).

## Proposed Changes

### [Backend & Pipeline]

#### [NEW] [spatial_alert.proto](file:///home/arvind/workspace/rtsa_webgpu/proto/rtsa/inference/v1/spatial_alert.proto)
- Define `SpatialAlert` message to support area-based reports (polygons) for coverage gaps.
- `rtsa.inference.v1.SpatialAlert`: `alert_id`, `area_polygon`, `anomaly_type` (GAP), `severity`, `explanation`.

#### [MODIFY] [sensor_observation.proto](file:///home/arvind/workspace/rtsa_webgpu/proto/rtsa/ingestion/v1/sensor_observation.proto)
- Standardize metadata keys for coverage propagation (e.g., `rtsa.coverage.range_nm`, `rtsa.coverage.bearing_start`).

#### [MODIFY] [sensorstate.go](file:///home/arvind/workspace/rtsa_webgpu/pkg/ingestion/sensorstate.go)
- Add `coverage atomic.Value` to [SensorStateTracker](file:///home/arvind/workspace/rtsa_webgpu/pkg/ingestion/sensorstate.go#26-53) to store `ingestionv1.SensorCoverage`.

#### [NEW] [Redpanda Topics]
- **`alerts.spatial.gaps`**: New topic for broadcasting geometric gap alerts.
- **`sensors.coverage.updates`**: Optional topic for dedicated coverage geometry changes (if high-frequency updates are needed later).

#### [NEW] [Coverage Analyzer Service]
- A new analytics stage (or enhancement to `svc-anomaly-detection`) to compute geometric gaps by comparing sensor footprints with track density and emitting `SpatialAlert` messages.

---

### [UI Strategy: The Three Levels]

#### 1. Level 1: Sensor Dashboard Overview
**Objective**: Situational awareness via a grid of health cards and a miniature map.
- **[MODIFY] [SensorHealthDashboard.tsx](file:///home/arvind/workspace/rtsa_webgpu/web-cop-gpu/src/components/dashboard/SensorHealthDashboard.tsx)**:
    - Embed a restricted miniature **Overview Map** (pinned to a specific region) beneath or alongside the sensor grid.
    - Apply "Premium Theme": Radial gradients (indigo/charcoal) and glassmorphic card containers.
- **[MODIFY] [SensorStatusCard.tsx](file:///home/arvind/workspace/rtsa_webgpu/web-cop-gpu/src/components/dashboard/SensorStatusCard.tsx)**:
    - Add "Neon Status Pulse" for connected sensors.
    - Upgrade sparklines with neon gradients.

#### 2. Level 2: Sensor Drill-Down
**Objective**: Detailed diagnostics and health metrics for a specific sensor.
- **[MODIFY] [SensorDiagnosticView.tsx](file:///home/arvind/workspace/rtsa_webgpu/web-cop-gpu/src/components/dashboard/SensorDiagnosticView.tsx)**:
    - Embed a **Targeted Mini-Map** focused specifically on the selected sensor's coverage area.
    - Implement high-res gradient charts for throughput and DLQ analytics.

#### 3. Level 3: Full Coverage Map
**Objective**: Strategic view focusing on global footprints and coverage gaps.
- **[MODIFY] [App.tsx](file:///home/arvind/workspace/rtsa_webgpu/web-cop-gpu/src/App.tsx)**:
    - Ensure selecting "Map" (Dashboard: [sensor](file:///home/arvind/workspace/rtsa_webgpu/web-cop-gpu/src/components/dashboard/SensorStatusCard.tsx#26-37)) for a `sensor_operator` defaults to the **Coverage Optimization** mode.
    - Highlight "Coverage Gap" alerts directly on this map with high-visibility pulsing halos.

---

### [WebGPU Rendering Engine]

#### [NEW] [coverage.wgsl](file:///home/arvind/workspace/rtsa_webgpu/web-cop-gpu/src/shaders/coverage.wgsl)
- Implement neon-styled shaders for coverage footprints:
    - **Radar/ELINT**: Fan sector arcs with gradient fills.
    - **EW/AIS**: Concentric range rings.
    - **ISR**: Swath polygons with dashed perimeters.
- Support status-driven opacity (30% active, 10% offline).

#### [NEW] [coverage.ts](file:///home/arvind/workspace/rtsa_webgpu/web-cop-gpu/src/gpu/coverage.ts)
- Manage WebGPU pipeline, vertex buffers, and bind groups for the coverage layer.

#### [MODIFY] [renderer.ts](file:///home/arvind/workspace/rtsa_webgpu/web-cop-gpu/src/gpu/renderer.ts)
- Integrate the `coverage` pass into the main render frame (Step 7.5: between geofences and icons).

---

### [E2E Test Suite]

#### [MODIFY] [sensor-health.spec.ts](file:///home/arvind/workspace/rtsa_webgpu/web-cop-gpu/e2e/sensor-health.spec.ts)
- **Level 1 Test**: Verify miniature map presence and card pulsing.
- **Level 2 Test**: Verify diagnostic mini-map correctly zooms to the selected sensor.
- **Level 3 Test**: Verify "Coverage Gap" alerts trigger visual highlights on the full map.

---

## Full Pipeline E2E Strategy

The following "Full-Pipeline" flow will be used for development and final validation:

1.  **Simulation Layer**: `tools/simulator` is updated to generate "Sensor Footprints" (Coverage) and "Tactical Gaps" (e.g., by moving entities into unmonitored zones).
2.  **Ingestion & State**: Real `svc-radar-ingestion` (and other type-specific services) will run in Docker/Local, consuming from the simulator and producing to Redpanda.
3.  **Analytics Layer**: The `Coverage Analyzer` will consume from Redpanda, detect gaps, and emit `SpatialAlert` messages back to Redpanda.
4.  **Presentation Layer**: The Web UI will consume real-time tracks via WebTransport and Spatial Alerts via gRPC-Web (Cold Path).

## Verification Plan

### Full-Pipeline E2E Test
```bash
# 1. Start Full Pipeline
./scripts/dev/start-pipeline.sh --with-simulator 

# 2. Run E2E Test (Full Mode)
cd web-cop-gpu
VITE_MOCK_SENSORS=false npx playwright test e2e/sensor-pipeline.spec.ts
```

### Manual Verification
- **Scenario A**: Launch simulator with "Coverage Gap" scenario. Observe Level 3 Map highlighting the gap and triggering a strategic alert.
- **Scenario B**: Drill down into a specific sensor. Observe Level 2 Mini-Map showing the precise footprint dictated by the simulator's metadata.
- **Scenario C**: Verify Level 1 Overview reflects real-time status changes (Neon Pulse) based on simulator throughput.
