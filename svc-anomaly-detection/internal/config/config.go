// CLASSIFICATION: UNCLASSIFIED
package config

import (
"encoding/json"
"fmt"
"os"
"strconv"
"strings"
"time"
)

// ExclusionZoneConfig holds configuration for a single exclusion zone.
type ExclusionZoneConfig struct {
Name      string  `json:"name"`
CenterLat float64 `json:"center_lat"`
CenterLon float64 `json:"center_lon"`
RadiusNM  float64 `json:"radius_nm"`
}

// DetectorConfig holds enable/disable flags and thresholds for each detector.
type DetectorConfig struct {
SpeedEnabled       bool    `json:"speed_enabled"`
SpeedSigmaThresh   float64 `json:"speed_sigma_threshold"`
RouteEnabled       bool    `json:"route_enabled"`
RouteDeviationDeg  float64 `json:"route_deviation_degrees"`
RouteSustainedN    int     `json:"route_sustained_updates"`
AISEnabled         bool    `json:"ais_enabled"`
AISDiscrepancyNM   float64 `json:"ais_discrepancy_nm"`
BehavioralEnabled  bool    `json:"behavioral_enabled"`
BehavioralConfThr  float64 `json:"behavioral_confidence_threshold"`
TemporalEnabled    bool    `json:"temporal_enabled"`
TemporalPValue     float64 `json:"temporal_p_value_threshold"`
ProximityEnabled   bool    `json:"proximity_enabled"`
}

// Config holds the complete service configuration.
type Config struct {
ServiceName    string
RedpandaBrokers []string
ConsumerGroup  string
InputTopics    []string
HealthAddr     string
LogLevel       string
LogFormat      string
ModelVersion   string
TrackHistoryMaxEntries int
TrackHistoryMaxAge     time.Duration
Detectors      DetectorConfig
ExclusionZones []ExclusionZoneConfig
}

// Load reads configuration from environment variables and returns a Config.
func Load() (*Config, error) {
cfg := &Config{
ServiceName:     getEnv("RTSA_SERVICE_NAME", "svc-anomaly-detection"),
RedpandaBrokers: strings.Split(getEnv("RTSA_REDPANDA_BROKERS", "localhost:19092"), ","),
ConsumerGroup:   getEnv("RTSA_ANOMALY_CONSUMER_GROUP", "anomaly-detection"),
InputTopics: []string{
"tracks.fused.surface",
"tracks.fused.air",
"tracks.fused.subsurface",
"tracks.fused.land",
"tracks.fused.cyber",
},
HealthAddr:   getEnv("RTSA_HEALTH_ADDR", ":8081"),
LogLevel:     getEnv("RTSA_LOG_LEVEL", "info"),
LogFormat:    getEnv("RTSA_LOG_FORMAT", "json"),
ModelVersion: getEnv("RTSA_MODEL_VERSION", "rules-v1.0.0"),
TrackHistoryMaxEntries: getEnvInt("RTSA_TRACK_HISTORY_MAX_ENTRIES", 100),
TrackHistoryMaxAge:     getEnvDuration("RTSA_TRACK_HISTORY_MAX_AGE", 2*time.Hour),
Detectors: DetectorConfig{
SpeedEnabled:      getEnvBool("RTSA_DETECTOR_SPEED_ENABLED", true),
SpeedSigmaThresh:  getEnvFloat("RTSA_DETECTOR_SPEED_SIGMA_THRESHOLD", 3.0),
RouteEnabled:      getEnvBool("RTSA_DETECTOR_ROUTE_ENABLED", true),
RouteDeviationDeg: getEnvFloat("RTSA_DETECTOR_ROUTE_DEVIATION_DEG", 30.0),
RouteSustainedN:   getEnvInt("RTSA_DETECTOR_ROUTE_SUSTAINED_N", 3),
AISEnabled:        getEnvBool("RTSA_DETECTOR_AIS_ENABLED", true),
AISDiscrepancyNM:  getEnvFloat("RTSA_DETECTOR_AIS_DISCREPANCY_NM", 0.5),
BehavioralEnabled: getEnvBool("RTSA_DETECTOR_BEHAVIORAL_ENABLED", true),
BehavioralConfThr: getEnvFloat("RTSA_DETECTOR_BEHAVIORAL_CONFIDENCE_THRESHOLD", 0.75),
TemporalEnabled:   getEnvBool("RTSA_DETECTOR_TEMPORAL_ENABLED", true),
TemporalPValue:    getEnvFloat("RTSA_DETECTOR_TEMPORAL_P_VALUE", 0.05),
ProximityEnabled:  getEnvBool("RTSA_DETECTOR_PROXIMITY_ENABLED", true),
},
}

// Load exclusion zones from JSON env var or file.
if err := cfg.loadExclusionZones(); err != nil {
return nil, fmt.Errorf("[config.Load]: %w", err)
}

return cfg, nil
}

// loadExclusionZones loads exclusion zone configuration from env or file.
func (c *Config) loadExclusionZones() error {
// Try JSON env var first.
if raw := os.Getenv("RTSA_EXCLUSION_ZONES_JSON"); raw != "" {
if err := json.Unmarshal([]byte(raw), &c.ExclusionZones); err != nil {
return fmt.Errorf("[config.loadExclusionZones] invalid RTSA_EXCLUSION_ZONES_JSON: %w", err)
}
return nil
}
// Try config file.
if path := os.Getenv("RTSA_EXCLUSION_ZONES_FILE"); path != "" {
data, err := os.ReadFile(path)
if err != nil {
return fmt.Errorf("[config.loadExclusionZones] read file %q: %w", path, err)
}
if err := json.Unmarshal(data, &c.ExclusionZones); err != nil {
return fmt.Errorf("[config.loadExclusionZones] parse file %q: %w", path, err)
}
return nil
}
// Default: empty list (no exclusion zones configured).
return nil
}

func getEnv(key, defaultVal string) string {
if v := os.Getenv(key); v != "" {
return v
}
return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
v := os.Getenv(key)
if v == "" {
return defaultVal
}
b, err := strconv.ParseBool(v)
if err != nil {
return defaultVal
}
return b
}

func getEnvFloat(key string, defaultVal float64) float64 {
v := os.Getenv(key)
if v == "" {
return defaultVal
}
f, err := strconv.ParseFloat(v, 64)
if err != nil {
return defaultVal
}
return f
}

func getEnvInt(key string, defaultVal int) int {
v := os.Getenv(key)
if v == "" {
return defaultVal
}
i, err := strconv.Atoi(v)
if err != nil {
return defaultVal
}
return i
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
v := os.Getenv(key)
if v == "" {
return defaultVal
}
d, err := time.ParseDuration(v)
if err != nil {
return defaultVal
}
return d
}
