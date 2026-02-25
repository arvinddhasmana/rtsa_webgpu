// CLASSIFICATION: UNCLASSIFIED
// src/components/forensics/QueryBuilder.tsx

import React, { useState } from "react";
import { HistoricalQueryRequest } from "../../api/query-client";

interface QueryBuilderProps {
  onQuery: (req: HistoricalQueryRequest) => void;
  isLoading: boolean;
}

const ENTITY_TYPES = ["SURFACE", "AIR", "SUBSURFACE", "LAND", "CYBER"];
const ANOMALY_TYPES = [
  "SPEED",
  "ROUTE_DEVIATION",
  "AIS_MANIPULATION",
  "BEHAVIORAL",
  "TEMPORAL",
  "PROXIMITY",
];

/**
 * QueryBuilder — form to build historical queries.
 * Time range: max 30 days.
 */
export const QueryBuilder: React.FC<QueryBuilderProps> = ({
  onQuery,
  isLoading,
}) => {
  const now = new Date();
  const yesterday = new Date(now.getTime() - 24 * 60 * 60 * 1000);

  const [startTime, setStartTime] = useState(
    yesterday.toISOString().slice(0, 16)
  );
  const [endTime, setEndTime] = useState(now.toISOString().slice(0, 16));
  const [entityTypes, setEntityTypes] = useState<string[]>([]);
  const [anomalyTypes, setAnomalyTypes] = useState<string[]>([]);
  const [minSeverity, setMinSeverity] = useState("WATCH");

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const start = new Date(startTime);
    const end = new Date(endTime);
    const maxRange = 30 * 24 * 60 * 60 * 1000;
    if (end.getTime() - start.getTime() > maxRange) {
      return; // Enforce max 30-day range
    }
    onQuery({
      startTime: start,
      endTime: end,
      entityTypes: entityTypes.length > 0 ? entityTypes : undefined,
      anomalyTypes: anomalyTypes.length > 0 ? anomalyTypes : undefined,
      minSeverity,
    });
  };

  const toggleSelection = (
    value: string,
    selected: string[],
    setSelected: (v: string[]) => void
  ) => {
    setSelected(
      selected.includes(value)
        ? selected.filter((v) => v !== value)
        : [...selected, value]
    );
  };

  return (
    <form
      data-testid="query-builder"
      onSubmit={(e) => {
        handleSubmit(e);
      }}
      style={{ padding: "8px", display: "flex", flexDirection: "column", gap: "8px" }}
    >
      <div style={{ display: "flex", gap: "8px" }}>
        <div style={{ flex: 1 }}>
          <label style={{ fontSize: "0.7rem", color: "#9CA3AF" }}>Start (UTC)</label>
          <input
            type="datetime-local"
            data-testid="start-time"
            value={startTime}
            onChange={(e) => setStartTime(e.target.value)}
            style={{
              width: "100%",
              padding: "4px",
              backgroundColor: "#0F172A",
              color: "#F1F5F9",
              border: "1px solid #334155",
              borderRadius: "4px",
              fontSize: "0.75rem",
              boxSizing: "border-box",
            }}
          />
        </div>
        <div style={{ flex: 1 }}>
          <label style={{ fontSize: "0.7rem", color: "#9CA3AF" }}>End (UTC)</label>
          <input
            type="datetime-local"
            data-testid="end-time"
            value={endTime}
            onChange={(e) => setEndTime(e.target.value)}
            style={{
              width: "100%",
              padding: "4px",
              backgroundColor: "#0F172A",
              color: "#F1F5F9",
              border: "1px solid #334155",
              borderRadius: "4px",
              fontSize: "0.75rem",
              boxSizing: "border-box",
            }}
          />
        </div>
      </div>

      <div>
        <div style={{ fontSize: "0.7rem", color: "#9CA3AF", marginBottom: "4px" }}>
          Entity Types
        </div>
        <div style={{ display: "flex", flexWrap: "wrap", gap: "4px" }}>
          {ENTITY_TYPES.map((et) => (
            <button
              key={et}
              type="button"
              data-testid={`entity-type-${et.toLowerCase()}`}
              onClick={() => toggleSelection(et, entityTypes, setEntityTypes)}
              style={{
                padding: "2px 8px",
                fontSize: "0.65rem",
                backgroundColor: entityTypes.includes(et) ? "#1D4ED8" : "#374151",
                color: "#F1F5F9",
                border: "1px solid #475569",
                borderRadius: "4px",
                cursor: "pointer",
              }}
            >
              {et}
            </button>
          ))}
        </div>
      </div>

      <div>
        <div style={{ fontSize: "0.7rem", color: "#9CA3AF", marginBottom: "4px" }}>
          Anomaly Types
        </div>
        <div style={{ display: "flex", flexWrap: "wrap", gap: "4px" }}>
          {ANOMALY_TYPES.map((at) => (
            <button
              key={at}
              type="button"
              data-testid={`anomaly-type-${at.toLowerCase()}`}
              onClick={() => toggleSelection(at, anomalyTypes, setAnomalyTypes)}
              style={{
                padding: "2px 8px",
                fontSize: "0.65rem",
                backgroundColor: anomalyTypes.includes(at) ? "#7C3AED" : "#374151",
                color: "#F1F5F9",
                border: "1px solid #475569",
                borderRadius: "4px",
                cursor: "pointer",
              }}
            >
              {at.replace("_", " ")}
            </button>
          ))}
        </div>
      </div>

      <div>
        <label style={{ fontSize: "0.7rem", color: "#9CA3AF" }}>
          Min Severity
        </label>
        <select
          data-testid="min-severity"
          value={minSeverity}
          onChange={(e) => setMinSeverity(e.target.value)}
          style={{
            marginLeft: "8px",
            padding: "2px 4px",
            backgroundColor: "#0F172A",
            color: "#F1F5F9",
            border: "1px solid #334155",
            borderRadius: "4px",
            fontSize: "0.75rem",
          }}
        >
          {["NORMAL", "WATCH", "ELEVATED", "CRITICAL"].map((s) => (
            <option key={s} value={s}>
              {s}
            </option>
          ))}
        </select>
      </div>

      <button
        type="submit"
        data-testid="run-query"
        disabled={isLoading}
        style={{
          padding: "6px",
          backgroundColor: isLoading ? "#374151" : "#1D4ED8",
          color: "#F1F5F9",
          border: "none",
          borderRadius: "4px",
          cursor: isLoading ? "not-allowed" : "pointer",
          fontSize: "0.75rem",
          fontWeight: "bold",
        }}
      >
        {isLoading ? "Running..." : "Run Query"}
      </button>
    </form>
  );
};
