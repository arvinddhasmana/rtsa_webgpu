// CLASSIFICATION: UNCLASSIFIED
// src/types/track.ts

import {
  ClassificationLevel,
  EntityType,
  HostileClassification,
  TrackStatus,
} from "./common";

export interface Position {
  latitude: number;
  longitude: number;
  altitudeMeters?: number;
  speedKnots?: number;
  headingDegrees?: number;
}

export interface SourceAttribution {
  sensorId: string;
  sensorType: string;
  confidence: number;
  lastContribution: Date;
}

export interface FusedTrack {
  trackId: string;
  entityType: EntityType;
  hostileClass: HostileClassification;
  position: Position;
  confidenceScore: number;
  sourceCount: number;
  sources: SourceAttribution[];
  status: TrackStatus;
  classification: ClassificationLevel;
  createdAt: Date;
  updatedAt: Date;
}
