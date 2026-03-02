# ClickHouse Guidelines

> **CLASSIFICATION: UNCLASSIFIED**
> **Document Type**: Technology-Specific Standard
> **Parent**: `00_master_policy.md`
> **Last Updated**: 2026-03-02

---

## 1. Purpose

ClickHouse is a columnar OLAP database designed for real-time analytical queries over large volumes of time-series and event data. This document defines best practices for schema design, query patterns, engine selection, partitioning, data retention, and operational management when using ClickHouse as an analytical data store.

## 2. Architecture Role

ClickHouse is best suited as the analytical and historical query layer in an event-driven architecture. It is not a transactional database and should not be used for OLTP workloads.

**Typical data flow:**
- Event streaming platform (e.g., Redpanda/Kafka) → ETL/CDC connector → ClickHouse
- Application services query ClickHouse for historical analysis, forensics, and aggregations
- Dashboards and monitoring tools connect to ClickHouse for visualization

**Best practices:**
- Use ClickHouse for read-heavy analytical workloads, not write-heavy transactional workloads
- Batch inserts for optimal performance — avoid single-row inserts
- Use a connector or ETL pipeline for streaming data into ClickHouse — do not write directly from high-throughput application services

## 3. Schema Design

### 3.1 Naming Conventions

| Element | Convention | Example |
|---|---|---|
| Databases | `snake_case`, short, descriptive | `analytics`, `audit` |
| Tables | `snake_case`, plural nouns | `sensor_events`, `audit_logs` |
| Columns | `snake_case` | `event_time`, `sensor_type` |
| Materialized Views | `mv_<purpose>` | `mv_hourly_counts` |
| Dictionaries | `dict_<entity>` | `dict_sensor_metadata` |

### 3.2 Column Type Selection

| Data Characteristic | Recommended Type | Rationale |
|---|---|---|
| Timestamps | `DateTime64(3, 'UTC')` | Millisecond precision; always store in UTC |
| Low-cardinality strings | `LowCardinality(String)` or `Enum8`/`Enum16` | Dramatically reduce storage and query time |
| UUIDs / identifiers | `String` or `UUID` | Use `UUID` type when UUIDs are the only format |
| Numeric IDs | `UInt32` / `UInt64` | Prefer unsigned integers for IDs and counts |
| Floating point | `Float64` | Use for coordinates, scores, measurements |
| Boolean flags | `UInt8` | ClickHouse has no native `BOOLEAN`; use `0`/`1` |
| Nested structures | `Nested` or `Array(Tuple(...))` | Avoid deeply nested structures — flatten where possible |
| JSON-like data | `String` with JSON functions | Use `JSONExtract*` functions for occasional JSON parsing |

**Rules:**
- Always specify timezone on `DateTime` columns — use `'UTC'` unless there is a documented reason not to
- Prefer `Enum8` or `LowCardinality(String)` for columns with < 256 distinct values
- Avoid `Nullable` types unless null values have genuine semantic meaning — they add overhead
- Use `DEFAULT` expressions for computed or auto-populated columns

### 3.3 Materialized Views

Materialized views in ClickHouse are triggers that transform data during insertion — they are not cached queries:

**Best practices:**
- Use `SummingMergeTree` or `AggregatingMergeTree` engines for pre-aggregated materialized views
- Design materialized views for your most frequent query patterns (e.g., hourly/daily rollups)
- Materialized views execute during insert time — keep transformations lightweight
- Partition materialized views at a coarser granularity than their source tables (e.g., monthly for hourly aggregates)
- Name materialized views with the `mv_` prefix to distinguish them from base tables
- Always define an explicit `ORDER BY` on the materialized view's target table

## 4. MergeTree Engine Family

### 4.1 Engine Selection Guide

| Engine | Use Case | Key Behavior |
|---|---|---|
| `MergeTree` | Default for most tables | No deduplication; append-only |
| `ReplacingMergeTree` | Latest-state tables (upserts) | Deduplicates by sort key on merge; keeps latest version |
| `SummingMergeTree` | Pre-aggregated counters | Sums numeric columns with same sort key on merge |
| `AggregatingMergeTree` | Complex pre-aggregations | Stores intermediate aggregation states |
| `CollapsingMergeTree` | Row-level updates/deletes via sign column | Use `sign` column (+1 insert, -1 cancel) |
| `VersionedCollapsingMergeTree` | Same as above, with out-of-order data | Adds `version` column for ordering |

**Rules:**
- Use `MergeTree` as the default unless you have a specific need for deduplication or aggregation
- Use `ReplacingMergeTree` when you need "latest state" semantics (e.g., current entity position)
- Remember: deduplication in `ReplacingMergeTree` only occurs during background merges — query with `FINAL` for guaranteed deduplication
- Use `SummingMergeTree` for counter/metric tables where you only need totals

### 4.2 ORDER BY Key Design

The `ORDER BY` clause determines both the sort order on disk and the primary index. This is the single most important schema decision:

**Best practices:**
- Lead with columns used in `WHERE` filters (most frequently filtered first)
- Place high-cardinality columns after low-cardinality columns
- Include a timestamp column for time-range queries
- Keep the primary key (ORDER BY) as short as possible — each additional column increases memory usage
- For time-series data, a common pattern: `ORDER BY (category, entity_id, timestamp)`

```sql
-- GOOD: low-cardinality first, then entity, then time
ORDER BY (sensor_type, sensor_id, event_time)

-- BAD: high-cardinality field first — poor index selectivity
ORDER BY (event_id, event_time)
```

## 5. Partitioning Strategy

### 5.1 Partition Expression Best Practices

Partitioning controls how data is physically organized on disk. Each partition becomes a directory containing data parts.

| Guideline | Explanation |
|---|---|
| Partition by time | Use `toYYYYMM(timestamp)` for monthly or `toYYYYMMDD(timestamp)` for daily |
| Target 50–300 active partitions | Too many partitions degrade merge performance; too few reduce partition pruning benefits |
| Daily partitions for high-volume data | Event streams with > 1M rows/day benefit from daily partitioning |
| Monthly partitions for moderate volume | Tables with < 1M rows/day typically work well with monthly partitions |
| Never partition by high-cardinality columns | Partitioning by user_id or event_id creates millions of tiny partitions |

### 5.2 Partition Pruning

Queries that filter on the partition expression can skip entire partitions, dramatically improving performance:

```sql
-- GOOD — partition pruning applies (partition is toYYYYMM(event_time))
SELECT count() FROM events
WHERE event_time >= '2026-01-01' AND event_time < '2026-02-01'

-- BAD — no partition pruning (filter on non-partition column)
SELECT count() FROM events
WHERE sensor_id = 'abc'
```

### 5.3 Merge Considerations

- ClickHouse continuously merges data parts within each partition
- Very large partitions slow down merges and increase write latency
- Very small partitions create overhead from too many parts
- Monitor the `system.parts` table for partition health and merge activity

## 6. Data Retention (TTL)

### 6.1 TTL Best Practices

ClickHouse supports automatic data expiration via TTL expressions:

```sql
-- Basic TTL: delete data after 90 days
CREATE TABLE events (
    event_time DateTime64(3, 'UTC'),
    ...
) ENGINE = MergeTree()
PARTITION BY toYYYYMMDD(event_time)
ORDER BY (category, event_time)
TTL event_time + INTERVAL 90 DAY;
```

**Best practices:**
- Always define TTL on time-series tables to prevent unbounded storage growth
- Use TTL expressions based on the partition column for efficient expiration (ClickHouse drops entire parts)
- Consider multi-tier TTL for "hot/warm/cold" storage strategies:

```sql
-- Multi-tier TTL: move to warm storage after 30 days, delete after 365 days
TTL event_time + INTERVAL 30 DAY TO VOLUME 'warm',
    event_time + INTERVAL 365 DAY DELETE
```

### 6.2 Environment-Specific Retention

Adjust TTL values based on deployment environment:
- **Data centre**: Longer retention (months to years) — ample storage
- **Edge/constrained**: Short retention (hours to days) — limited disk; sync to central before expiry
- Use `ALTER TABLE ... MODIFY TTL` to adjust retention without rebuilding tables

## 7. Query Patterns

### 7.1 Parameterized Queries Only

**Always** use parameterized queries — never concatenate user input into SQL strings:

```go
// GOOD — parameterized query
rows, err := ch.Query(ctx,
    `SELECT entity_id, event_time, score
     FROM events
     WHERE category = @category
       AND event_time BETWEEN @startTime AND @endTime
     ORDER BY event_time DESC
     LIMIT @limit`,
    clickhouse.Named("category", category),
    clickhouse.Named("startTime", startTime),
    clickhouse.Named("endTime", endTime),
    clickhouse.Named("limit", maxRows),
)

// BAD — SQL injection vulnerability
query := fmt.Sprintf("SELECT * FROM events WHERE category = '%s'", category)
```

### 7.2 Query Guardrails

Set query limits to prevent runaway queries from impacting the cluster:

| Guardrail | Recommended Value | Purpose |
|---|---|---|
| `max_execution_time` | 30 seconds | Prevent long-running queries |
| `max_rows_to_read` | 10,000,000 | Prevent full table scans |
| `max_result_rows` | 100,000 | Limit result set size |
| `max_memory_usage` | 1 GB per query | Prevent OOM |
| `max_threads` | Cluster-appropriate | Limit CPU usage per query |

Apply guardrails per user profile or per query:

```sql
SET max_execution_time = 30;
SET max_rows_to_read = 10000000;
```

### 7.3 Query Optimization Tips

- Filter on `ORDER BY` columns first — ClickHouse uses the primary index for these
- Filter on partition key for partition pruning
- Avoid `SELECT *` — specify only the columns you need (columnar storage benefits)
- Use `PREWHERE` for light filters before heavy column reads
- Use `LIMIT` to cap result sets
- Use `FINAL` with `ReplacingMergeTree` only when you need guaranteed deduplication — it has a performance cost

## 8. Data Skipping Indices

Use data skipping indices for columns not in the primary key but frequently used in filters:

```sql
-- Bloom filter index for string column lookups
ALTER TABLE events
  ADD INDEX idx_entity_id entity_id TYPE bloom_filter GRANULARITY 4;

-- Min-max index for numeric range queries
ALTER TABLE events
  ADD INDEX idx_score score TYPE minmax GRANULARITY 4;

-- Set index for low-cardinality column filtering
ALTER TABLE events
  ADD INDEX idx_status status TYPE set(100) GRANULARITY 4;
```

| Index Type | Best For | Notes |
|---|---|---|
| `bloom_filter` | Equality checks on strings/IDs | Probabilistic; false positives possible |
| `minmax` | Range queries on numeric/date columns | Eliminates granules where range doesn't overlap |
| `set(N)` | Low-cardinality columns (< N distinct values) | Stores complete set of values per granule |
| `ngrambf_v1` | Full-text search (`LIKE '%pattern%'`) | Limited use; consider external search engine for heavy FTS |

## 9. Codec and Compression

### 9.1 Compression Best Practices

ClickHouse supports per-column compression codecs:

| Codec | Best For | Compression Ratio |
|---|---|---|
| `LZ4` (default) | General purpose; fast decompression | Good |
| `ZSTD` | Higher compression ratio; slower | Better |
| `Delta` + `ZSTD` | Monotonically increasing values (timestamps, IDs) | Excellent |
| `DoubleDelta` + `ZSTD` | Timestamps with regular intervals | Excellent |
| `T64` | Integer columns with small value ranges | Good |

```sql
CREATE TABLE events (
    event_time DateTime64(3, 'UTC') CODEC(DoubleDelta, ZSTD),
    sensor_id String CODEC(ZSTD),
    score Float64 CODEC(Delta, ZSTD),
    raw_data String CODEC(ZSTD(3))  -- Higher ZSTD level for rarely-read columns
) ENGINE = MergeTree()
...
```

### 9.2 Guidelines

- Use `DoubleDelta` + `ZSTD` for timestamp columns (typically 90%+ compression)
- Use `Delta` + `ZSTD` for monotonically increasing numeric columns
- Use `ZSTD` with higher levels (3–5) for large text/blob columns that are rarely read
- Test compression ratios with `OPTIMIZE TABLE ... FINAL` and check `system.parts` for size data

## 10. Operational Best Practices

### 10.1 Monitoring

Key metrics to monitor:
- `system.merges` — Active merge operations
- `system.parts` — Number of active parts per partition (should stay low; > 300 indicates issues)
- `system.query_log` — Query performance and errors
- `system.replicas` — Replication lag (in replicated setups)
- `system.disks` — Disk usage and free space

### 10.2 Backup and Restore

- Use `BACKUP TABLE ... TO ...` (ClickHouse native backup) for regular backups
- For large datasets, use incremental backups with `FROM PREVIOUS`
- Test restore procedures periodically — untested backups are not backups
- In replicated setups, backup from a replica to avoid impacting the primary

### 10.3 Resource Configuration by Environment

| Setting | Data Centre | Edge / Constrained |
|---|---|---|
| `max_server_memory_usage_to_ram_ratio` | 0.8 | 0.5 |
| `max_concurrent_queries` | 100 | 10 |
| `max_connections` | 500 | 50 |
| `mark_cache_size` | 1 GB | 256 MB |
| `max_partition_size_to_drop` | 50 GB | 1 GB |

## 11. AI Agent Instructions

When generating ClickHouse-related code:

1. ALWAYS use parameterized queries with `clickhouse.Named()` — NEVER concatenate
2. Include query guardrails (`max_execution_time`, `max_rows_to_read`)
3. Choose the appropriate MergeTree engine variant (Section 4) based on the use case
4. Design `ORDER BY` keys with low-cardinality columns first, then entity ID, then timestamp
5. Partition by time using `toYYYYMM()` (monthly) or `toYYYYMMDD()` (daily) based on data volume
6. Include TTL for data retention management — adjust by deployment environment
7. Use `Enum8` or `LowCardinality(String)` for low-cardinality columns
8. Use `DateTime64(3, 'UTC')` for all timestamp columns (millisecond precision, UTC)
9. Create materialized views for frequently-queried aggregations
10. Apply appropriate compression codecs per column type (Section 9)
