<!-- CLASSIFICATION: UNCLASSIFIED -->
# svc-ais-ingestion

AIS/BFT (Automatic Identification System / Blue Force Tracking) Sensor Ingestion Service.

## Configuration

| Variable | Default | Description |
|---|---|---|
| GRPC_PORT | 50055 | gRPC listen port |
| REDPANDA_BROKERS | localhost:9092 | Redpanda broker addresses |
| OUTPUT_TOPIC | sensors.ais.positions | Output topic |
| DLQ_TOPIC | dlq.sensors.ais | Dead-letter queue topic |
| MAX_CLASSIFICATION | CLASSIFICATION_LEVEL_SECRET | Maximum classification ceiling |
| LOG_LEVEL | info | Log level |

## Running

```bash
go run ./cmd/ais-ingestion
```
