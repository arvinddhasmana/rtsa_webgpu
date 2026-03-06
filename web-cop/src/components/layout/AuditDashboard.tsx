// CLASSIFICATION: UNCLASSIFIED
// src/components/layout/AuditDashboard.tsx
//
// Security Officer — Audit & Feedback Dashboard.
//
// Displays:
//   - Feedback Log: all operator submissions, trust scores, anti-poisoning
//   - Trust Score Histogram: distribution across operators
//   - Audit Log: immutable event stream from ClickHouse
//   - Filters: operator ID, feedback type, time range, trust threshold

import React, { useMemo, useState } from "react";
import { auditClient } from "../../api/audit-client";
import { FeedbackLogEntry, FeedbackLogTable } from "../audit/FeedbackLogTable";
import { TrustScoreHistogram } from "../audit/TrustScoreHistogram";

/* ── Demo seed data (until real backend wires up) ─── */
const DEMO_ENTRIES: FeedbackLogEntry[] = [
  {
    feedbackId: "fb-001",
    operatorId: "op-alpha",
    trackId: "TRK-0042",
    feedbackType: "CONFIRM_HOSTILE",
    trustScore: 0.92,
    accepted: true,
    submittedAt: new Date(Date.now() - 180_000),
  },
  {
    feedbackId: "fb-002",
    operatorId: "op-bravo",
    trackId: "TRK-0017",
    feedbackType: "RECLASSIFY",
    trustScore: 0.45,
    accepted: false,
    submittedAt: new Date(Date.now() - 600_000),
  },
  {
    feedbackId: "fb-003",
    operatorId: "op-alpha",
    trackId: "TRK-0088",
    feedbackType: "REJECT_ANOMALY",
    trustScore: 0.78,
    accepted: true,
    submittedAt: new Date(Date.now() - 1_200_000),
  },
];

/**
 * AuditDashboard — Security Officer default view.
 *
 * Sections:
 *   LEFT: Feedback log table + filters
 *   RIGHT: Trust histogram + audit log stream
 */
export const AuditDashboard: React.FC = () => {
  const [activeTab, setActiveTab] = useState<"feedback" | "audit">("feedback");
  const [auditEvents, setAuditEvents] = useState<
    Array<{ eventId: string; description: string; timestamp: Date }>
  >([]);
  const [loading, setLoading] = useState(false);

  // Load audit log on mount (best-effort — backend may not be running)
  React.useEffect(() => {
    let mounted = true;
    setLoading(true);
    auditClient
      .queryAuditLog({})
      .then((res) => {
        if (mounted) setAuditEvents(res.events);
      })
      .catch(() => {
        /* backend unavailable in dev — silently ignore */
      })
      .finally(() => {
        if (mounted) setLoading(false);
      });
    return () => {
      mounted = false;
    };
  }, []);

  const avgTrust = useMemo(() => {
    if (!DEMO_ENTRIES.length) return 0;
    return (
      DEMO_ENTRIES.reduce((s, e) => s + e.trustScore, 0) / DEMO_ENTRIES.length
    );
  }, []);

  const acceptedCount = DEMO_ENTRIES.filter((e) => e.accepted).length;
  const rejectedCount = DEMO_ENTRIES.length - acceptedCount;

  return (
    <div
      data-testid="audit-dashboard"
      style={{
        flex: 1,
        display: "flex",
        flexDirection: "column",
        overflow: "hidden",
        backgroundColor: "var(--ds-bg-primary, #0F172A)",
      }}
    >
      {/* ── Header ──────────────────────────────────────── */}
      <div
        style={{
          display: "flex",
          alignItems: "center",
          gap: "24px",
          padding: "10px 20px",
          borderBottom: "1px solid #334155",
          backgroundColor: "#0F172A",
        }}
      >
        <span
          style={{
            fontSize: "0.85rem",
            fontWeight: "bold",
            color: "#A855F7",
            letterSpacing: "0.06em",
          }}
        >
          🔐 SECURITY AUDIT
        </span>

        {/* KPIs */}
        <div style={{ display: "flex", gap: "20px" }}>
          <KpiPill label="Submissions" value={DEMO_ENTRIES.length} color="#60A5FA" />
          <KpiPill
            label="Avg Trust"
            value={`${(avgTrust * 100).toFixed(0)}%`}
            color={avgTrust >= 0.7 ? "#10B981" : "#F59E0B"}
          />
          <KpiPill label="Accepted" value={acceptedCount} color="#10B981" />
          <KpiPill label="Rejected" value={rejectedCount} color="#EF4444" />
        </div>

        {/* Tab switcher */}
        <div style={{ marginLeft: "auto", display: "flex", gap: "4px" }}>
          <TabBtn
            label="Feedback Log"
            active={activeTab === "feedback"}
            onClick={() => setActiveTab("feedback")}
          />
          <TabBtn
            label="Audit Log"
            active={activeTab === "audit"}
            onClick={() => setActiveTab("audit")}
          />
        </div>
      </div>

      {/* ── Body ─────────────────────────────────────────── */}
      {activeTab === "feedback" ? (
        <div
          style={{
            flex: 1,
            display: "flex",
            overflow: "hidden",
            gap: 0,
          }}
        >
          {/* Left: Feedback log table */}
          <div
            style={{
              flex: 1,
              overflow: "hidden",
              display: "flex",
              flexDirection: "column",
            }}
          >
            <FeedbackLogTable entries={DEMO_ENTRIES} />
          </div>

          {/* Right: Trust histogram */}
          <div
            style={{
              width: "280px",
              borderLeft: "1px solid #334155",
              padding: "16px",
              display: "flex",
              flexDirection: "column",
              gap: "16px",
            }}
          >
            <div
              style={{
                fontSize: "0.65rem",
                color: "#64748B",
                letterSpacing: "0.05em",
                fontWeight: "bold",
              }}
            >
              TRUST SCORE DISTRIBUTION
            </div>
            <TrustScoreHistogram entries={DEMO_ENTRIES} height={120} />
            <div
              style={{
                fontSize: "0.7rem",
                color: "#94A3B8",
                lineHeight: "1.6",
              }}
            >
              <p>Scores below 40% flag potential anti-poisoning intervention.</p>
              <p style={{ marginTop: "8px" }}>
                Threshold alerts are raised automatically and routed to the
                security review queue.
              </p>
            </div>
          </div>
        </div>
      ) : (
        <div
          style={{ flex: 1, overflow: "auto", padding: "16px" }}
          data-testid="audit-log"
        >
          {loading ? (
            <div style={{ color: "#64748B", fontSize: "0.75rem" }}>
              Loading audit events…
            </div>
          ) : auditEvents.length === 0 ? (
            <div
              style={{
                color: "#64748B",
                fontSize: "0.75rem",
                fontFamily: "monospace",
                lineHeight: "1.8",
              }}
            >
              <div style={{ color: "#10B981" }}>
                [AUDIT] Backend audit log requires ClickHouse connection.
              </div>
              <div>[AUDIT] No events loaded — backend may be offline.</div>
              <div>[AUDIT] Start backend with: scripts/cop-dev/start-backend.sh</div>
            </div>
          ) : (
            auditEvents.map((ev) => (
              <div
                key={ev.eventId}
                style={{
                  fontFamily: "monospace",
                  fontSize: "0.7rem",
                  color: "#CBD5E1",
                  borderBottom: "1px solid rgba(255,255,255,0.03)",
                  padding: "4px 0",
                }}
              >
                [{ev.timestamp.toISOString().slice(11, 19)}Z] {ev.description}
              </div>
            ))
          )}
        </div>
      )}
    </div>
  );
};

const KpiPill: React.FC<{
  label: string;
  value: string | number;
  color: string;
}> = ({ label, value, color }) => (
  <div style={{ display: "flex", alignItems: "center", gap: "6px" }}>
    <span style={{ fontSize: "0.6rem", color: "#64748B" }}>{label}</span>
    <span style={{ fontSize: "0.8rem", fontWeight: "bold", color }}>{value}</span>
  </div>
);

const TabBtn: React.FC<{
  label: string;
  active: boolean;
  onClick: () => void;
}> = ({ label, active, onClick }) => (
  <button
    onClick={onClick}
    style={{
      padding: "4px 12px",
      fontSize: "0.7rem",
      fontWeight: active ? "bold" : "normal",
      backgroundColor: active ? "rgba(168, 85, 247, 0.15)" : "transparent",
      color: active ? "#A855F7" : "#64748B",
      border: `1px solid ${active ? "#A855F7" : "transparent"}`,
      borderRadius: "4px",
      cursor: "pointer",
    }}
  >
    {label}
  </button>
);
