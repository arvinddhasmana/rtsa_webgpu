<!-- CLASSIFICATION: UNCLASSIFIED -->

# NATO Liaison Officer — Role Guide

> **CLASSIFICATION: UNCLASSIFIED**
> **Audience**: NATO Liaison Officers, NATO Interoperability Officers
> **Version**: 1.0
> **Last Updated**: 2026-02-26

---

## Your Role in RTSA

As a **NATO Liaison Officer (LO)**, you are responsible for managing the bidirectional exchange of tactical data between RTSA and NATO-allied systems. You work at the intersection of Canadian and allied situational awareness, ensuring the right information is shared with allied partners while rigorous classification controls are maintained.

Your primary responsibilities are:

- Monitoring **NATO data exchange** health — are outbound and inbound links functioning?
- Reviewing the **NATO export audit log** — what track data has been shared with allies?
- **Manually nominating tracks** for NATO sharing when automatic export is insufficient
- Coordinating **exercise mode** data exchange distinct from live operations
- Responding to **link degradation** or exchange failures

---

## Your Quick-Start Checklist

At the start of each shift:

- [ ] Open the **NATO Exchange Panel** (🌐 NATO in toolbar)
- [ ] Verify both **Link 16** and **NFFI/MIP** link status are **Active**
- [ ] Check the **export throughput** — are tracks being shared at expected rate?
- [ ] Review the **export audit log** for the last hour — spot check for anomalies
- [ ] Confirm **Exercise Mode** status matches your current operational context

---

## Your Guide Contents

| Document | What It Covers |
|---|---|
| [NATO Data Exchange](01_nato_data_exchange.md) | How RTSA exchanges data with NATO; monitoring link health and export logs |
| [Manual Track Nomination](02_track_nomination.md) | Manually nominating specific tracks for NATO sharing |

---

## Key Business Use Cases You Cover

| Use Case | Guide Section |
|---|---|
| UC014 — NATO Outbound Data Exchange | [NATO Data Exchange](01_nato_data_exchange.md) + [Track Nomination](02_track_nomination.md) |
| UC015 — NATO Inbound Data Exchange | [NATO Data Exchange](01_nato_data_exchange.md) |

---

## Classification Responsibilities

All NATO data exchange is subject to strict classification controls enforced automatically by the system. Your role is oversight, not manual enforcement. Key rules to understand:

| Rule | Description |
|---|---|
| **Maximum outbound classification** | Only tracks ≤ NATO SECRET are exported |
| **NOFORN blocks** | Tracks marked NOFORN are never exported (automatic) |
| **Source sanitization** | Specific sensor attribution is removed; only "MULTI-SOURCE" is shared |
| **Confidence threshold** | Only tracks with confidence ≥ 0.6 are exported automatically |
| **Releasability caveats** | REL TO NATO is applied to all outbound data |

If you observe data being exported that appears inconsistent with these rules, immediately contact your Security Officer.

---

## Important Reminders

> 🌐 **All NATO exchanges are immutably audited.** Every track export and every inbound reception is logged with track ID, format, timestamp, and classification.

> ⚠️ **You cannot override classification controls.** The cross-domain guard is system-enforced. Manual track nominations are still subject to the same classification rules.

> 📋 **Exercise mode is separate from live operations.** Always verify exercise mode status before beginning an exercise to avoid contaminating the live operational picture.

---

> **CLASSIFICATION: UNCLASSIFIED** — This document contains no classified information.
