<!-- CLASSIFICATION: UNCLASSIFIED -->
# svc-elint-ingestion

ELINT/COMINT Sensor Ingestion Service.

## Configuration

| Variable | Default | Description |
|---|---|---|
| GRPC_PORT | 50053 | gRPC listen port |
| REDPANDA_BROKERS | localhost:9092 | Redpanda broker addresses |
| OUTPUT_TOPIC | sensors.elint.detections | Output topic |
| DLQ_TOPIC | dlq.sensors.elint | Dead-letter queue topic |
| MAX_CLASSIFICATION | CLASSIFICATION_LEVEL_SECRET | Maximum classification ceiling |
| LOG_LEVEL | info | Log level |

## Running

```bash
go run ./cmd/elint-ingestion
```
