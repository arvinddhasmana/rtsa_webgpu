<!-- CLASSIFICATION: UNCLASSIFIED -->
# svc-isr-ingestion

ISR (Intelligence, Surveillance, and Reconnaissance) Sensor Ingestion Service.

## Configuration

| Variable | Default | Description |
|---|---|---|
| GRPC_PORT | 50054 | gRPC listen port |
| REDPANDA_BROKERS | localhost:9092 | Redpanda broker addresses |
| OUTPUT_TOPIC | sensors.isr.observations | Output topic |
| DLQ_TOPIC | dlq.sensors.isr | Dead-letter queue topic |
| MAX_CLASSIFICATION | CLASSIFICATION_LEVEL_SECRET | Maximum classification ceiling |
| LOG_LEVEL | info | Log level |

## Running

```bash
go run ./cmd/isr-ingestion
```
