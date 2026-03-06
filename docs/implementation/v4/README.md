<!-- CLASSIFICATION: UNCLASSIFIED -->
# v4 Implementation Plan — WebGPU COP

> **Document**: RTSA v4 Implementation Plan Overview
> **Version**: 1.0
> **Classification**: UNCLASSIFIED
> **Last Updated**: 2026-03-05
> **Authoritative Source**: `docs/architecture/v1/RTSA_WebGPU_Architecture_v1.md`

---

## 1. Purpose

This document organises the implementation of the WebGPU Common Operating Picture (COP) into five time-boxed phases. Each phase is self-contained, delivers a testable milestone, and has an explicit "done" gate.

> **Scope**: This plan covers the **browser frontend and supporting backend additions** only. Existing Go microservices, Redpanda topics, ClickHouse schemas, and the gRPC cold path are unchanged unless explicitly noted.

---

## 2. Phase Overview

```mermaid
gantt
  title v4 WebGPU COP Implementation
  dateFormat YYYY-MM-DD
  axisFormat %b %d

  section Phase 0 — Foundation
  Project scaffolding         :p0a, 2026-03-10, 5d
  Capability gate + build     :p0b, after p0a, 3d
  SharedArrayBuffer setup     :p0c, after p0a, 3d
  Wasm decoder (Rust)         :p0d, after p0c, 5d
  Phase 0 gate                :milestone, after p0d, 0d

  section Phase 1 — Core Rendering
  WebGPU device + pipelines   :p1a, after p0d, 5d
  Instanced track icons       :p1b, after p1a, 5d
  Compute interpolation       :p1c, after p1a, 4d
  Trail lines + halos         :p1d, after p1b, 4d
  SDF labels                  :p1e, after p1d, 3d
  Pick buffer                 :p1f, after p1e, 3d
  Phase 1 gate                :milestone, after p1f, 0d

  section Phase 2 — Backend Integration
  FlatBuffer serializer (Go)  :p2a, after p0d, 5d
  WebTransport server (Go)    :p2b, after p2a, 5d
  Data Worker integration     :p2c, after p2b, 3d
  Priority shedding           :p2d, after p2c, 3d
  Phase 2 gate                :milestone, after p2d, 0d

  section Phase 3 — UI & Interaction
  SolidJS shell + toolbar     :p3a, after p1f, 4d
  Track detail panel          :p3b, after p3a, 3d
  Alert sidebar + feedback    :p3c, after p3b, 4d
  Search + timeline           :p3d, after p3c, 4d
  Dashboard views             :p3e, after p3d, 3d
  Phase 3 gate                :milestone, after p3e, 0d

  section Phase 4 — Hardening & Cutover
  Perf profiling + tuning     :p4a, after p3e, 5d
  Security audit              :p4b, after p4a, 3d
  E2E + visual regression     :p4c, after p4a, 5d
  Documentation               :p4d, after p4c, 3d
  Cutover                     :p4e, after p4d, 2d
  Phase 4 gate                :milestone, after p4e, 0d
```

> **Note**: Phase 1 and Phase 2 run in parallel once Phase 0 is complete.

---

## 3. Phase Index

| Phase | Title | Document | Key Deliverables |
|---|---|---|---|
| **0** | Foundation | [phase0_foundation.md](phase0_foundation.md) | `web-cop-gpu/` scaffold, Vite + SolidJS, capability gate, SharedArrayBuffer, Rust Wasm decoder |
| **1** | Core Rendering | [phase1_core_rendering.md](phase1_core_rendering.md) | WebGPU device, all pipelines (icons, trails, halos, labels), compute shaders, pick buffer |
| **2** | Backend Integration | [phase2_backend_integration.md](phase2_backend_integration.md) | Go FlatBuffer serializer, Go WebTransport server, Data Worker, priority shedding |
| **3** | UI & Interaction | [phase3_ui_interaction.md](phase3_ui_interaction.md) | SolidJS overlay components, track detail, alerts, feedback, search, dashboards |
| **4** | Hardening & Cutover | [phase4_hardening_cutover.md](phase4_hardening_cutover.md) | Perf tuning, security audit, E2E tests, visual regression, documentation, production cutover |

---

## 4. Dependencies Between Phases

```mermaid
flowchart TD
  P0["Phase 0<br/>Foundation"] --> P1["Phase 1<br/>Core Rendering"]
  P0 --> P2["Phase 2<br/>Backend Integration"]
  P1 --> P3["Phase 3<br/>UI & Interaction"]
  P2 --> P3
  P3 --> P4["Phase 4<br/>Hardening & Cutover"]

  style P0 fill:#4a9eff,color:#fff
  style P1 fill:#ff9f43,color:#fff
  style P2 fill:#ff9f43,color:#fff
  style P3 fill:#2ed573,color:#fff
  style P4 fill:#ff4757,color:#fff
```

- **Phase 1 and Phase 2 are parallel** — rendering pipeline and backend pipeline can be developed independently
- Phase 1 uses **mock data** (generated in Render Worker) until Phase 2 delivers real WebTransport data
- Phase 3 requires both Phase 1 (canvas to click on) and Phase 2 (live data) to be complete
- Phase 4 is the final integration, hardening, and cutover phase

---

## 5. Architecture References

All implementation must conform to:

| Document | Path | Sections |
|---|---|---|
| v1 Architecture (canonical) | `docs/architecture/v1/RTSA_WebGPU_Architecture_v1.md` | All sections |
| High-Level Architecture | `docs/architecture/high_level_architecture.md` | Dual-protocol, tech stack |
| Component Design | `docs/architecture/component_design.md` | Thread model, pipelines |
| Data Architecture | `docs/architecture/data_architecture.md` | §12 hot-path wire format |
| Security Architecture | `docs/architecture/security_architecture.md` | §13–16 browser security |
| SolidJS Standards | `docs/sdlc_guidelines/04_coding_standards/solidjs_standards.md` | All sections |
| WebGPU Guidelines | `docs/sdlc_guidelines/08_tech_specific/webgpu_guidelines.md` | All sections |
| WGSL Shader Standards | `docs/sdlc_guidelines/08_tech_specific/wgsl_shader_standards.md` | All sections |
| FlatBuffers Guidelines | `docs/sdlc_guidelines/08_tech_specific/flatbuffers_guidelines.md` | All sections |
| WebTransport Guidelines | `docs/sdlc_guidelines/08_tech_specific/webtransport_guidelines.md` | All sections |

---

## 6. Testing Strategy Per Phase

| Phase | Test Type | Coverage Target |
|---|---|---|
| 0 | Unit tests for Wasm decoder, capability gate | 80%+ line coverage |
| 1 | Compute shader output tests, visual regression baselines | All shaders tested |
| 2 | Go unit tests, integration tests (WebTransport + FlatBuffer round-trip) | 80%+ line coverage |
| 3 | SolidJS component tests (`@solidjs/testing-library`), Playwright E2E | 80%+ components |
| 4 | Full E2E suite, perf benchmarks (50k tracks @ 60 FPS), security scan | Pass all gates |

---

## 7. Risk Register

| Risk | Impact | Mitigation |
|---|---|---|
| Browser WebGPU support gaps | High | Capability gate (Phase 0), Chrome/Edge priority, degrade gracefully |
| WebTransport blocked by enterprise proxy | Medium | Fallback to gRPC-Web streaming (existing cold path) |
| 50k track perf target not met | High | Profiling in Phase 4, reduce visible layers, LOD system |
| Rust Wasm decoder size | Low | `opt-level = "s"`, tree-shaking, target < 100 KB |
| SharedArrayBuffer requires cross-origin isolation | Medium | COOP/COEP headers in Phase 0, test in CI |
