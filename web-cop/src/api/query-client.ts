// CLASSIFICATION: UNCLASSIFIED
// src/api/query-client.ts

import { Timestamp } from "@bufbuild/protobuf";
import { createPromiseClient } from "@connectrpc/connect";
import {
  BoundingBox,
  ClassificationLevel as GenClassificationLevel,
  EntityType as GenEntityType,
  HostileClassification as GenHostileClassification,
  SensorType as GenSensorType,
  TrackStatus as GenTrackStatus,
  PaginationRequest,
  TimeRange,
} from "../../../gen/ts/rtsa/common/v1/types_pb";
import { QueryService } from "../../../gen/ts/rtsa/query/v1/query_service_connect";
import { QueryTracksRequest } from "../../../gen/ts/rtsa/query/v1/query_service_pb";
import { AnomalyAlert } from "../types/alert";
import {
  ClassificationLevel,
  EntityType,
  HostileClassification,
  TrackStatus,
} from "../types/common";
import { FusedTrack } from "../types/track";
import { transport } from "./grpc-client";

export interface HistoricalQueryRequest {
  startTime: Date;
  endTime: Date;
  entityTypes?: string[];
  anomalyTypes?: string[];
  minSeverity?: string;
  boundingBox?: {
    minLat: number;
    maxLat: number;
    minLon: number;
    maxLon: number;
  };
  pageSize?: number;
  pageToken?: string;
}

export interface HistoricalQueryResponse {
  tracks: FusedTrack[];
  alerts: AnomalyAlert[];
  nextPageToken?: string;
  totalCount: number;
}

function cleanEnum(val: string, prefix: string): string {
  if (!val) return val;
  return val.replace(prefix, "");
}

/**
 * QueryClient wraps the QueryService gRPC-Web unary endpoints.
 * Production: connects to svc-query via Envoy.
 */
export class QueryClient {
  private client = createPromiseClient(QueryService, transport);

  async queryHistory(
    req: HistoricalQueryRequest,
  ): Promise<HistoricalQueryResponse> {
    try {
      // Query tracks
      const trackReq = new QueryTracksRequest({
        timeRange: new TimeRange({
          startTime: Timestamp.fromDate(req.startTime),
          endTime: Timestamp.fromDate(req.endTime),
        }),
        pagination: new PaginationRequest({
          pageSize: req.pageSize || 1000,
          pageToken: req.pageToken || "",
        }),
      });

      // TODO: Handle entity types enum mapping properly if needed in future

      if (req.boundingBox) {
        trackReq.boundingBox = new BoundingBox({
          minLatitude: req.boundingBox.minLat,
          maxLatitude: req.boundingBox.maxLat,
          minLongitude: req.boundingBox.minLon,
          maxLongitude: req.boundingBox.maxLon,
        });
      }

      const trackRes = await this.client.queryTracks(trackReq);

      // Map tracks
      const tracks: FusedTrack[] = trackRes.tracks.map((t) => {
        const pos = t.estimatedPosition;
        return {
          trackId: t.trackId,
          entityType: (cleanEnum(GenEntityType[t.entityType], "ENTITY_TYPE_") ||
            "UNKNOWN") as EntityType,
          hostileClass: (cleanEnum(
            GenHostileClassification[t.hostileClass],
            "HOSTILE_CLASSIFICATION_",
          ) || "UNKNOWN") as HostileClassification,
          position: {
            latitude: pos?.latitude || 0,
            longitude: pos?.longitude || 0,
            altitudeMeters: pos?.altitudeMeters,
            speedKnots: pos?.speedKnots,
            headingDegrees: pos?.headingDegrees,
          },
          status: (cleanEnum(GenTrackStatus[t.status], "TRACK_STATUS_") ||
            "ACTIVE") as TrackStatus,
          confidenceScore: t.confidenceScore,
          classification: (cleanEnum(
            GenClassificationLevel[t.classification],
            "CLASSIFICATION_LEVEL_",
          ) || "UNCLASSIFIED") as ClassificationLevel,
          sourceCount: t.sources.length,
          sources: t.sources.map((s) => ({
            sensorId: s.sensorId,
            sensorType: cleanEnum(GenSensorType[s.sensorType], "SENSOR_TYPE_"),
            confidence: s.confidence,
            lastContribution: s.lastContribution
              ? s.lastContribution.toDate()
              : new Date(),
          })),
          createdAt: t.createdAt ? t.createdAt.toDate() : new Date(),
          updatedAt: t.updatedAt ? t.updatedAt.toDate() : new Date(),
        };
      });

      // TODO: Implement anomaly query similarly if needed
      // Currently just return tracks

      return {
        tracks,
        alerts: [],
        totalCount: trackRes.pagination
          ? trackRes.pagination.totalCount
          : tracks.length,
        nextPageToken: trackRes.pagination?.nextPageToken,
      };
    } catch (err: any) {
      console.error("Query failed", err);
      throw err;
    }
  }
}

export const queryClient = new QueryClient();
