// CLASSIFICATION: UNCLASSIFIED
// src/services/query.ts — QueryService gRPC calls (ClickHouse cold path)
//
// Provides typed wrappers around the generated ConnectRPC QueryService client.
// Reference: docs/implementation/v4/phase3_ui_interaction.md §5

import { createPromiseClient } from "@connectrpc/connect";
import { QueryService } from "@gen/rtsa/query/v1/query_service_connect.js";
import type { FusedTrack } from "@gen/rtsa/entity/v1/fused_track_pb.js";
import { transport } from "./grpc-client";
import type { TrackDetail } from "../signals/track";
import {
  ClassificationLevel,
  EntityType,
  HostileClassification,
  TrackStatus,
} from "@gen/rtsa/common/v1/types_pb.js";

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
  };
}

/**
 * Fetch full track detail for a given track ID from ClickHouse via gRPC.
 * Returns null if no matching track is found.
 */
export async function fetchTrackDetail(trackId: string): Promise<TrackDetail | null> {
  const response = await client.queryTracks({
    trackId,
    clearanceLevel: ClassificationLevel.UNCLASSIFIED,
    pagination: { pageSize: 1, pageToken: "" },
  });

  const track = response.tracks[0];
  if (!track) return null;
  return mapTrack(track);
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
  return client.getEventTimeline({
    trackId,
    clearanceLevel: ClassificationLevel.UNCLASSIFIED,
    maxEvents,
  });
}
