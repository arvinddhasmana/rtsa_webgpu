package config
// CLASSIFICATION: UNCLASSIFIED
// Package config provides environment-variable-driven configuration for
// svc-webtransport.
//
// svc-webtransport is the RTSA hot-path fan-out service: it consumes fused
// track updates from Redpanda, serialises them to 128-byte GPU-aligned records
// and broadcasts them to browser Common Operating Picture (COP) clients over
// WebTransport (QUIC datagrams).
//
// Feature: FEAT-13 Situational Awareness UI (hot path)
// UC: UC012 Situational Awareness UI
// Requirements: CR-UI-001, CR-UI-002, CR-SEC-001, NFR-PERF-001
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all runtime configuration for svc-webtransport.
// All fields are populated from RTSA_* environment variables. No secrets are
// ever embedded in source — the JWT signing secret must be supplied via the
// environment (injected from Key Vault via the Secrets Store CSI driver).
type Config struct {
	// ServiceName is the service identifier for telemetry.
	ServiceName string
	// Environment is the deployment environment (dev, staging, prod).
	Environment string

	// WTListenAddr is the WebTransport/QUIC listener address.
	WTListenAddr string
	// TLSCert is the path to the TLS certificate (WebTransport requires TLS).
	TLSCert string
	// TLSKey is the path to the TLS private key.
	TLSKey string
	// TLSCACert is the CA bundle used to verify the Redpanda broker when
	// RedpandaTLSEnabled is true.
	TLSCACert string
	// JWTSecret is the HMAC secret used to validate operator session tokens.
	// Required — the process fails closed if it is not provided.
	JWTSecret []byte
	// AllowedOrigins is the set of permitted browser Origin header values.
	// Empty disables origin checking and is permitted for dev only.
	AllowedOrigins []string
	// MaxSessions caps concurrent WebTransport sessions (0 = unlimited).
	MaxSessions int
	// DatagramBatchSize is the number of 128-byte records per QUIC datagram.
	DatagramBatchSize int

	// RedpandaBrokers is a comma-separated list of Redpanda broker addresses.
	RedpandaBrokers []string
	// RedpandaTLSEnabled toggles mTLS on the broker connection.
	RedpandaTLSEnabled bool
	// ConsumerGroup is the Kafka consumer group for this service.
	ConsumerGroup string
	// Topics is the set of fused-track topics to consume.
	Topics []string
	// StartOffset selects "latest" (hot path default) or "earliest".
	StartOffset string
	// SubscriberBufferSize is the per-session broadcast channel buffer.
	SubscriberBufferSize int

	// HealthPort is the HTTP health/readiness server port.
	HealthPort string
	// MetricsPort is the Prometheus metrics server port.
	MetricsPort string
	// OTelEndpoint is the OpenTelemetry collector gRPC endpoint (empty = off).
	OTelEndpoint string
	// LogLevel is the minimum log level (debug, info, warn, error).
	LogLevel string
}

// Load reads configuration from environment variables and validates it.
func Load() (*Config, error) {
	maxSessions, err := strconv.Atoi(getEnv("RTSA_WT_MAX_SESSIONS", "0"))
	if err != nil || maxSessions < 0 {
		return nil, fmt.Errorf("config.Load: invalid RTSA_WT_MAX_SESSIONS: %q", getEnv("RTSA_WT_MAX_SESSIONS", "0"))
	}

	batch, err := strconv.Atoi(getEnv("RTSA_WT_DATAGRAM_BATCH", "9"))
	if err != nil || batch < 1 || batch > 9 {
		return nil, fmt.Errorf("config.Load: RTSA_WT_DATAGRAM_BATCH must be 1..9, got %q", getEnv("RTSA_WT_DATAGRAM_BATCH", "9"))
	}

	subBuf, err := strconv.Atoi(getEnv("RTSA_WT_SUBSCRIBER_BUFFER", "1024"))
	if err != nil || subBuf <= 0 {
		return nil, fmt.Errorf("config.Load: RTSA_WT_SUBSCRIBER_BUFFER must be > 0, got %q", getEnv("RTSA_WT_SUBSCRIBER_BUFFER", "1024"))
	}

	redpandaTLS, err := strconv.ParseBool(getEnv("RTSA_REDPANDA_TLS_ENABLED", "false"))
	if err != nil {
		return nil, fmt.Errorf("config.Load: invalid RTSA_REDPANDA_TLS_ENABLED: %w", err)
	}

	cfg := &Config{
		ServiceName:          getEnv("RTSA_SERVICE_NAME", "svc-webtransport"),
		Environment:          getEnv("RTSA_ENV", "dev"),
		WTListenAddr:         getEnv("RTSA_WT_LISTEN_ADDR", ":4443"),
		TLSCert:              getEnv("RTSA_TLS_SERVER_CERT", "./certs/dev/server.crt"),
		TLSKey:               getEnv("RTSA_TLS_SERVER_KEY", "./certs/dev/server.key"),
		TLSCACert:            getEnv("RTSA_TLS_CA_CERT", "./certs/dev/ca.crt"),
		JWTSecret:            []byte(os.Getenv("RTSA_WT_JWT_SECRET")),
		AllowedOrigins:       splitCSV(getEnv("RTSA_WT_ALLOWED_ORIGINS", "")),
		MaxSessions:          maxSessions,
		DatagramBatchSize:    batch,
		RedpandaBrokers:      splitCSV(getEnv("RTSA_REDPANDA_BROKERS", "localhost:19092")),
		RedpandaTLSEnabled:   redpandaTLS,
		ConsumerGroup:        getEnv("RTSA_CONSUMER_GROUP", "webtransport-service"),
		Topics:               splitCSV(getEnv("RTSA_WT_TOPICS", "tracks.fused")),
		StartOffset:          getEnv("RTSA_WT_START_OFFSET", "latest"),
		SubscriberBufferSize: subBuf,
		HealthPort:           ":" + getEnv("RTSA_HEALTH_PORT", "8081"),
		MetricsPort:          ":" + getEnv("RTSA_METRICS_PORT", "9090"),
		OTelEndpoint:         getEnv("RTSA_OTEL_ENDPOINT", ""),
		LogLevel:             getEnv("RTSA_LOG_LEVEL", "info"),
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config.Load: %w", err)
	}
	return cfg, nil
}

// validate checks that configuration values are self-consistent and that the
// security-critical JWT secret is present.
func (c *Config) validate() error {
	if len(c.JWTSecret) == 0 {
		return fmt.Errorf("RTSA_WT_JWT_SECRET must be set (no default is permitted for a signing secret)")
	}
	if len(c.RedpandaBrokers) == 0 {
		return fmt.Errorf("RTSA_REDPANDA_BROKERS must not be empty")
	}
	for _, b := range c.RedpandaBrokers {
		if b == "" {
			return fmt.Errorf("RTSA_REDPANDA_BROKERS contains an empty broker address")
		}
	}
	if len(c.Topics) == 0 {
		return fmt.Errorf("RTSA_WT_TOPICS must not be empty")
	}
	if c.ConsumerGroup == "" {
		return fmt.Errorf("RTSA_CONSUMER_GROUP must not be empty")
	}
	if c.StartOffset != "latest" && c.StartOffset != "earliest" {
		return fmt.Errorf("RTSA_WT_START_OFFSET must be latest or earliest, got %q", c.StartOffset)
	}
	logLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !logLevels[c.LogLevel] {
		return fmt.Errorf("RTSA_LOG_LEVEL must be one of debug/info/warn/error, got %q", c.LogLevel)
	}
	return nil
}

// getEnv returns the value of the environment variable named by key, or
// defaultVal if the variable is unset or empty.
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// splitCSV splits a comma-separated string into a trimmed, non-empty slice.
func splitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
