# React Frontend Standards

> **CLASSIFICATION: UNCLASSIFIED**
> **Document Type**: Coding Standard
> **Parent**: `04_coding_standards/general_coding.md`
> **Dependencies**: `04_coding_standards/secure_coding.md`
> **Last Updated**: 2026-02-23

---

## 1. Purpose

This document defines React/TypeScript coding standards for the RTSA situational awareness dashboard. The frontend provides real-time visualization of entity tracks, AI anomaly detections, operator feedback submission, and historical analysis interfaces.

## 2. Technology Stack

| Technology | Version | Purpose |
|---|---|---|
| React | 18+ | UI framework |
| TypeScript | 5+ | Type-safe JavaScript |
| gRPC-Web | latest | gRPC communication with backend services |
| WebSocket | native | Real-time streaming (fallback from gRPC-Web streaming) |
| Leaflet / MapLibre | latest | Geospatial map rendering |
| Zustand / Redux Toolkit | latest | State management |
| Vitest | latest | Unit/component testing |
| React Testing Library | latest | Component testing |

## 3. Project Structure

```
src/
├── app/
│   ├── App.tsx                      # Root component, router, providers
│   ├── routes.tsx                   # Route definitions
│   └── providers/
│       ├── AuthProvider.tsx         # Authentication context
│       ├── GrpcProvider.tsx         # gRPC client provider
│       └── WebSocketProvider.tsx    # WebSocket connection provider
├── components/
│   ├── common/                      # Reusable UI components
│   │   ├── Button/
│   │   ├── Card/
│   │   ├── Modal/
│   │   └── DataTable/
│   ├── map/                         # Geospatial map components
│   │   ├── SituationalMap.tsx       # Main map view
│   │   ├── EntityMarker.tsx         # Entity track marker
│   │   ├── TrackLine.tsx            # Track history line
│   │   └── ThreatZone.tsx           # Threat/risk overlay
│   ├── timeline/                    # Timeline components
│   │   ├── EventTimeline.tsx        # Chronological event view
│   │   └── TimelineFilter.tsx       # Time range controls
│   ├── feedback/                    # Operator feedback components
│   │   ├── FeedbackPanel.tsx        # Feedback submission form
│   │   └── FeedbackHistory.tsx      # Previous feedback view
│   ├── alerts/                      # Alert/anomaly components
│   │   ├── AlertPanel.tsx           # Active alerts list
│   │   └── AnomalyDetail.tsx        # Anomaly detail view
│   └── analytics/                   # Historical analysis components
│       ├── QueryBuilder.tsx         # ClickHouse query interface
│       └── AnalyticsChart.tsx       # Data visualization charts
├── hooks/
│   ├── useEntityStream.ts           # Real-time entity stream hook
│   ├── useWebSocket.ts              # WebSocket connection hook
│   ├── useGrpcClient.ts             # gRPC-Web client hook
│   └── useAuth.ts                   # Authentication hook
├── services/
│   ├── grpc/                        # gRPC-Web client wrappers
│   │   ├── ingestionClient.ts
│   │   ├── entityClient.ts
│   │   ├── feedbackClient.ts
│   │   └── auditClient.ts
│   └── api/
│       └── analyticsApi.ts          # REST API for analytics (if needed)
├── store/
│   ├── entityStore.ts               # Entity track state
│   ├── alertStore.ts                # Alert/anomaly state
│   ├── feedbackStore.ts             # Feedback state
│   └── uiStore.ts                   # UI state (filters, selections)
├── types/
│   ├── entity.ts                    # Entity domain types
│   ├── sensor.ts                    # Sensor types
│   └── feedback.ts                  # Feedback types
├── utils/
│   ├── coordinates.ts               # WGS 84 coordinate utilities
│   ├── classification.ts            # Classification level utilities
│   ├── formatters.ts                # Display formatting
│   └── validators.ts                # Input validation
└── generated/                       # protoc-gen-grpc-web output
    └── proto/
```

## 4. Component Design Rules

### 4.1 Component Types

| Type | File Convention | Purpose |
|---|---|---|
| Page | `*Page.tsx` | Route-level container; composes layout + features |
| Feature | `*Panel.tsx`, `*View.tsx` | Self-contained feature unit with its own state |
| Presentational | `*Display.tsx`, `*Card.tsx` | Pure display; props-only, no side effects |
| Common | Named by function | Reusable UI atoms (Button, Modal, Table) |

### 4.2 Component Template

```tsx
// CLASSIFICATION: UNCLASSIFIED

/**
 * EntityMarker — Renders an entity track on the situational map.
 *
 * Traceability:
 *   Feature: F-005 (Real-Time Dashboard)
 *   Use Case: UC-005 (Situational Awareness Display)
 *   Requirements: FR-VIS-001, FR-VIS-003
 */

import React, { memo } from 'react';
import type { Entity } from '@/types/entity';

interface EntityMarkerProps {
  /** Entity track data to display */
  entity: Entity;
  /** Whether this entity is currently selected */
  isSelected: boolean;
  /** Callback when entity is clicked */
  onSelect: (entityId: string) => void;
}

export const EntityMarker = memo(function EntityMarker({
  entity,
  isSelected,
  onSelect,
}: EntityMarkerProps) {
  // Component implementation
  return (
    // JSX
  );
});
```

### 4.3 Component Rules

- Use functional components with hooks (no class components)
- Use `memo` for components receiving entity/track data (high-frequency updates)
- Props interfaces must be explicitly typed (no `any`)
- Side effects only in hooks (`useEffect`, `useCallback`, custom hooks)
- No direct DOM manipulation — use refs only when React abstractions are insufficient

## 5. State Management

### 5.1 State Categories

| Category | Location | Example |
|---|---|---|
| Server state | gRPC streams + store | Entity tracks, alerts, sensor status |
| Local UI state | Component `useState` | Modal open/closed, form inputs |
| Global UI state | Zustand store | Selected entity, map viewport, filters |
| Auth state | Context provider | User identity, roles, tokens |

### 5.2 Real-Time Data Pattern

```tsx
// CLASSIFICATION: UNCLASSIFIED

import { useEffect, useCallback } from 'react';
import { useEntityStore } from '@/store/entityStore';

/**
 * useEntityStream — Subscribes to real-time entity track updates
 * via gRPC-Web server streaming.
 *
 * Traceability: UC-005 (Situational Awareness Display)
 */
export function useEntityStream(filters: EntityFilter) {
  const updateEntity = useEntityStore((s) => s.updateEntity);
  const removeEntity = useEntityStore((s) => s.removeEntity);

  useEffect(() => {
    const stream = entityClient.streamEntityUpdates(filters);

    stream.on('data', (update) => {
      updateEntity(mapProtoToEntity(update));
    });

    stream.on('error', (err) => {
      // Log structured error — no classified data
      logger.error('Entity stream error', {
        code: err.code,
        message: err.message,
      });
    });

    return () => {
      stream.cancel();
    };
  }, [filters, updateEntity]);
}
```

## 6. Security Rules for Frontend

### 6.1 Input Validation

- Validate all operator input before submission (feedback text, query parameters)
- Sanitize display of any user-generated content (XSS prevention)
- Never trust data from URL parameters without validation

### 6.2 Authentication

- MFA enforced via identity provider integration
- JWT tokens stored in HTTP-only secure cookies (not localStorage)
- Token refresh handled by AuthProvider
- Unauthorized routes redirect to login

### 6.3 Classification Display

- Display the classification marking prominently on every page/view
- Color-code by classification level (green=UNCLASSIFIED, yellow=PROTECTED, red=SECRET)
- Never display data above the user's authorized classification level

### 6.4 No Classified Data in Client

- The frontend never stores classified data in localStorage/sessionStorage
- Browser cache headers must prevent caching of classified responses
- Console logging must never output entity position data or sensor payloads

## 7. Offline / Edge Capability

For tactical edge deployments where connectivity is intermittent:

### 7.1 Offline Queue

Feedback submissions queue locally when offline and sync when reconnected:

```tsx
const submitFeedback = useCallback(async (feedback: Feedback) => {
  try {
    await feedbackClient.submitFeedback(feedback);
  } catch (err) {
    if (isNetworkError(err)) {
      offlineQueue.enqueue('feedback', feedback);
      showNotification('Feedback queued for sync');
    } else {
      throw err;
    }
  }
}, []);
```

### 7.2 Graceful Degradation

| Feature | Online | Offline/Edge |
|---|---|---|
| Entity map | Full real-time | Last known positions; reduced update rate |
| Feedback | Real-time submission | Queued for sync |
| Historical queries | Full ClickHouse | Unavailable (show message) |
| Alerts | Real-time | Local alerts only |
| NATO interop | Live | Unavailable |

## 8. Accessibility (Operational UI)

Military operational UIs have unique accessibility requirements:

- High contrast themes for daylight and night-vision (NVG-compatible dark mode)
- Keyboard navigable — critical for operators wearing gloves
- Large click targets — minimum 44×44px for touch/glove operation
- Screen reader support for alert notifications
- No reliance on color alone for status indication (use icons + text)

## 9. Testing Standards

| Test Type | Tool | Coverage Target | What to Test |
|---|---|---|---|
| Unit | Vitest | 80%+ | Utility functions, formatters, validators |
| Component | React Testing Library | Key interactions | User interactions, state changes, rendering |
| Hook | renderHook (RTL) | All custom hooks | State management, side effects |
| E2E | Playwright | Critical paths | Login → dashboard → feedback submission |

## 10. AI Agent Instructions

When generating React/TypeScript code:

1. Start every file with `// CLASSIFICATION: UNCLASSIFIED`
2. Include traceability comments (Feature, UC, Requirements) on components
3. Use functional components with TypeScript interfaces for props
4. Use `memo` for components receiving high-frequency entity data
5. Never log or store classified data in the browser
6. Validate all operator input before submission
7. Handle offline scenarios with queued operations
8. Include keyboard navigation and high-contrast support
9. Write tests for all custom hooks and utility functions
10. Use Zustand stores for shared state; React state for local UI state
