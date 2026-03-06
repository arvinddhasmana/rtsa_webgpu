# Prompt Templates

> **CLASSIFICATION: UNCLASSIFIED**
> **Document Type**: Governance Standard
> **Parent**: `09_governance/agent_governance.md`
> **Last Updated**: 2026-02-23

---

## 1. Purpose

This document provides reusable prompt templates for common RTSA development tasks when working with AI coding agents. Using consistent prompts ensures the agent loads the correct policy files and produces compliant output.

## 2. Template Structure

Every prompt should follow this structure:

```
1. Context:     What you're working on (service, feature, use case)
2. Task:        What you want the agent to do
3. Constraints: Specific requirements or limitations
4. References:  Requirement IDs, existing code, architecture docs
5. Output:      Expected format and deliverables
```

## 3. Prompt Templates

### 3.1 New Go gRPC Service

```markdown
## Task

Create a new Go gRPC service for [SERVICE_NAME] that [PURPOSE].

## Context

- This service is part of the RTSA system (see docs/architecture/component_design.md)
- It implements use case [UC_ID] (see docs/business/usecases/[UC_ID].md)
- It consumes from Redpanda topic [INPUT_TOPIC] and produces to [OUTPUT_TOPIC]

## Requirements

- Follow docs/sdlc_guidelines/04_coding_standards/go_standards.md
- Follow docs/sdlc_guidelines/08_tech_specific/grpc_service_guidelines.md
- Follow docs/sdlc_guidelines/04_coding_standards/secure_coding.md
- Include full interceptor chain (Recovery → OTel → Metrics → Logging → Auth → Audit → Validation)
- Include mTLS configuration
- Include health check endpoints
- Include graceful shutdown
- Include Dockerfile (distroless base)

## Deliverables

1. cmd/[service-name]/main.go — Entry point
2. internal/service/service.go — Business logic
3. internal/service/service_test.go — Unit tests (table-driven, 80%+ coverage)
4. proto/[service].proto — Service definition
5. Dockerfile — Multi-stage build
```

### 3.2 New Protobuf Service Definition

```markdown
## Task

Define the Protobuf service and message types for [SERVICE_NAME].

## Context

- Part of bounded context: [CONTEXT_NAME]
- Implements RPCs: [LIST_RPCS]
- Ref: docs/sdlc_guidelines/04_coding_standards/protobuf_grpc_standards.md

## Requirements

- Proto3 syntax
- Package: rtsa.[context].v1
- Include common types (Classification, Position, Kinematics)
- Follow field numbering strategy (1-15 high frequency)
- Include service deadline comments
- Backward-compatible design

## Deliverables

1. proto/rtsa/[context]/v1/[service].proto
```

### 3.3 New ClickHouse Table

```markdown
## Task

Create a ClickHouse table for [DATA_TYPE] with appropriate partitioning and TTL.

## Context

- Stores [DESCRIPTION]
- Ingested via Redpanda Connect from topic [TOPIC]
- Queried by [QUERY_SERVICE] for [USE_CASES]
- Ref: docs/sdlc_guidelines/08_tech_specific/clickhouse_guidelines.md

## Requirements

- MergeTree engine with ORDER BY optimized for query patterns
- Daily partitioning by time column
- TTL: [DC_RETENTION] for data centre, [EDGE_RETENTION] for edge
- Enum8 for low-cardinality columns
- DateTime64(3, 'UTC') for timestamps
- Include materialized view for [AGGREGATION] if applicable
- Include Redpanda Connect pipeline YAML

## Deliverables

1. SQL schema (CREATE TABLE)
2. Redpanda Connect pipeline configuration
3. Example parameterized Go query
```

### 3.4 New SolidJS Component

```markdown
## Task

Create a SolidJS component for [COMPONENT_NAME] that [PURPOSE].

## Context

- Part of the RTSA WebGPU COP overlay UI
- Implements [UC_ID] requirement [REQ_ID]
- Ref: docs/sdlc_guidelines/04_coding_standards/solidjs_standards.md

## Requirements

- TypeScript, named export function component
- Include traceability comment (// Implements: [REQ_ID])
- Props as signal accessors where applicable (never destructure)
- Cold-path data via ConnectRPC/gRPC-Web; hot-path data via Worker postMessage signals
- Support offline/edge degradation
- Include classification badge display
- Accessible (keyboard navigation, ARIA labels)
- Dark mode compatible (NVG mode)

## Deliverables

1. Component file (.tsx) in src/components/
2. Test file (.test.tsx) — behavior tests with @solidjs/testing-library
```

### 3.5 New Wasm Data Transform

```markdown
## Task

Create a Wasm data transform for [TRANSFORM_NAME] that [PURPOSE].

## Context

- Runs on Redpanda broker
- Input topic: [INPUT_TOPIC]
- Output topics: [OUTPUT_TOPIC], [DLQ_TOPIC]
- Ref: docs/sdlc_guidelines/08_tech_specific/wasm_transforms.md

## Requirements

- Go compiled to WASI (GOOS=wasip1 GOARCH=wasm)
- Validate [VALIDATION_RULES]
- Route invalid messages to DLQ with reason header
- Include classification guard check
- < 1ms per message
- No network calls, no disk I/O, no external dependencies
- Stateless (no state between messages)

## Deliverables

1. Transform source (transform.go)
2. Unit tests (transform_test.go)
3. Build script
4. Deploy command (rpk transform deploy)
```

### 3.6 Bug Fix

```markdown
## Task

Fix [BUG_DESCRIPTION] in [FILE/SERVICE].

## Context

- Current behavior: [WHAT_HAPPENS]
- Expected behavior: [WHAT_SHOULD_HAPPEN]
- Ref: [ISSUE_ID], [RELATED_UC]

## Requirements

- Root cause analysis (explain why the bug exists)
- Minimal change to fix the issue
- Regression test that would have caught the bug
- No behavior changes outside the fix scope
- Follow docs/sdlc_guidelines/04_coding_standards/[relevant_standard].md

## Deliverables

1. Code fix with explanation
2. Regression test
3. Updated documentation (if behavior changes)
```

### 3.7 Security Review

```markdown
## Task

Review [FILE/SERVICE/PR] for security issues.

## Context

- Component: [COMPONENT_NAME]
- Classification: [DATA_CLASSIFICATION_LEVEL]
- Ref: docs/sdlc_guidelines/04_coding_standards/secure_coding.md
- Ref: docs/sdlc_guidelines/03_architecture_design/threat_modeling.md

## Check

1. Input validation on all external data
2. No hardcoded secrets
3. Parameterized queries (no SQL injection)
4. mTLS configured correctly
5. Classification markings present and correct
6. Error handling (no panics, no swallowed errors)
7. Logging (no PII, no classified data)
8. Anti-poisoning controls (if feedback path)
9. Rate limiting (if external-facing)
10. Dependency versions (no known CVEs)

## Deliverables

1. Finding list with severity (Critical/High/Medium/Low)
2. Recommended fixes for each finding
3. ITSG-33/NIST control references
```

## 4. AI Agent Instructions

When using these templates:

1. Fill in all `[PLACEHOLDER]` values before starting work
2. Load the referenced policy files (Ref: links) into context
3. Follow the deliverables list — produce all listed items
4. Apply the output validation checklist from `09_governance/agent_governance.md`
5. These templates are starting points — adapt as needed for the specific task
