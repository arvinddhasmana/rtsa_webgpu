// CLASSIFICATION: UNCLASSIFIED
// src/hooks/useSensorCoverage.ts
//
// TODO: IngestionService.ListSensorStatuses is not yet routed through Envoy
// (no /rtsa.ingestion.v1.IngestionService/ route entry). Until that route is
// added the RPC returns HTTP 404, which the browser logs at the network level
// regardless of try/catch. Stub returns empty until the Envoy route is wired.

import { SensorStatusResponse } from "../../../gen/ts/rtsa/ingestion/v1/ingestion_service_pb";

// eslint-disable-next-line @typescript-eslint/no-unused-vars
export function useSensorCoverage(): SensorStatusResponse[] {
  return [];
}
