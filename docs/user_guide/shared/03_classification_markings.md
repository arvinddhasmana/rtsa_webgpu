<!-- CLASSIFICATION: UNCLASSIFIED -->

# Classification Markings in RTSA

> **CLASSIFICATION: UNCLASSIFIED**
> **Audience**: All RTSA Users
> **Version**: 1.0
> **Last Updated**: 2026-02-26

---

## Why Classification Matters

RTSA handles data up to the **Protected C / Secret** classification ceiling. The system enforces classification at every layer — from sensor ingestion to what appears on your screen. You will **only ever see data you are cleared to access**. This is enforced by the system, not by honour system.

Despite the automatic enforcement, you must understand classification markings so you can:

- Interpret the security level of what you are viewing
- Recognize a potential spillage incident immediately
- Handle any exported outputs appropriately

---

## The Classification Banner

Every screen in RTSA shows a **classification banner** at the top and bottom of the display. This banner reflects the **highest classification level of any data currently visible on your screen**.

| Banner Colour | Classification Level | Description |
|---|---|---|
| 🟢 **Green** | UNCLASSIFIED | No classified information visible |
| 🔵 **Blue** | PROTECTED A | Contains low-sensitivity protected data |
| 🟡 **Yellow** | PROTECTED B | Contains medium-sensitivity protected data |
| 🟠 **Orange** | PROTECTED C | Contains high-sensitivity protected data |
| 🔴 **Red** | SECRET | Contains secret-level data |

> **If the banner colour is higher than your expected clearance level**, stop immediately and contact your Security Officer. Do not continue working. This may indicate a classification spillage.

---

## Classification Badges on Tracks

Every entity track on the map carries a **classification badge** that indicates its current assessment status. These badges use NATO standard colour coding:

| Badge Colour | Hostile Status | Meaning |
|---|---|---|
| 🔴 **Red** | HOSTILE | Confirmed or assessed hostile entity |
| 🟠 **Orange** | SUSPECT | Entity behaviour is suspicious; under assessment |
| 🟡 **Yellow** | UNKNOWN | Identity and intent not yet determined |
| 🟢 **Green** | FRIENDLY | Confirmed friendly (CAF or allied) entity |
| ⚪ **White** | NEUTRAL | Assessed as non-threatening |
| 🔵 **Blue** | PENDING | Awaiting assessment |

These badges are separate from the security classification of the data — they represent the **assessed identity** of the real-world entity.

---

## Anomaly Severity Indicators

Anomaly alerts also carry severity colour coding:

| Severity | Colour | Anomaly Score | Required Action |
|---|---|---|---|
| NORMAL | Grey | 0.0–0.3 | No action; logged automatically |
| WATCH | Yellow | 0.3–0.6 | Monitor; no immediate action required |
| ELEVATED | Orange | 0.6–0.8 | Operator attention required; review alert |
| CRITICAL | Red (pulsing) | 0.8–1.0 | Immediate operator acknowledgment required |

---

## Handling Potential Spillage

A **classification spillage** occurs when classified data is exposed to a system or person not authorized to access it at that classification level. If you suspect a spillage:

1. **Do not attempt to clean it up yourself**
2. **Do not copy, screenshot, or forward the information**
3. **Immediately contact your Security Officer**
4. **Note the time, what was visible, and who was present**
5. **Leave the screen as-is** until the Security Officer instructs otherwise

RTSA logs all user sessions, so the Security Officer can review exactly what data was presented and when.

---

## Releasability Caveats

Some tracks in RTSA carry releasability caveats that control whether data can be shared with NATO allies:

| Caveat | Meaning |
|---|---|
| **REL TO NATO** | Data may be shared with NATO alliance partners |
| **REL TO [country code]** | Data releasable to specific named nations |
| **NOFORN** | Data must not be shared with foreign nationals or systems |
| **CAN EYES ONLY** | Restricted to Canadian nationals only |

These caveats are enforced automatically by the NATO Adapter when data is exported. You will see caveats displayed in the entity detail panel.

---

## Classification in Offline / Edge Mode

When operating in **tactical edge mode** or **offline mode**, the classification enforcement rules remain fully in effect:

- The system only displays data up to your clearance level
- Classification banners continue to show the appropriate level
- All actions remain logged locally and synced to the data centre when connectivity is restored

---

> **Next**: [UI Navigation →](04_ui_navigation.md)

---

> **CLASSIFICATION: UNCLASSIFIED** — This document contains no classified information.
