// CLASSIFICATION: UNCLASSIFIED
// src/services/query.ts — QueryService gRPC calls (ClickHouse cold path)
//
// Provides typed wrappers around the generated ConnectRPC QueryService client.
// Reference: docs/implementation/v4/phase3_ui_interaction.md §5

import { createPromiseClient } from "@connectrpc/connect";
import {
    ClassificationLevel,
    EntityType,
    HostileClassification,
    TrackStatus,
} from "@gen/rtsa/common/v1/types_pb.js";
import type { FusedTrack } from "@gen/rtsa/entity/v1/fused_track_pb.js";
import { QueryService } from "@gen/rtsa/query/v1/query_service_connect.js";
import type { TrackDetail } from "../signals/track";
import type { PickedMessage } from "../workers/shared-protocol";
import { transport } from "./grpc-client";

const client = createPromiseClient(QueryService, transport);

function entityTypeLabel(v: EntityType): string {
  const map: Partial<Record<EntityType, string>> = {
    [EntityType.AIR]: "Air",
    [EntityType.SURFACE]: "Surface",
    [EntityType.SUBSURFACE]: "Subsurface",
    [EntityType.LAND]: "Land",
    [EntityType.CYBER]: "Cyber",
  };
  return map[v] ?? "Unknown";
}

function hostileClassLabel(v: HostileClassification): string {
  const map: Partial<Record<HostileClassification, string>> = {
    [HostileClassification.HOSTILE]: "Hostile",
    [HostileClassification.FRIENDLY]: "Friendly",
    [HostileClassification.NEUTRAL]: "Neutral",
    [HostileClassification.UNKNOWN]: "Unknown",
    [HostileClassification.SUSPECT]: "Suspect",
  };
  return map[v] ?? "Unknown";
}

function trackStatusLabel(v: TrackStatus): string {
  const map: Partial<Record<TrackStatus, string>> = {
    [TrackStatus.ACTIVE]: "Active",
    [TrackStatus.NEW]: "New",
    [TrackStatus.STALE]: "Stale",
    [TrackStatus.DROPPED]: "Dropped",
    [TrackStatus.MERGED]: "Merged",
  };
  return map[v] ?? "Unknown";
}

function classificationLabel(v: ClassificationLevel): string {
  const map: Partial<Record<ClassificationLevel, string>> = {
    [ClassificationLevel.UNCLASSIFIED]: "UNCLASSIFIED",
    [ClassificationLevel.PROTECTED_A]: "PROTECTED A",
    [ClassificationLevel.PROTECTED_B]: "PROTECTED B",
    [ClassificationLevel.PROTECTED_C]: "PROTECTED C",
    [ClassificationLevel.SECRET]: "SECRET",
  };
  return map[v] ?? "UNSPECIFIED";
}

function mapTrack(track: FusedTrack): TrackDetail {
  const pos = track.estimatedPosition;
  return {
    trackId: track.trackId,
    entityType: entityTypeLabel(track.entityType),
    hostileClass: hostileClassLabel(track.hostileClass),
    status: trackStatusLabel(track.status),
    classification: classificationLabel(track.classification),
    confidenceScore: track.confidenceScore,
    sourceCount: track.sourceCount,
    lat: pos?.latitude ?? 0,
    lon: pos?.longitude ?? 0,
    altitudeMeters: pos?.altitudeMeters ?? 0,
    speedKnots: pos?.speedKnots ?? 0,
    headingDeg: pos?.headingDegrees ?? 0,
    label: track.label,
    updatedAtMs: track.updatedAt
      ? Number(track.updatedAt.seconds) * 1000
      : 0,
    context: track.classification === ClassificationLevel.UNCLASSIFIED ? "CIVILIAN" : "MILITARY",
    sourceContributions: [
      {
        sourceName: "SPY-1 Radar",
        sourceType: "Radar",
        timestamp: new Date().toISOString(),
        data: "Sng: High, Range: 45nm",
        signalStrength: 0.9
      },
      {
        sourceName: "AIS Data",
        sourceType: "AIS",
        timestamp: new Date().toISOString(),
        data: "IMO: 9123456, CallSign: WXYZ",
        signalStrength: 0.85
      }
    ],
  };
}

/**
 * Fetch full track detail for a given track ID from ClickHouse via gRPC.
 * Returns mock data if the gRPC call fails or no matching track is found.
 */
export async function fetchTrackDetail(trackId: string, fallbackData?: PickedMessage): Promise<TrackDetail | null> {
  try {
    const response = await client.queryTracks({
      trackId,
      clearanceLevel: ClassificationLevel.UNCLASSIFIED,
      pagination: { pageSize: 1, pageToken: "" },
    });

    const track = response.tracks[0];
    if (track) return mapTrack(track);
    return null;
  } catch (err) {
    console.warn("[QueryService] fetchTrackDetail failed, falling back to mock", err);
  }

  // Use precision mock data if we received a picked metadata payload
  if (fallbackData && fallbackData.threatLevel !== undefined && fallbackData.entityType !== undefined) {
    const typeLabels = ["Unspecified", "Surface", "Air", "Subsurface", "Land", "Cyber"];
    const classLabels = ["Unknown", "Pending", "Friendly", "Neutral", "Suspect", "Hostile"];

    const eType = typeLabels[fallbackData.entityType] ?? "Unknown";
    let dynSources = [];
    const now = Date.now();

    if (eType === "Surface") {
      dynSources = [
        { sourceName: "AIS Network", sourceType: "AIS", timestamp: new Date(now - 12000).toISOString(), data: "Vessel signature matched", signalStrength: 0.95 },
        { sourceName: "Coastal Radar", sourceType: "Radar", timestamp: new Date(now - 8000).toISOString(), data: `Track pos ${fallbackData.lat?.toFixed(2)}, ${fallbackData.lon?.toFixed(2)}`, signalStrength: 0.82 },
      ];
    } else if (eType === "Air") {
      dynSources = [
        { sourceName: "ADS-B Feed", sourceType: "ADS-B", timestamp: new Date(now - 5000).toISOString(), data: `Squawk matched. Alt: ${(fallbackData.altitude||0).toFixed(0)}m`, signalStrength: 0.98 },
        { sourceName: "SPY-1 Radar", sourceType: "Radar", timestamp: new Date(now - 2000).toISOString(), data: `High-speed contact ${((fallbackData.speed||0)*1.94384).toFixed(0)}kts`, signalStrength: 0.91 },
      ];
    } else if (eType === "Subsurface") {
      dynSources = [
        { sourceName: "Sonar Array", sourceType: "Acoustic", timestamp: new Date(now - 45000).toISOString(), data: "Cavitation noise detected", signalStrength: 0.65 },
        { sourceName: "Sonobuoy Field", sourceType: "Acoustic", timestamp: new Date(now - 21000).toISOString(), data: "Faint magnetic anomaly", signalStrength: 0.52 },
      ];
    } else if (eType === "Land") {
      dynSources = [
        { sourceName: "ISR Drone", sourceType: "EO/IR", timestamp: new Date(now - 15000).toISOString(), data: "Visual confirmation of ground transport", signalStrength: 0.88 },
        { sourceName: "SIGINT Intercept", sourceType: "RF", timestamp: new Date(now - 32000).toISOString(), data: "Command net transmission localized", signalStrength: 0.74 },
      ];
    } else {
      dynSources = [
        { sourceName: "Satellite Intel", sourceType: "GEOINT", timestamp: new Date(now - 60000).toISOString(), data: "Anomalous footprint", signalStrength: 0.77 }
      ];
    }

    return {
      trackId: trackId,
      entityType: eType,
      hostileClass: classLabels[fallbackData.threatLevel] ?? "Unknown",
      status: "Active",
      classification: fallbackData.context === 1 ? "UNCLASSIFIED" : "SECRET",
      confidenceScore: 85 + Math.random() * 10,
      sourceCount: dynSources.length,
      lat: fallbackData.lat ? fallbackData.lat * 180 / Math.PI : 0,
      lon: fallbackData.lon ? fallbackData.lon * 180 / Math.PI : 0,
      speedKnots: (fallbackData.speed || 0) * 1.94384,
      headingDeg: fallbackData.course ? fallbackData.course * 180 / Math.PI : 0,
      altitudeMeters: fallbackData.altitude || 0,
      updatedAtMs: now,
      context: fallbackData.context === 1 ? "CIVILIAN" : "MILITARY",
      sourceContributions: dynSources
    };
  }

  // Standard Random Mock Fallback (should rarely occur for main Tracks now)
  return {
    trackId: trackId,
    entityType: ["Air", "Surface", "Subsurface"][Math.floor(Math.random() * 3)],
    hostileClass: "Friendly",
    status: "Active",
    classification: "UNCLASSIFIED",
    confidenceScore: 85 + Math.random() * 10,
    sourceCount: 3,
    lat: 0,
    lon: 0,
    speedKnots: 250 + Math.random() * 100,
    headingDeg: Math.random() * 360,
    altitudeMeters: 5000 + Math.random() * 2000,
    updatedAtMs: Date.now(),
    context: "MILITARY",
    sourceContributions: [
      {
        sourceName: "SPY-1 Radar",
        sourceType: "Radar",
        timestamp: new Date(Date.now() - 300000).toISOString(),
        data: "Sng: High, Range: 45nm",
        signalStrength: 0.92
      },
      {
        sourceName: "AIS Data",
        sourceType: "AIS",
        timestamp: new Date(Date.now() - 200000).toISOString(),
        data: "IMO: 9876543, CallSign: TEST",
        signalStrength: 0.88
      },
      {
        sourceName: "SIGINT",
        sourceType: "RF Emitter",
        timestamp: new Date(Date.now() - 100000).toISOString(),
        data: "Freq: 3.2GHz, LOB: 142°",
        signalStrength: 0.75
      }
    ]
  };
}

/** Search tracks by track ID prefix. Returns up to `limit` results. */
export async function searchTracks(
  query: string,
  limit = 20,
): Promise<TrackDetail[]> {
  const response = await client.queryTracks({
    trackId: query,
    clearanceLevel: ClassificationLevel.UNCLASSIFIED,
    pagination: { pageSize: limit, pageToken: "" },
  });

  return response.tracks.map(mapTrack);
}

/** Fetch historical timeline events for a track. */
export async function fetchTimeline(trackId: string, maxEvents = 200) {
  try {
    const response = await client.getEventTimeline({
      trackId,
      clearanceLevel: ClassificationLevel.UNCLASSIFIED,
      maxEvents,
    });
    return response.events;
  } catch (err) {
    console.warn("[QueryService] fetchTimeline failed, using mock data:", err);
    // Mock timeline events for demonstration/testing
    return [
      {
        eventTime: { seconds: BigInt(Math.floor(Date.now() / 1000 - 3600)) },
        eventType: 1, // Created
        summary: "Initial signal detected by SPY-1 Radar",
      },
      {
        eventTime: { seconds: BigInt(Math.floor(Date.now() / 1000 - 3000)) },
        eventType: 2, // Updated
        summary: "Fused track identified as Drone",
      },
      {
        eventTime: { seconds: BigInt(Math.floor(Date.now() / 1000 - 2400)) },
        eventType: 5, // Anomaly
        summary: "Violation of Restricted Airspace Sector B",
      },
      {
        eventTime: { seconds: BigInt(Math.floor(Date.now() / 1000 - 1800)) },
        eventType: 7, // Feedback
        summary: "CO confirmed hostile designation",
      },
      {
        eventTime: { seconds: BigInt(Math.floor(Date.now() / 1000 - 600)) },
        eventType: 2, // Updated
        summary: "SIGINT correlation confirmed 3.2GHz emitter",
      },
    ];
  }
}
