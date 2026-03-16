// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/MultiDomainCommanderDashboard.tsx

import { For, createSignal, type JSX } from "solid-js";

interface MultiDomainCommanderDashboardProps {
  mapContent: JSX.Element;
}

interface LayerToggle {
  id: string;
  label: string;
  enabled: boolean;
}

const INITIAL_LAYER_TOGGLES: LayerToggle[] = [
  { id: "fused-tracks", label: "Fused Tracks", enabled: true },
  { id: "observations", label: "Observations", enabled: true },
  { id: "sensor-coverage", label: "Sensor Coverage", enabled: true },
  { id: "alerts", label: "Alerts", enabled: true },
];

/** Operations Commander Multi-Domain dashboard skeleton. */
export function MultiDomainCommanderDashboard(
  props: MultiDomainCommanderDashboardProps,
) {
  const [layerToggles, setLayerToggles] = createSignal(INITIAL_LAYER_TOGGLES);

  function toggleLayer(layerId: string) {
    setLayerToggles((curr) =>
      curr.map((item) =>
        item.id === layerId ? { ...item, enabled: !item.enabled } : item,
      ),
    );
  }

  return (
    <div
      data-testid="commander-multi-domain-dashboard"
      style={{
        width: "100%",
        height: "100%",
        position: "relative",
        background: "#090f1a",
        overflow: "hidden",
      }}
    >
      <div
        data-testid="commander-multidomain-map-container"
        style={{ width: "100%", height: "100%" }}
      >
        {props.mapContent}
      </div>

      <div
        data-testid="commander-multidomain-kpi-overlay"
        aria-label="Domain KPI overlay"
        style={{
          position: "absolute",
          top: "0.75rem",
          left: "0.75rem",
          display: "grid",
          "grid-template-columns": "repeat(4, minmax(0, 1fr))",
          gap: "0.5rem",
          "max-width": "46rem",
          padding: "0.5rem",
          background: "rgba(11, 19, 33, 0.8)",
          border: "1px solid rgba(148,163,184,0.2)",
          "border-radius": "8px",
          "backdrop-filter": "blur(3px)",
        }}
      >
        <div style={{ color: "#cbd5e1", "font-size": "0.75rem" }}>Air: 0</div>
        <div style={{ color: "#cbd5e1", "font-size": "0.75rem" }}>Land: 0</div>
        <div style={{ color: "#cbd5e1", "font-size": "0.75rem" }}>
          Maritime: 0
        </div>
        <div style={{ color: "#cbd5e1", "font-size": "0.75rem" }}>Cyber: 0</div>
      </div>

      <div
        data-testid="commander-multidomain-layer-toggles"
        aria-label="Layer toggle controls"
        style={{
          position: "absolute",
          top: "0.75rem",
          right: "0.75rem",
          display: "flex",
          "flex-direction": "column",
          gap: "0.4rem",
          padding: "0.6rem",
          background: "rgba(11, 19, 33, 0.8)",
          border: "1px solid rgba(148,163,184,0.2)",
          "border-radius": "8px",
          "backdrop-filter": "blur(3px)",
        }}
      >
        <For each={layerToggles()}>
          {(item) => (
            <label
              style={{
                display: "flex",
                "align-items": "center",
                gap: "0.4rem",
                color: "#e2e8f0",
                "font-size": "0.72rem",
              }}
            >
              <input
                type="checkbox"
                checked={item.enabled}
                onChange={() => toggleLayer(item.id)}
              />
              <span>{item.label}</span>
            </label>
          )}
        </For>
      </div>
    </div>
  );
}
