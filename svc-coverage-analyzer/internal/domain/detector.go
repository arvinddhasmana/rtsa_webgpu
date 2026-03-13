// CLASSIFICATION: UNCLASSIFIED
package domain

import (
	"context"
	"fmt"
	"time"

	commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
	inferencev1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/inference/v1"
	ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GapDetector struct {
	logger *zap.Logger
}

func NewGapDetector(logger *zap.Logger) *GapDetector {
	return &GapDetector{logger: logger}
}

// Analyze processes an observation and returns a SpatialAlert if a gap is detected.
func (d *GapDetector) Analyze(ctx context.Context, obs *ingestionv1.SensorObservation) (*inferencev1.SpatialAlert, error) {
	meta := obs.GetMetadata()
	if meta == nil {
		return nil, nil
	}

	// For demonstration/E2E testing, we look for a simulated gap injection
	if gapInject, ok := meta["rtsa.sim_inject_gap"]; ok && gapInject == "true" {
		d.logger.Info("simulated coverage gap detected", zap.String("sensor_id", obs.GetSensorId()))

		// Create a mock polygon for the gap area
		// In a real scenario, this would be computed by comparing sensor footprints
		return &inferencev1.SpatialAlert{
			AlertId:      fmt.Sprintf("gap-%d", time.Now().UnixNano()),
			AnomalyType:  commonv1.AnomalyType_ANOMALY_TYPE_UNSPECIFIED,
			Severity:     commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL,
			Explanation:  "Tactical coverage gap detected in sector " + obs.GetSensorId(),
			DetectedAt:   timestamppb.Now(),
			ModelVersion: "geometric-v1",
			AreaPolygon: []*commonv1.Position{
				{Latitude: 44.7, Longitude: -63.6},
				{Latitude: 44.8, Longitude: -63.6},
				{Latitude: 44.8, Longitude: -63.5},
				{Latitude: 44.7, Longitude: -63.5},
			},
		}, nil
	}

	return nil, nil
}
