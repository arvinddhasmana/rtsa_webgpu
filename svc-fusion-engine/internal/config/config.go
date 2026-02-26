// CLASSIFICATION: UNCLASSIFIED
package config

import (
	pkgconfig "github.com/arvinddhasmana/RTSA_VS_Opus/pkg/config"
)

// Config holds all configuration for the fusion engine service.
// All fields are loaded from environment variables.
type Config struct {
	pkgconfig.BaseConfig

	// ── Consumer ──
	InputTopics   []string // RTSA_FUSION_INPUT_TOPICS (default: all sensors.* topics)
	ConsumerGroup string   // RTSA_FUSION_CONSUMER_GROUP (default: "fusion-engine")

	// ── Gating Thresholds ──
	GateSurfaceDistanceNM float64 // RTSA_FUSION_GATE_SURFACE_DISTANCE (default: 5.0)
	GateSurfaceTimeSec    int     // RTSA_FUSION_GATE_SURFACE_TIME (default: 30)
	GateAirDistanceNM     float64 // RTSA_FUSION_GATE_AIR_DISTANCE (default: 20.0)
	GateAirTimeSec        int     // RTSA_FUSION_GATE_AIR_TIME (default: 15)
	GateSubDistanceNM     float64 // RTSA_FUSION_GATE_SUB_DISTANCE (default: 2.0)
	GateSubTimeSec        int     // RTSA_FUSION_GATE_SUB_TIME (default: 60)

	// ── Correlation Weights ──
	WeightPosition float64 // RTSA_FUSION_WEIGHT_POSITION (default: 0.35)
	WeightVelocity float64 // RTSA_FUSION_WEIGHT_VELOCITY (default: 0.25)
	WeightType     float64 // RTSA_FUSION_WEIGHT_TYPE (default: 0.20)
	WeightTemporal float64 // RTSA_FUSION_WEIGHT_TEMPORAL (default: 0.20)

	// ── Correlation Thresholds ──
	AutoCorrelateThreshold      float64 // RTSA_FUSION_AUTO_THRESHOLD (default: 0.85)
	TentativeCorrelateThreshold float64 // RTSA_FUSION_TENTATIVE_THRESHOLD (default: 0.60)

	// ── Track Lifecycle ──
	StaleTimeoutSec    int // RTSA_FUSION_STALE_TIMEOUT (default: 60)
	DropTimeoutSec     int // RTSA_FUSION_DROP_TIMEOUT (default: 300)
	StaleCheckInterval int // RTSA_FUSION_STALE_CHECK_INTERVAL (default: 10)

	// ── Output ──
	OutputTopicPrefix string // RTSA_FUSION_OUTPUT_PREFIX (default: "tracks.fused")
}

// Load reads fusion engine configuration from environment variables.
func Load() *Config {
	return &Config{
		BaseConfig: pkgconfig.LoadBase(),
		InputTopics: pkgconfig.GetEnvStrSlice("RTSA_FUSION_INPUT_TOPICS", []string{
			"sensors.radar.tracks", "sensors.ew.intercepts", "sensors.elint.detections",
			"sensors.isr.observations", "sensors.ais.positions", "sensors.cyber.iocs",
		}),
		ConsumerGroup:               pkgconfig.GetEnv("RTSA_FUSION_CONSUMER_GROUP", "fusion-engine"),
		GateSurfaceDistanceNM:       pkgconfig.GetEnvFloat("RTSA_FUSION_GATE_SURFACE_DISTANCE", 5.0),
		GateSurfaceTimeSec:          pkgconfig.GetEnvInt("RTSA_FUSION_GATE_SURFACE_TIME", 30),
		GateAirDistanceNM:           pkgconfig.GetEnvFloat("RTSA_FUSION_GATE_AIR_DISTANCE", 20.0),
		GateAirTimeSec:              pkgconfig.GetEnvInt("RTSA_FUSION_GATE_AIR_TIME", 15),
		GateSubDistanceNM:           pkgconfig.GetEnvFloat("RTSA_FUSION_GATE_SUB_DISTANCE", 2.0),
		GateSubTimeSec:              pkgconfig.GetEnvInt("RTSA_FUSION_GATE_SUB_TIME", 60),
		WeightPosition:              pkgconfig.GetEnvFloat("RTSA_FUSION_WEIGHT_POSITION", 0.35),
		WeightVelocity:              pkgconfig.GetEnvFloat("RTSA_FUSION_WEIGHT_VELOCITY", 0.25),
		WeightType:                  pkgconfig.GetEnvFloat("RTSA_FUSION_WEIGHT_TYPE", 0.20),
		WeightTemporal:              pkgconfig.GetEnvFloat("RTSA_FUSION_WEIGHT_TEMPORAL", 0.20),
		AutoCorrelateThreshold:      pkgconfig.GetEnvFloat("RTSA_FUSION_AUTO_THRESHOLD", 0.85),
		TentativeCorrelateThreshold: pkgconfig.GetEnvFloat("RTSA_FUSION_TENTATIVE_THRESHOLD", 0.60),
		StaleTimeoutSec:             pkgconfig.GetEnvInt("RTSA_FUSION_STALE_TIMEOUT", 60),
		DropTimeoutSec:              pkgconfig.GetEnvInt("RTSA_FUSION_DROP_TIMEOUT", 300),
		StaleCheckInterval:          pkgconfig.GetEnvInt("RTSA_FUSION_STALE_CHECK_INTERVAL", 10),
		OutputTopicPrefix:           pkgconfig.GetEnv("RTSA_FUSION_OUTPUT_PREFIX", "tracks.fused"),
	}
}
