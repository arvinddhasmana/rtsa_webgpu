// CLASSIFICATION: UNCLASSIFIED
// src/api/query-client.ts

import { FusedTrack } from "../types/track";
import { AnomalyAlert } from "../types/alert";

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

/**
 * QueryClient wraps the QueryService gRPC-Web unary endpoints.
 * Production: connects to svc-query via Envoy.
 */
export class QueryClient {
  async queryHistory(
    req: HistoricalQueryRequest
  ): Promise<HistoricalQueryResponse> {
    void req;
    return { tracks: [], alerts: [], totalCount: 0 };
  }
}

export const queryClient = new QueryClient();
