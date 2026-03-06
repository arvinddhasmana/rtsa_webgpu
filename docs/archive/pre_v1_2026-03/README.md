<!-- CLASSIFICATION: UNCLASSIFIED -->

# Archived Documents — Pre-v1 React/MapLibre/Zustand Era

> **Archived**: 2026-03-05
> **Reason**: Superseded by RTSA WebGPU Architecture v1
> **Authoritative Document**: `docs/architecture/v1/RTSA_WebGPU_Architecture_v1.md`

## Contents

These documents describe the original React 18 + MapLibre GL JS + Zustand + gRPC-Web streaming frontend architecture which has been replaced by:

- **SolidJS** — Command UI layer (replaces React 18)
- **WebGPU** — Tactical display rendering (replaces MapLibre GL / WebGL)
- **FlatBuffers + WebTransport** — Real-time hot path (replaces gRPC-Web streaming + Protobuf decode)
- **SharedArrayBuffer + Web Workers** — Off-main-thread data pipeline (replaces Zustand state management)

## Archived Files

| Original Location                                        | Description                                                           |
| -------------------------------------------------------- | --------------------------------------------------------------------- |
| `architecture/high_level_architecture.md`                | C4 high-level architecture (React UI references)                      |
| `architecture/component_design.md`                       | Component design with Zustand 4-store model                           |
| `architecture/dependency_graph.md`                       | React/Zustand/MapLibre dependency tree                                |
| `implementation/15-cop-web-app.md`                       | Module 15 — React COP Web Application                                 |
| `implementation/v2/`                                     | React UI v2 implementation phases (design system, dashboards, polish) |
| `implementation/v3/`                                     | React UI v3 role-based implementation phases                          |
| `status/`                                                | React-era implementation status tracking                              |
| `sdlc_guidelines/04_coding_standards/react_standards.md` | React 18 coding standards (replaced by `solidjs_standards.md`)        |

## Retention Policy

These documents are retained for reference during the WebGPU COP migration only. **Delete this entire directory after v4 implementation is validated and the old `web-cop/` codebase is removed.**
