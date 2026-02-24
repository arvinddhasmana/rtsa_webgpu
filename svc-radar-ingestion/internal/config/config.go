// CLASSIFICATION: UNCLASSIFIED
package config

import (
"fmt"
"os"
"strconv"
"strings"

pkgconfig "github.com/arvinddhasmana/RTSA_VS_Opus/pkg/config"
)

// Config extends BaseConfig with radar-specific settings.
type Config struct {
pkgconfig.BaseConfig

// Output topic for validated observations.
OutputTopic string // RTSA_RADAR_OUTPUT_TOPIC (default: "sensors.radar.tracks")
// DLQ topic for invalid messages.
DLQTopic string // RTSA_RADAR_DLQ_TOPIC (default: "dlq.sensors.radar")
// Maximum speed in knots for surface tracks.
MaxSurfaceSpeedKnots float64 // RTSA_RADAR_MAX_SURFACE_SPEED (default: 999)
// Maximum speed in knots for air tracks.
MaxAirSpeedKnots float64 // RTSA_RADAR_MAX_AIR_SPEED (default: 2500)
// Maximum accepted future time offset (seconds).
MaxFutureOffsetSec int // RTSA_RADAR_MAX_FUTURE_OFFSET (default: 300)
// Maximum accepted past time offset (seconds).
MaxPastOffsetSec int // RTSA_RADAR_MAX_PAST_OFFSET (default: 86400)
}

// MustLoad loads and validates radar ingestion config.
func MustLoad() *Config {
cfg, err := Load()
if err != nil {
panic(fmt.Sprintf("config: failed to load radar config: %v", err))
}
return cfg
}

// Load reads radar-specific configuration from environment variables.
func Load() (*Config, error) {
base, err := pkgconfig.Load()
if err != nil {
return nil, fmt.Errorf("config: base config: %w", err)
}

cfg := &Config{
BaseConfig:           *base,
OutputTopic:          getEnvOrDefault("RTSA_RADAR_OUTPUT_TOPIC", "sensors.radar.tracks"),
DLQTopic:             getEnvOrDefault("RTSA_RADAR_DLQ_TOPIC", "dlq.sensors.radar"),
MaxSurfaceSpeedKnots: parseFloat("RTSA_RADAR_MAX_SURFACE_SPEED", 999.0),
MaxAirSpeedKnots:     parseFloat("RTSA_RADAR_MAX_AIR_SPEED", 2500.0),
MaxFutureOffsetSec:   parseInt("RTSA_RADAR_MAX_FUTURE_OFFSET", 300),
MaxPastOffsetSec:     parseInt("RTSA_RADAR_MAX_PAST_OFFSET", 86400),
}
return cfg, nil
}

func getEnvOrDefault(key, defaultVal string) string {
if val, ok := os.LookupEnv(key); ok && val != "" {
return val
}
return defaultVal
}

func parseFloat(key string, defaultVal float64) float64 {
val := os.Getenv(key)
if val == "" {
return defaultVal
}
f, err := strconv.ParseFloat(strings.TrimSpace(val), 64)
if err != nil {
return defaultVal
}
return f
}

func parseInt(key string, defaultVal int) int {
val := os.Getenv(key)
if val == "" {
return defaultVal
}
n, err := strconv.Atoi(strings.TrimSpace(val))
if err != nil {
return defaultVal
}
return n
}
