// CLASSIFICATION: UNCLASSIFIED
package config

import (
"fmt"
"os"
"strconv"
"strings"

pkgconfig "github.com/arvinddhasmana/RTSA_VS_Opus/pkg/config"
)

// Config extends BaseConfig with ISR-specific settings.
type Config struct {
pkgconfig.BaseConfig

// Output topic for validated ISR observations.
OutputTopic string // RTSA_ISR_OUTPUT_TOPIC (default: "sensors.isr.observations")
// DLQ topic for invalid messages.
DLQTopic string // RTSA_ISR_DLQ_TOPIC (default: "dlq.sensors.isr")
// Minimum polygon vertices required.
MinPolygonVertices int // RTSA_ISR_MIN_POLYGON_VERTS (default: 3)
}

// MustLoad loads and validates ISR ingestion config.
func MustLoad() *Config {
cfg, err := Load()
if err != nil {
panic(fmt.Sprintf("config: failed to load isr config: %v", err))
}
return cfg
}

// Load reads ISR-specific configuration from environment variables.
func Load() (*Config, error) {
base, err := pkgconfig.Load()
if err != nil {
return nil, fmt.Errorf("config: base config: %w", err)
}

cfg := &Config{
BaseConfig:         *base,
OutputTopic:        getEnvOrDefault("RTSA_ISR_OUTPUT_TOPIC", "sensors.isr.observations"),
DLQTopic:           getEnvOrDefault("RTSA_ISR_DLQ_TOPIC", "dlq.sensors.isr"),
MinPolygonVertices: parseInt("RTSA_ISR_MIN_POLYGON_VERTS", 3),
}
return cfg, nil
}

func getEnvOrDefault(key, defaultVal string) string {
if val, ok := os.LookupEnv(key); ok && val != "" {
return val
}
return defaultVal
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
