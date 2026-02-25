<!-- CLASSIFICATION: UNCLASSIFIED -->

# RTSA COP Web Application

> **Module**: 15 — COP Web Application (React)
> **Classification**: UNCLASSIFIED

## Overview

The Common Operating Picture (COP) Web Application provides real-time situational
awareness to military operators. It connects to backend services via gRPC-Web through
the Envoy API Gateway.

## Features

- Real-time track map (MapLibre GL JS, Mid-Atlantic default: -60°, 45°)
- Priority alert queue with severity filtering
- Entity detail panel with operator feedback submission
- Historical forensics query and map replay
- Classification banner (top + bottom, always visible)
- Offline capability via Service Worker

## Technology Stack

| Technology | Version | Purpose |
|---|---|---|
| React | 18.x | UI framework |
| TypeScript | 5.x | Type safety |
| Vite | 5.x | Build tool |
| Vitest | 1.x | Unit testing |
| Zustand | 4.x | State management |
| MapLibre GL JS | 4.x | Map rendering |
| Tailwind CSS | 3.x | Utility CSS |

## Development

```bash
npm install
npm run dev          # Start dev server on :3000
npm run test         # Run unit tests
npm run test:coverage # Run tests with coverage report
npm run build        # Production build
```

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `VITE_GRPC_WEB_URL` | `https://localhost:8443` | Envoy gateway URL |
| `VITE_MAP_TILE_URL` | (offline) | Map tile server URL |
| `VITE_APP_TITLE` | `RTSA COP` | Application title |
| `VITE_CLASSIFICATION_CEILING` | `PROTECTED_B` | Default classification banner |

## Classification

All displayed data is classification-marked. The banner at top and bottom of the
viewport always reflects the highest classification of currently displayed data.
