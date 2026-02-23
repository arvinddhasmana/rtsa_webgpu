<!-- CLASSIFICATION: UNCLASSIFIED -->
# Business Requirements Document — RTSA

> **Project**: Real-Time Situational Awareness & Risk Assessment (RTSA)
> **Sponsor**: Canadian Department of National Defence (DND)
> **Classification**: UNCLASSIFIED
> **Version**: 1.0
> **Last Updated**: 2026-02-23

---

## 1. Executive Summary

The Real-Time Situational Awareness & Risk Assessment (RTSA) system provides Canadian Armed Forces (CAF) operators with a unified, real-time operational picture by fusing data from six sensor categories, applying AI-driven anomaly detection, and enabling human-in-the-loop feedback for continuous model improvement. The system operates across data centre and tactical edge environments, supports NATO interoperability, and enforces Government of Canada security classification requirements.

## 2. Business Objectives

| ID | Objective | Success Metric |
|---|---|---|
| BO-001 | Provide a unified operational picture across all sensor types | Single fused view integrating 6 sensor categories |
| BO-002 | Detect anomalous entities and behaviors in real time | Anomaly detection within 500ms of sensor event |
| BO-003 | Enable operator feedback to improve detection accuracy | Measurable accuracy improvement per feedback cycle |
| BO-004 | Support disconnected tactical edge operations | Full autonomous operation without data centre connectivity |
| BO-005 | Ensure NATO interoperability | Compliant STANAG 5516, NFFI, and MIP data exchange |
| BO-006 | Maintain GC security classification compliance | Zero classification spillage incidents |
| BO-007 | Provide forensic and historical analysis capability | Queryable history with configurable retention |

## 3. Stakeholders

| Stakeholder | Role | Concerns |
|---|---|---|
| CAF Operations Command | End user / decision authority | Timely, accurate situational awareness |
| Intelligence Analysts | End user | Historical analysis, pattern identification |
| Sensor Operators | End user | Sensor health, data quality monitoring |
| System Administrators | Operations | Deployment, monitoring, maintenance |
| Security Authority (SA) | Governance | Classification compliance, ITSG-33 |
| NATO Partners | Interoperability | STANAG/NFFI/MIP data exchange |
| DND IT Security | Governance | Accreditation, vulnerability management |
| Project Management | Oversight | Schedule, budget, risk |

## 4. Scope

### 4.1 In Scope

- Real-time sensor data ingestion from 6 sensor categories
- Multi-source data fusion and entity track management
- AI-driven anomaly and threat detection
- Human-in-the-loop feedback with anti-poisoning safeguards
- Situational awareness UI with map display
- Historical data storage and forensic query capability
- NATO interoperability (STANAG 5516, NFFI, MIP)
- Tactical edge deployment (disconnected operation)
- Immutable audit trail for all state-changing operations
- Security classification enforcement at data flow level

### 4.2 Out of Scope

- Sensor hardware provisioning or maintenance
- Weapon engagement systems or fire control
- Personnel management or human resources systems
- Financial or logistics systems
- Classification above SECRET (TOP SECRET requires separate system)
- Offensive cyber operations

## 5. Capability Requirements

### 5.1 Sensor Ingestion (CR-ING)

| ID | Requirement | Priority |
|---|---|---|
| CR-ING-001 | The system shall ingest radar sensor data (track reports, plot data) in real time | MUST |
| CR-ING-002 | The system shall ingest Electronic Warfare / SIGINT sensor data in real time | MUST |
| CR-ING-003 | The system shall ingest ELINT / COMINT sensor data in real time | MUST |
| CR-ING-004 | The system shall ingest ISR platform metadata (imagery metadata, not raw imagery) in real time | MUST |
| CR-ING-005 | The system shall ingest AIS (Automatic Identification System) and BFT (Blue Force Tracking) data in real time | MUST |
| CR-ING-006 | The system shall ingest cyber threat indicator data in real time | MUST |
| CR-ING-007 | The system shall validate all sensor data before processing (coordinates, timestamps, required fields) | MUST |
| CR-ING-008 | The system shall reject invalid sensor data and route it to a dead-letter queue for analysis | MUST |
| CR-ING-009 | The system shall support a sustained ingestion rate of 50,000 events/sec (data centre) | MUST |
| CR-ING-010 | The system shall support a sustained ingestion rate of 5,000 events/sec (tactical edge) | MUST |

### 5.2 Data Fusion (CR-FUS)

| ID | Requirement | Priority |
|---|---|---|
| CR-FUS-001 | The system shall correlate sensor reports from multiple sensors into unified entity tracks | MUST |
| CR-FUS-002 | The system shall assign entity type classification (Air, Surface, Subsurface, Land, Space, Cyber) | MUST |
| CR-FUS-003 | The system shall assign hostile status assessment (Unknown, Pending, Friendly, Neutral, Hostile, Suspect) | MUST |
| CR-FUS-004 | The system shall compute position, kinematics, and confidence for each fused track | MUST |
| CR-FUS-005 | The system shall maintain track identity across sensor handoffs | SHOULD |
| CR-FUS-006 | The system shall de-duplicate reports from overlapping sensor coverage | MUST |
| CR-FUS-007 | The system shall complete fusion within 100ms of receiving sensor data | MUST |

### 5.3 Anomaly Detection & Inference (CR-INF)

| ID | Requirement | Priority |
|---|---|---|
| CR-INF-001 | The system shall detect anomalous entity behavior using AI/ML models | MUST |
| CR-INF-002 | The system shall produce anomaly scores with confidence levels | MUST |
| CR-INF-003 | The system shall provide human-readable explanations for anomaly detections | SHOULD |
| CR-INF-004 | The system shall complete inference within 150ms of receiving a fused track | MUST |
| CR-INF-005 | The system shall support multiple anomaly detection models concurrently | SHOULD |
| CR-INF-006 | The system shall use pre-trained models at the tactical edge (no live training at edge) | MUST |
| CR-INF-007 | The system shall support model versioning and rollback | MUST |

### 5.4 Human-in-the-Loop Feedback (CR-FB)

| ID | Requirement | Priority |
|---|---|---|
| CR-FB-001 | Operators shall be able to confirm or reject anomaly detections | MUST |
| CR-FB-002 | Operators shall be able to reclassify entity hostile status | MUST |
| CR-FB-003 | The system shall assign trust scores to all operator feedback | MUST |
| CR-FB-004 | The system shall reject feedback that fails anti-poisoning validation | MUST |
| CR-FB-005 | The system shall route validated feedback to model retraining pipeline | MUST |
| CR-FB-006 | The system shall provide complete audit trail for all feedback operations | MUST |
| CR-FB-007 | The system shall queue feedback at the edge and sync to data centre when connected | MUST |

### 5.5 Situational Awareness UI (CR-UI)

| ID | Requirement | Priority |
|---|---|---|
| CR-UI-001 | The system shall display entity tracks on a geographic map in real time | MUST |
| CR-UI-002 | The system shall display anomaly alerts with severity indicators | MUST |
| CR-UI-003 | The system shall allow operators to filter by entity type, hostile status, and sensor type | MUST |
| CR-UI-004 | The system shall display sensor coverage and health status | SHOULD |
| CR-UI-005 | The system shall support classified display markings (banners, badges) | MUST |
| CR-UI-006 | The system shall operate in reduced mode when disconnected from data centre | MUST |
| CR-UI-007 | The system shall support NVG-compatible dark mode | SHOULD |
| CR-UI-008 | The system shall provide keyboard-only navigation for tactical environments | SHOULD |

### 5.6 Historical Analysis (CR-HIS)

| ID | Requirement | Priority |
|---|---|---|
| CR-HIS-001 | The system shall store all sensor events, fused tracks, and anomaly scores in an analytical database | MUST |
| CR-HIS-002 | The system shall support time-range queries over historical data | MUST |
| CR-HIS-003 | The system shall support aggregation queries (counts, trends, patterns) | MUST |
| CR-HIS-004 | The system shall retain sensor data for 90 days (data centre) | MUST |
| CR-HIS-005 | The system shall retain audit events for 2 years (data centre) | MUST |
| CR-HIS-006 | Simple queries shall complete within 500ms | MUST |
| CR-HIS-007 | Complex aggregation queries shall complete within 5 seconds | SHOULD |

### 5.7 NATO Interoperability (CR-NATO)

| ID | Requirement | Priority |
|---|---|---|
| CR-NATO-001 | The system shall exchange tactical data via STANAG 5516 (Link 16 J-Series) | MUST |
| CR-NATO-002 | The system shall exchange formatted entity data via NFFI (NATO Friendly Force Information) | MUST |
| CR-NATO-003 | The system shall support MIP Information Exchange Requirements (IERs) | SHOULD |
| CR-NATO-004 | The system shall map between NATO and GC classification levels | MUST |
| CR-NATO-005 | The system shall enforce cross-domain security at NATO exchange boundaries | MUST |

### 5.8 Security & Compliance (CR-SEC)

| ID | Requirement | Priority |
|---|---|---|
| CR-SEC-001 | The system shall enforce GC security classification on all data flows | MUST |
| CR-SEC-002 | The system shall use mTLS for all inter-service communication | MUST |
| CR-SEC-003 | The system shall maintain immutable audit logs for all state-changing operations | MUST |
| CR-SEC-004 | The system shall comply with ITSG-33 controls (CCCS) | MUST |
| CR-SEC-005 | The system shall comply with NIST 800-53 Rev 5 HIGH baseline | MUST |
| CR-SEC-006 | The system shall encrypt all data at rest using AES-256 | MUST |
| CR-SEC-007 | The system shall prevent classification spillage between security domains | MUST |
| CR-SEC-008 | The system shall support certificate-based authentication (no password auth) | MUST |

## 6. Non-Functional Requirements

### 6.1 Performance

| ID | Requirement | Target |
|---|---|---|
| NFR-PERF-001 | End-to-end latency (sensor → UI display) | < 500ms (P99) |
| NFR-PERF-002 | Sensor ingestion throughput (data centre) | 50,000 events/sec sustained |
| NFR-PERF-003 | Sensor ingestion throughput (edge) | 5,000 events/sec sustained |
| NFR-PERF-004 | Fusion engine latency | < 100ms (P99) |
| NFR-PERF-005 | Inference engine latency | < 150ms (P99) |
| NFR-PERF-006 | Concurrent WebSocket connections | 500 (DC), 50 (edge) |

### 6.2 Availability

| ID | Requirement | Target |
|---|---|---|
| NFR-AVAIL-001 | Data centre availability | 99.9% (8.77 hours downtime/year) |
| NFR-AVAIL-002 | Edge autonomous operation | Indefinite (until power loss) |
| NFR-AVAIL-003 | Recovery Time Objective (RTO) | < 15 minutes |
| NFR-AVAIL-004 | Recovery Point Objective (RPO) | < 1 minute |

### 6.3 Scalability

| ID | Requirement | Target |
|---|---|---|
| NFR-SCALE-001 | Horizontal scaling (data centre) | 2x capacity via pod scaling |
| NFR-SCALE-002 | Sensor type extensibility | Add new sensor type without code changes to core |
| NFR-SCALE-003 | Historical data growth | Predictable storage growth with TTL management |

## 7. Constraints

| ID | Constraint | Impact |
|---|---|---|
| CON-001 | Must operate on GC-approved infrastructure (no public cloud) | On-premise deployment only |
| CON-002 | Must use CSE-approved cryptographic standards | Limits algorithm choices |
| CON-003 | Edge hardware is resource-constrained (see edge deployment guide) | Limits per-service resource budgets |
| CON-004 | Team size is 3-5 developers | Limits parallel development capacity |
| CON-005 | Trunk-based development with 2-day branch lifetime | Requires small, incremental changes |
| CON-006 | No classified data in the code repository | All test data must be synthetic |

## 8. Assumptions

| ID | Assumption | Risk if Invalid |
|---|---|---|
| ASM-001 | Sensor feeds provide data in documented formats | Requires adapter development |
| ASM-002 | Network infrastructure between sensors and RTSA exists | Cannot ingest data |
| ASM-003 | PKI infrastructure is available for certificate management | Cannot implement mTLS |
| ASM-004 | Operators have appropriate security clearances | Classification access issues |
| ASM-005 | NATO partners use compliant STANAG implementations | Interoperability failures |

## 9. Acceptance Criteria Summary

The RTSA system is considered acceptable when:

1. All 6 sensor types can be ingested simultaneously at specified throughput
2. Entity tracks are fused and displayed within 500ms end-to-end
3. Anomaly detection produces scored alerts with explanations
4. Operator feedback is captured, trust-scored, and routed correctly
5. The system operates autonomously at the tactical edge
6. NATO data exchange functions correctly for STANAG 5516 and NFFI
7. No classification spillage occurs under any tested scenario
8. All audit events are immutably recorded
9. Performance targets are met under steady-state load
