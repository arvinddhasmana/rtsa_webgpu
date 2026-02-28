// CLASSIFICATION: UNCLASSIFIED
// src/hooks/useSensorCoverage.ts

import { createPromiseClient } from "@connectrpc/connect";
import { useEffect, useState } from "react";
import { IngestionService } from "../../../gen/ts/rtsa/ingestion/v1/ingestion_service_connect";
import { ListSensorStatusesRequest, SensorStatusResponse } from "../../../gen/ts/rtsa/ingestion/v1/ingestion_service_pb";
import { transport } from "../api/grpc-client";

export function useSensorCoverage(): SensorStatusResponse[] {
  const [sensors, setSensors] = useState<SensorStatusResponse[]>([]);

  useEffect(() => {
    let isMounted = true;
    const client = createPromiseClient(IngestionService, transport);

    const fetchCoverage = async () => {
      try {
        const res = await client.listSensorStatuses(new ListSensorStatusesRequest());
        if (isMounted) {
          setSensors(res.sensors);
        }
      } catch (err) {
        console.error("Failed to fetch sensor coverage:", err);
      }
    };

    fetchCoverage();

    // Refresh coverage every 30 seconds
    const interval = setInterval(fetchCoverage, 30000);
    return () => {
      isMounted = false;
      clearInterval(interval);
    };
  }, []);

  return sensors;
}
