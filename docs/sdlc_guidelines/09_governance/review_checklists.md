# Review Checklists

> **CLASSIFICATION: UNCLASSIFIED**
> **Document Type**: Governance Standard
> **Parent**: `09_governance/agent_governance.md`
> **Last Updated**: 2026-02-23

---

## 1. Purpose

This document provides structured checklists for code reviews, architecture reviews, and release reviews. These checklists ensure consistency, completeness, and compliance across all RTSA reviews.

## 2. Code Review Checklist

### 2.1 General (All Languages)

```markdown
## Code Review Checklist — General

### Compliance

- [ ] Classification header present on all new files
- [ ] No hardcoded secrets, passwords, API keys, or tokens
- [ ] No PII in code, comments, or log statements
- [ ] No references to real operational data
- [ ] Approved dependencies only (checked against supply_chain_security.md)

### Quality

- [ ] Code compiles without warnings
- [ ] Linter passes (golangci-lint / eslint)
- [ ] Unit tests included for new/changed code
- [ ] Test coverage maintained at ≥ 80%
- [ ] Tests use synthetic data only
- [ ] Naming conventions followed (general_coding.md)
- [ ] Comments are meaningful (explain "why", not "what")

### Security

- [ ] Input validation on all external data
- [ ] Error handling: no silenced errors, no panics (Go)
- [ ] No SQL string concatenation (parameterized queries)
- [ ] No use of `math/rand` for security purposes
- [ ] No `innerHTML` directive with untrusted data (SolidJS)
- [ ] Structured logging (no fmt.Println)
- [ ] Classification markings enforced in data flow

### Architecture

- [ ] Follows established patterns (no unnecessary novelty)
- [ ] Backward-compatible changes (especially Protobuf)
- [ ] Traceability: requirement IDs referenced in code/tests
- [ ] No circular dependencies introduced
```

### 2.2 Go-Specific

```markdown
## Code Review Checklist — Go

- [ ] `go vet` passes
- [ ] `golangci-lint run` passes
- [ ] `go test -race` passes (no data races)
- [ ] Errors wrapped with context (`fmt.Errorf("op: %w", err)`)
- [ ] Context propagated through all function calls
- [ ] Interfaces used for dependency injection
- [ ] `sync.Pool` used for hot-path allocations (if applicable)
- [ ] `GOMEMLIMIT` considered for edge deployment
- [ ] gRPC deadlines set on all outgoing calls
- [ ] Graceful shutdown implemented
- [ ] Conventional commit message format
```

### 2.3 Protobuf-Specific

```markdown
## Code Review Checklist — Protobuf

- [ ] `buf lint` passes
- [ ] `buf breaking` passes (backward compatibility)
- [ ] Package follows `rtsa.<context>.v1` convention
- [ ] UNSPECIFIED = 0 for all enums
- [ ] Field numbers follow strategy (1-15 high-freq)
- [ ] Service deadlines documented in comments
- [ ] Common types reused (Classification, Position, etc.)
- [ ] No removed or renumbered fields
```

### 2.4 SolidJS + WebGPU COP

```markdown
## Code Review Checklist — SolidJS + WebGPU

- [ ] TypeScript strict mode (no `any` types)
- [ ] Props not destructured (preserves SolidJS reactivity)
- [ ] Classification banner displayed correctly
- [ ] Accessibility: ARIA labels, keyboard navigation
- [ ] Dark mode / NVG mode compatibility
- [ ] Offline/edge degradation handled
- [ ] No `innerHTML` directive with untrusted data
- [ ] Component tests with `@solidjs/testing-library`
- [ ] Signal-based state management (no global stores)
- [ ] Worker communication uses typed `postMessage` protocol
- [ ] WGSL bindings match TypeScript `GPUBindGroupLayout`
- [ ] No per-frame GPU buffer allocations
```

## 3. Architecture Review Checklist

```markdown
## Architecture Review Checklist

### Design

- [ ] C4 diagrams included (Context, Container, Component as applicable)
- [ ] ADR created for significant decisions
- [ ] Follows event-driven architecture patterns
- [ ] Service boundaries align with bounded contexts
- [ ] Data flow diagrams accurate

### Security

- [ ] Threat model updated for new components
- [ ] Trust boundaries identified
- [ ] Classification levels assigned to data flows
- [ ] mTLS configured for all gRPC channels
- [ ] Cross-classification guards in place
- [ ] Anti-poisoning controls for feedback paths
- [ ] ITSG-33 controls referenced

### Operational

- [ ] Health check endpoints specified
- [ ] Monitoring metrics defined
- [ ] Alert conditions specified
- [ ] Edge deployment profile defined
- [ ] Graceful degradation behavior specified
- [ ] Data retention policies defined

### Compliance

- [ ] ITSG-33 controls mapped
- [ ] NIST 800-53 controls mapped (where applicable)
- [ ] NATO interoperability requirements addressed (if applicable)
- [ ] SBOM requirements satisfied
```

## 4. Release Review Checklist

```markdown
## Release Review Checklist

### Build

- [ ] All CI pipeline stages passed (SG-1 through SG-5)
- [ ] Container images built and signed (cosign)
- [ ] SBOM generated and attached
- [ ] Container images scanned (no Critical/High CVEs)

### Testing

- [ ] Unit test coverage ≥ 80%
- [ ] Integration tests passed
- [ ] E2E smoke tests passed
- [ ] Performance tests met SLA targets
- [ ] Security tests passed (SAST, dependency scan)

### Documentation

- [ ] Release notes written
- [ ] API changes documented
- [ ] Breaking changes documented with migration guide
- [ ] Updated architecture diagrams (if applicable)

### Deployment

- [ ] Helm chart updated
- [ ] Environment-specific values validated
- [ ] Rollback procedure tested
- [ ] Edge deployment bundle prepared (if edge release)
- [ ] Air-gap transfer media encrypted

### Approval

- [ ] At least one PR reviewer approved all included PRs
- [ ] Security reviewer approved (if security-critical changes)
- [ ] Release manager approved deployment to production
```

## 5. Threat Model Review Trigger Checklist

Use this checklist to determine if a change requires a threat model update:

```markdown
## Threat Model Update Triggers

Answer YES to any → threat model update required:

- [ ] New external data source or sensor type added?
- [ ] New gRPC service or endpoint added?
- [ ] New trust boundary crossed or modified?
- [ ] Change to authentication or authorization logic?
- [ ] Change to classification guard or cross-domain flow?
- [ ] Change to feedback processing or trust scoring?
- [ ] New or modified NATO/NFFI data exchange?
- [ ] Change to data retention or storage patterns?
- [ ] New deployment environment or target?
- [ ] Change to cryptographic algorithm or key management?
```

## 6. AI Agent Instructions

When performing reviews:

1. Use the appropriate checklist from this document based on the review type
2. Check EVERY item — do not skip any
3. Flag items as PASS, FAIL, or N/A with brief justification
4. For FAIL items, reference the specific SDLC guideline that is violated
5. Provide concrete fix suggestions for each FAIL
6. Determine if a threat model update is needed using Section 5
