// CLASSIFICATION: UNCLASSIFIED
// Package config provides environment-variable-driven configuration for svc-track.
//
// Feature: FEAT-13 Situational Awareness UI
// UC: UC012 Situational Awareness UI
// Requirements: CR-UI-001, CR-SEC-001
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all runtime configuration for svc-track.
// All fields are populated from RTSA_* environment variables.
type Config struct {
	// GRPCAddr is the address the gRPC server listens on.
	GRPCAddr string
	// HealthPort is the HTTP health check server port.
	HealthPort string
	// MetricsPort is the Prometheus metrics server port.
	MetricsPort string
	// RedpandaBrokers is a comma-separated list of Redpanda broker addresses.
	RedpandaBrokers []string
	// ConsumerGroupID is the Kafka consumer group for this service.
	ConsumerGroupID string
	// TLSCACert is the path to the CA certificate for mTLS.
	TLSCACert string
	// TLSServerCert is the path to the server certificate for mTLS.
	TLSServerCert string
	// TLSServerKey is the path to the server private key for mTLS.
	TLSServerKey string
	// TLSEnabled controls whether mTLS is enforced (default: true).
	TLSEnabled bool
	// LogLevel is the minimum log level (debug, info, warn, error).
	LogLevel string
	// LogFormat is the log output format (json, text).
	LogFormat string
	// OTelEndpoint is the OpenTelemetry collector gRPC endpoint.
	OTelEndpoint string
	// ServiceName is the service identifier for telemetry.
	ServiceName string
	// HistoryMaxPoints is the maximum number of position history points per track.
	HistoryMaxPoints int
	// StreamChannelBufferSize is the buffer size for per-client update channels.
	StreamChannelBufferSize int
}

// Load reads configuration from environment variables.
// Returns an error if required values are missing or invalid.
func Load() (*Config, error) {
	grpcPort := getEnv("RTSA_GRPC_PORT", "50051")
	healthPort := getEnv("RTSA_HEALTH_PORT", "8081")
	metricsPort := getEnv("RTSA_METRICS_PORT", "9090")
	brokerStr := getEnv("RTSA_REDPANDA_BROKERS", "localhost:19092")
	brokers := strings.Split(brokerStr, ",")
	for i, b := range brokers {
		brokers[i] = strings.TrimSpace(b)
	}

	tlsEnabled, err := strconv.ParseBool(getEnv("RTSA_TLS_ENABLED", "false"))
	if err != nil {
		return nil, fmt.Errorf("config.Load: invalid RTSA_TLS_ENABLED: %w", err)
	}

	historyMax, err := strconv.Atoi(getEnv("RTSA_HISTORY_MAX_POINTS", "100"))
	if err != nil {
		return nil, fmt.Errorf("config.Load: invalid RTSA_HISTORY_MAX_POINTS: %w", err)
	}
	if historyMax <= 0 {
		return nil, fmt.Errorf("config.Load: RTSA_HISTORY_MAX_POINTS must be > 0, got %d", historyMax)
	}

	chanBuf, err := strconv.Atoi(getEnv("RTSA_STREAM_CHANNEL_BUFFER", "256"))
	if err != nil {
		return nil, fmt.Errorf("config.Load: invalid RTSA_STREAM_CHANNEL_BUFFER: %w", err)
	}
	if chanBuf <= 0 {
		return nil, fmt.Errorf("config.Load: RTSA_STREAM_CHANNEL_BUFFER must be > 0, got %d", chanBuf)
	}

	cfg := &Config{
		GRPCAddr:                ":" + grpcPort,
		HealthPort:              ":" + healthPort,
		MetricsPort:             ":" + metricsPort,
		RedpandaBrokers:         brokers,
		ConsumerGroupID:         getEnv("RTSA_CONSUMER_GROUP", "track-service"),
		TLSCACert:               getEnv("RTSA_TLS_CA_CERT", "./certs/dev/ca.crt"),
		TLSServerCert:           getEnv("RTSA_TLS_SERVER_CERT", "./certs/dev/server.crt"),
		TLSServerKey:            getEnv("RTSA_TLS_SERVER_KEY", "./certs/dev/server.key"),
		TLSEnabled:              tlsEnabled,
		LogLevel:                getEnv("RTSA_LOG_LEVEL", "info"),
		LogFormat:               getEnv("RTSA_LOG_FORMAT", "json"),
		OTelEndpoint:            getEnv("RTSA_OTEL_ENDPOINT", "localhost:4317"),
		ServiceName:             getEnv("RTSA_SERVICE_NAME", "svc-track"),
		HistoryMaxPoints:        historyMax,
		StreamChannelBufferSize: chanBuf,
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config.Load: %w", err)
	}

	return cfg, nil
}

// validate checks that configuration values are self-consistent.
func (c *Config) validate() error {
	if len(c.RedpandaBrokers) == 0 {
		return fmt.Errorf("RTSA_REDPANDA_BROKERS must not be empty")
	}
	for _, b := range c.RedpandaBrokers {
		if b == "" {
			return fmt.Errorf("RTSA_REDPANDA_BROKERS contains empty broker address")
		}
	}
	if c.ConsumerGroupID == "" {
		return fmt.Errorf("RTSA_CONSUMER_GROUP must not be empty")
	}
	logLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !logLevels[c.LogLevel] {
		return fmt.Errorf("RTSA_LOG_LEVEL must be one of debug/info/warn/error, got %q", c.LogLevel)
	}
	return nil
}

// getEnv returns the value of the environment variable named by key,
// or defaultVal if the variable is not set or empty.
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
