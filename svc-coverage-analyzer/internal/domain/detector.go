// CLASSIFICATION: UNCLASSIFIED
package domain

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
	inferencev1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/inference/v1"
	ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// staleDuration is kept for future Phase B multi-tier staleness detection.
// Phase A only uses offlineDuration to decide when to emit a gap alert.
const staleDuration = 30 * time.Second

// offlineDuration is the threshold after which a sensor is considered OFFLINE.
const offlineDuration = 120 * time.Second

// sensorCoverageRecord holds the last known coverage geometry and heartbeat for a sensor.
type sensorCoverageRecord struct {
	sensorID     string
	rangeNm      float64
	bearingStart float64
	bearingEnd   float64
	centerLat    float64
	centerLon    float64
	lastSeen     time.Time
}

// GapDetector maintains in-memory sensor coverage state and emits SpatialAlerts
// when sensor outages create uncovered geographic sectors.
//
// Phase A heuristic: when a sensor transitions to OFFLINE/STALE, we check
// whether any configured monitored sector relies solely on that sensor.
// Full computational geometry (polygon union/difference) is Phase C.
type GapDetector struct {
	logger   *zap.Logger
	mu       sync.Mutex
	coverage map[string]*sensorCoverageRecord // sensorID → last known coverage
}

// NewGapDetector creates a GapDetector with an empty coverage state.
func NewGapDetector(logger *zap.Logger) *GapDetector {
	return &GapDetector{
		logger:   logger,
		coverage: make(map[string]*sensorCoverageRecord),
	}
}

// Analyze processes an incoming observation, updates the coverage state, and
// returns a SpatialAlert if a gap is detected.  Returns nil, nil when no gap
// is warranted.
func (d *GapDetector) Analyze(ctx context.Context, obs *ingestionv1.SensorObservation) (*inferencev1.SpatialAlert, error) {
	meta := obs.GetMetadata()
	if meta == nil {
		return nil, nil
	}

	sensorID := obs.GetSensorId()
	now := time.Now()

	// ── Simulated gap injection (E2E / demo mode) ────────────────────────────
	if gapInject, ok := meta["rtsa.sim_inject_gap"]; ok && gapInject == "true" {
		d.logger.Info("simulated coverage gap detected", zap.String("sensor_id", sensorID))
		return d.buildGapAlert(sensorID, "SIM-SECTOR", now), nil
	}

	// ── Update in-memory coverage state ─────────────────────────────────────
	rec := d.updateCoverageRecord(sensorID, meta, now)
	if rec == nil {
		// No coverage metadata in this observation — nothing to analyse.
		return nil, nil
	}

	// ── Heuristic gap detection: check all known sensors for staleness ───────
	d.mu.Lock()
	defer d.mu.Unlock()

	for id, r := range d.coverage {
		if id == sensorID {
			// Skip the sensor that just updated its heartbeat — its lastSeen is now,
			// so it cannot be considered offline in this analysis cycle.
			continue
		}
		age := now.Sub(r.lastSeen)
		if age >= offlineDuration {
			d.logger.Warn("sensor OFFLINE — emitting gap alert",
				zap.String("sensor_id", id),
				zap.Duration("age", age),
			)
			// Phase A: Analyze() returns a single SpatialAlert per call.
			// If multiple sensors are simultaneously offline, subsequent sensors
			// will trigger their own gap alert on the next observation received
			// from any remaining active sensor.
			// Full multi-sensor batch emission is a Phase B enhancement.
			return d.buildGapAlertFromRecord(r, now), nil
		}
	}

	return nil, nil
}

// updateCoverageRecord parses coverage metadata from the observation and
// refreshes (or creates) the in-memory record for the sensor.
// Returns nil when no coverage metadata is present.
func (d *GapDetector) updateCoverageRecord(sensorID string, meta map[string]string, now time.Time) *sensorCoverageRecord {
	rangeNm, okR := parseFloat(meta, "rtsa.coverage.range_nm")
	centerLat, okLat := parseFloat(meta, "rtsa.coverage.center_lat")
	centerLon, okLon := parseFloat(meta, "rtsa.coverage.center_lon")

	if !okR || !okLat || !okLon {
		return nil
	}

	bearingStart, _ := parseFloat(meta, "rtsa.coverage.bearing_start")
	bearingEnd, _ := parseFloat(meta, "rtsa.coverage.bearing_end")

	d.mu.Lock()
	defer d.mu.Unlock()

	rec, exists := d.coverage[sensorID]
	if !exists {
		rec = &sensorCoverageRecord{sensorID: sensorID}
		d.coverage[sensorID] = rec
	}
	rec.rangeNm = rangeNm
	rec.bearingStart = bearingStart
	rec.bearingEnd = bearingEnd
	rec.centerLat = centerLat
	rec.centerLon = centerLon
	rec.lastSeen = now
	return rec
}

// buildGapAlert creates a SpatialAlert referencing a named sector and a
// default rectangular polygon around the sensor's last known position.
func (d *GapDetector) buildGapAlert(sensorID, sectorID string, now time.Time) *inferencev1.SpatialAlert {
	d.mu.Lock()
	rec, ok := d.coverage[sensorID]
	d.mu.Unlock()

	if !ok {
		// Fallback polygon when we have no coverage record
		return &inferencev1.SpatialAlert{
			AlertId:      fmt.Sprintf("gap-%d", now.UnixNano()),
			AnomalyType:  commonv1.AnomalyType_ANOMALY_TYPE_UNSPECIFIED,
			Severity:     commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL,
			Explanation:  fmt.Sprintf("Tactical coverage gap detected — sensor %s offline", sensorID),
			DetectedAt:   timestamppb.New(now),
			ModelVersion: "gap-heuristic-v1",
			AreaPolygon: []*commonv1.Position{
				{Latitude: 44.7, Longitude: -63.6},
				{Latitude: 44.8, Longitude: -63.6},
				{Latitude: 44.8, Longitude: -63.5},
				{Latitude: 44.7, Longitude: -63.5},
			},
		}
	}
	return d.buildGapAlertFromRecord(rec, now)
}

// buildGapAlertFromRecord creates a SpatialAlert using a stale sensor's last
// known coverage to approximate the gap polygon.
func (d *GapDetector) buildGapAlertFromRecord(rec *sensorCoverageRecord, now time.Time) *inferencev1.SpatialAlert {
	// Approximate the gap area as a bounding box around the sensor coverage
	// centroid ± range (1 NM ≈ 1/60 degree latitude).
	halfDeg := rec.rangeNm / 60.0
	polygon := []*commonv1.Position{
		{Latitude: rec.centerLat - halfDeg, Longitude: rec.centerLon - halfDeg},
		{Latitude: rec.centerLat + halfDeg, Longitude: rec.centerLon - halfDeg},
		{Latitude: rec.centerLat + halfDeg, Longitude: rec.centerLon + halfDeg},
		{Latitude: rec.centerLat - halfDeg, Longitude: rec.centerLon + halfDeg},
	}

	return &inferencev1.SpatialAlert{
		AlertId:      fmt.Sprintf("gap-%s-%d", rec.sensorID, now.UnixNano()),
		AnomalyType:  commonv1.AnomalyType_ANOMALY_TYPE_UNSPECIFIED,
		Severity:     commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL,
		Explanation:  fmt.Sprintf("Coverage gap: sensor %s OFFLINE (last seen %.0fs ago)", rec.sensorID, now.Sub(rec.lastSeen).Seconds()),
		DetectedAt:   timestamppb.New(now),
		ModelVersion: "gap-heuristic-v1",
		AreaPolygon:  polygon,
	}
}

// parseFloat is a helper that extracts and parses a float64 from a metadata map.
func parseFloat(meta map[string]string, key string) (float64, bool) {
	v, ok := meta[key]
	if !ok {
		return 0, false
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0, false
	}
	return f, true
}

