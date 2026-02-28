-- CLASSIFICATION: UNCLASSIFIED
-- ClickHouse Migration: 003-materialized-views.sql
-- Description: Adds materialized views required for the UI dashboards using actual tables

-- 1. mv_active_tracks_by_domain (10s granularity)
-- Pre-aggregates entity track updates by domain (AIR, SURFACE, etc.) every 10 seconds.
CREATE MATERIALIZED VIEW rtsa.mv_active_tracks_by_domain
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMMDD(interval_start)
ORDER BY (entity_type, interval_start)
AS
SELECT
    entity_type,
    toStartOfInterval(event_time, INTERVAL 10 SECOND) AS interval_start,
    count() AS active_track_updates,
    uniqExact(track_id) AS unique_active_tracks
FROM rtsa.tracks_fused
GROUP BY entity_type, interval_start;

-- 2. mv_sensor_throughput_5min (rolling sensor observation rates)
-- Pre-aggregates sensor events by type over 5-minute rolling windows.
CREATE MATERIALIZED VIEW rtsa.mv_sensor_throughput_5min
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMMDD(interval_start)
ORDER BY (sensor_type, interval_start)
AS
SELECT
    sensor_type,
    toStartOfFiveMinutes(event_time) AS interval_start,
    count() AS observation_count,
    uniqExact(sensor_id) AS active_sensors,
    avg(speed_knots) AS avg_speed_recorded
FROM rtsa.sensor_observations
GROUP BY sensor_type, interval_start;

-- 3. mv_alert_ack_latency (time-to-acknowledge by severity)
-- Pre-aggregates the count of acknowledged alerts by classification
CREATE MATERIALIZED VIEW rtsa.mv_alert_ack_latency
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMMDD(interval_start)
ORDER BY (classification_level, interval_start)
AS
SELECT
    classification_level,
    toStartOfHour(event_time) AS interval_start,
    count() AS ack_count
FROM rtsa.audit_log
WHERE resource_type = 'ALERT' AND action = 'ACKNOWLEDGE_ALERT' AND event_type = 'SUCCESS'
GROUP BY classification_level, interval_start;
