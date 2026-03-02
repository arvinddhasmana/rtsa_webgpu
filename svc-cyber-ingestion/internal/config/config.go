// CLASSIFICATION: UNCLASSIFIED
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
	ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
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
	// Optional static sensor coverage geometry.
	Coverage *ingestionv1.SensorCoverage
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

	// Parse coverage
	if rangeNmStr := os.Getenv("RTSA_CYBER_RANGE_NM"); rangeNmStr != "" {
		rangeNm := parseFloat("RTSA_CYBER_RANGE_NM", -1.0)
		if rangeNm >= 0 {
			if cfg.Coverage == nil {
				cfg.Coverage = &ingestionv1.SensorCoverage{}
			}
			cfg.Coverage.RangeNm = &rangeNm
		}
	}
	if bStartStr := os.Getenv("RTSA_CYBER_BEARING_START"); bStartStr != "" {
		bStart := parseFloat("RTSA_CYBER_BEARING_START", -1.0)
		if bStart >= 0 {
			if cfg.Coverage == nil {
				cfg.Coverage = &ingestionv1.SensorCoverage{}
			}
			cfg.Coverage.BearingStartDegrees = &bStart
		}
	}
	if bEndStr := os.Getenv("RTSA_CYBER_BEARING_END"); bEndStr != "" {
		bEnd := parseFloat("RTSA_CYBER_BEARING_END", -1.0)
		if bEnd >= 0 {
			if cfg.Coverage == nil {
				cfg.Coverage = &ingestionv1.SensorCoverage{}
			}
			cfg.Coverage.BearingEndDegrees = &bEnd
		}
	}
	if latStr := os.Getenv("RTSA_CYBER_LAT"); latStr != "" {
		if lonStr := os.Getenv("RTSA_CYBER_LON"); lonStr != "" {
			lat := parseFloat("RTSA_CYBER_LAT", -999.0)
			lon := parseFloat("RTSA_CYBER_LON", -999.0)
			if lat != -999.0 && lon != -999.0 {
				if cfg.Coverage == nil {
					cfg.Coverage = &ingestionv1.SensorCoverage{}
				}
				cfg.Coverage.SensorPosition = &commonv1.Position{
					Latitude:  lat,
					Longitude: lon,
				}
			}
		}
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
