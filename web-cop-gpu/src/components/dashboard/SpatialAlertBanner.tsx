// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/SpatialAlertBanner.tsx — Level 3: Bottom Alert Strip
//
// Fixed bottom-of-screen alert banner showing active coverage gaps with RESOLVE action.
// Cycles through multiple alerts with arrow controls.
//
// Reference: docs/implementation/v5/sensordashboard_three_level_plan.md §C2

import { createSignal, Show } from "solid-js";
import type { SpatialAlertPayload } from "../../signals/spatial-alerts";

interface SpatialAlertBannerProps {
  alerts: SpatialAlertPayload[];
  onResolve?: (alertId: string) => void;
}

const SEVERITY_STYLE: Record<SpatialAlertPayload["severity"], { color: string; border: string; bg: string }> = {
  CRITICAL: {
    color: "#fca5a5",
    border: "#ef4444",
    bg: "rgba(239, 68, 68, 0.12)",
  },
  ELEVATED: {
    color: "#fdba74",
    border: "#f97316",
    bg: "rgba(249, 115, 22, 0.12)",
  },
  WATCH: {
    color: "#fcd34d",
    border: "#f59e0b",
    bg: "rgba(245, 158, 11, 0.12)",
  },
};

/**
 * Bottom alert strip that cycles through active spatial alerts.
 * Auto-dismisses when alerts are resolved on backend.
 */
export function SpatialAlertBanner(props: SpatialAlertBannerProps) {
  const [currentIndex, setCurrentIndex] = createSignal(0);

  // Filter to only unacknowledged alerts
  const activeAlerts = () => props.alerts.filter((a) => !a.acknowledged);

  // Clamp index to valid range
  const clampedIndex = () => Math.min(currentIndex(), Math.max(0, activeAlerts().length - 1));

  const currentAlert = () => activeAlerts()[clampedIndex()];

  const canGoPrevious = () => clampedIndex() > 0;
  const canGoNext = () => clampedIndex() < activeAlerts().length - 1;

  const handlePrevious = () => {
    if (canGoPrevious()) {
      setCurrentIndex(clampedIndex() - 1);
    }
  };

  const handleNext = () => {
    if (canGoNext()) {
      setCurrentIndex(clampedIndex() + 1);
    }
  };

  const handleResolve = () => {
    const alert = currentAlert();
    if (alert && props.onResolve) {
      props.onResolve(alert.alertId);
    }
    // Reset to first alert after resolving
    setCurrentIndex(0);
  };

  return (
    <Show when={activeAlerts().length > 0 && currentAlert()}>
      {(alert) => {
        const style = SEVERITY_STYLE[alert().severity];
        return (
          <div
            data-testid="spatial-alert-banner"
            style={{
              position: "fixed",
              bottom: "0",
              left: "0",
              right: "0",
              "z-index": "1000",
              background: style.bg,
              "border-top": `2px solid ${style.border}`,
              "border-left": `4px solid ${style.border}`,
              padding: "12px 20px",
              display: "flex",
              "align-items": "center",
              "justify-content": "space-between",
              gap: "16px",
              "box-shadow": "0 -4px 12px rgba(0, 0, 0, 0.3)",
              "font-family": "system-ui, sans-serif",
            }}
          >
            {/* Alert content */}
            <div
              style={{
                flex: "1",
                display: "flex",
                "align-items": "center",
                gap: "12px",
                "min-width": "0",
              }}
            >
              <span
                style={{
                  "font-size": "0.85rem",
                  "font-weight": "700",
                  color: style.color,
                  "text-transform": "uppercase",
                  "letter-spacing": "0.05em",
                  "white-space": "nowrap",
                }}
              >
                ⚠ SENSOR ALERT
              </span>
              <span
                style={{
                  "font-size": "0.9rem",
                  color: "#e2e8f0",
                  "font-weight": "600",
                  "white-space": "nowrap",
                  "text-overflow": "ellipsis",
                  overflow: "hidden",
                }}
              >
                {alert().affectedSensorId} OFFLINE
              </span>
              <span style={{ color: "#64748b", "font-size": "0.85rem" }}>|</span>
              <span
                style={{
                  "font-size": "0.85rem",
                  color: "#cbd5e1",
                  "white-space": "nowrap",
                  "text-overflow": "ellipsis",
                  overflow: "hidden",
                }}
              >
                {alert().description}
              </span>
              <span style={{ color: "#64748b", "font-size": "0.85rem" }}>|</span>
              <span
                style={{
                  "font-size": "0.75rem",
                  color: "#94a3b8",
                  "font-family": "monospace",
                  "white-space": "nowrap",
                }}
              >
                Last contact: {new Date(alert().lastContactUtc).toISOString().slice(11, 19)}Z
              </span>
            </div>

            {/* Controls */}
            <div
              style={{
                display: "flex",
                "align-items": "center",
                gap: "8px",
                "flex-shrink": "0",
              }}
            >
              {/* Alert counter and navigation */}
              <Show when={activeAlerts().length > 1}>
                <div
                  style={{
                    display: "flex",
                    "align-items": "center",
                    gap: "6px",
                  }}
                >
                  <button
                    onClick={handlePrevious}
                    disabled={!canGoPrevious()}
                    aria-label="Previous alert"
                    style={{
                      background: "transparent",
                      border: `1px solid ${canGoPrevious() ? style.border : "#4b5563"}`,
                      color: canGoPrevious() ? style.color : "#4b5563",
                      "border-radius": "4px",
                      padding: "4px 8px",
                      "font-size": "0.75rem",
                      cursor: canGoPrevious() ? "pointer" : "not-allowed",
                      opacity: canGoPrevious() ? "1" : "0.5",
                      "line-height": "1",
                    }}
                  >
                    ←
                  </button>
                  <span
                    style={{
                      "font-size": "0.75rem",
                      color: "#94a3b8",
                      "font-family": "monospace",
                      "white-space": "nowrap",
                    }}
                  >
                    {clampedIndex() + 1} / {activeAlerts().length}
                  </span>
                  <button
                    onClick={handleNext}
                    disabled={!canGoNext()}
                    aria-label="Next alert"
                    style={{
                      background: "transparent",
                      border: `1px solid ${canGoNext() ? style.border : "#4b5563"}`,
                      color: canGoNext() ? style.color : "#4b5563",
                      "border-radius": "4px",
                      padding: "4px 8px",
                      "font-size": "0.75rem",
                      cursor: canGoNext() ? "pointer" : "not-allowed",
                      opacity: canGoNext() ? "1" : "0.5",
                      "line-height": "1",
                    }}
                  >
                    →
                  </button>
                </div>
              </Show>

              {/* RESOLVE button */}
              <button
                onClick={handleResolve}
                data-testid="resolve-alert-button"
                style={{
                  background: style.border,
                  border: "none",
                  color: "#0a0f1a",
                  "border-radius": "4px",
                  padding: "6px 16px",
                  "font-size": "0.8rem",
                  "font-weight": "700",
                  cursor: "pointer",
                  "text-transform": "uppercase",
                  "letter-spacing": "0.05em",
                  transition: "all 0.2s ease",
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.opacity = "0.8";
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.opacity = "1";
                }}
              >
                RESOLVE
              </button>
            </div>
          </div>
        );
      }}
    </Show>
  );
}
