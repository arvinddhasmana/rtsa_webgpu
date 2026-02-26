<!-- CLASSIFICATION: UNCLASSIFIED -->

# Intelligence Analyst — Role Guide

> **CLASSIFICATION: UNCLASSIFIED**
> **Audience**: Intelligence Analysts, Forensic Investigators
> **Version**: 1.0
> **Last Updated**: 2026-02-26

---

## Your Role in RTSA

As an **Intelligence Analyst**, you use RTSA's historical database to investigate incidents, identify patterns in entity behaviour, and produce intelligence assessments. While Operations Commanders focus on what is happening *right now*, you focus on:

- **What happened** — reconstructing events from historical data
- **Why it happened** — identifying patterns and correlations
- **What it means** — producing analysis that informs future operations

RTSA provides you with up to **90 days of sensor data** and **2 years of audit history** through a powerful query engine backed by ClickHouse, a high-performance analytical database.

---

## Your Quick-Start Checklist

When you begin an investigation:

- [ ] Confirm your clearance level matches the expected classification banner
- [ ] Define your **investigation scope** — entity, time range, geographic area
- [ ] Start with the **Historical Query** panel (accessible from the toolbar: 🔍)
- [ ] Use broad queries first, then narrow down to specific entities or events
- [ ] Review the anomaly history alongside the track history for context
- [ ] Generate a forensic report if required for record-keeping

---

## Your Guide Contents

| Document | What It Covers |
|---|---|
| [Historical Queries](01_historical_queries.md) | How to query the historical database; query types and parameters |
| [Forensic Investigation](02_forensic_analysis.md) | Investigating specific incidents; entity deep-dive analysis |
| [Anomaly Pattern Analysis](03_pattern_analysis.md) | Statistical trend analysis; recurring anomaly type identification |

---

## Key Business Use Cases You Cover

| Use Case | Guide Section |
|---|---|
| UC013 — Historical Query & Forensic Analysis | [Historical Queries](01_historical_queries.md) + [Forensic Investigation](02_forensic_analysis.md) |
| UC009 — Anomaly Detection (review) | [Pattern Analysis](03_pattern_analysis.md) |

---

## Data Available to You

| Data Type | Retention (Data Centre) | Retention (Edge) |
|---|---|---|
| Raw sensor events | 90 days | 7 days |
| Fused entity tracks | 90 days | 7 days |
| Anomaly detection scores | 90 days | 7 days |
| Operator feedback records | 90 days | 7 days |
| Audit events | 2 years | 7 days |

> All data access is **classification-filtered server-side** based on your clearance level. You cannot access data above your clearance — this is enforced by the system, not by any manual process.

---

## Important Reminders

> 🔍 **All queries are logged.** Every query you run — including the parameters and result count — is recorded in the immutable audit trail. This is required by security policy and protects the chain of custody for any investigation.

> ⚠️ **Query guardrails are in place.** The system limits query scope (max 30-day time range per query, max 100,000 result rows, 30-second timeout) to protect performance. If you need to analyse a longer period, run multiple queries.

> 📋 **Parameterize your queries.** All system queries use parameterized inputs to prevent injection attacks. Use the query builder UI rather than raw SQL entry.

---

> **CLASSIFICATION: UNCLASSIFIED** — This document contains no classified information.
