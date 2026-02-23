# General Coding Standards

> **CLASSIFICATION: UNCLASSIFIED**
> **Document Type**: Coding Standard
> **Parent**: `00_master_policy.md`
> **Dependencies**: `01_security_compliance/security_classification.md`
> **Last Updated**: 2026-02-23

---

## 1. Purpose

This document defines language-agnostic coding rules that apply to ALL source code in the RTSA project — Go, TypeScript/React, Protobuf, SQL, configuration files, and scripts.

## 2. Classification Headers

Every source file SHALL begin with a classification header as the first comment (after shebang lines):

| Language | Header |
|---|---|
| Go, Protobuf, TypeScript, JavaScript | `// CLASSIFICATION: UNCLASSIFIED` |
| YAML, TOML, Shell, Dockerfile, Makefile | `# CLASSIFICATION: UNCLASSIFIED` |
| SQL | `-- CLASSIFICATION: UNCLASSIFIED` |
| HTML, XML | `<!-- CLASSIFICATION: UNCLASSIFIED -->` |
| CSS | `/* CLASSIFICATION: UNCLASSIFIED */` |

## 3. Naming Conventions

### 3.1 Files and Directories

| Item | Convention | Example |
|---|---|---|
| Go files | `snake_case.go` | `radar_handler.go` |
| Go test files | `snake_case_test.go` | `radar_handler_test.go` |
| Protobuf files | `snake_case.proto` | `sensor_event.proto` |
| TypeScript files | `PascalCase.tsx` (components), `camelCase.ts` (utilities) | `SituationalMap.tsx`, `formatTrack.ts` |
| Test files (TS) | `*.test.ts`, `*.test.tsx` | `SituationalMap.test.tsx` |
| Config files | `kebab-case.yaml` | `redpanda-config.yaml` |
| SQL migrations | `NNN_description.sql` | `001_create_sensor_events.sql` |
| Documentation | `kebab-case.md` or `snake_case.md` | `high_level_architecture.md` |
| Directories | `snake_case` (Go), `kebab-case` (other) | `internal/sensor_ingestion/`, `src/components/` |

### 3.2 Code Symbols

| Symbol | Go Convention | TypeScript Convention | Protobuf Convention |
|---|---|---|---|
| Exported types | `PascalCase` | `PascalCase` | `PascalCase` |
| Unexported types | `camelCase` | `camelCase` | — |
| Functions/methods | `PascalCase` (exported), `camelCase` (unexported) | `camelCase` | `PascalCase` (service methods) |
| Constants | `PascalCase` (exported) | `UPPER_SNAKE_CASE` | `UPPER_SNAKE_CASE` |
| Variables | `camelCase` | `camelCase` | — |
| Enum values | — | `PascalCase` | `UPPER_SNAKE_CASE` with prefix |
| Package/module | `lowercase` single word | `camelCase` | `lowercase.dot.separated` |

## 4. Error Handling

### 4.1 Universal Rules

- **NEVER** swallow errors silently. Every error must be: handled, logged, or propagated.
- **NEVER** use panic/throw in production code for expected conditions.
- **ALWAYS** add context when wrapping errors — include the operation name and key parameters.
- **ALWAYS** use structured error types that map to gRPC status codes.

### 4.2 Error Context Pattern

```
[package].[function]([key-params]): [underlying error]
```
Example: `ingestion.ProcessRadarEvent(sensorID=SEN-042): redpanda publish: connection refused`

## 5. Logging Standards

### 5.1 Structured Logging Only

All services must use structured logging (JSON format). No `fmt.Println` or `console.log` in production code.

| Language | Library | Format |
|---|---|---|
| Go | `log/slog` (stdlib) | JSON |
| TypeScript | Structured logger (e.g., `pino`) | JSON |

### 5.2 Log Levels

| Level | Usage | Examples |
|---|---|---|
| **ERROR** | Unrecoverable failures; service degradation | Connection lost; message publish failed; auth failure |
| **WARN** | Recoverable issues; approaching limits | Rate limit approached; retry succeeded; deprecated API used |
| **INFO** | Key operational events (business-level) | Service started; sensor registered; model loaded |
| **DEBUG** | Diagnostic detail | Event processing steps; intermediate calculation values |

### 5.3 Log Field Standards

Every log entry must include:

| Field | Type | Required | Notes |
|---|---|---|---|
| `timestamp` | ISO 8601 UTC | YES | `2026-02-23T14:30:00.000Z` |
| `level` | string | YES | `ERROR`, `WARN`, `INFO`, `DEBUG` |
| `service` | string | YES | Service name: `ingestion`, `inference` |
| `message` | string | YES | Human-readable description |
| `correlation_id` | string | YES (if request) | Trace ID for request correlation |
| `component` | string | SHOULD | Internal component: `radar_handler` |
| `error` | string | IF ERROR | Error message (not stack trace at INFO) |

### 5.4 Prohibited Log Content

NEVER log:
- Classified data (sensor payloads, entity positions, intelligence products)
- PII (operator names, user IDs at INFO or above)
- Credentials (tokens, passwords, certificates, private keys)
- Raw Protobuf message bytes at INFO or above
- Full stack traces at INFO or above (ERROR level only)

## 6. Documentation Standards

### 6.1 Package/Module Documentation

Every Go package and TypeScript module must have a doc comment explaining:
- What the package/module does
- What bounded context it belongs to
- Traceability references (Feature, Use Case, Requirements)

### 6.2 Function/Method Documentation

Document any function that:
- Is exported/public
- Has non-obvious behavior
- Has preconditions or postconditions
- Has side effects (publishes events, writes to stores)

### 6.3 TODO/FIXME Convention

```
// TODO(author): [description] — [ticket/issue ID]
// FIXME(author): [description] — [ticket/issue ID]
// SECURITY: [description] — [ITSG-33/NIST control ID]
```

## 7. Configuration Management

### 7.1 Environment Variables

All runtime configuration from environment variables. Naming convention:

```
RTSA_[SERVICE]_[CONFIG_NAME]

Examples:
RTSA_INGESTION_GRPC_PORT=50051
RTSA_INGESTION_REDPANDA_BROKERS=broker1:9092,broker2:9092
RTSA_INFERENCE_MODEL_PATH=/models/anomaly_v1.onnx
RTSA_DEPLOYMENT_MODE=full|edge|hybrid
```

### 7.2 No Secrets in Code

- NEVER hardcode passwords, API keys, tokens, connection strings, or certificates
- Use Kubernetes Secrets, HashiCorp Vault, or environment variables
- Configuration templates (with placeholder values) are acceptable in the repo
- Actual secret values are NEVER committed, even in test configurations

## 8. Git Commit Standards

### 8.1 Conventional Commits

```
type(scope): description

Types:
  feat     — New feature
  fix      — Bug fix
  refactor — Code restructuring (no behavior change)
  test     — Adding/updating tests
  docs     — Documentation only
  chore    — Build, CI, tooling changes
  security — Security-related changes
  perf     — Performance improvements

Scope: service name, component, or area
  ingestion, inference, feedback, dashboard,
  proto, redpanda, clickhouse, ci, deploy

Examples:
  feat(ingestion): add radar event validation
  fix(inference): handle nil confidence score
  security(auth): rotate mTLS certificates
  test(feedback): add trust score validation tests
```

### 8.2 Commit Rules

- Each commit should be atomic — one logical change per commit
- Commits must compile and pass unit tests independently
- No commits with only commented-out code
- No merge commits on feature branches (rebase before PR)

## 9. AI Agent Instructions

When generating any source code:

1. ALWAYS start the file with the classification header (Section 2)
2. Follow naming conventions precisely (Section 3)
3. Never swallow errors; always wrap with context (Section 4)
4. Use structured logging only; never log classified data (Section 5)
5. Include package-level and function-level documentation (Section 6)
6. Configuration from environment variables only; no hardcoded secrets (Section 7)
7. Use conventional commit format for commit messages (Section 8)
