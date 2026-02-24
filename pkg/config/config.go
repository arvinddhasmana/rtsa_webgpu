// CLASSIFICATION: UNCLASSIFIED
package config

import (
	"os"
	"strconv"
	"strings"
)

// BaseConfig contains configuration fields shared by all RTSA services.
// All fields are loaded from environment variables with the RTSA_ prefix.
type BaseConfig struct {
	ServiceName    string
	ServiceVersion string
	Environment    string

	GRPCPort int
	TLSCert  string
	TLSKey   string
	TLSCA    string

	RedpandaBrokers    []string
	RedpandaTLSEnabled bool

	OTelEndpoint string
	LogLevel     string
	LogFormat    string

	MaxClassification string
}

// LoadBase loads BaseConfig from environment variables with safe defaults.
func LoadBase() BaseConfig {
	return BaseConfig{
		ServiceName:        GetEnv("RTSA_SERVICE_NAME", "unknown"),
		ServiceVersion:     GetEnv("RTSA_SERVICE_VERSION", "dev"),
		Environment:        GetEnv("RTSA_ENVIRONMENT", "development"),
		GRPCPort:           GetEnvInt("RTSA_GRPC_PORT", 50051),
		TLSCert:            GetEnv("RTSA_TLS_CERT", "/certs/server.crt"),
		TLSKey:             GetEnv("RTSA_TLS_KEY", "/certs/server.key"),
		TLSCA:              GetEnv("RTSA_TLS_CA", "/certs/ca.crt"),
		RedpandaBrokers:    GetEnvStrSlice("RTSA_REDPANDA_BROKERS", []string{"localhost:9092"}),
		RedpandaTLSEnabled: GetEnvBool("RTSA_REDPANDA_TLS_ENABLED", true),
		OTelEndpoint:       GetEnv("RTSA_OTEL_ENDPOINT", "localhost:4317"),
		LogLevel:           GetEnv("RTSA_LOG_LEVEL", "info"),
		LogFormat:          GetEnv("RTSA_LOG_FORMAT", "json"),
		MaxClassification:  GetEnv("RTSA_MAX_CLASSIFICATION", "UNCLASSIFIED"),
	}
}

// GetEnv returns the environment variable value or the default.
func GetEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// GetEnvInt returns the environment variable as int or the default.
func GetEnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultVal
}

// GetEnvBool returns the environment variable as bool or the default.
func GetEnvBool(key string, defaultVal bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return defaultVal
}

// GetEnvFloat returns the environment variable as float64 or the default.
func GetEnvFloat(key string, defaultVal float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return defaultVal
}

// GetEnvStrSlice returns the environment variable as a comma-separated slice or the default.
func GetEnvStrSlice(key string, defaultVal []string) []string {
	if v := os.Getenv(key); v != "" {
		return strings.Split(v, ",")
	}
	return defaultVal
}
