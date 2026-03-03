// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/stores/sensorHealthStore.test.ts

import { beforeEach, describe, expect, it } from "vitest";
import {
    getDLQSummary,
    getFilteredDLQEvents,
    getSensorKPIs,
    getSortedSensors,
    useSensorHealthStore,
} from "../../stores/sensorHealthStore";
import type { DLQEvent, SensorStatus } from "../../types/sensor";

const makeSensor = (
  overrides: Partial<SensorStatus> = {}
): SensorStatus => ({
  sensorId: "RADAR-01",
  sensorType: "RADAR",
  connected: true,
  connectionStatus: "connected",
  totalReceived: 1000,
  totalAccepted: 950,
  totalRejected: 50,
  lastObservationTime: new Date(),
  eventsPerSecond: 5.0,
  acceptanceRate: 95.0,
  rateHistory: [4, 5, 5, 6, 5],
  latencyMs: 120,
  ...overrides,
});

const makeDLQEvent = (overrides: Partial<DLQEvent> = {}): DLQEvent => ({
  eventId: `dlq-${Math.random()}`,
  sensorId: "RADAR-01",
  sensorType: "RADAR",
  timestamp: new Date(),
  rejectionReason: "Schema validation failed",
  ...overrides,
});

describe("sensorHealthStore", () => {
  beforeEach(() => {
    useSensorHealthStore.getState().clearAll();
    useSensorHealthStore.setState({
      sortField: "sensorId",
      sortDirection: "asc",
      activeTab: "grid",
      dlqFilterSensorType: "",
      dlqFilterReason: "",
      dlqFilterTimeRange: "1h",
      dlqPopupSensorId: null,
      isLoading: true,
      error: null,
    });
  });

  describe("upsertSensors", () => {
    it("adds new sensors to the map", () => {
      const s = makeSensor();
      useSensorHealthStore.getState().upsertSensors([s]);
      expect(useSensorHealthStore.getState().sensors.get("RADAR-01")).toBeDefined();
    });

    it("accumulates rate history across updates", () => {
      const s1 = makeSensor({ eventsPerSecond: 5.0, rateHistory: [] });
      useSensorHealthStore.getState().upsertSensors([s1]);
      const s2 = makeSensor({ eventsPerSecond: 6.0, rateHistory: [] });
      useSensorHealthStore.getState().upsertSensors([s2]);
      const stored = useSensorHealthStore.getState().sensors.get("RADAR-01");
      expect(stored?.rateHistory.length).toBe(2);
      expect(stored?.rateHistory).toContain(5.0);
      expect(stored?.rateHistory).toContain(6.0);
    });

    it("caps rate history at 30 entries", () => {
      for (let i = 0; i < 35; i++) {
        useSensorHealthStore
          .getState()
          .upsertSensors([makeSensor({ eventsPerSecond: i, rateHistory: [] })]);
      }
      const stored = useSensorHealthStore.getState().sensors.get("RADAR-01");
      expect(stored!.rateHistory.length).toBeLessThanOrEqual(30);
    });

    it("derives disconnected status when connected=false", () => {
      useSensorHealthStore.getState().upsertSensors([
        makeSensor({ connected: false, connectionStatus: "connected" }),
      ]);
      const stored = useSensorHealthStore.getState().sensors.get("RADAR-01");
      expect(stored?.connectionStatus).toBe("disconnected");
    });
  });

  describe("appendDLQEvents", () => {
    it("prepends new events to the DLQ list", () => {
      const e1 = makeDLQEvent({ eventId: "a" });
      const e2 = makeDLQEvent({ eventId: "b" });
      useSensorHealthStore.getState().appendDLQEvents([e1]);
      useSensorHealthStore.getState().appendDLQEvents([e2]);
      const events = useSensorHealthStore.getState().dlqEvents;
      expect(events[0].eventId).toBe("b");
      expect(events[1].eventId).toBe("a");
    });

    it("caps DLQ list at 500 events", () => {
      const many = Array.from({ length: 600 }, (_, i) =>
        makeDLQEvent({ eventId: `ev-${i}` })
      );
      useSensorHealthStore.getState().appendDLQEvents(many);
      expect(useSensorHealthStore.getState().dlqEvents.length).toBeLessThanOrEqual(500);
    });
  });

  describe("setSortField", () => {
    it("sets ascending on first click", () => {
      useSensorHealthStore.getState().setSortField("eventsPerSecond");
      expect(useSensorHealthStore.getState().sortField).toBe("eventsPerSecond");
      expect(useSensorHealthStore.getState().sortDirection).toBe("asc");
    });

    it("toggles to descending on second click of same field", () => {
      useSensorHealthStore.getState().setSortField("eventsPerSecond");
      useSensorHealthStore.getState().setSortField("eventsPerSecond");
      expect(useSensorHealthStore.getState().sortDirection).toBe("desc");
    });

    it("resets to ascending when switching fields", () => {
      useSensorHealthStore.getState().setSortField("eventsPerSecond");
      useSensorHealthStore.getState().setSortField("eventsPerSecond"); // desc
      useSensorHealthStore.getState().setSortField("sensorId");       // new field → asc
      expect(useSensorHealthStore.getState().sortDirection).toBe("asc");
    });
  });

  describe("setDLQFilter", () => {
    it("updates partial DLQ filters", () => {
      useSensorHealthStore.getState().setDLQFilter({ sensorType: "RADAR" });
      expect(useSensorHealthStore.getState().dlqFilterSensorType).toBe("RADAR");
      expect(useSensorHealthStore.getState().dlqFilterReason).toBe(""); // untouched
    });
  });

  describe("setDLQPopupSensorId", () => {
    it("sets and clears the popup sensor ID", () => {
      useSensorHealthStore.getState().setDLQPopupSensorId("EW-01");
      expect(useSensorHealthStore.getState().dlqPopupSensorId).toBe("EW-01");
      useSensorHealthStore.getState().setDLQPopupSensorId(null);
      expect(useSensorHealthStore.getState().dlqPopupSensorId).toBeNull();
    });
  });
});

describe("getSortedSensors", () => {
  const sensors: Map<string, SensorStatus> = new Map([
    ["C", makeSensor({ sensorId: "C", eventsPerSecond: 1 })],
    ["A", makeSensor({ sensorId: "A", eventsPerSecond: 3 })],
    ["B", makeSensor({ sensorId: "B", eventsPerSecond: 2 })],
  ]);

  it("sorts by sensorId ascending", () => {
    const result = getSortedSensors(sensors, "sensorId", "asc");
    expect(result.map((s) => s.sensorId)).toEqual(["A", "B", "C"]);
  });

  it("sorts by sensorId descending", () => {
    const result = getSortedSensors(sensors, "sensorId", "desc");
    expect(result.map((s) => s.sensorId)).toEqual(["C", "B", "A"]);
  });

  it("sorts by eventsPerSecond ascending", () => {
    const result = getSortedSensors(sensors, "eventsPerSecond", "asc");
    expect(result[0].eventsPerSecond).toBe(1);
    expect(result[2].eventsPerSecond).toBe(3);
  });

  it("sorts by connectionStatus (connected < degraded < disconnected)", () => {
    const mixed = new Map<string, SensorStatus>([
      ["x", makeSensor({ sensorId: "x", connectionStatus: "disconnected" })],
      ["y", makeSensor({ sensorId: "y", connectionStatus: "connected" })],
      ["z", makeSensor({ sensorId: "z", connectionStatus: "degraded" })],
    ]);
    const result = getSortedSensors(mixed, "connectionStatus", "asc");
    expect(result[0].connectionStatus).toBe("connected");
    expect(result[1].connectionStatus).toBe("degraded");
    expect(result[2].connectionStatus).toBe("disconnected");
  });
});

describe("getDLQSummary", () => {
  const events: DLQEvent[] = [
    makeDLQEvent({ sensorId: "R1", rejectionReason: "Schema" }),
    makeDLQEvent({ sensorId: "R1", rejectionReason: "Schema" }),
    makeDLQEvent({ sensorId: "R2", rejectionReason: "Timeout" }),
  ];

  it("counts total events", () => {
    expect(getDLQSummary(events).totalCount).toBe(3);
  });

  it("groups by sensor", () => {
    const s = getDLQSummary(events);
    expect(s.bySensor["R1"]).toBe(2);
    expect(s.bySensor["R2"]).toBe(1);
  });

  it("groups by reason", () => {
    const s = getDLQSummary(events);
    expect(s.byReason["Schema"]).toBe(2);
    expect(s.byReason["Timeout"]).toBe(1);
  });

  it("filters to a specific sensor", () => {
    const s = getDLQSummary(events, "R2");
    expect(s.totalCount).toBe(1);
    expect(s.bySensor["R1"]).toBeUndefined();
  });

  it("returns isolated pattern for few spread-out events", () => {
    const sparse: DLQEvent[] = [
      makeDLQEvent({ timestamp: new Date(Date.now() - 120_000) }),
      makeDLQEvent({ timestamp: new Date(Date.now() - 60_000) }),
    ];
    expect(getDLQSummary(sparse).pattern).toBe("isolated");
  });
});

describe("getFilteredDLQEvents", () => {
  const now = Date.now();
  const events: DLQEvent[] = [
    makeDLQEvent({ sensorId: "R1", sensorType: "RADAR", rejectionReason: "Schema", timestamp: new Date(now - 1000) }),
    makeDLQEvent({ sensorId: "E1", sensorType: "EW",    rejectionReason: "Timeout", timestamp: new Date(now - 1000) }),
    makeDLQEvent({ sensorId: "R1", sensorType: "RADAR", rejectionReason: "Schema", timestamp: new Date(now - 7200_000) }),
  ];

  it("returns all events when no filters applied", () => {
    const result = getFilteredDLQEvents(events, {
      sensorType: "",
      reason: "",
      timeRange: "all",
    });
    expect(result).toHaveLength(3);
  });

  it("filters by sensor type", () => {
    const result = getFilteredDLQEvents(events, {
      sensorType: "EW",
      reason: "",
      timeRange: "all",
    });
    expect(result).toHaveLength(1);
    expect(result[0].sensorId).toBe("E1");
  });

  it("filters by rejection reason", () => {
    const result = getFilteredDLQEvents(events, {
      sensorType: "",
      reason: "Timeout",
      timeRange: "all",
    });
    expect(result).toHaveLength(1);
  });

  it("filters by time range (1h excludes 2h-old events)", () => {
    const result = getFilteredDLQEvents(events, {
      sensorType: "",
      reason: "",
      timeRange: "1h",
    });
    expect(result).toHaveLength(2);
  });
});

describe("getSensorKPIs", () => {
  it("computes active/degraded/offline counts", () => {
    const sensors = new Map<string, SensorStatus>([
      ["a", makeSensor({ connectionStatus: "connected" })],
      ["b", makeSensor({ sensorId: "b", connectionStatus: "degraded" })],
      ["c", makeSensor({ sensorId: "c", connectionStatus: "disconnected" })],
    ]);
    const kpis = getSensorKPIs(sensors);
    expect(kpis.active).toBe(1);
    expect(kpis.degraded).toBe(1);
    expect(kpis.offline).toBe(1);
    expect(kpis.total).toBe(3);
  });

  it("sums throughput across all sensors", () => {
    const sensors = new Map<string, SensorStatus>([
      ["a", makeSensor({ eventsPerSecond: 5 })],
      ["b", makeSensor({ sensorId: "b", eventsPerSecond: 10 })],
    ]);
    const kpis = getSensorKPIs(sensors);
    expect(kpis.totalThroughput).toBe(15);
  });

  it("returns zeros for empty sensor map", () => {
    const kpis = getSensorKPIs(new Map());
    expect(kpis.active).toBe(0);
    expect(kpis.total).toBe(0);
    expect(kpis.avgLatency).toBe(0);
  });
});
