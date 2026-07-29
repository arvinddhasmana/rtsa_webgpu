// CLASSIFICATION: UNCLASSIFIED
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	pkgconfig "github.com/arvinddhasmana/rtsa_webgpu/pkg/config"
)

type Config struct {
	pkgconfig.BaseConfig
	InputTopic  string // Default: "sensors.*.tracks" (using wildcard if supported or specific topics)
	AlertTopic  string // Default: "alerts.spatial.gaps"
	GapThreshold float64 // Seconds of silence before a gap is declared
}

func MustLoad() *Config {
	base, err := pkgconfig.Load()
	if err != nil {
		panic(fmt.Sprintf("config: failed to load base config: %v", err))
	}

	return &Config{
		BaseConfig:   *base,
		InputTopic:    getEnvOrDefault("RTSA_COVERAGE_INPUT_TOPIC", "sensors.radar.tracks"),
		AlertTopic:    getEnvOrDefault("RTSA_COVERAGE_ALERT_TOPIC", "alerts.spatial.gaps"),
		GapThreshold: parseFloat("RTSA_COVERAGE_GAP_THRESHOLD", 60.0),
	}
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
