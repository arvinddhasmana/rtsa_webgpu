<!-- CLASSIFICATION: UNCLASSIFIED -->

# RTSA User Guide — Master Index

> **CLASSIFICATION: UNCLASSIFIED**
> **Project**: Real-Time Situational Awareness & Risk Assessment (RTSA)
> **Audience**: End Users (All Stakeholders)
> **Version**: 1.0
> **Last Updated**: 2026-02-26
> **Classification Ceiling**: Protected C / Secret

---

## Welcome to RTSA

The **Real-Time Situational Awareness & Risk Assessment (RTSA)** system gives Canadian Armed Forces (CAF) operators a unified, real-time operational picture of the battlespace. It fuses data from six sensor categories — Radar, Electronic Warfare/SIGINT, ELINT/COMINT, ISR platforms, AIS/Blue Force Tracking, and Cyber threat feeds — applies AI-driven anomaly detection, and presents everything through an interactive Common Operating Picture (COP) dashboard built with **SolidJS + WebGPU**.

This guide helps you understand and use the system effectively, regardless of your role.

---

## How to Use This Guide

This guide is organized in two ways:

1. **Shared Content** — Foundational knowledge every RTSA user needs, regardless of role.
2. **Role-Specific Guides** — Tailored walkthroughs for your specific job function.

Start with the shared content if you are new to RTSA, then go to your role-specific guide.

---

## 📚 Shared Content (Start Here)

All users should read these sections first.

| Document                                                        | Description                                                               |
| --------------------------------------------------------------- | ------------------------------------------------------------------------- |
| [System Overview](shared/01_system_overview.md)                 | What RTSA does, how the pieces fit together, and key terminology          |
| [Getting Started](shared/02_getting_started.md)                 | Logging in, navigating the interface, and understanding your workspace    |
| [Classification Markings](shared/03_classification_markings.md) | Understanding security classification indicators in the UI                |
| [UI Navigation](shared/04_ui_navigation.md)                     | Common interface patterns, keyboard shortcuts, and accessibility features |

---

## 👤 Role-Specific Guides

Find your role and go to the guide that matches your job function.

### 🎯 Operations Commander / Watch Officer

**You are responsible for**: Tactical decision-making, situational awareness, managing alerts, and guiding force posture in real time.

| Document                                                                                | Description                                                                |
| --------------------------------------------------------------------------------------- | -------------------------------------------------------------------------- |
| [Role Overview](operations_commander/README.md)                                         | Your workspace, key responsibilities, and quick-start checklist            |
| [Situational Awareness — The COP Map](operations_commander/01_situational_awareness.md) | Viewing, filtering, and interpreting real-time entity tracks               |
| [Anomaly Alerts — Detection & Response](operations_commander/02_anomaly_alerts.md)      | Understanding AI-generated alerts and managing your alert queue            |
| [Operator Feedback](operations_commander/03_operator_feedback.md)                       | Confirming, rejecting, or reclassifying AI detections                      |
| [Tactical Edge Operations](operations_commander/04_tactical_edge.md)                    | Operating the system in disconnected or bandwidth-constrained environments |

---

### 🔎 Intelligence Analyst

**You are responsible for**: Historical analysis, incident investigation, pattern identification, and producing intelligence assessments.

| Document                                                                | Description                                                           |
| ----------------------------------------------------------------------- | --------------------------------------------------------------------- |
| [Role Overview](intelligence_analyst/README.md)                         | Your workspace, key responsibilities, and quick-start checklist       |
| [Historical Queries](intelligence_analyst/01_historical_queries.md)     | Querying the historical database for tracks, alerts, and events       |
| [Forensic Investigation](intelligence_analyst/02_forensic_analysis.md)  | Deep-diving into specific entities, incidents, and time periods       |
| [Anomaly Pattern Analysis](intelligence_analyst/03_pattern_analysis.md) | Identifying trends, recurring anomaly types, and statistical patterns |

---

### 🔒 Security Officer

**You are responsible for**: Classification compliance, audit trail review, managing classification markings, and overseeing the integrity of operator feedback.

| Document                                                                      | Description                                                       |
| ----------------------------------------------------------------------------- | ----------------------------------------------------------------- |
| [Role Overview](security_officer/README.md)                                   | Your workspace, key responsibilities, and quick-start checklist   |
| [Audit Trail Review](security_officer/01_audit_trail.md)                      | Accessing and interpreting immutable audit records                |
| [Classification Management](security_officer/02_classification_management.md) | Enforcing and verifying classification markings across the system |
| [Feedback Integrity Review](security_officer/03_feedback_review.md)           | Reviewing flagged operator feedback and anti-poisoning alerts     |

---

### 📡 Sensor Operator

**You are responsible for**: Monitoring sensor health, validating data quality, and responding to sensor degradation events.

| Document                                                               | Description                                                     |
| ---------------------------------------------------------------------- | --------------------------------------------------------------- |
| [Role Overview](sensor_operator/README.md)                             | Your workspace, key responsibilities, and quick-start checklist |
| [Sensor Health Monitoring](sensor_operator/01_sensor_health.md)        | Viewing sensor status, throughput, and availability indicators  |
| [Data Quality & Dead-Letter Queue](sensor_operator/02_data_quality.md) | Understanding rejected sensor data and data quality indicators  |

---

### 🌐 NATO Liaison Officer

**You are responsible for**: Managing data exchange with NATO allied systems, overseeing cross-domain data release, and tracking nomination for NATO sharing.

| Document                                                       | Description                                                     |
| -------------------------------------------------------------- | --------------------------------------------------------------- |
| [Role Overview](nato_liaison/README.md)                        | Your workspace, key responsibilities, and quick-start checklist |
| [NATO Data Exchange](nato_liaison/01_nato_data_exchange.md)    | Understanding how RTSA exchanges data with NATO partners        |
| [Manual Track Nomination](nato_liaison/02_track_nomination.md) | Manually nominating tracks for NATO data sharing                |

---

## 📋 Business Use Case Reference Map

The table below maps each business use case to the guide sections that cover it.

| Use Case    | Title                        | Guide Section                                                                           |
| ----------- | ---------------------------- | --------------------------------------------------------------------------------------- |
| UC001       | System Initialization        | [Getting Started](shared/02_getting_started.md)                                         |
| UC002–UC007 | Sensor Ingestion (all types) | [Sensor Health Monitoring](sensor_operator/01_sensor_health.md)                         |
| UC008       | Multi-Source Fusion          | [System Overview — Data Fusion](shared/01_system_overview.md)                           |
| UC009       | Anomaly Detection            | [Anomaly Alerts](operations_commander/02_anomaly_alerts.md)                             |
| UC010       | Operator Feedback            | [Operator Feedback](operations_commander/03_operator_feedback.md)                       |
| UC011       | Model Retraining             | [Feedback Integrity Review](security_officer/03_feedback_review.md)                     |
| UC012       | Situational Awareness UI     | [Situational Awareness — The COP Map](operations_commander/01_situational_awareness.md) |
| UC013       | Historical Query             | [Historical Queries](intelligence_analyst/01_historical_queries.md)                     |
| UC014       | NATO Outbound                | [NATO Data Exchange](nato_liaison/01_nato_data_exchange.md)                             |
| UC015       | NATO Inbound                 | [NATO Data Exchange](nato_liaison/01_nato_data_exchange.md)                             |

---

## 🆘 Quick Reference — Who to Contact

| Issue                                                  | Contact                                                                      |
| ------------------------------------------------------ | ---------------------------------------------------------------------------- |
| Cannot log in                                          | Your unit IT security officer                                                |
| Display appears stale / disconnected                   | Sensor Operator or system administrator                                      |
| Anomaly alert not behaving correctly                   | Submit feedback (Operations Commander guide) or escalate to Security Officer |
| Data above your clearance visible (potential spillage) | **Immediately** contact your Security Officer                                |
| NATO link not transmitting                             | NATO Liaison Officer + system administrator                                  |
| Suspected model manipulation                           | Security Officer (feedback poisoning review)                                 |

---

> **CLASSIFICATION: UNCLASSIFIED** — This document contains no classified information.
