// CLASSIFICATION: UNCLASSIFIED
// src/components/layout/SensorHealthDashboard.tsx

import React, { useCallback, useEffect, useRef, useState } from "react";
import { useSensorHealth } from "../../hooks/useSensorHealth";
import {
    getSensorKPIs,
    getSortedSensors,
    useSensorHealthStore,
} from "../../stores/sensorHealthStore";
import type { SensorStatus, SortField } from "../../types/sensor";
import { MapView } from "../map/MapView";
import { DLQPopup } from "../sensor/DLQPopup";
import { DLQViewer } from "../sensor/DLQViewer";
import { Sparkline } from "../sensor/Sparkline";

const DEFAULT_LEFT_PERCENT = 40;
const MIN_LEFT_PERCENT = 20;
const MAX_LEFT_PERCENT = 70;

/**
 * SensorHealthDashboard — Variant C split-pane layout.
 * LEFT: Tabbed (Sensor Grid / Dead Letter Queue) with KPI tiles, sortable table, inline DLQ popup.
 * RIGHT: Map with sensor coverage overlays.
 * Resizable panes with drag handle. Escape resets to default.
 */
export const SensorHealthDashboard: React.FC = () => {
  // Start sensor health polling
  useSensorHealth();

  // Store state
  const sensors = useSensorHealthStore((s) => s.sensors);
  const activeTab = useSensorHealthStore((s) => s.activeTab);
  const setActiveTab = useSensorHealthStore((s) => s.setActiveTab);
  const sortField = useSensorHealthStore((s) => s.sortField);
  const sortDirection = useSensorHealthStore((s) => s.sortDirection);
  const setSortField = useSensorHealthStore((s) => s.setSortField);
  const selectedSensorId = useSensorHealthStore((s) => s.selectedSensorId);
  const selectSensor = useSensorHealthStore((s) => s.selectSensor);
  const setDLQPopupSensorId = useSensorHealthStore(
    (s) => s.setDLQPopupSensorId
  );
  const isLoading = useSensorHealthStore((s) => s.isLoading);

  // Derived data
  const kpis = getSensorKPIs(sensors);
  const sortedSensors = getSortedSensors(sensors, sortField, sortDirection);

  // Resizable pane state
  const [leftPercent, setLeftPercent] = useState(DEFAULT_LEFT_PERCENT);
  const containerRef = useRef<HTMLDivElement>(null);
  const isDraggingRef = useRef(false);

  // Escape to reset layout
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        setLeftPercent(DEFAULT_LEFT_PERCENT);
        selectSensor(null);
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [selectSensor]);

  // Drag resize handler
  const onMouseDown = useCallback(
    (e: React.MouseEvent) => {
      e.preventDefault();
      isDraggingRef.current = true;
      document.body.style.cursor = "col-resize";
      document.body.style.userSelect = "none";

      const onMouseMove = (ev: MouseEvent) => {
        if (!isDraggingRef.current || !containerRef.current) return;
        const rect = containerRef.current.getBoundingClientRect();
        const pct = ((ev.clientX - rect.left) / rect.width) * 100;
        setLeftPercent(
          Math.max(MIN_LEFT_PERCENT, Math.min(MAX_LEFT_PERCENT, pct))
        );
      };

      const onMouseUp = () => {
        isDraggingRef.current = false;
        document.body.style.cursor = "";
        document.body.style.userSelect = "";
        document.removeEventListener("mousemove", onMouseMove);
        document.removeEventListener("mouseup", onMouseUp);
      };

      document.addEventListener("mousemove", onMouseMove);
      document.addEventListener("mouseup", onMouseUp);
    },
    []
  );

  return (
    <div
      ref={containerRef}
      data-testid="sensor-health-dashboard"
      style={{
        flex: 1,
        display: "flex",
        overflow: "hidden",
        position: "relative",
      }}
    >
      {/* ── Left Panel ────────────────────────────── */}
      <div
        style={{
          width: `${leftPercent}%`,
          minWidth: "280px",
          display: "flex",
          flexDirection: "column",
          background: "var(--ds-bg-primary)",
          borderRight: "1px solid var(--ds-border-default)",
          overflow: "hidden",
        }}
      >
        {/* Tabs */}
        <div className="ds-tabs">
          <button
            className={`ds-tab ${activeTab === "grid" ? "ds-tab--active" : ""}`}
            onClick={() => setActiveTab("grid")}
            data-testid="tab-sensor-grid"
          >
            Sensor Grid
          </button>
          <button
            className={`ds-tab ${activeTab === "dlq" ? "ds-tab--active" : ""}`}
            onClick={() => setActiveTab("dlq")}
            data-testid="tab-dlq"
          >
            Dead Letter Queue
          </button>
        </div>

        {/* Tab Content */}
        {activeTab === "grid" ? (
          <SensorGridTab
            kpis={kpis}
            sensors={sortedSensors}
            sortField={sortField}
            sortDirection={sortDirection}
            setSortField={setSortField}
            selectedSensorId={selectedSensorId}
            selectSensor={selectSensor}
            setDLQPopupSensorId={setDLQPopupSensorId}
            isLoading={isLoading}
          />
        ) : (
          <DLQViewer />
        )}
      </div>

      {/* ── Resize Handle ─────────────────────────── */}
      <div
        className="ds-resize-handle"
        onMouseDown={onMouseDown}
        role="separator"
        aria-label="Resize panels"
        tabIndex={0}
        data-testid="resize-handle"
      />

      {/* ── Right Panel (Map) ─────────────────────── */}
      <div
        style={{
          flex: 1,
          overflow: "hidden",
          position: "relative",
        }}
        aria-label="Map View"
        role="region"
        tabIndex={0}
      >
        <MapView />
      </div>

      {/* ── DLQ Popup Overlay ─────────────────────── */}
      <DLQPopup />
    </div>
  );
};

// ── SensorGridTab ─────────────────────────────────────

interface SensorGridTabProps {
  kpis: ReturnType<typeof getSensorKPIs>;
  sensors: SensorStatus[];
  sortField: SortField;
  sortDirection: "asc" | "desc";
  setSortField: (field: SortField) => void;
  selectedSensorId: string | null;
  selectSensor: (id: string | null) => void;
  setDLQPopupSensorId: (id: string | null) => void;
  isLoading: boolean;
}

const SensorGridTab: React.FC<SensorGridTabProps> = ({
  kpis,
  sensors,
  sortField,
  sortDirection,
  setSortField,
  selectedSensorId,
  selectSensor,
  setDLQPopupSensorId,
  isLoading,
}) => {
  if (isLoading) {
    return (
      <div
        style={{
          flex: 1,
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          color: "var(--ds-text-muted)",
        }}
      >
        <div style={{ textAlign: "center" }}>
          <div style={{ fontSize: "24px", marginBottom: "8px" }}>📡</div>
          <div>Awaiting sensor telemetry...</div>
        </div>
      </div>
    );
  }

  return (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        flex: 1,
        overflow: "hidden",
      }}
    >
      {/* KPI Tiles */}
      <div
        style={{
          display: "grid",
          gridTemplateColumns: "1fr 1fr 1fr 1fr",
          gap: "var(--ds-space-sm)",
          padding: "var(--ds-space-md) var(--ds-space-lg)",
        }}
      >
        <div className="ds-kpi-tile" data-testid="kpi-active">
          <span
            className="ds-kpi-value"
            style={{ color: "var(--ds-status-ok)" }}
          >
            {kpis.active}
          </span>
          <span className="ds-kpi-label">Active</span>
        </div>
        <div className="ds-kpi-tile" data-testid="kpi-degraded">
          <span
            className="ds-kpi-value"
            style={{
              color:
                kpis.degraded + kpis.offline > 0
                  ? "var(--ds-status-error)"
                  : "var(--ds-text-muted)",
            }}
          >
            {kpis.degraded + kpis.offline}
          </span>
          <span className="ds-kpi-label">Degraded</span>
        </div>
        <div className="ds-kpi-tile" data-testid="kpi-throughput">
          <span
            className="ds-kpi-value"
            style={{ color: "var(--ds-accent-blue)" }}
          >
            {kpis.totalThroughput.toFixed(0)}
          </span>
          <span className="ds-kpi-label">EPS</span>
        </div>
        <div className="ds-kpi-tile" data-testid="kpi-latency">
          <span
            className="ds-kpi-value"
            style={{ color: "var(--ds-accent-amber)" }}
          >
            {Math.round(kpis.avgLatency)}
            <span style={{ fontSize: "var(--ds-text-xs)" }}>ms</span>
          </span>
          <span className="ds-kpi-label">Latency</span>
        </div>
      </div>

      {/* Sensor Table */}
      <div className="ds-scroll" style={{ flex: 1, overflow: "auto" }}>
        <table className="ds-table" data-testid="sensor-table">
          <thead>
            <tr>
              <SortHeader
                field="connectionStatus"
                label="Status"
                current={sortField}
                direction={sortDirection}
                onSort={setSortField}
              />
              <SortHeader
                field="sensorId"
                label="Sensor ID"
                current={sortField}
                direction={sortDirection}
                onSort={setSortField}
              />
              <SortHeader
                field="sensorType"
                label="Type"
                current={sortField}
                direction={sortDirection}
                onSort={setSortField}
              />
              <SortHeader
                field="eventsPerSecond"
                label="Rate"
                current={sortField}
                direction={sortDirection}
                onSort={setSortField}
              />
              <SortHeader
                field="totalRejected"
                label="DLQ"
                current={sortField}
                direction={sortDirection}
                onSort={setSortField}
              />
              <SortHeader
                field="latencyMs"
                label="Latency"
                current={sortField}
                direction={sortDirection}
                onSort={setSortField}
              />
              <SortHeader
                field="lastObservationTime"
                label="Last Seen"
                current={sortField}
                direction={sortDirection}
                onSort={setSortField}
              />
            </tr>
          </thead>
          <tbody>
            {sensors.map((sensor) => (
              <React.Fragment key={sensor.sensorId}>
                <tr
                  className={
                    selectedSensorId === sensor.sensorId
                      ? "ds-row-selected"
                      : ""
                  }
                  onClick={() =>
                    selectSensor(
                      selectedSensorId === sensor.sensorId
                        ? null
                        : sensor.sensorId
                    )
                  }
                  data-testid={`sensor-row-${sensor.sensorId}`}
                >
                  <td>
                    <span
                      className={`ds-status-dot ds-status-dot--${
                        sensor.connectionStatus === "connected"
                          ? "ok"
                          : sensor.connectionStatus === "degraded"
                            ? "warn"
                            : "error"
                      }`}
                    />
                  </td>
                  <td>
                    <span style={{ fontWeight: 600 }}>{sensor.sensorId}</span>
                  </td>
                  <td>
                    <span style={{ color: "var(--ds-text-secondary)" }}>
                      {getSensorIcon(sensor.sensorType)} {sensor.sensorType}
                    </span>
                  </td>
                  <td>
                    <span
                      style={{
                        fontFamily: "var(--ds-font-mono)",
                        fontSize: "var(--ds-text-xs)",
                      }}
                    >
                      {sensor.eventsPerSecond.toFixed(1)}
                    </span>
                    <Sparkline
                      data={sensor.rateHistory}
                      width={50}
                      height={16}
                    />
                  </td>
                  <td>
                    <div
                      style={{
                        display: "flex",
                        alignItems: "center",
                        gap: "4px",
                      }}
                    >
                      <span
                        style={{
                          fontFamily: "var(--ds-font-mono)",
                          fontSize: "var(--ds-text-xs)",
                          color:
                            sensor.totalRejected > 0
                              ? "var(--ds-status-error)"
                              : "var(--ds-text-muted)",
                        }}
                      >
                        {sensor.totalRejected}
                      </span>
                      {sensor.totalRejected > 0 && (
                        <button
                          className="ds-btn-icon"
                          onClick={(e) => {
                            e.stopPropagation();
                            setDLQPopupSensorId(sensor.sensorId);
                          }}
                          aria-label={`View DLQ details for ${sensor.sensorId}`}
                          title="View DLQ breakdown"
                          data-testid={`dlq-icon-${sensor.sensorId}`}
                          style={{ padding: "2px 4px", fontSize: "12px" }}
                        >
                          📊
                        </button>
                      )}
                    </div>
                  </td>
                  <td>
                    <span
                      style={{
                        fontFamily: "var(--ds-font-mono)",
                        fontSize: "var(--ds-text-xs)",
                        color:
                          sensor.latencyMs > 500
                            ? "var(--ds-status-error)"
                            : sensor.latencyMs > 200
                              ? "var(--ds-status-warn)"
                              : "var(--ds-text-primary)",
                      }}
                    >
                      {sensor.latencyMs}ms
                    </span>
                  </td>
                  <td>
                    <span
                      style={{
                        fontFamily: "var(--ds-font-mono)",
                        fontSize: "var(--ds-text-2xs)",
                        color: "var(--ds-text-muted)",
                      }}
                    >
                      {sensor.lastObservationTime
                        ? formatTimeCompact(sensor.lastObservationTime)
                        : "—"}
                    </span>
                  </td>
                </tr>

                {/* Inline Expansion */}
                {selectedSensorId === sensor.sensorId && (
                  <tr className="ds-expand-row">
                    <td colSpan={7}>
                      <SensorDrillDown sensor={sensor} />
                    </td>
                  </tr>
                )}
              </React.Fragment>
            ))}

            {sensors.length === 0 && (
              <tr>
                <td
                  colSpan={7}
                  style={{
                    textAlign: "center",
                    color: "var(--ds-text-muted)",
                    padding: "40px",
                  }}
                >
                  No sensors registered
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
};

// ── SortHeader ────────────────────────────────────────

interface SortHeaderProps {
  field: SortField;
  label: string;
  current: SortField;
  direction: "asc" | "desc";
  onSort: (field: SortField) => void;
}

const SortHeader: React.FC<SortHeaderProps> = ({
  field,
  label,
  current,
  direction,
  onSort,
}) => (
  <th
    className={current === field ? "ds-sort-active" : ""}
    onClick={() => onSort(field)}
  >
    {label}{" "}
    {current === field ? (direction === "asc" ? "▲" : "▼") : ""}
  </th>
);

// ── SensorDrillDown ───────────────────────────────────

const SensorDrillDown: React.FC<{ sensor: SensorStatus }> = ({ sensor }) => {
  return (
    <div
      data-testid={`sensor-detail-${sensor.sensorId}`}
      style={{
        display: "grid",
        gridTemplateColumns: "1fr 1fr 1fr",
        gap: "var(--ds-space-lg)",
      }}
    >
      {/* Identity */}
      <div>
        <div className="ds-section-header">Identity</div>
        <DetailRow label="Sensor ID" value={sensor.sensorId} />
        <DetailRow label="Type" value={`${getSensorIcon(sensor.sensorType)} ${sensor.sensorType}`} />
        <DetailRow
          label="Status"
          value={sensor.connectionStatus.toUpperCase()}
          valueColor={
            sensor.connectionStatus === "connected"
              ? "var(--ds-status-ok)"
              : sensor.connectionStatus === "degraded"
                ? "var(--ds-status-warn)"
                : "var(--ds-status-error)"
          }
        />
      </div>

      {/* Metrics */}
      <div>
        <div className="ds-section-header">Metrics</div>
        <DetailRow
          label="Events/sec"
          value={sensor.eventsPerSecond.toFixed(1)}
        />
        <DetailRow
          label="Total Received"
          value={sensor.totalReceived.toLocaleString()}
        />
        <DetailRow
          label="Total Accepted"
          value={sensor.totalAccepted.toLocaleString()}
        />
        <DetailRow
          label="Total Rejected"
          value={sensor.totalRejected.toLocaleString()}
          valueColor={
            sensor.totalRejected > 0 ? "var(--ds-status-error)" : undefined
          }
        />
        <DetailRow
          label="Acceptance Rate"
          value={`${sensor.acceptanceRate.toFixed(1)}%`}
        />
      </div>

      {/* Coverage */}
      <div>
        <div className="ds-section-header">Coverage</div>
        {sensor.coverage?.sensorPosition ? (
          <>
            <DetailRow
              label="Position"
              value={`${sensor.coverage.sensorPosition.latitude.toFixed(2)}°N, ${Math.abs(sensor.coverage.sensorPosition.longitude).toFixed(2)}°${sensor.coverage.sensorPosition.longitude < 0 ? "W" : "E"}`}
            />
            {sensor.coverage.rangeNm && (
              <DetailRow
                label="Range"
                value={`${sensor.coverage.rangeNm} NM`}
              />
            )}
            {sensor.coverage.bearingStartDegrees !== undefined &&
              sensor.coverage.bearingEndDegrees !== undefined && (
                <DetailRow
                  label="Sector"
                  value={`${sensor.coverage.bearingStartDegrees}°–${sensor.coverage.bearingEndDegrees}°`}
                />
              )}
          </>
        ) : (
          <div
            style={{
              fontSize: "var(--ds-text-xs)",
              color: "var(--ds-text-muted)",
            }}
          >
            No coverage data
          </div>
        )}
        {sensor.lastObservationTime && (
          <DetailRow
            label="Last Seen"
            value={sensor.lastObservationTime.toISOString().substring(11, 19) + "Z"}
          />
        )}
      </div>
    </div>
  );
};

// ── DetailRow ─────────────────────────────────────────

const DetailRow: React.FC<{
  label: string;
  value: string;
  valueColor?: string;
}> = ({ label, value, valueColor }) => (
  <div
    style={{
      display: "flex",
      justifyContent: "space-between",
      padding: "3px 0",
      fontSize: "var(--ds-text-xs)",
    }}
  >
    <span style={{ color: "var(--ds-text-muted)" }}>{label}</span>
    <span
      style={{
        fontFamily: "var(--ds-font-mono)",
        color: valueColor || "var(--ds-text-primary)",
        fontWeight: 500,
      }}
    >
      {value}
    </span>
  </div>
);

// ── Helpers ──────────────────────────────────────────

function getSensorIcon(type: string): string {
  switch (type.toUpperCase()) {
    case "RADAR":
      return "📡";
    case "EW":
      return "📻";
    case "ELINT":
      return "🔊";
    case "ISR":
      return "🛰️";
    case "AIS":
      return "🚢";
    case "CYBER":
      return "💻";
    default:
      return "📡";
  }
}

function formatTimeCompact(date: Date): string {
  const diff = Date.now() - date.getTime();
  if (diff < 60_000) return `${Math.floor(diff / 1000)}s ago`;
  if (diff < 3600_000) return `${Math.floor(diff / 60_000)}m ago`;
  return date.toISOString().substring(11, 19) + "Z";
}
