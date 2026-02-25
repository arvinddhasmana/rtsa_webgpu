// CLASSIFICATION: UNCLASSIFIED
package config

import (
"fmt"
"os"
"strconv"
"strings"

pkgconfig "github.com/arvinddhasmana/RTSA_VS_Opus/pkg/config"
)

// Config extends BaseConfig with Cyber-specific settings.
type Config struct {
pkgconfig.BaseConfig

// Output topic for validated Cyber IOC observations.
OutputTopic string // RTSA_CYBER_OUTPUT_TOPIC (default: "sensors.cyber.iocs")
// DLQ topic for invalid messages.
DLQTopic string // RTSA_CYBER_DLQ_TOPIC (default: "dlq.sensors.cyber")
// Deduplication cache size for IOC dedup.
DedupCacheSize int // RTSA_CYBER_DEDUP_CACHE_SIZE (default: 1000)
}

// MustLoad loads and validates Cyber ingestion config.
func MustLoad() *Config {
cfg, err := Load()
if err != nil {
panic(fmt.Sprintf("config: failed to load cyber config: %v", err))
}
return cfg
}

// Load reads Cyber-specific configuration from environment variables.
func Load() (*Config, error) {
base, err := pkgconfig.Load()
if err != nil {
return nil, fmt.Errorf("config: base config: %w", err)
}

cfg := &Config{
BaseConfig:     *base,
OutputTopic:    getEnvOrDefault("RTSA_CYBER_OUTPUT_TOPIC", "sensors.cyber.iocs"),
DLQTopic:       getEnvOrDefault("RTSA_CYBER_DLQ_TOPIC", "dlq.sensors.cyber"),
DedupCacheSize: parseInt("RTSA_CYBER_DEDUP_CACHE_SIZE", 1000),
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
