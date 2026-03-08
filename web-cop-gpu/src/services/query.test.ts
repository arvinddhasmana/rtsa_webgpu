// CLASSIFICATION: UNCLASSIFIED
// src/services/query.test.ts — Unit tests for QueryService gRPC wrapper
//
// Tests:
//   - fetchTrackDetail returns a correctly typed TrackDetail object
//   - fetchTrackDetail returns null when no matching track is found
//   - searchTracks handles an empty result set without error
//   - searchTracks handles pagination fields in proto response
//   - Label mapping functions (entity type, hostile class, status, classification)

import { describe, it, expect, vi, beforeEach } from "vitest";

// ── Mock ConnectRPC client at the transport boundary ──────────────────────────

const mockQueryTracks = vi.hoisted(() => vi.fn());
const mockGetEventTimeline = vi.hoisted(() => vi.fn());

vi.mock("@connectrpc/connect", () => ({
  createPromiseClient: () => ({
    queryTracks: mockQueryTracks,
    getEventTimeline: mockGetEventTimeline,
  }),
}));

vi.mock("./grpc-client", () => ({ transport: {} }));

vi.mock("@gen/rtsa/query/v1/query_service_connect.js", () => ({
  QueryService: {},
}));

vi.mock("@gen/rtsa/common/v1/types_pb.js", () => ({
  ClassificationLevel: {
    UNCLASSIFIED: 1,
    PROTECTED_A: 2,
    PROTECTED_B: 3,
    PROTECTED_C: 4,
    SECRET: 5,
  },
  EntityType: {
    AIR: 2,
    SURFACE: 1,
    SUBSURFACE: 3,
    LAND: 4,
    CYBER: 5,
  },
  HostileClassification: {
    HOSTILE: 5,
    FRIENDLY: 2,
    NEUTRAL: 3,
    UNKNOWN: 0,
    SUSPECT: 4,
  },
  TrackStatus: {
    ACTIVE: 1,
    NEW: 2,
    STALE: 3,
    DROPPED: 4,
    MERGED: 5,
  },
  FeedbackType: {},
}));

// Mock the FusedTrack protobuf type (just a plain object for tests)
vi.mock("@gen/rtsa/entity/v1/fused_track_pb.js", () => ({}));

import { fetchTrackDetail, searchTracks } from "./query";

// ── Helper to build a mock FusedTrack proto response ─────────────────────────

function makeMockTrack(overrides: Record<string, unknown> = {}): Record<string, unknown> {
  return {
    trackId: "TRK-001",
    entityType: 2,         // AIR
    hostileClass: 5,       // HOSTILE
    status: 1,             // ACTIVE
    classification: 1,     // UNCLASSIFIED
    confidenceScore: 0.87,
    sourceCount: 3,
    estimatedPosition: {
      latitude: 45.5,
      longitude: -75.7,
      altitudeMeters: 10000,
      speedKnots: 420,
      headingDegrees: 270,
    },
    label: "BOGEY-01",
    updatedAt: { seconds: BigInt(1_700_000_000) },
    ...overrides,
  };
}

// ── fetchTrackDetail ──────────────────────────────────────────────────────────

describe("fetchTrackDetail", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("returns a correctly typed TrackDetail mapped from proto response", async () => {
    mockQueryTracks.mockResolvedValue({ tracks: [makeMockTrack()] });

    const detail = await fetchTrackDetail("TRK-001");

    expect(detail).not.toBeNull();
    expect(detail!.trackId).toBe("TRK-001");
    expect(detail!.entityType).toBe("Air");
    expect(detail!.hostileClass).toBe("Hostile");
    expect(detail!.status).toBe("Active");
    expect(detail!.classification).toBe("UNCLASSIFIED");
    expect(detail!.confidenceScore).toBe(0.87);
    expect(detail!.sourceCount).toBe(3);
    expect(detail!.lat).toBe(45.5);
    expect(detail!.lon).toBe(-75.7);
    expect(detail!.altitudeMeters).toBe(10000);
    expect(detail!.speedKnots).toBe(420);
    expect(detail!.headingDeg).toBe(270);
    expect(detail!.label).toBe("BOGEY-01");
    expect(detail!.updatedAtMs).toBe(1_700_000_000_000);
  });

  it("returns null when no matching track is found (empty tracks array)", async () => {
    mockQueryTracks.mockResolvedValue({ tracks: [] });

    const detail = await fetchTrackDetail("TRK-MISSING");

    expect(detail).toBeNull();
  });

  it("calls queryTracks with pageSize=1 and the given trackId", async () => {
    mockQueryTracks.mockResolvedValue({ tracks: [] });

    await fetchTrackDetail("TRK-TEST");

    const [args] = mockQueryTracks.mock.calls[0] as [Record<string, unknown>];
    expect(args["trackId"]).toBe("TRK-TEST");
    const pagination = args["pagination"] as Record<string, unknown>;
    expect(pagination["pageSize"]).toBe(1);
  });

  it("handles missing estimatedPosition gracefully (lat/lon/alt default to 0)", async () => {
    mockQueryTracks.mockResolvedValue({
      tracks: [
        makeMockTrack({ estimatedPosition: undefined }),
      ],
    });

    const detail = await fetchTrackDetail("TRK-NOPOS");

    expect(detail!.lat).toBe(0);
    expect(detail!.lon).toBe(0);
    expect(detail!.altitudeMeters).toBe(0);
  });

  it("handles missing updatedAt by returning 0 for updatedAtMs", async () => {
    mockQueryTracks.mockResolvedValue({
      tracks: [makeMockTrack({ updatedAt: null })],
    });

    const detail = await fetchTrackDetail("TRK-NOTIME");

    expect(detail!.updatedAtMs).toBe(0);
  });
});

// ── searchTracks ──────────────────────────────────────────────────────────────

describe("searchTracks", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("returns an empty array without error when tracks is empty", async () => {
    mockQueryTracks.mockResolvedValue({ tracks: [] });

    const results = await searchTracks("TRK");

    expect(results).toEqual([]);
  });

  it("returns multiple correctly mapped TrackDetail objects", async () => {
    mockQueryTracks.mockResolvedValue({
      tracks: [
        makeMockTrack({ trackId: "TRK-A" }),
        makeMockTrack({ trackId: "TRK-B", hostileClass: 2 }), // FRIENDLY
      ],
    });

    const results = await searchTracks("TRK");

    expect(results).toHaveLength(2);
    expect(results[0].trackId).toBe("TRK-A");
    expect(results[1].trackId).toBe("TRK-B");
    expect(results[1].hostileClass).toBe("Friendly");
  });

  it("passes pagination pageSize = limit argument to queryTracks", async () => {
    mockQueryTracks.mockResolvedValue({ tracks: [] });

    await searchTracks("TRK", 50);

    const [args] = mockQueryTracks.mock.calls[0] as [Record<string, unknown>];
    const pagination = args["pagination"] as Record<string, unknown>;
    expect(pagination["pageSize"]).toBe(50);
  });

  it("uses default limit of 20 when not specified", async () => {
    mockQueryTracks.mockResolvedValue({ tracks: [] });

    await searchTracks("TRK");

    const [args] = mockQueryTracks.mock.calls[0] as [Record<string, unknown>];
    const pagination = args["pagination"] as Record<string, unknown>;
    expect(pagination["pageSize"]).toBe(20);
  });

  it("includes pageToken in the request for pagination", async () => {
    mockQueryTracks.mockResolvedValue({ tracks: [] });

    await searchTracks("TRK", 10);

    const [args] = mockQueryTracks.mock.calls[0] as [Record<string, unknown>];
    const pagination = args["pagination"] as Record<string, unknown>;
    expect(pagination).toHaveProperty("pageToken");
  });

  it("maps all entity types correctly", async () => {
    const entityTypes = [
      { value: 1, expected: "Surface" },
      { value: 2, expected: "Air" },
      { value: 3, expected: "Subsurface" },
      { value: 4, expected: "Land" },
      { value: 5, expected: "Cyber" },
    ];

    for (const { value, expected } of entityTypes) {
      mockQueryTracks.mockResolvedValue({
        tracks: [makeMockTrack({ entityType: value })],
      });
      const results = await searchTracks("TRK", 1);
      expect(results[0].entityType).toBe(expected);
    }
  });
});
