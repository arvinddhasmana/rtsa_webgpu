-- CLASSIFICATION: UNCLASSIFIED
-- ClickHouse Migration: 001_initial_schema.sql
-- Creates the core RTSA tables used by ingestion, fusion, audit, and queries.

CREATE DATABASE IF NOT EXISTS rtsa;

CREATE TABLE IF NOT EXISTS rtsa.sensor_observations
(
    observation_id       String,
    sensor_id            String,
    sensor_type          LowCardinality(String),
    track_id             String DEFAULT '',
    latitude             Float64,
    longitude            Float64,
    altitude_meters      Float64,
    speed_knots          Float64,
    heading_degrees      Float64,
    confidence           Float32 DEFAULT 0,
    classification_level LowCardinality(String),
    metadata_json        String,
    event_time           DateTime64(3, 'UTC'),
    ingestion_time       DateTime64(3, 'UTC') DEFAULT now64(3)
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(event_time)
ORDER BY (sensor_type, sensor_id, event_time)
TTL toDateTime(event_time) + INTERVAL 90 DAY
SETTINGS index_granularity = 8192;

CREATE TABLE IF NOT EXISTS rtsa.tracks_fused
(
    track_id              String,
    entity_type           LowCardinality(String),
    hostile_classification LowCardinality(String),
    label                 String DEFAULT '',
    latitude              Float64,
    longitude             Float64,
    altitude_meters       Float64,
    speed_knots           Float64,
    heading_degrees       Float64,
    confidence_score      Float32,
    source_count          UInt16,
    source_sensors        Array(String),
    classification_level  LowCardinality(String),
    track_status          LowCardinality(String),
    event_time            DateTime64(3, 'UTC'),
    ingestion_time        DateTime64(3, 'UTC') DEFAULT now64(3)
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(event_time)
ORDER BY (entity_type, track_id, event_time)
TTL toDateTime(event_time) + INTERVAL 90 DAY
SETTINGS index_granularity = 8192;

CREATE TABLE IF NOT EXISTS rtsa.anomaly_detections
(
    alert_id             String,
    track_id             String,
    anomaly_type         LowCardinality(String),
    severity             LowCardinality(String),
    confidence_score     Float32,
    explanation          String,
    model_version        String DEFAULT '',
    alert_status         LowCardinality(String) DEFAULT 'ACTIVE',
    classification_level LowCardinality(String),
    event_time           DateTime64(3, 'UTC'),
    ingestion_time       DateTime64(3, 'UTC') DEFAULT now64(3)
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(event_time)
ORDER BY (track_id, event_time)
TTL toDateTime(event_time) + INTERVAL 90 DAY
SETTINGS index_granularity = 8192;

CREATE TABLE IF NOT EXISTS rtsa.audit_log
(
    audit_id             String,
    service_id           LowCardinality(String),
    event_type           LowCardinality(String),
    actor_id             String,
    actor_type           LowCardinality(String),
    resource_type        LowCardinality(String),
    resource_id          String,
    action               LowCardinality(String),
    detail_json          String,
    classification_level LowCardinality(String),
    event_time           DateTime64(3, 'UTC'),
    ingestion_time       DateTime64(3, 'UTC') DEFAULT now64(3)
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(event_time)
ORDER BY (service_id, event_type, event_time)
SETTINGS index_granularity = 8192;

CREATE TABLE IF NOT EXISTS rtsa.operator_feedback
(
    feedback_id          String,
    track_id             String,
    operator_id          String,
    alert_id             String DEFAULT '',
    feedback_type        LowCardinality(String),
    justification        String,
    trust_score          Float32,
    clearance_score      Float32,
    accuracy_score       Float32,
    temporal_score       Float32,
    deviation_score      Float32,
    validated            UInt8 DEFAULT 0,
    classification_level LowCardinality(String),
    event_time           DateTime64(3, 'UTC'),
    ingestion_time       DateTime64(3, 'UTC') DEFAULT now64(3)
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(event_time)
ORDER BY (operator_id, event_time)
SETTINGS index_granularity = 8192;
