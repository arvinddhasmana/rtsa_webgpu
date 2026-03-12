// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/SensorGrid.tsx — Filtered sensor card grid
//
// Reference: docs/business/usecases/UC017_sensor_health_monitoring.md

import { createMemo, For, Show } from "solid-js";
import { SensorStatus } from "../../services/sensor-health";
import { selectedStatuses, selectedTypes } from "../../signals/sensor-filters";
import { SensorStatusCard } from "./SensorStatusCard";

interface SensorGridProps {
  sensors: SensorStatus[];
}

/**
 * Responsive grid that displays sensor cards filtered by the global filter signals.
 */
export function SensorGrid(props: SensorGridProps) {
  const filteredSensors = createMemo(() => {
    return props.sensors.filter(s =>
      selectedStatuses().includes(s.status) &&
      selectedTypes().includes(s.sensorType)
    );
  });

  return (
    <div style={{
        padding: "24px",
        display: "flex",
        "flex-direction": "column",
        height: "100%",
        "background-color": "rgba(10, 15, 28, 0.2)",
    }}>
      <Show
        when={filteredSensors().length > 0}
        fallback={
          <div style={{
            flex: 1,
            display: "flex",
            "flex-direction": "column",
            "align-items": "center",
            "justify-content": "center",
            color: "#64748b",
            gap: "1rem"
          }}>
            <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1" opacity="0.5">
                <circle cx="12" cy="12" r="10" />
                <line x1="8" y1="12" x2="16" y2="12" />
            </svg>
            <div style={{ "font-size": "1.1rem" }}>No sensors match the selected filters</div>
            <div style={{ "font-size": "0.8rem" }}>Try adjusting your status or type selections in the sidebar</div>
          </div>
        }
      >
        <div style={{
            display: "grid",
            "grid-template-columns": "repeat(auto-fill, minmax(320px, 1fr))",
            gap: "24px",
            "align-content": "start",
            overflow: "auto",
            "padding-bottom": "40px"
        }}>
          <For each={filteredSensors()}>
            {(sensor) => <SensorStatusCard sensor={sensor} />}
          </For>
        </div>
      </Show>
    </div>
  );
}
