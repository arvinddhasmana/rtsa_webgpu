<!-- CLASSIFICATION: UNCLASSIFIED -->

# Manual Track Nomination for NATO Sharing

> **CLASSIFICATION: UNCLASSIFIED**
> **Audience**: NATO Liaison Officers
> **Version**: 1.0
> **Last Updated**: 2026-02-26

---

## Overview

RTSA automatically exports fused tracks to NATO allies based on configured release policies (classification, confidence threshold, releasability caveats). However, there are situations where an Operations Commander or NATO Liaison Officer needs to **manually nominate a specific track** for NATO sharing — for example, a track that is below the automatic confidence threshold but is operationally significant, or a track that requires immediate allied awareness.

**Manual nomination** allows authorized NATO Liaison Officers to explicitly identify tracks for export, subject to the same mandatory classification guard that applies to automatic exports.

---

## When to Use Manual Nomination

Use manual track nomination when:

- A track with confidence < 0.6 needs to be shared due to operational urgency
- An Operations Commander has directed that a specific track should be shared with allies
- A track has been manually assessed as significant (by operator feedback) but automatic export has not triggered yet
- You are conducting a coordinated operation with an allied partner that requires specific track coverage

**You cannot use manual nomination to:**
- Share data above NATO SECRET classification
- Override NOFORN caveats
- Share data that would otherwise violate GC security policy

---

## How to Nominate a Track

### Step 1 — Find the Track

Option A: From the COP Map
- Click the entity on the map to open the entity detail panel
- Scroll to the **NATO Actions** section at the bottom

Option B: From the NATO Export Panel
- Click 🌐 **NATO** → **Nominate Track**
- Use the search box to find the track by ID

Option C: From an active alert card
- Open the alert (click **Inspect**)
- Select **NATO Actions → Nominate for NATO Sharing**

---

### Step 2 — Review the Track Before Nominating

Before submitting a nomination, verify:

| Check | How to Verify |
|---|---|
| Track ID | Shown in the detail panel header |
| Classification level | Shown in the Identity section — must be ≤ NATO SECRET |
| Releasability caveats | Check for NOFORN, CAN EYES ONLY — if present, you cannot nominate |
| Confidence score | Note if below automatic threshold (< 0.6); requires justification |
| Entity type and hostile status | Understand what you are sharing |

---

### Step 3 — Complete the Nomination Form

```
┌──────────────────────────────────────────────────────────┐
│  MANUAL NATO TRACK NOMINATION                            │
│                                                          │
│  Track ID:        TRK-5721                               │
│  Entity Type:     SURFACE (vessel)                       │
│  Hostile Status:  SUSPECT                                │
│  Confidence:      0.52  ⚠️ (below auto threshold)        │
│  Classification:  NATO RESTRICTED                        │
│  Releasability:   REL TO NATO [no NOFORN] ✅             │
│                                                          │
│  Export Format:                                          │
│  ☑ Link 16 (STANAG 5516)                                 │
│  ☑ NFFI XML                                              │
│  ☐ MIP (observations)                                    │
│                                                          │
│  Justification (required):                               │
│  ┌────────────────────────────────────────────────────┐  │
│  │ Directed by Ops Comd. Vessel showing location      │  │
│  │ anomaly near allied patrol area. Allied awareness  │  │
│  │ required for coordination. Comd authorization: OC  │  │
│  │ Smith, 2026-02-20 14:30 UTC.                       │  │
│  └────────────────────────────────────────────────────┘  │
│                                                          │
│  Release Authority:  [Your name auto-filled from cert]  │
│                                                          │
│  [Cancel]                          [Submit Nomination]   │
└──────────────────────────────────────────────────────────┘
```

| Field | Description |
|---|---|
| **Export Format** | Select which NATO protocol(s) to use. Choose based on the receiving allied system's capability. |
| **Justification** | Required free-text field. Explain why this track should be manually nominated. Include the directing authority if applicable. |
| **Release Authority** | Automatically populated from your certificate identity. You take personal accountability for this nomination. |

---

### Step 4 — Cross-Domain Guard Review

After you submit, the nomination passes through the **cross-domain guard** — the same automated review applied to automatic exports. This takes 1–5 seconds.

**Possible outcomes:**

| Outcome | What It Means |
|---|---|
| ✅ **Approved and Exported** | Guard passed; track transmitted to NATO |
| ❌ **Blocked — NOFORN restriction** | Track has a caveat that prevents export; cannot override |
| ❌ **Blocked — classification too high** | Track is above NATO SECRET; cannot export |
| ❌ **Blocked — source sanitation failed** | Track could not be sanitized sufficiently; export unsafe |
| ✅ **Approved with modification** | Guard sanitized some fields (source details removed); track exported in modified form |

---

## After Nomination — Audit Record

Every manual nomination is recorded in the immutable audit trail with:

- Your identity (from certificate)
- Track ID and its classification at time of nomination
- Justification text
- Cross-domain guard decision
- Format and destination of the export
- Timestamp

This record is permanent and cannot be modified. Your Security Officer has access to the full nomination audit log.

---

## Recalling a Nominated Track

Once a track is exported to NATO systems, **RTSA cannot recall it from the allied systems** — once transmitted, the data exists on the receiving side. If you believe a track was incorrectly nominated:

1. Contact your Security Officer immediately
2. Contact the NATO Liaison at the receiving allied system through your chain of command
3. Document the incident and initiate a spillage report if classification was involved

To **stop ongoing automatic re-export** of a track (if the track continues to be automatically exported and should not be):
1. Contact your system administrator to add the track ID to the export exclusion list
2. Document the exclusion and reason in your duty log

---

> **Back to Role Overview**: [NATO Liaison Officer Guide →](README.md)

---

> **CLASSIFICATION: UNCLASSIFIED** — This document contains no classified information.
