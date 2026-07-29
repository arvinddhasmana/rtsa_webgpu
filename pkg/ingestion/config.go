// CLASSIFICATION: UNCLASSIFIED
package ingestion

import (
	"fmt"
	"os"
	"strconv"

	commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
	ingestionv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/ingestion/v1"
)

// Config holds common ingestion service configuration.
type Config struct {
	GRPCPort          int
	RedpandaBrokers   string
	OutputTopic       string
	DLQTopic          string
	ServiceName       string
	MaxClassification commonv1.ClassificationLevel
	LogLevel          string
	Coverage          *ingestionv1.SensorCoverage
}

// MustLoad loads config from environment variables, panics on missing required vars.
func MustLoad(serviceName, defaultOutputTopic, defaultDLQTopic string, defaultPort int) *Config {
	cfg := &Config{
		ServiceName:     serviceName,
		OutputTopic:     getEnvOrDefault("OUTPUT_TOPIC", defaultOutputTopic),
		DLQTopic:        getEnvOrDefault("DLQ_TOPIC", defaultDLQTopic),
		RedpandaBrokers: getEnvOrDefault("REDPANDA_BROKERS", "localhost:9092"),
		LogLevel:        getEnvOrDefault("LOG_LEVEL", "info"),
	}
	portStr := getEnvOrDefault("GRPC_PORT", strconv.Itoa(defaultPort))
	port, err := strconv.Atoi(portStr)
	if err != nil {
		panic(fmt.Sprintf("invalid GRPC_PORT: %s", portStr))
	}
	cfg.GRPCPort = port

	maxClass := getEnvOrDefault("MAX_CLASSIFICATION", "CLASSIFICATION_LEVEL_SECRET")
	level, ok := commonv1.ClassificationLevel_value[maxClass]
	if !ok {
		panic(fmt.Sprintf("invalid MAX_CLASSIFICATION: %s", maxClass))
	}
	cfg.MaxClassification = commonv1.ClassificationLevel(level)

	// Optional Coverage parsing based on typical ENV vars for sensors
	if rangeNmStr := os.Getenv("SENSOR_RANGE_NM"); rangeNmStr != "" {
		rangeNm, err := strconv.ParseFloat(rangeNmStr, 64)
		if err == nil {
			if cfg.Coverage == nil {
				cfg.Coverage = &ingestionv1.SensorCoverage{}
			}
			cfg.Coverage.RangeNm = &rangeNm
		}
	}
	if bStartStr := os.Getenv("SENSOR_BEARING_START"); bStartStr != "" {
		bStart, err := strconv.ParseFloat(bStartStr, 64)
		if err == nil {
			if cfg.Coverage == nil {
				cfg.Coverage = &ingestionv1.SensorCoverage{}
			}
			cfg.Coverage.BearingStartDegrees = &bStart
		}
	}
	if bEndStr := os.Getenv("SENSOR_BEARING_END"); bEndStr != "" {
		bEnd, err := strconv.ParseFloat(bEndStr, 64)
		if err == nil {
			if cfg.Coverage == nil {
				cfg.Coverage = &ingestionv1.SensorCoverage{}
			}
			cfg.Coverage.BearingEndDegrees = &bEnd
		}
	}
	if latStr := os.Getenv("SENSOR_LAT"); latStr != "" {
		if lonStr := os.Getenv("SENSOR_LON"); lonStr != "" {
			lat, er1 := strconv.ParseFloat(latStr, 64)
			lon, er2 := strconv.ParseFloat(lonStr, 64)
			if er1 == nil && er2 == nil {
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

	return cfg
}

func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
