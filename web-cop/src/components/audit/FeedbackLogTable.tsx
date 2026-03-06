// CLASSIFICATION: UNCLASSIFIED
// src/components/audit/FeedbackLogTable.tsx
//
// Sortable/filterable table of operator feedback submissions.
// Used in the Security Officer AuditDashboard.

import React, { useMemo, useState } from "react";

export interface FeedbackLogEntry {
  feedbackId: string;
  operatorId: string;
  trackId: string;
  feedbackType: string;
  trustScore: number;
  accepted: boolean;
  submittedAt: Date;
}

interface FeedbackLogTableProps {
  entries: FeedbackLogEntry[];
}

type SortKey = "submittedAt" | "operatorId" | "feedbackType" | "trustScore";

function trustColor(score: number): string {
  if (score >= 0.8) return "#10B981";
  if (score >= 0.5) return "#F59E0B";
  return "#EF4444";
}

/**
 * FeedbackLogTable — sortable, filterable log of operator feedback submissions.
 * Displays trust scores and accepted/rejected decisions.
 */
export const FeedbackLogTable: React.FC<FeedbackLogTableProps> = ({
  entries,
}) => {
  const [sortKey, setSortKey] = useState<SortKey>("submittedAt");
  const [sortAsc, setSortAsc] = useState(false);
  const [filterType, setFilterType] = useState("");
  const [filterOperator, setFilterOperator] = useState("");

  const sorted = useMemo(() => {
    let list = [...entries];
    if (filterType) {
      list = list.filter((e) =>
        e.feedbackType.toLowerCase().includes(filterType.toLowerCase())
      );
    }
    if (filterOperator) {
      list = list.filter((e) =>
        e.operatorId.toLowerCase().includes(filterOperator.toLowerCase())
      );
    }
    list.sort((a, b) => {
      let cmp = 0;
      switch (sortKey) {
        case "submittedAt":
          cmp = a.submittedAt.getTime() - b.submittedAt.getTime();
          break;
        case "operatorId":
          cmp = a.operatorId.localeCompare(b.operatorId);
          break;
        case "feedbackType":
          cmp = a.feedbackType.localeCompare(b.feedbackType);
          break;
        case "trustScore":
          cmp = a.trustScore - b.trustScore;
          break;
      }
      return sortAsc ? cmp : -cmp;
    });
    return list;
  }, [entries, sortKey, sortAsc, filterType, filterOperator]);

  const handleSort = (key: SortKey) => {
    if (key === sortKey) setSortAsc((v) => !v);
    else {
      setSortKey(key);
      setSortAsc(false);
    }
  };

  const SortBtn: React.FC<{ field: SortKey; label: string }> = ({
    field,
    label,
  }) => (
    <th
      onClick={() => handleSort(field)}
      style={{
        cursor: "pointer",
        padding: "6px 10px",
        fontSize: "0.65rem",
        fontWeight: "bold",
        color: sortKey === field ? "#60A5FA" : "#94A3B8",
        whiteSpace: "nowrap",
        textAlign: "left",
        borderBottom: "1px solid #334155",
        userSelect: "none",
      }}
    >
      {label} {sortKey === field ? (sortAsc ? "↑" : "↓") : ""}
    </th>
  );

  return (
    <div
      data-testid="feedback-log-table"
      style={{
        display: "flex",
        flexDirection: "column",
        height: "100%",
        overflow: "hidden",
      }}
    >
      {/* Filter row */}
      <div
        style={{
          display: "flex",
          gap: "8px",
          padding: "8px 12px",
          borderBottom: "1px solid #334155",
        }}
      >
        <input
          placeholder="Filter by type…"
          value={filterType}
          onChange={(e) => setFilterType(e.target.value)}
          data-testid="feedback-filter-type"
          style={{
            flex: 1,
            backgroundColor: "#1E293B",
            border: "1px solid #334155",
            borderRadius: "4px",
            color: "#F1F5F9",
            padding: "4px 8px",
            fontSize: "0.7rem",
          }}
        />
        <input
          placeholder="Filter by operator…"
          value={filterOperator}
          onChange={(e) => setFilterOperator(e.target.value)}
          data-testid="feedback-filter-operator"
          style={{
            flex: 1,
            backgroundColor: "#1E293B",
            border: "1px solid #334155",
            borderRadius: "4px",
            color: "#F1F5F9",
            padding: "4px 8px",
            fontSize: "0.7rem",
          }}
        />
      </div>

      {/* Table */}
      <div style={{ flex: 1, overflowY: "auto" }}>
        <table
          style={{ width: "100%", borderCollapse: "collapse" }}
        >
          <thead style={{ position: "sticky", top: 0, backgroundColor: "#0F172A", zIndex: 1 }}>
            <tr>
              <SortBtn field="submittedAt" label="Time" />
              <SortBtn field="operatorId" label="Operator" />
              <th style={{ padding: "6px 10px", fontSize: "0.65rem", color: "#94A3B8", textAlign: "left", borderBottom: "1px solid #334155" }}>
                Track
              </th>
              <SortBtn field="feedbackType" label="Type" />
              <SortBtn field="trustScore" label="Trust" />
              <th style={{ padding: "6px 10px", fontSize: "0.65rem", color: "#94A3B8", textAlign: "left", borderBottom: "1px solid #334155" }}>
                Decision
              </th>
            </tr>
          </thead>
          <tbody>
            {sorted.length === 0 ? (
              <tr>
                <td
                  colSpan={6}
                  style={{
                    padding: "24px",
                    textAlign: "center",
                    color: "#64748B",
                    fontSize: "0.75rem",
                  }}
                >
                  No feedback entries found
                </td>
              </tr>
            ) : (
              sorted.map((entry, idx) => (
                <tr
                  key={entry.feedbackId}
                  data-testid={`feedback-row-${entry.feedbackId}`}
                  style={{
                    backgroundColor:
                      idx % 2 === 0
                        ? "rgba(255,255,255,0.02)"
                        : "transparent",
                    borderBottom: "1px solid rgba(255,255,255,0.03)",
                  }}
                >
                  <td
                    style={{
                      padding: "6px 10px",
                      fontSize: "0.65rem",
                      fontFamily: "monospace",
                      color: "#64748B",
                      whiteSpace: "nowrap",
                    }}
                  >
                    {entry.submittedAt.toISOString().slice(11, 19)}Z
                  </td>
                  <td
                    style={{
                      padding: "6px 10px",
                      fontSize: "0.7rem",
                      color: "#CBD5E1",
                    }}
                  >
                    {entry.operatorId}
                  </td>
                  <td
                    style={{
                      padding: "6px 10px",
                      fontSize: "0.65rem",
                      fontFamily: "monospace",
                      color: "#60A5FA",
                    }}
                  >
                    {entry.trackId.slice(-8)}
                  </td>
                  <td
                    style={{
                      padding: "6px 10px",
                      fontSize: "0.65rem",
                      color: "#94A3B8",
                    }}
                  >
                    {entry.feedbackType.replace(/_/g, " ")}
                  </td>
                  <td
                    style={{
                      padding: "6px 10px",
                      fontSize: "0.7rem",
                      fontFamily: "monospace",
                      fontWeight: "bold",
                      color: trustColor(entry.trustScore),
                    }}
                  >
                    {(entry.trustScore * 100).toFixed(0)}%
                  </td>
                  <td
                    style={{
                      padding: "6px 10px",
                      fontSize: "0.65rem",
                      color: entry.accepted ? "#10B981" : "#EF4444",
                      fontWeight: "bold",
                    }}
                  >
                    {entry.accepted ? "✓ ACCEPTED" : "✗ REJECTED"}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
};
