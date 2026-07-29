# Plan: RTSA Defence SDLC Framework & Project Documentation

## TL;DR

Build a two-layer documentation system for a Protected C/Secret Real-Time Situational Awareness system ingesting 6 sensor types (Radar, EW/SIGINT, ELINT/COMINT, ISR, AIS/BFT, Cyber) with NATO STANAG 5516 interoperability. **Layer 1** is the SDLC Guidelines Framework (`docs/sdlc_guidelines/`) — a hierarchical policy engine with 5 dependency tiers (Generic → Security → Phase → Technology → Pattern), designed as Copilot-loadable instruction files with Mermaid diagrams. **Layer 2** is the Project Documentation (`docs/business/`, `docs/architecture/`) — comprehensive requirements, 15 detailed use cases, and full architecture documentation following the C4 model. All documents are natural-language, AI-agent-consumable, with embedded governance rules. A `.github/copilot-instructions.md` file will serve as the root loader pointing Copilot to the relevant policy files.

---

## File Structure

```
c:\Src\rtsa_webgpu\
├── .github/
│   └── copilot-instructions.md              # Root Copilot policy loader
├── docs/
│   ├── sdlc_guidelines/
│   │   ├── README.md                         # Index + dependency graph + loading order
│   │   ├── 00_master_policy.md               # Root governance — all agents load this
│   │   ├── 01_security_compliance/
│   │   │   ├── security_classification.md    # Protected C/Secret data handling
│   │   │   ├── itsg33_controls.md            # ITSG-33 control mapping
│   │   │   ├── nist800_53_controls.md        # NIST 800-53 control mapping
│   │   │   ├── nato_stanag_compliance.md     # NATO STANAG 5516 / NFFI / MIP
│   │   │   └── supply_chain_security.md      # SBOM, dependency vetting
│   │   ├── 02_requirements/
│   │   │   ├── requirements_engineering.md   # How to capture/manage requirements
│   │   │   └── traceability.md               # Traceability matrix standards
│   │   ├── 03_architecture_design/
│   │   │   ├── architecture_guidelines.md    # ADR format, C4 model usage
│   │   │   ├── design_guidelines.md          # Design principles & patterns
│   │   │   └── threat_modeling.md            # STRIDE/DREAD methodology
│   │   ├── 04_coding_standards/
│   │   │   ├── general_coding.md             # Language-agnostic rules
│   │   │   ├── go_standards.md               # Go microservice conventions
│   │   │   ├── protobuf_grpc_standards.md    # Proto3, gRPC patterns
│   │   │   ├── react_standards.md            # React/TypeScript frontend
│   │   │   └── secure_coding.md              # OWASP, input validation, crypto
│   │   ├── 05_testing/
│   │   │   ├── testing_strategy.md           # Test pyramid, coverage targets
│   │   │   ├── security_testing.md           # SAST/DAST/pen-test/fuzz
│   │   │   └── performance_testing.md        # Latency targets, edge constraints
│   │   ├── 06_integration_cicd/
│   │   │   ├── ci_cd_pipeline.md             # Pipeline stages, gates
│   │   │   ├── branching_strategy.md         # Git flow for small team
│   │   │   └── artifact_management.md        # Container images, signing, SBOM
│   │   ├── 07_deployment_operations/
│   │   │   ├── deployment_guidelines.md      # Procedures, rollback, approval
│   │   │   ├── edge_tactical_deployment.md   # Disconnected ops, resource limits
│   │   │   └── monitoring_observability.md   # Logging, metrics, alerting
│   │   ├── 08_tech_specific/
│   │   │   ├── redpanda_guidelines.md        # Topics, partitioning, tiered storage
│   │   │   ├── clickhouse_guidelines.md      # Schema design, query patterns
│   │   │   ├── grpc_service_guidelines.md    # Service mesh, interceptors
│   │   │   └── wasm_transforms.md            # Wasm filter/validation pipelines
│   │   └── 09_governance/
│   │       ├── agent_governance.md           # AI agent output validation rules
│   │       ├── prompt_templates.md           # Structured prompt templates
│   │       └── review_checklists.md          # Phase-gate review checklists
│   ├── business/
│   │   ├── requirements.md                   # Full BRD
│   │   ├── feature_list.md                   # Feature registry with priorities
│   │   └── usecases/
│   │       ├── UC001_sensor_event_ingestion.md
│   │       ├── UC002_multi_sensor_fusion.md
│   │       ├── UC003_realtime_entity_detection.md
│   │       ├── UC004_ai_anomaly_threat_detection.md
│   │       ├── UC005_situational_awareness_display.md
│   │       ├── UC006_human_feedback_loop.md
│   │       ├── UC007_model_reinforcement_learning.md
│   │       ├── UC008_anti_poisoning_detection.md
│   │       ├── UC009_historical_analysis_forensics.md
│   │       ├── UC010_tiered_storage_management.md
│   │       ├── UC011_edge_tactical_operations.md
│   │       ├── UC012_audit_trail_immutability.md
│   │       ├── UC013_risk_assessment_scoring.md
│   │       ├── UC014_nato_data_interoperability.md
│   │       └── UC015_cross_domain_guard.md
│   └── architecture/
│       ├── high_level_architecture.md        # C4 L1-L2, system context
│       ├── component_design.md               # C4 L3, all components
│       ├── data_architecture.md              # Data models, flows, schemas
│       ├── security_architecture.md          # Security zones, crypto, ZTA
│       ├── deployment_architecture.md        # Edge/on-prem/hybrid topologies
│       ├── integration_architecture.md       # NATO interop, external feeds
│       └── dependency_graph.md               # Feature ↔ UseCase ↔ Component
```

---

## Steps

### Phase 1: SDLC Guidelines Framework (Policy Engine)

**Step 1. Create `.github/copilot-instructions.md`** — Root Copilot instruction file that: establishes this as a Protected C/Secret defence project; lists mandatory policy files to load per task type (coding → load `04_coding_standards/*`, design → load `03_architecture_design/*`); declares global constraints (no secrets in code, classification marking, approved libraries only).

**Step 2. Create `docs/sdlc_guidelines/README.md`** — Master index with a Mermaid dependency graph showing the 5-tier hierarchy. Contains a loading-order table mapping agent task type → required policy files. This is the "table of contents" for all governance.

**Step 3. Create `docs/sdlc_guidelines/00_master_policy.md`** — Root governance policy: project identity, classification level (Protected C/Secret), compliance mandates (ITSG-33, NIST 800-53, NATO STANAG), team structure (3-5 devs), core tech stack declaration, universal rules (all code must be unclassified in repo, no PII, mandatory review gates). Includes Mermaid SDLC lifecycle diagram with security gates at each phase.

**Step 4. Create `docs/sdlc_guidelines/01_security_compliance/` (5 files):**

- `security_classification.md` — Data classification tiers (Unclassified→Secret), handling rules per tier, marking requirements, spillage procedures, TEMPEST considerations, air-gap constraints, crypto requirements (CSE-approved algorithms), personnel clearance requirements.
- `itsg33_controls.md` — Map relevant ITSG-33 security control families (AC, AU, CM, IA, SC, SI, etc.) to project-specific implementation guidance. Mermaid diagram of control family relationships.
- `nist800_53_controls.md` — NIST 800-53 Rev 5 controls mapped to project with cross-reference to ITSG-33 equivalents. Focus on: AC (Access Control), AU (Audit), IA (Identification), SC (System Communications), SI (System Integrity).
- `nato_stanag_compliance.md` — STANAG 5516 (Link 16 data), NFFI (NATO Friendly Force Information), MIP (Multilateral Interoperability Programme) data exchange formats. Mermaid interoperability diagram.
- `supply_chain_security.md` — Approved dependency sources, SBOM generation requirements, vulnerability scanning gates, container image provenance, Wasm module signing.

**Step 5. Create `docs/sdlc_guidelines/02_requirements/` (2 files):**

- `requirements_engineering.md` — Requirement format (SHALL/SHOULD/MAY per RFC 2119), classification marking on requirements, security requirements derivation from ITSG-33, acceptance criteria format, User Story templates for AI agents.
- `traceability.md` — Bidirectional traceability: Requirement → Feature → UseCase → Component → Test. Mermaid traceability diagram template.

**Step 6. Create `docs/sdlc_guidelines/03_architecture_design/` (3 files):**

- `architecture_guidelines.md` — C4 model usage (Context, Container, Component, Code), ADR (Architecture Decision Record) template, architecture review checklist, mandatory diagrams per design doc, event-driven architecture patterns for this stack.
- `design_guidelines.md` — Microservice design principles (single responsibility, API-first, idempotency), data partitioning strategy, event schema evolution rules, circuit breaker patterns, graceful degradation for edge/tactical.
- `threat_modeling.md` — STRIDE methodology applied to each system boundary, DREAD risk scoring, threat model template with Mermaid data flow diagrams, mandatory threat modeling triggers (new service, new data flow, new external interface).

**Step 7. Create `docs/sdlc_guidelines/04_coding_standards/` (5 files):**

- `general_coding.md` — Universal rules: classification header comments, error handling patterns, logging standards (no classified data in logs), naming conventions, file structure conventions, documentation requirements.
- `go_standards.md` — Go module structure, error handling (no panics in production), context propagation, gRPC interceptor patterns, structured logging (slog), testing conventions (table-driven tests), linting (golangci-lint config), memory management for constrained environments.
- `protobuf_grpc_standards.md` — Proto3 style guide, package naming, field numbering strategy, backward compatibility rules, streaming patterns, deadline/timeout policies, error code usage, interceptor chain design.
- `react_standards.md` — Component structure, state management, WebSocket/gRPC-Web integration, real-time data rendering patterns, accessibility for operational UIs, offline-capable patterns for tactical edge.
- `secure_coding.md` — Input validation (all sensor data untrusted), output encoding, cryptographic standards (CSE-approved), secret management, secure gRPC channel configuration (mTLS), SQL injection prevention for ClickHouse queries.

**Step 8. Create `docs/sdlc_guidelines/05_testing/` (3 files):**

- `testing_strategy.md` — Test pyramid (unit 70%, integration 20%, E2E 10%), coverage targets (80%+ line coverage), test naming conventions, mock strategy for Redpanda/ClickHouse/gRPC, test data classification handling, CI gate thresholds.
- `security_testing.md` — SAST tools and gates, DAST for gRPC endpoints, fuzz testing for sensor input parsers, dependency vulnerability scanning, container image scanning, penetration testing schedule, red team exercise requirements.
- `performance_testing.md` — Latency targets per pipeline stage, throughput requirements (events/sec per sensor type), resource budgets for tactical edge hardware, load test scenarios, Redpanda partition performance validation, ClickHouse query performance baselines.

**Step 9. Create `docs/sdlc_guidelines/06_integration_cicd/` (3 files):**

- `ci_cd_pipeline.md` — Pipeline stages (lint → build → unit test → SAST → integration test → DAST → security gate → staging → approval → deploy), air-gap deployment packaging, security gate definitions, pipeline-as-code standards.
- `branching_strategy.md` — Trunk-based development with short-lived feature branches (optimized for 3-5 dev team), mandatory PR reviews, conventional commit messages, release tagging strategy.
- `artifact_management.md` — Container registry (air-gap mirror), SBOM generation (CycloneDX), artifact signing (cosign/sigstore), Go module proxy, Protobuf registry, Helm chart versioning.

**Step 10. Create `docs/sdlc_guidelines/07_deployment_operations/` (3 files):**

- `deployment_guidelines.md` — Deployment approval workflow, canary/blue-green strategies, rollback procedures, configuration management (no secrets in repos), environment parity requirements, air-gap deployment procedures.
- `edge_tactical_deployment.md` — Hardware constraints (CPU/RAM/disk budgets), Redpanda single-binary edge config, disconnected operation mode, data sync on reconnect, reduced-capability graceful degradation, field update procedures.
- `monitoring_observability.md` — Structured logging standards, metrics collection (Prometheus-compatible), distributed tracing for gRPC, alerting rules for sensor pipeline health, security event monitoring, audit log forwarding.

**Step 11. Create `docs/sdlc_guidelines/08_tech_specific/` (4 files):**

- `redpanda_guidelines.md` — Topic naming conventions, partitioning strategy per sensor type, retention policies (hot/cold tiering), consumer group design, schema registry usage, Redpanda Connect pipeline patterns, tactical edge configuration.
- `clickhouse_guidelines.md` — Table engine selection (MergeTree family), partition key design, materialized view patterns, TTL for data lifecycle, query optimization rules, cluster vs. single-node topology, ClickHouse Keeper configuration.
- `grpc_service_guidelines.md` — Service definition patterns, interceptor chain (auth → logging → tracing → rate-limit), health check standards, load balancing, connection pooling, deadline propagation, error handling conventions.
- `wasm_transforms.md` — Wasm module development standards, trust score validation logic, anti-poisoning filter design, module testing in isolation, deployment and versioning within Redpanda.

**Step 12. Create `docs/sdlc_guidelines/09_governance/` (3 files):**

- `agent_governance.md` — Rules engine for validating AI agent output: classification marking checks, banned patterns (hardcoded secrets, unsafe crypto, direct DB access bypassing ORM), mandatory security headers, output schema validation, mandatory test generation with code.
- `prompt_templates.md` — Structured prompt templates for: new microservice creation, new Protobuf definition, new React component, new ClickHouse table, security review, threat model update. Each template forces compliance with relevant guidelines.
- `review_checklists.md` — Phase-gate checklists: requirements review, design review, code review, security review, deployment readiness review. Each checklist references specific guideline sections.

---

### Phase 2: Project Documentation (Business)

**Step 13. Create `docs/business/requirements.md`** — Full Business Requirements Document: project vision, stakeholders (military operators, analysts, system admins, NATO partners), capability requirements derived from all 6 sensor types, non-functional requirements (latency <100ms for real-time, 99.9% availability on-prem, graceful degradation at edge), compliance requirements, data sovereignty requirements. Includes Mermaid stakeholder map and capability decomposition tree.

**Step 14. Create `docs/business/feature_list.md`** — Feature registry with ~15 features, each with: ID, name, description, priority (MoSCoW), complexity estimate, dependency list, mapped use cases, mapped compliance controls. Features include: Sensor Ingestion Pipeline, Multi-Sensor Fusion Engine, AI Inference Engine, Real-Time Dashboard, Feedback Processing, Anti-Poisoning Middleware, Analytics Engine, Tiered Storage, Edge Deployment, Audit System, NATO Interop Adapter, Risk Scoring Engine, Cross-Domain Guard, Model Training Pipeline, Alert Management. Includes Mermaid feature dependency graph.

**Step 15. Create 15 Use Case files in `docs/business/usecases/`** — Each file contains: UC ID, title, actors, preconditions, main flow (numbered steps), alternative flows, exception flows, postconditions, security considerations, data classification of inputs/outputs, non-functional requirements, related features, Mermaid sequence diagram, compliance control references. The 15 use cases are:

- `UC001_sensor_event_ingestion.md` — Ingest events from 6 sensor types via gRPC into Redpanda
- `UC002_multi_sensor_fusion.md` — Fuse data from multiple sensors to create correlated entity tracks
- `UC003_realtime_entity_detection.md` — Detect and classify entities (assets/threats) in real-time
- `UC004_ai_anomaly_threat_detection.md` — AI-driven anomaly detection on sensor streams
- `UC005_situational_awareness_display.md` — Real-time operational picture on React dashboard
- `UC006_human_feedback_loop.md` — Operator provides feedback on AI classifications
- `UC007_model_reinforcement_learning.md` — Use validated feedback to retrain models
- `UC008_anti_poisoning_detection.md` — Detect and filter malicious feedback before retraining
- `UC009_historical_analysis_forensics.md` — Query historical data in ClickHouse for forensic analysis
- `UC010_tiered_storage_management.md` — Manage hot (Redpanda) to cold (S3/ClickHouse) data lifecycle
- `UC011_edge_tactical_operations.md` — Deploy and operate in tactical edge environments
- `UC012_audit_trail_immutability.md` — Maintain immutable audit trail of all events and actions
- `UC013_risk_assessment_scoring.md` — Compute and display threat risk scores
- `UC014_nato_data_interoperability.md` — Exchange data with NATO systems via STANAG 5516/NFFI
- `UC015_cross_domain_guard.md` — Controlled data exchange between classification domains

---

### Phase 3: Project Documentation (Architecture & Design)

**Step 16. Create `docs/architecture/high_level_architecture.md`** — C4 Level 1 (System Context): RTSA system interacting with Sensor Networks, NATO Partner Systems, Operators, Analysts, ML Training Infrastructure. C4 Level 2 (Container): Go Ingestion Services, Redpanda Cluster, AI Inference Engine, ClickHouse Cluster, React Dashboard, Feedback Service, Redpanda Connect, Wasm Transform Layer, Model Training Service. All as Mermaid C4 diagrams. Includes architectural principles and key design decisions.

**Step 17. Create `docs/architecture/component_design.md`** — C4 Level 3 for each container: internal components, interfaces, data flows. Detailed Mermaid component diagrams for: Ingestion Service (per-sensor-type adapters, normalizer, partitioner), AI Inference Engine (model loader, inference pipeline, confidence scorer), Feedback Service (trust scorer, validation engine, routing), Dashboard (map renderer, timeline, alert panel, entity tracker). Includes gRPC service interface definitions.

**Step 18. Create `docs/architecture/data_architecture.md`** — Data models for all 6 sensor types, normalized entity model, feedback event schema, Redpanda topic design (naming, partitioning keys, retention), ClickHouse table schemas (MergeTree engines, partition keys, materialized views), data flow diagrams showing complete lifecycle from sensor → ingestion → Redpanda → AI + Archiver → ClickHouse. Includes Mermaid ER diagrams and data flow diagrams.

**Step 19. Create `docs/architecture/security_architecture.md`** — Security zones (RED/BLACK network separation), encryption at rest and in transit (CSE-approved), mTLS for all gRPC, RBAC model, authentication architecture, audit logging architecture, anti-poisoning trust model, key management, certificate lifecycle, TEMPEST considerations, security monitoring architecture. Includes Mermaid security zone diagram and trust boundary diagram.

**Step 20. Create `docs/architecture/deployment_architecture.md`** — Three deployment topologies with Mermaid diagrams: (1) Full on-premise data centre — all containers, full HA, multi-node Redpanda and ClickHouse clusters. (2) Tactical edge — single-node Redpanda, embedded ClickHouse, reduced AI model, minimal footprint. (3) Hybrid — edge nodes syncing to central on-prem when connected. Includes hardware sizing estimates, network topology, and Kubernetes/container orchestration design.

**Step 21. Create `docs/architecture/integration_architecture.md`** — External system interfaces: NATO STANAG 5516 adapter (Link 16 data ingest/egress), NFFI format translator, sensor-specific protocol adapters (radar data formats, SIGINT protocols, ISR video feed ingest), cross-domain guard architecture, API gateway design. Mermaid integration diagrams.

**Step 22. Create `docs/architecture/dependency_graph.md`** — Comprehensive dependency mapping: Feature → UseCase matrix, UseCase → Component matrix, Component → Technology matrix, Feature → Compliance Control matrix. All as Mermaid diagrams showing bidirectional traceability. This is the keystone document tying everything together.

---

## Execution Order & Dependencies

```
Step 1-3   → Foundation (can start immediately)
Step 4     → Depends on Step 3 (master policy defines compliance scope)
Steps 5-12 → Depend on Step 4 (security baseline established)
Steps 5-12 → Can be parallelized across guideline areas
Step 13-14 → Depend on Steps 2-3 (requirements follow guidelines)
Step 15    → Depends on Steps 13-14 (use cases derive from requirements/features)
Steps 16-21 → Depend on Step 15 (architecture serves use cases)
Step 22    → Depends on all above (captures full dependency graph)
```

## Diagrams Included (Total: ~45+ Mermaid diagrams)

| Document          | Diagram Type               | Content                               |
| ----------------- | -------------------------- | ------------------------------------- |
| README.md         | Mermaid flowchart          | 5-tier guideline dependency hierarchy |
| 00_master_policy  | Mermaid flowchart          | SDLC lifecycle with security gates    |
| itsg33_controls   | Mermaid mindmap            | Control family relationships          |
| nato_stanag       | Mermaid C4                 | Interoperability context              |
| threat_modeling   | Mermaid DFD                | Data flow + trust boundaries          |
| ci_cd_pipeline    | Mermaid flowchart          | Pipeline stages with gates            |
| Each UC file      | Mermaid sequence           | Actor interaction flows               |
| high_level_arch   | Mermaid C4 (L1+L2)         | System context + containers           |
| component_design  | Mermaid C4 (L3)            | Component internals (×6)              |
| data_architecture | Mermaid ER + flowchart     | Data models + data flows              |
| security_arch     | Mermaid flowchart          | Security zones + trust boundaries     |
| deployment_arch   | Mermaid deployment         | 3 topology diagrams                   |
| integration_arch  | Mermaid C4                 | External system interfaces            |
| dependency_graph  | Mermaid flowchart + matrix | Full traceability map                 |

## Verification

- **Structural**: Every use case traces to at least one feature; every feature maps to at least one component; every component has at least one ITSG-33 control
- **Copilot Integration**: Load `.github/copilot-instructions.md` in Copilot and verify it correctly references guideline paths
- **Completeness**: Cross-reference the dependency graph against all documents to ensure no orphan artifacts
- **AI Consumability**: Each document starts with a YAML-style metadata block (classification, doc-type, dependencies) so AI agents can programmatically parse and load
- **Compliance**: Verify all 6 sensor types appear in at least one use case; verify ITSG-33, NIST 800-53, and NATO STANAG are mapped to specific project controls

## Decisions

- **Copilot-first**: Guidelines formatted as natural language instruction files with YAML metadata headers, optimized for GitHub Copilot workspace context loading
- **Mermaid over PlantUML**: Better native rendering in GitHub, VS Code, and AI agent consumption
- **C4 model**: Industry standard for defence software architecture documentation, maps cleanly to ITSG-33 authorization boundaries
- **Trunk-based development**: Optimal for 3-5 developer team, reduces merge complexity while maintaining security gates
- **15 use cases**: Comprehensive coverage of all sensor types, AI/ML lifecycle, tactical ops, and NATO interop without duplication
- **Protected C/Secret baseline**: All guidelines default to Secret-level handling; documents can be individually downgraded where appropriate

## Key Parameters (from user answers)

| Parameter                  | Value                                               |
| -------------------------- | --------------------------------------------------- |
| Security Classification    | Protected C / Secret                                |
| Compliance Frameworks      | ITSG-33 (CCCS), NIST 800-53, NATO STANAG            |
| Deployment Environments    | Tactical Edge, On-premise Data Centre, Hybrid       |
| AI Agent Tooling           | GitHub Copilot (Workspace/Chat)                     |
| Sensor Types               | Radar, EW/SIGINT, ELINT/COMINT, ISR, AIS/BFT, Cyber |
| Interoperability Standards | NFFI / MIP / STANAG 5516                            |
| Team Size                  | Small (3-5 developers)                              |
| Documentation Depth        | Comprehensive — thorough before coding              |
