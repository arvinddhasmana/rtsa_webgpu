<!-- CLASSIFICATION: UNCLASSIFIED -->

# RTSA — Real-Time Situational Awareness & Risk Assessment

> **CLASSIFICATION: UNCLASSIFIED**
> **Project**: Real-Time Situational Awareness & Risk Assessment (RTSA)
> **Domain**: Canadian Defence — Situational Awareness & AI-driven Anomaly Detection
> **Classification Ceiling**: Protected C / Secret
> **Compliance**: ITSG-33 (CCCS), NIST 800-53 Rev 5, NATO STANAG 5516
> **Status**: Active Development

---

## Overview

RTSA provides Canadian Armed Forces (CAF) operators with a unified, real-time operational picture by fusing data from six sensor categories, applying AI-driven anomaly detection, and enabling human-in-the-loop feedback for continuous model improvement. The system supports full data centre deployments and hardware-constrained tactical edge environments, and maintains NATO interoperability via STANAG 5516 / NFFI / MIP.

```
Sensor Networks          Event Streaming           AI & Fusion             Operators
─────────────────        ────────────────          ───────────────         ─────────
Radar                    ┌──────────────┐          Fusion Engine           React COP
EW / SIGINT      ──────► │   Redpanda   │ ──────►  Anomaly Detection  ──►  Dashboard
ELINT / COMINT           │   Cluster    │          Model Training
ISR                      │  (Wasm Xfm)  │
AIS / BFT                └──────────────┘
Cyber Feeds                     │
                         ClickHouse OLAP
                         Historical Storage
```

---

## Key Capabilities

| Capability                 | Description                                                                            |
| -------------------------- | -------------------------------------------------------------------------------------- |
| **Multi-Sensor Ingestion** | Real-time ingestion from Radar, EW/SIGINT, ELINT/COMINT, ISR, AIS/BFT, and Cyber feeds |
| **Data Fusion**            | Multi-source track correlation, entity state estimation, confidence scoring            |
| **AI Anomaly Detection**   | Behavioral, spatial, and temporal anomaly detection with human-readable explanations   |
| **Human-in-the-Loop**      | Operator feedback with anti-poisoning trust scoring for safe model retraining          |
| **NATO Interoperability**  | Bidirectional STANAG 5516 / NFFI / MIP data exchange                                   |
| **Historical Analysis**    | ClickHouse OLAP with classification-aware query filtering                              |
| **Immutable Audit Trail**  | All state-changing operations routed through Redpanda audit topic                      |
| **Tactical Edge Support**  | Full autonomous operation in disconnected, resource-constrained environments           |

---

## Technology Stack

| Layer             | Technology                                          | Purpose                                             |
| ----------------- | --------------------------------------------------- | --------------------------------------------------- |
| Event Streaming   | [Redpanda](https://redpanda.com/)                   | Real-time event log, audit trail, inter-service bus |
| Microservices     | Go 1.22+ with gRPC / Protobuf                       | High-performance, type-safe service communication   |
| OLAP Storage      | ClickHouse                                          | Historical storage, forensics, analytical queries   |
| Frontend          | React 18 + TypeScript + gRPC-Web                    | Real-time common operating picture (COP)            |
| Data Pipeline     | Redpanda Connect                                    | Batch ETL: stream → ClickHouse / S3                 |
| Anti-Poisoning    | Wasm Data Transforms                                | In-broker feedback trust validation                 |
| Interoperability  | STANAG 5516 / NFFI / MIP adapters                   | NATO data exchange                                  |
| Observability     | OpenTelemetry + Prometheus + Grafana + Loki + Tempo | Structured telemetry across all services            |
| Container Runtime | Docker / Kubernetes / K3s (edge)                    | Dev via Docker Compose, staging/prod via Helm       |

---

## Repository Structure

```
rtsa/
├── .github/
│   ├── workflows/            # GitHub Actions — CI/CD pipeline
│   ├── prompts/              # AI agent profiles (greatest-ever-developer, meanest-ever-reviewer)
│   └── copilot-instructions.md
├── docs/
│   ├── architecture/         # High-level, component, data, security, deployment architecture
│   ├── business/             # Requirements, feature list, use cases (UC001–UC015)
│   └── sdlc_guidelines/      # Coding standards, testing strategy, deployment, governance
├── deploy/
│   ├── docker-compose.yml    # Local development stack
│   ├── docker-compose.test.yml
│   └── charts/               # Helm charts — rtsa-platform umbrella + per-service subcharts
├── proto/
│   └── rtsa/                 # Protobuf service contracts (versioned)
├── services/
│   ├── radar-ingestion/      # Go microservice — Radar sensor ingestion
│   ├── ew-ingestion/         # Go microservice — EW/SIGINT ingestion
│   ├── elint-ingestion/      # Go microservice — ELINT/COMINT ingestion
│   ├── isr-ingestion/        # Go microservice — ISR metadata ingestion
│   ├── ais-ingestion/        # Go microservice — AIS/BFT ingestion
│   ├── cyber-ingestion/      # Go microservice — Cyber threat feed ingestion
│   ├── nato-adapter/         # Go microservice — STANAG 5516/NFFI/MIP adapter
│   ├── fusion-engine/        # Go microservice — Multi-source track fusion
│   ├── anomaly-detection/    # Go microservice — AI/ML inference
│   ├── feedback-service/     # Go microservice — Operator feedback & trust scoring
│   ├── model-training/       # Go + Python — Reinforcement learning pipeline
│   ├── track-service/        # Go microservice — Track state management & streaming
│   ├── alert-service/        # Go microservice — Alert lifecycle management
│   ├── query-service/        # Go microservice — Historical ClickHouse queries
│   └── audit-service/        # Go microservice — Immutable audit trail
├── ui/                       # React 18 + TypeScript COP dashboard
├── scripts/
│   ├── setup/                # Developer environment setup automation
│   └── dev/                  # Development utility scripts
├── GETTING_STARTED.md
└── README.md
```

---

## Architecture

See [docs/architecture/high_level_architecture.md](docs/architecture/high_level_architecture.md) for the full C4 Context and Container diagrams.

### Security Zones

All traffic between zones uses mTLS with CSE-approved cipher suites. Components operate under a zero-trust model.

```
Zone 0 — External (Untrusted): Sensor networks, NATO allied systems, cyber feeds
Zone 1 — DMZ: Cross-Domain Guards, Link 16 terminal, NFFI/MIP gateway
Zone 2 — Ingestion (Restricted): Ingestion services, Wasm Data Transforms
Zone 3 — Processing (Confidential): Redpanda, Fusion, Anomaly Detection, Feedback
Zone 4 — Storage (Confidential): ClickHouse, Redpanda Connect, Model Registry
Zone 5 — Presentation (Controlled): API Gateway (Envoy), Track/Alert/Query services
Zone 6 — Operator (User-Facing): COP Web Application, Operator workstations
Zone 7 — Management (Administrative): Audit service, Observability stack
```

---

## Getting Started

| Platform / Workflow     | Guide                                                                 |
| ----------------------- | --------------------------------------------------------------------- |
| **Linux / macOS**       | **[GETTING_STARTED.md](GETTING_STARTED.md)**                          |
| **Windows (WSL2)**      | **[GETTING_STARTED_WINDOWS.md](GETTING_STARTED_WINDOWS.md)**          |
| **Demo (Setup + Run)**  | **[docs/demo/demo_setup_run_showcase.md](docs/demo/demo_setup_run_showcase.md)** |

Quick start (Linux / macOS):

```bash
# Clone the repository
git clone https://github.com/<org>/rtsa.git
cd rtsa

# Run automated developer setup (installs tools, configures environment)
./scripts/setup/setup-dev.sh

# Start the full local development stack
docker compose -f deploy/docker-compose.yml up -d

# Verify all services are healthy
./scripts/dev/health-check.sh
```

Quick start (Windows) — open **PowerShell as Administrator**:

```powershell
# 1. Enable WSL2 and install Ubuntu 24.04 (restart when prompted)
wsl --install -d Ubuntu-24.04

# 2. Install Docker Desktop manually: https://www.docker.com/products/docker-desktop/

# 3. Then open Ubuntu terminal and run:
#    ./scripts/setup/setup-dev.sh
#    See GETTING_STARTED_WINDOWS.md for the full step-by-step guide.
```

---

## Development Workflow

This project uses **trunk-based development** with short-lived feature branches.

```
main (trunk)
 ├── feature/RTSA-001-radar-ingestion    (max 2 days)
 ├── feature/RTSA-002-fusion-engine      (max 2 days)
 └── bugfix/RTSA-042-mtls-cert-rotation  (max 1 day)
```

### Branch Rules

- No direct commits to `main`
- All changes via Pull Request with at least one reviewer
- CI pipeline must pass before merge (all 5 security gates)
- Squash merge strategy enforced

### CI Pipeline Security Gates

| Gate             | Checks                                                          |
| ---------------- | --------------------------------------------------------------- |
| SG-1 Pre-Build   | Secret scan (gitleaks), classification headers, code formatting |
| SG-2 Build       | Go compile, proto generation, TypeScript compile, SBOM          |
| SG-3 Test        | Unit tests ≥ 80% coverage, contract tests                       |
| SG-4 Security    | SAST (semgrep/gosec), dependency scan, container scan           |
| SG-5 Integration | Integration tests, E2E smoke tests                              |

---

## Security & Compliance

> All code artifacts in this repository are **UNCLASSIFIED**. No classified data, credentials, PII, or operationally sensitive information is stored here.

| Requirement           | Standard                 | Reference                                                                                                                                      |
| --------------------- | ------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| Security Controls     | ITSG-33 (CCCS)           | [docs/sdlc_guidelines/01_security_compliance/itsg33_controls.md](docs/sdlc_guidelines/01_security_compliance/itsg33_controls.md)               |
| Security Controls     | NIST 800-53 Rev 5        | [docs/sdlc_guidelines/01_security_compliance/nist800_53_controls.md](docs/sdlc_guidelines/01_security_compliance/nist800_53_controls.md)       |
| NATO Interoperability | STANAG 5516 / NFFI / MIP | [docs/sdlc_guidelines/01_security_compliance/nato_stanag_compliance.md](docs/sdlc_guidelines/01_security_compliance/nato_stanag_compliance.md) |
| Supply Chain          | Approved libraries only  | [docs/sdlc_guidelines/01_security_compliance/supply_chain_security.md](docs/sdlc_guidelines/01_security_compliance/supply_chain_security.md)   |

### Mandatory Rules (All Contributors)

1. **No secrets in code.** Use environment variables or the approved secrets manager.
2. **Classification header required** in every file (first comment line).
3. **No PII in logs.** Structured JSON logging only.
4. **mTLS on all gRPC channels.** No plaintext inter-service communication.
5. **Error propagation required.** No silent swallowing of errors.
6. **80%+ unit test coverage** on all code changes.
7. **No panic() in production Go code** outside `main.go` initialization.

---

## SDLC Guidelines

| Topic                   | Document                                                                                                                                         |
| ----------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| Master Policy           | [docs/sdlc_guidelines/00_master_policy.md](docs/sdlc_guidelines/00_master_policy.md)                                                             |
| Go Standards            | [docs/sdlc_guidelines/04_coding_standards/go_standards.md](docs/sdlc_guidelines/04_coding_standards/go_standards.md)                             |
| Protobuf/gRPC Standards | [docs/sdlc_guidelines/04_coding_standards/protobuf_grpc_standards.md](docs/sdlc_guidelines/04_coding_standards/protobuf_grpc_standards.md)       |
| React Standards         | [docs/sdlc_guidelines/04_coding_standards/react_standards.md](docs/sdlc_guidelines/04_coding_standards/react_standards.md)                       |
| Secure Coding           | [docs/sdlc_guidelines/04_coding_standards/secure_coding.md](docs/sdlc_guidelines/04_coding_standards/secure_coding.md)                           |
| Testing Strategy        | [docs/sdlc_guidelines/05_testing/testing_strategy.md](docs/sdlc_guidelines/05_testing/testing_strategy.md)                                       |
| Redpanda Guidelines     | [docs/sdlc_guidelines/08_tech_specific/redpanda_guidelines.md](docs/sdlc_guidelines/08_tech_specific/redpanda_guidelines.md)                     |
| ClickHouse Guidelines   | [docs/sdlc_guidelines/08_tech_specific/clickhouse_guidelines.md](docs/sdlc_guidelines/08_tech_specific/clickhouse_guidelines.md)                 |
| Deployment Guidelines   | [docs/sdlc_guidelines/07_deployment_operations/deployment_guidelines.md](docs/sdlc_guidelines/07_deployment_operations/deployment_guidelines.md) |

---

## AI Agent Profiles

This repository includes purpose-built GitHub Copilot agent profiles in `.github/prompts/`:

| Agent                       | Invocation                               | Role                                                |
| --------------------------- | ---------------------------------------- | --------------------------------------------------- |
| **Greatest Ever Developer** | `@greatest-ever-developer <description>` | Full lifecycle: design → implement → test → PR      |
| **Meanest Ever Reviewer**   | `@meanest-ever-reviewer Review PR #<n>`  | Exhaustive security, compliance, and quality review |

---

## Use Cases

| ID    | Title                    | Document                                                                                                             |
| ----- | ------------------------ | -------------------------------------------------------------------------------------------------------------------- |
| UC001 | System Initialization    | [docs/business/usecases/UC001_system_initialization.md](docs/business/usecases/UC001_system_initialization.md)       |
| UC002 | Radar Ingestion          | [docs/business/usecases/UC002_radar_ingestion.md](docs/business/usecases/UC002_radar_ingestion.md)                   |
| UC003 | EW/SIGINT Ingestion      | [docs/business/usecases/UC003_ew_sigint_ingestion.md](docs/business/usecases/UC003_ew_sigint_ingestion.md)           |
| UC004 | ELINT/COMINT Ingestion   | [docs/business/usecases/UC004_elint_comint_ingestion.md](docs/business/usecases/UC004_elint_comint_ingestion.md)     |
| UC005 | ISR Metadata Ingestion   | [docs/business/usecases/UC005_isr_metadata_ingestion.md](docs/business/usecases/UC005_isr_metadata_ingestion.md)     |
| UC006 | AIS/BFT Ingestion        | [docs/business/usecases/UC006_ais_bft_ingestion.md](docs/business/usecases/UC006_ais_bft_ingestion.md)               |
| UC007 | Cyber Threat Ingestion   | [docs/business/usecases/UC007_cyber_threat_ingestion.md](docs/business/usecases/UC007_cyber_threat_ingestion.md)     |
| UC008 | Multi-Source Fusion      | [docs/business/usecases/UC008_multi_source_fusion.md](docs/business/usecases/UC008_multi_source_fusion.md)           |
| UC009 | Anomaly Detection        | [docs/business/usecases/UC009_anomaly_detection.md](docs/business/usecases/UC009_anomaly_detection.md)               |
| UC010 | Operator Feedback        | [docs/business/usecases/UC010_operator_feedback.md](docs/business/usecases/UC010_operator_feedback.md)               |
| UC011 | Model Retraining         | [docs/business/usecases/UC011_model_retraining.md](docs/business/usecases/UC011_model_retraining.md)                 |
| UC012 | Situational Awareness UI | [docs/business/usecases/UC012_situational_awareness_ui.md](docs/business/usecases/UC012_situational_awareness_ui.md) |
| UC013 | Historical Query         | [docs/business/usecases/UC013_historical_query.md](docs/business/usecases/UC013_historical_query.md)                 |
| UC014 | NATO Outbound            | [docs/business/usecases/UC014_nato_outbound.md](docs/business/usecases/UC014_nato_outbound.md)                       |
| UC015 | NATO Inbound             | [docs/business/usecases/UC015_nato_inbound.md](docs/business/usecases/UC015_nato_inbound.md)                         |

---

## Contributing

1. Read [docs/sdlc_guidelines/00_master_policy.md](docs/sdlc_guidelines/00_master_policy.md) before contributing.
2. Follow the [branching strategy](docs/sdlc_guidelines/06_integration_cicd/branching_strategy.md).
3. Ensure your branch passes all 5 CI security gates.
4. Open a Pull Request and request review.
5. No merge without at least one approver and a green CI run.

---

## License

This project is proprietary to the Canadian Department of National Defence (DND). Unauthorized access, distribution, or use is prohibited.

---

> **CLASSIFICATION: UNCLASSIFIED** — This document contains no classified information.
