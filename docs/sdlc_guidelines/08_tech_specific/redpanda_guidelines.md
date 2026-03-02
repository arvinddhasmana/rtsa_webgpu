# Redpanda Guidelines

> **CLASSIFICATION: UNCLASSIFIED**
> **Document Type**: Technology-Specific Standard
> **Parent**: `00_master_policy.md`
> **Last Updated**: 2026-03-02

---

## 1. Purpose

Redpanda is a Kafka-compatible event streaming platform designed for low-latency, high-throughput workloads. This document defines best practices for topic design, producer/consumer patterns, schema management, operational procedures, and performance tuning when using Redpanda as the event streaming backbone.

## 2. Architecture Role

Redpanda serves as the central event streaming platform in an event-driven microservices architecture. It functions as:

- **Event log**: Immutable, ordered record of all state changes
- **Inter-service communication bus**: Decoupled publish/subscribe messaging between services
- **Audit trail**: Append-only audit event stream
- **Data pipeline source**: Streaming data to analytical stores via connectors

**Best practices:**
- Treat Redpanda topics as the source of truth for event state — services derive state from event streams
- Use Redpanda for asynchronous, event-driven communication — not for synchronous request/response patterns
- Pair Redpanda with an analytical store (e.g., ClickHouse) for historical queries — do not query Redpanda for analytics

## 3. Topic Design

### 3.1 Topic Naming Convention

Use a hierarchical, dot-separated naming convention for topics:

```
<domain>.<entity>.<qualifier>

Examples:
  sensor.raw.radar
  sensor.validated.radar
  entity.tracks.fused
  inference.anomaly.scores
  feedback.submissions
  feedback.validated
  audit.events
  interop.exchange.inbound
  interop.exchange.outbound
  <any-topic>.dlq              (dead-letter queue suffix)
```

**Rules:**
- Use lowercase, dot-separated names
- Domain-first naming enables topic-level ACLs by domain
- Use `.dlq` suffix for dead-letter queue topics
- Avoid overly specific names that couple topics to implementation details

### 3.2 Topic Configuration Best Practices

#### Partition Count

| Guideline | Explanation |
|---|---|
| Start with 6–12 partitions | Sufficient for most workloads; can increase later |
| Match partitions to expected consumer concurrency | Each consumer in a group reads from one or more partitions |
| Higher partitions for high-throughput topics | 16–64 for topics with > 10K msg/sec |
| Lower partitions for low-throughput topics | 1–3 for control/admin topics |
| Partitions can be increased but never decreased | Plan for growth but do not over-provision initially |

**Caution:** Increasing partition count rebalances consumers and may temporarily break key-based ordering guarantees for new partitions.

#### Retention Policy

| Retention Strategy | When to Use |
|---|---|
| Time-based (`retention.ms`) | Default — retain messages for a fixed duration (24h, 7d, 30d) |
| Size-based (`retention.bytes`) | Constrained environments — limit total storage per topic |
| Compacted (`cleanup.policy=compact`) | Latest-state topics — retain only the last value per key |
| Compact + Delete | Retain latest state AND expire old tombstones |

**Best practices:**
- Set retention based on the consumer's processing SLA — retain long enough for all consumers to process
- Use shorter retention on edge/constrained deployments
- Use compaction for topics that represent current state (e.g., entity positions, configuration)

#### Replication

| Environment | Replication Factor | Min In-Sync Replicas |
|---|---|---|
| Production (3+ brokers) | 3 | 2 |
| Staging (3 brokers) | 3 | 2 |
| Single-node / Edge | 1 | 1 |

**Rules:**
- Never set replication factor higher than the number of brokers
- Set `min.insync.replicas` to `replication.factor - 1` for durability without blocking writes when one broker is down
- For critical data (audit, transactions), use `acks=all` with `min.insync.replicas >= 2`

### 3.3 Partition Key Selection

Choose partition keys that balance ordering requirements with parallelism:

| Consideration | Guideline |
|---|---|
| **Ordering** | Messages with the same key go to the same partition — maintaining order |
| **Parallelism** | More distinct keys → more even distribution across partitions |
| **Hot partitions** | Avoid keys with skewed distribution (e.g., a single high-volume entity) |
| **Compound keys** | Use `<entity_type>:<entity_id>` for finer-grained ordering with better distribution |

**Pattern:** For event streams, use the primary entity identifier as the partition key (e.g., `sensor_id`, `entity_id`, `operator_id`).

## 4. Producer Standards

### 4.1 Configuration Best Practices

```go
// CLASSIFICATION: UNCLASSIFIED

cfg := kgo.NewClient(
    kgo.SeedBrokers(brokers...),
    kgo.DefaultProduceTopic(topic),
    kgo.ProducerBatchMaxBytes(1 * 1024 * 1024),  // 1 MB batch
    kgo.ProducerLinger(5 * time.Millisecond),     // Batch for 5ms
    kgo.RecordPartitioner(kgo.StickyKeyPartitioner(nil)),
    kgo.RequiredAcks(kgo.AllISRAcks()),           // Wait for all ISR
    kgo.ProducerOnDataLossDetected(func(topic string, partition int32) {
        slog.Error("data loss detected",
            "topic", topic,
            "partition", partition,
        )
    }),
)
```

### 4.2 Producer Rules

1. **Always set a partition key** — ensures ordering per entity
2. **Always wait for ACK** — use `AllISRAcks()` for critical data; `LeaderAck()` for lower-latency, less-critical data
3. **Include schema version** in message headers for schema evolution
4. **Serialize with Protobuf** — prefer binary serialization over JSON for performance and type safety
5. **Handle produce errors** — retry transient errors with exponential backoff; route to DLQ on persistent failure
6. **Include correlation ID** in message headers for distributed tracing
7. **Batch messages** — configure `ProducerLinger` and `ProducerBatchMaxBytes` for throughput vs. latency trade-off

### 4.3 Message Headers

Use headers for metadata that is orthogonal to the message payload:

| Header | Required | Purpose |
|---|---|---|
| `correlation-id` | Yes | Distributed tracing across services |
| `schema-version` | Yes | Schema evolution tracking |
| `source-service` | Yes | Producer identification for debugging |
| `timestamp` | Yes | Production timestamp (ISO 8601) |
| `classification` | Conditional | Data classification level (if applicable) |

## 5. Consumer Standards

### 5.1 Configuration Best Practices

```go
// CLASSIFICATION: UNCLASSIFIED

cfg := kgo.NewClient(
    kgo.SeedBrokers(brokers...),
    kgo.ConsumerGroup(consumerGroup),
    kgo.ConsumeTopics(topics...),
    kgo.FetchMaxWait(100 * time.Millisecond),
    kgo.FetchMaxBytes(5 * 1024 * 1024),
    kgo.DisableAutoCommit(),  // Manual commit after processing
    kgo.BlockRebalanceOnPoll(),
)
```

### 5.2 Consumer Rules

1. **Manual offset commit** — commit only after successful processing to prevent data loss
2. **Idempotent processing** — handle duplicate deliveries gracefully (at-least-once is the default)
3. **Monitor consumer lag** — alert when lag is increasing for > 10 minutes
4. **Graceful shutdown** — commit offsets and leave consumer group cleanly on SIGTERM
5. **Dead-letter queue** — route unprocessable messages to `<topic>.dlq` instead of blocking the consumer
6. **Backpressure handling** — if downstream is slow, pause consumption rather than accumulating in memory
7. **Rebalance safety** — use `BlockRebalanceOnPoll()` to prevent processing during rebalance

### 5.3 Dead-Letter Queue Pattern

Route messages that cannot be processed after retries to a dead-letter queue:

```go
// CLASSIFICATION: UNCLASSIFIED

func processRecord(ctx context.Context, record *kgo.Record) error {
    err := handler.Process(ctx, record)
    if err != nil {
        if isRetryable(err) {
            return err // Will be retried
        }
        // Non-retryable: send to DLQ
        dlqRecord := &kgo.Record{
            Topic: record.Topic + ".dlq",
            Key:   record.Key,
            Value: record.Value,
            Headers: append(record.Headers,
                kgo.RecordHeader{Key: "dlq-reason", Value: []byte(err.Error())},
                kgo.RecordHeader{Key: "dlq-timestamp", Value: []byte(time.Now().Format(time.RFC3339))},
                kgo.RecordHeader{Key: "dlq-original-topic", Value: []byte(record.Topic)},
            ),
        }
        return dlqProducer.ProduceSync(ctx, dlqRecord).FirstErr()
    }
    return nil
}
```

**DLQ best practices:**
- Include the original topic, partition, offset, and error reason in DLQ headers
- Set longer retention on DLQ topics (7–30 days) than on source topics
- Implement a DLQ reprocessing tool for replaying messages after fixes
- Monitor DLQ message count as an alert metric

## 6. Schema Management

### 6.1 Serialization Best Practices

- Use Protobuf serialization for all non-human-readable topics — type-safe, compact, versionable
- Use JSON serialization only for audit/debug topics where human readability is required
- Track schema version via message headers — consumers must handle at least the current and previous version

### 6.2 Schema Evolution Rules

- Add new optional fields — old consumers ignore unknown fields
- Never remove fields — mark as reserved in Protobuf
- Never change field types or numbers — add new fields instead
- Maintain backward compatibility for at least one version (consumers upgraded before producers)
- Use a schema registry (if available) for centralized schema management

## 7. Redpanda Connect (ETL Pipelines)

### 7.1 Best Practices for Streaming ETL

Redpanda Connect (formerly Benthos) provides streaming ETL pipelines from Redpanda to downstream stores:

**Configuration best practices:**
- Use `kafka_franz` input for Redpanda consumption (franz-go based, highest performance)
- Configure batching on the output — batch size and flush interval for optimal throughput
- Use Bloblang or Protobuf processors for message transformation
- Implement error handling with DLQ routing for failed transformations
- Use consumer groups for offset management and horizontal scaling

```yaml
# CLASSIFICATION: UNCLASSIFIED
# Generic Redpanda Connect pipeline pattern

input:
  kafka_franz:
    seed_brokers: ["${BROKER_ADDRESSES}"]
    topics: ["source.topic"]
    consumer_group: "connect-pipeline"

pipeline:
  processors:
    - protobuf:
        operator: to_json
        message: my.package.v1.MyMessage
    - mapping: |
        root = this
        root.ingestion_time = now()

output:
  sql_insert:
    driver: clickhouse
    dsn: "${CLICKHOUSE_DSN}"
    table: target_table
    columns: [col1, col2, ingestion_time]
    batching:
      count: 1000
      period: "1s"
```

### 7.2 Error Handling in Pipelines

- Configure a DLQ output for messages that fail transformation or insertion
- Set retry policies with exponential backoff for transient downstream failures
- Monitor pipeline consumer lag as a health metric
- Use Redpanda Connect's built-in metrics endpoint for Prometheus scraping

## 8. Operational Best Practices

### 8.1 ACL Management

- Use topic-level ACLs to restrict which services can produce to or consume from each topic
- Follow the principle of least privilege — services should only access topics they need
- Use service identity (mTLS CN or SASL principal) as the ACL principal
- Review ACLs periodically and remove unused entries

### 8.2 Monitoring

Key metrics to monitor:

| Metric | Alert Threshold | Rationale |
|---|---|---|
| Consumer lag (per group) | Increasing for > 10 min | Consumer or downstream issue |
| Under-replicated partitions | > 0 for > 5 min | Broker health or network issue |
| Topic disk usage | > 80% of allocated | Retention or partition issues |
| Producer error rate | > 0.1% | Network, ACL, or broker issue |
| Broker CPU / memory | > 80% sustained | Capacity planning trigger |

### 8.3 Topic Lifecycle Management

| Operation | Procedure |
|---|---|
| Create a new topic | Define in IaC (Terraform, Helm, or scripts); deploy via CI/CD |
| Increase partition count | Update IaC; apply carefully (triggers consumer rebalance) |
| Delete a topic | Remove all consumers first; delete via IaC; update ACLs |
| Adjust retention | Update topic config via IaC; no consumer impact |

### 8.4 Tiered Storage

Redpanda supports tiered storage to offload older data to object storage (S3-compatible):

**Best practices:**
- Enable tiered storage for topics with long retention requirements
- Configure local retention (`retention.local.target.bytes` or `retention.local.target.ms`) for hot data
- Use S3-compatible storage for cold data
- Tiered storage enables cost-effective long retention without scaling broker disk

### 8.5 Rack Awareness

In multi-rack or multi-availability-zone deployments:
- Enable rack awareness to distribute partition replicas across failure domains
- Configure `rack_id` per broker matching the physical/logical rack or availability zone
- Ensures that a single rack failure does not lose all replicas of any partition

## 9. Client Library Selection

### 9.1 Go Client

- Use `franz-go` (`kgo`) — highest performance Kafka-compatible Go client
- Do NOT use `sarama` (legacy, lower performance) or `confluent-kafka-go` (CGo dependency)

### 9.2 TypeScript/Node.js Client

- Use `kafkajs` or `@confluentinc/kafka-javascript` for Node.js applications
- Use `franz-go` via gRPC-Web proxy for browser-based consumption (not recommended for production)

### 9.3 Python Client

- Use `confluent-kafka-python` for performance-critical consumers
- Use `kafka-python` for simpler use cases and testing

## 10. AI Agent Instructions

When generating Redpanda-related code:

1. Use `franz-go` (`kgo`) client library for Go — not sarama or confluent-kafka-go
2. Always set `AllISRAcks()` for producers handling critical data
3. Always use manual offset commit for consumers — commit after processing
4. Include required message headers (correlation-id, schema-version, source-service, timestamp)
5. Route unprocessable messages to `<topic>.dlq` with reason, timestamp, and original topic headers
6. Use Protobuf serialization — not JSON (except for audit/debug topics)
7. Set partition key based on the primary entity identifier for ordering
8. Configure appropriate retention and replication based on the deployment environment
9. Monitor consumer lag and set alerts for diverging lag
10. Use Redpanda Connect for ETL pipelines — configure batching and error handling
