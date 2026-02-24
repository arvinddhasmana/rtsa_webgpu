<!-- CLASSIFICATION: UNCLASSIFIED -->
# svc-ew-ingestion

EW/SIGINT Sensor Ingestion Service.

## Configuration

| Variable | Default | Description |
|---|---|---|
| GRPC_PORT | 50052 | gRPC listen port |
| REDPANDA_BROKERS | localhost:9092 | Redpanda broker addresses |
| OUTPUT_TOPIC | sensors.ew.intercepts | Output topic |
| DLQ_TOPIC | dlq.sensors.ew | Dead-letter queue topic |
| MAX_CLASSIFICATION | CLASSIFICATION_LEVEL_SECRET | Maximum classification ceiling |
| LOG_LEVEL | info | Log level |

## Running

```bash
go run ./cmd/ew-ingestion
```
