// CLASSIFICATION: UNCLASSIFIED
package config

import (
"fmt"
"os"
"strconv"
"strings"

pkgconfig "github.com/arvinddhasmana/RTSA_VS_Opus/pkg/config"
)

// Config extends BaseConfig with ELINT/COMINT-specific settings.
type Config struct {
pkgconfig.BaseConfig

// Output topic for validated ELINT detections.
OutputTopic string // RTSA_ELINT_OUTPUT_TOPIC (default: "sensors.elint.detections")
// DLQ topic for invalid messages.
DLQTopic string // RTSA_ELINT_DLQ_TOPIC (default: "dlq.sensors.elint")
// Maximum CEP (circular error probable) in meters.
MaxCEPMeters float64 // RTSA_ELINT_MAX_CEP_METERS (default: 50000)
}

// MustLoad loads and validates ELINT ingestion config.
func MustLoad() *Config {
cfg, err := Load()
if err != nil {
panic(fmt.Sprintf("config: failed to load elint config: %v", err))
}
return cfg
}

// Load reads ELINT-specific configuration from environment variables.
func Load() (*Config, error) {
base, err := pkgconfig.Load()
if err != nil {
return nil, fmt.Errorf("config: base config: %w", err)
}

cfg := &Config{
BaseConfig:   *base,
OutputTopic:  getEnvOrDefault("RTSA_ELINT_OUTPUT_TOPIC", "sensors.elint.detections"),
DLQTopic:     getEnvOrDefault("RTSA_ELINT_DLQ_TOPIC", "dlq.sensors.elint"),
MaxCEPMeters: parseFloat("RTSA_ELINT_MAX_CEP_METERS", 50000.0),
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
