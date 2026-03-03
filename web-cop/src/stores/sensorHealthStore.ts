// CLASSIFICATION: UNCLASSIFIED
// src/stores/sensorHealthStore.ts

import { create } from "zustand";
import type {
    DLQEvent,
    DLQSummary,
    SensorConnectionStatus,
    SensorStatus,
    SortDirection,
    SortField,
} from "../types/sensor";

const STALE_THRESHOLD_MS = 15_000; // 15s without data → degraded
const RATE_HISTORY_LENGTH = 30;

interface SensorHealthState {
  /** All known sensor statuses */
  sensors: Map<string, SensorStatus>;
  /** Dead Letter Queue events */
  dlqEvents: DLQEvent[];
  /** Currently selected sensor ID for drill-down */
  selectedSensorId: string | null;
  /** Active tab: 'grid' or 'dlq' */
  activeTab: "grid" | "dlq";
  /** Sort state */
  sortField: SortField;
  sortDirection: SortDirection;
  /** DLQ filters */
  dlqFilterSensorType: string;
  dlqFilterReason: string;
  dlqFilterTimeRange: "1h" | "6h" | "24h" | "all";
  /** Whether DLQ popup is open (for inline drill-down) */
  dlqPopupSensorId: string | null;
  /** Loading state */
  isLoading: boolean;
  /** Error state */
  error: string | null;

  // Actions
  upsertSensors: (statuses: SensorStatus[]) => void;
  appendDLQEvents: (events: DLQEvent[]) => void;
  selectSensor: (sensorId: string | null) => void;
  setActiveTab: (tab: "grid" | "dlq") => void;
  setSortField: (field: SortField) => void;
  setDLQFilter: (filter: Partial<{
    sensorType: string;
    reason: string;
    timeRange: "1h" | "6h" | "24h" | "all";
  }>) => void;
  setDLQPopupSensorId: (sensorId: string | null) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  clearAll: () => void;
}

export const useSensorHealthStore = create<SensorHealthState>((set, get) => ({
  sensors: new Map(),
  dlqEvents: [],
  selectedSensorId: null,
  activeTab: "grid",
  sortField: "sensorId",
  sortDirection: "asc",
  dlqFilterSensorType: "",
  dlqFilterReason: "",
  dlqFilterTimeRange: "1h",
  dlqPopupSensorId: null,
  isLoading: true,
  error: null,

  upsertSensors: (statuses) => {
    set((state) => {
      const nextMap = new Map(state.sensors);
      for (const s of statuses) {
        const existing = nextMap.get(s.sensorId);
        // Append to rate history for sparkline
        const history = existing?.rateHistory ?? [];
        const nextHistory = [...history, s.eventsPerSecond].slice(
          -RATE_HISTORY_LENGTH
        );
        nextMap.set(s.sensorId, {
          ...s,
          rateHistory: nextHistory,
          connectionStatus: deriveConnectionStatus(s),
        });
      }
      return { sensors: nextMap };
    });
  },

  appendDLQEvents: (events) => {
    set((state) => ({
      dlqEvents: [...events, ...state.dlqEvents].slice(0, 500), // Keep last 500
    }));
  },

  selectSensor: (sensorId) => set({ selectedSensorId: sensorId }),

  setActiveTab: (tab) => set({ activeTab: tab }),

  setSortField: (field) => {
    const state = get();
    set({
      sortField: field,
      sortDirection:
        state.sortField === field && state.sortDirection === "asc"
          ? "desc"
          : "asc",
    });
  },

  setDLQFilter: (filter) =>
    set((state) => ({
      dlqFilterSensorType: filter.sensorType ?? state.dlqFilterSensorType,
      dlqFilterReason: filter.reason ?? state.dlqFilterReason,
      dlqFilterTimeRange: filter.timeRange ?? state.dlqFilterTimeRange,
    })),

  setDLQPopupSensorId: (sensorId) => set({ dlqPopupSensorId: sensorId }),

  setLoading: (loading) => set({ isLoading: loading }),
  setError: (error) => set({ error }),
  clearAll: () =>
    set({
      sensors: new Map(),
      dlqEvents: [],
      selectedSensorId: null,
      dlqPopupSensorId: null,
    }),
}));

// ── Derived Selectors ─────────────────────────────────

/** Derive connection status from sensor data */
function deriveConnectionStatus(s: SensorStatus): SensorConnectionStatus {
  if (!s.connected) return "disconnected";
  if (
    s.lastObservationTime &&
    Date.now() - s.lastObservationTime.getTime() > STALE_THRESHOLD_MS
  ) {
    return "degraded";
  }
  if (s.eventsPerSecond < 0.1 && s.totalReceived > 0) return "degraded";
  return "connected";
}

/** Get sorted sensor list */
export function getSortedSensors(
  sensors: Map<string, SensorStatus>,
  sortField: SortField,
  sortDirection: SortDirection
): SensorStatus[] {
  const list = Array.from(sensors.values());
  const dir = sortDirection === "asc" ? 1 : -1;

  return list.sort((a, b) => {
    switch (sortField) {
      case "sensorId":
        return dir * a.sensorId.localeCompare(b.sensorId);
      case "sensorType":
        return dir * a.sensorType.localeCompare(b.sensorType);
      case "eventsPerSecond":
        return dir * (a.eventsPerSecond - b.eventsPerSecond);
      case "totalRejected":
        return dir * (a.totalRejected - b.totalRejected);
      case "latencyMs":
        return dir * (a.latencyMs - b.latencyMs);
      case "connectionStatus": {
        const order = { connected: 0, degraded: 1, disconnected: 2 };
        return dir * (order[a.connectionStatus] - order[b.connectionStatus]);
      }
      case "lastObservationTime": {
        const aT = a.lastObservationTime?.getTime() ?? 0;
        const bT = b.lastObservationTime?.getTime() ?? 0;
        return dir * (aT - bT);
      }
      default:
        return 0;
    }
  });
}

/** Get DLQ summary for a sensor or all sensors */
export function getDLQSummary(
  events: DLQEvent[],
  sensorId?: string
): DLQSummary {
  const filtered = sensorId
    ? events.filter((e) => e.sensorId === sensorId)
    : events;

  const bySensor: Record<string, number> = {};
  const byReason: Record<string, number> = {};

  for (const e of filtered) {
    bySensor[e.sensorId] = (bySensor[e.sensorId] ?? 0) + 1;
    byReason[e.rejectionReason] = (byReason[e.rejectionReason] ?? 0) + 1;
  }

  // Determine pattern
  let pattern: DLQSummary["pattern"] = "isolated";
  if (filtered.length > 0) {
    const sorted = [...filtered].sort(
      (a, b) => a.timestamp.getTime() - b.timestamp.getTime()
    );
    const gaps: number[] = [];
    for (let i = 1; i < sorted.length; i++) {
      gaps.push(
        sorted[i].timestamp.getTime() - sorted[i - 1].timestamp.getTime()
      );
    }
    const avgGap = gaps.length > 0 ? gaps.reduce((a, b) => a + b, 0) / gaps.length : Infinity;
    if (avgGap < 5000 && filtered.length > 10) pattern = "sustained";
    else if (avgGap < 30000 && filtered.length > 3) pattern = "burst";
  }

  return {
    totalCount: filtered.length,
    bySensor,
    byReason,
    pattern,
  };
}

/** Get filtered DLQ events */
export function getFilteredDLQEvents(
  events: DLQEvent[],
  filters: {
    sensorType: string;
    reason: string;
    timeRange: "1h" | "6h" | "24h" | "all";
  }
): DLQEvent[] {
  const now = Date.now();
  const rangeMs: Record<string, number> = {
    "1h": 3600_000,
    "6h": 21600_000,
    "24h": 86400_000,
    all: Infinity,
  };
  const cutoff = now - (rangeMs[filters.timeRange] ?? Infinity);

  return events.filter((e) => {
    if (filters.sensorType && e.sensorType !== filters.sensorType) return false;
    if (filters.reason && e.rejectionReason !== filters.reason) return false;
    if (e.timestamp.getTime() < cutoff) return false;
    return true;
  });
}

/** Aggregate KPI metrics from sensor map */
export function getSensorKPIs(sensors: Map<string, SensorStatus>) {
  const list = Array.from(sensors.values());
  const active = list.filter((s) => s.connectionStatus === "connected").length;
  const degraded = list.filter(
    (s) => s.connectionStatus === "degraded"
  ).length;
  const offline = list.filter(
    (s) => s.connectionStatus === "disconnected"
  ).length;
  const totalThroughput = list.reduce((acc, s) => acc + s.eventsPerSecond, 0);
  const avgLatency =
    list.length > 0
      ? list.reduce((acc, s) => acc + s.latencyMs, 0) / list.length
      : 0;

  return { active, degraded, offline, totalThroughput, avgLatency, total: list.length };
}
