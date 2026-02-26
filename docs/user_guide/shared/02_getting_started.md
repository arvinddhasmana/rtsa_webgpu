<!-- CLASSIFICATION: UNCLASSIFIED -->

# Getting Started with RTSA

> **CLASSIFICATION: UNCLASSIFIED**
> **Audience**: All RTSA Users (New to System)
> **Version**: 1.0
> **Last Updated**: 2026-02-26

---

## Before You Begin

Before you can access RTSA, the following must be in place:

| Prerequisite | Who Provides It |
|---|---|
| Security clearance at the appropriate level | Your chain of command / Security Authority |
| User certificate issued and installed in your browser | Your unit IT security officer |
| Workstation approved for the classification level you will access | Your Security Officer |
| RTSA URL and access instructions for your unit | Your system administrator |

> ⚠️ **IMPORTANT**: RTSA uses certificate-based authentication. There is **no username/password login**. Your identity is established by the digital certificate installed in your browser. Do not share your workstation or certificate with anyone.

---

## Step 1 — Access the RTSA Dashboard

Open your approved web browser and navigate to the RTSA URL provided by your system administrator. The URL will typically look like:

```
https://rtsa.mil.ca/cop
```

Your browser will automatically present your certificate to authenticate. If prompted, select your RTSA operator certificate.

### What You Will See on First Load

```
┌─────────────────────────────────────────────────────────────────┐
│  [CLASSIFICATION BANNER — shown in colour matching your level]  │
├──────────────┬──────────────────────────────────────────────────┤
│  LEFT PANEL  │                                                  │
│              │                                                  │
│  Alert       │         MAIN MAP AREA                           │
│  Queue       │         (Interactive COP Map)                   │
│              │                                                  │
│  [Alerts     │                                                  │
│   listed     │                                                  │
│   by         │                                                  │
│   severity]  │                                                  │
│              │                                                  │
├──────────────┴──────────────────────────────────────────────────┤
│  STATUS BAR:  Connection ● | Tracks: 1,423 | Alerts: 7 | UTC   │
└─────────────────────────────────────────────────────────────────┘
```

---

## Step 2 — Understand the Status Bar

The **status bar** at the bottom of the screen gives you a real-time health summary:

| Indicator | Meaning |
|---|---|
| **Connection ●** (green) | Live connection to the RTSA backend; data is current |
| **Connection ●** (red) | Disconnected; displaying last known data (shows timestamp of last update) |
| **EDGE MODE** | System running in tactical edge mode with limited connectivity |
| **Tracks: N** | Number of active entity tracks currently on your display |
| **Alerts: N** | Number of unacknowledged anomaly alerts |
| **UTC timestamp** | Current time in Coordinated Universal Time |

> **Note on Disconnected Mode**: If the connection indicator turns red, the display shows the last known state with a "STALE" marker on each track. Data timestamps show you how old the information is. New local alerts from the on-device AI model will continue to appear.

---

## Step 3 — Know the Layout

The RTSA interface has four main areas:

### 🗺️ Map Area (Centre / Right)
The primary situational awareness display. Entity tracks are shown as icons on a geographic map. This is where you spend most of your time.

### 🚨 Alert Panel (Left Side)
Lists active anomaly alerts, sorted by severity (CRITICAL at the top). Each card shows the track ID, anomaly type, severity, confidence level, and time of detection.

### 📋 Entity Detail Panel (Appears when you click a track)
Shows detailed information about a selected entity: position, speed, heading, which sensors detected it, anomaly history, and feedback history.

### 🔍 Historical Query Panel (accessible from toolbar)
Provides access to the historical database for past tracks and events. Available to users with appropriate clearance and role.

---

## Step 4 — Understand Connection State

RTSA streams data to your browser in real time. The connection uses encrypted channels (you do not need to take any action for this — it is automatic). Tracks and alerts update continuously as new sensor data arrives.

| Update Rate | Condition |
|---|---|
| 10 Hz (10 updates/second) | Normal data centre operation |
| 1 Hz (1 update/second) | Edge / bandwidth-constrained mode |
| Frozen + "STALE" indicator | Disconnected; showing last known state |

---

## Step 5 — Session Security

RTSA enforces strict session security:

- **Session timeout**: Your session automatically expires after **30 minutes of inactivity**. You will be logged out and must re-authenticate.
- **No data in browser storage**: RTSA never stores classified data in your browser cache or local storage. Closing the browser window immediately clears all session data.
- **All actions are logged**: Every interaction you have with the system — viewing tracks, submitting feedback, running queries — is recorded in an immutable audit trail. This is not punitive; it is required for security compliance.

---

## Quick-Start Checklist

Use this checklist the first time you use RTSA:

- [ ] Confirm your certificate is installed and valid
- [ ] Navigate to the RTSA URL and authenticate
- [ ] Verify the classification banner colour matches your expected access level
- [ ] Confirm the connection indicator is green (live data)
- [ ] Familiarise yourself with the alert panel on the left
- [ ] Click on one entity track on the map to see its detail panel
- [ ] Read your role-specific guide (linked from the [Master Index](../README.md))

---

## Common First-Time Questions

**Q: I see a lot of tracks on the map. How do I focus on what matters?**
Use the filter toolbar to narrow down by entity type (Air, Surface, etc.), hostile status (Hostile, Unknown, etc.), or sensor type. See [UI Navigation](04_ui_navigation.md) for details.

**Q: What does the pulsing red ring around a track mean?**
That track has a CRITICAL anomaly alert — the AI has detected high-confidence unusual behaviour. Open the alert panel or click the track for details.

**Q: The banner says "STALE — DISCONNECTED". What do I do?**
Your browser has lost its live connection to the RTSA backend. Check with your system administrator. The data on screen is the last known state; note the timestamp to understand how old it is.

**Q: I submitted feedback and got a message "Feedback submitted for review". Did it work?**
Yes — your feedback was received but requires human review before affecting the AI model. This happens when the trust-scoring engine determines your feedback needs secondary validation. It is a normal part of the anti-manipulation safeguard.

**Q: Can I export data from the screen?**
Data export is controlled by classification policy. Contact your Security Officer. No export occurs above your authorized clearance level.

---

> **Next**: [Classification Markings →](03_classification_markings.md)

---

> **CLASSIFICATION: UNCLASSIFIED** — This document contains no classified information.
