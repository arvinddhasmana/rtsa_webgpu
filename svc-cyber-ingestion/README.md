<!-- CLASSIFICATION: UNCLASSIFIED -->
# svc-cyber-ingestion

Cyber IOC (Indicator of Compromise) Ingestion Service.

## Configuration

| Variable | Default | Description |
|---|---|---|
| GRPC_PORT | 50056 | gRPC listen port |
| REDPANDA_BROKERS | localhost:9092 | Redpanda broker addresses |
| OUTPUT_TOPIC | sensors.cyber.iocs | Output topic |
| DLQ_TOPIC | dlq.sensors.cyber | Dead-letter queue topic |
| MAX_CLASSIFICATION | CLASSIFICATION_LEVEL_SECRET | Maximum classification ceiling |
| LOG_LEVEL | info | Log level |

## Running

```bash
go run ./cmd/cyber-ingestion
```
