<!-- CLASSIFICATION: UNCLASSIFIED -->

# Classification Management

> **CLASSIFICATION: UNCLASSIFIED**
> **Audience**: Security Officers, Security Authorities
> **Version**: 1.0
> **Last Updated**: 2026-02-26

---

## Overview

RTSA enforces security classification at every data flow boundary — from sensor ingestion to operator display. This is **automatic and policy-driven**: the system never relies on users to enforce classification manually. Your role as Security Officer is to **verify** that enforcement is working correctly, **investigate** any potential incidents, and **respond** when a potential spillage is detected.

---

## How Classification Enforcement Works

### Classification at Every Layer

```
Sensor Data → [Classification tag added by ingestion service]
            → [Redpanda event stream — classification metadata on every message]
            → [Fusion Engine — classification propagated to fused track]
            → [Anomaly Detection — classification carried through]
            → [Track Service — classification metadata on every streamed update]
            → [API Gateway — clearance check before data reaches client]
            → [COP UI — classification banner reflects highest level on screen]
```

No component removes or downgrades classification metadata without an authorized downgrade operation recorded in the audit trail.

---

## Understanding Classification Levels in RTSA

| Level | Description | Display Colour |
|---|---|---|
| UNCLASSIFIED | Public or non-sensitive | Green |
| PROTECTED A | Low-sensitivity personal or operational data | Blue |
| PROTECTED B | Sensitive operational data | Yellow |
| PROTECTED C | Highly sensitive; significant harm if disclosed | Orange |
| SECRET | Very sensitive; serious harm if disclosed | Red |

---

## Verifying Classification Enforcement

### Daily Verification Checklist

- [ ] Open the Audit Trail Viewer and check for `classification_access_warning` events in the last 24 hours
- [ ] Verify that the COP dashboard shows the correct classification banner colour for its current data
- [ ] Check that no user has been granted access above their recorded clearance level (use User Management panel)
- [ ] Review any `nato_export_blocked` events — verify NOFORN and releasability rules are being applied correctly

---

## Managing User Clearance Profiles

User clearance levels in RTSA are derived from the certificate presented during authentication. The certificate is issued by your unit's PKI infrastructure and contains the user's authorized clearance level.

To view a user's current clearance profile:

1. Go to ⚙️ **Settings → Security → User Profiles**
2. Search for the user by name or certificate CN
3. The profile shows:
   - Current clearance level
   - Last authentication time
   - Active session(s)
   - Recent access events (last 10)

> **You cannot modify clearance levels directly in RTSA.** Clearance changes are made at the PKI certificate level (contact your unit's certificate authority). RTSA reads the clearance from the certificate at each authentication.

---

## Responding to a Potential Classification Spillage

A **classification spillage** occurs when classified data is exposed to a user, system, or channel that is not authorized for that classification level.

### Immediate Response Steps

1. **Do not attempt to delete the data** — it is immutably logged in the audit trail.
2. **Identify what was exposed**: Go to the Audit Trail Viewer → filter by the affected user's session → find all data access events in the suspected time window.
3. **Identify who was present**: Note the session ID and all authenticated users active during the window.
4. **Quarantine the workstation** if physical access to the screen is a concern — contact your system administrator to terminate active sessions remotely.
5. **Initiate a spillage report** per your unit's security incident procedures.
6. **Contact the Security Authority** in your chain of command.

### Investigating a Spillage in the Audit Trail

Filter the Audit Trail Viewer by:
- **User**: The affected user's certificate CN
- **Time Range**: The suspected spillage window ± 30 minutes
- **Event Category**: Track Access, Query

Look for events such as:
```
EVENT: classification_access_warning
Details: user_clearance=PROTECTED_B, data_level=PROTECTED_C, action=access_denied
```

If access was denied, no spillage occurred — the system blocked it. Document this finding.

If the access level on displayed data matches or exceeds the user's clearance (no `access_denied`), the system performed correctly — but investigate why the user encountered that data at all.

---

## Responding to a Classification Marking Error

If you believe a data item carries the **wrong classification marking** (e.g., marked UNCLASSIFIED when it should be PROTECTED B):

1. Document the affected track ID or data item
2. Do not attempt to correct the marking directly — classification marking changes require audit records
3. Contact the system administrator to initiate a formal reclassification operation, which is logged in the audit trail

---

## Monitoring NATO Export Classification

All data exported to NATO allies passes through a **cross-domain guard** that enforces:

- Maximum classification of NATO SECRET
- NOFORN blocking (no export of NOFORN-marked data)
- Source attribution removal (sensor specifics are stripped)
- Releasability caveats applied (REL TO NATO)

Monitor the NATO export audit log:

1. Open Audit Trail Viewer
2. Filter by **Event Category: NATO Exchange**
3. Look for:
   - `nato_export_blocked` — normal; verifies the guard is working
   - `nato_link16_exported` — each successful export; spot-check for unexpected volume
   - `nato_nffi_exported` — NFFI XML exports; verify frequency is expected

**If you see exports at unexpected volumes or for unexpected track types**, contact the NATO Liaison Officer and investigate the export policy configuration.

---

> **Next**: [Feedback Integrity Review →](03_feedback_review.md)

---

> **CLASSIFICATION: UNCLASSIFIED** — This document contains no classified information.
