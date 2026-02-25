// CLASSIFICATION: UNCLASSIFIED
package config

import (
"fmt"
"os"
"strconv"
"strings"

pkgconfig "github.com/arvinddhasmana/RTSA_VS_Opus/pkg/config"
)

// Config extends BaseConfig with AIS/BFT-specific settings.
type Config struct {
pkgconfig.BaseConfig

// Output topic for validated AIS position observations.
OutputTopic string // RTSA_AIS_OUTPUT_TOPIC (default: "sensors.ais.positions")
// DLQ topic for invalid messages.
DLQTopic string // RTSA_AIS_DLQ_TOPIC (default: "dlq.sensors.ais")
// Maximum speed jump in knots for spoofing detection.
MaxSpeedJumpKnots float64 // RTSA_AIS_MAX_SPEED_JUMP_KNOTS (default: 50)
}

// MustLoad loads and validates AIS ingestion config.
func MustLoad() *Config {
cfg, err := Load()
if err != nil {
panic(fmt.Sprintf("config: failed to load ais config: %v", err))
}
return cfg
}

// Load reads AIS-specific configuration from environment variables.
func Load() (*Config, error) {
base, err := pkgconfig.Load()
if err != nil {
return nil, fmt.Errorf("config: base config: %w", err)
}

cfg := &Config{
BaseConfig:        *base,
OutputTopic:       getEnvOrDefault("RTSA_AIS_OUTPUT_TOPIC", "sensors.ais.positions"),
DLQTopic:          getEnvOrDefault("RTSA_AIS_DLQ_TOPIC", "dlq.sensors.ais"),
MaxSpeedJumpKnots: parseFloat("RTSA_AIS_MAX_SPEED_JUMP_KNOTS", 50.0),
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
