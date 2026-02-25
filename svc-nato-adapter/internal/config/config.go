// CLASSIFICATION: UNCLASSIFIED
// Package config provides environment-variable-driven configuration for svc-nato-adapter.
//
// Feature: FEAT-15 NATO Interoperability
// UC: UC011 NATO Adapter
// Requirements: CR-NATO-001, CR-NATO-002, CR-NATO-003, CR-NATO-004, CR-NATO-005
package config

import (
"fmt"
"os"
"strconv"
"strings"
)

// Config holds all runtime configuration for svc-nato-adapter.
type Config struct {
// GRPCAddr is the address the gRPC server listens on.
GRPCAddr string
// HealthPort is the HTTP health check server port.
HealthPort string
// MetricsPort is the Prometheus metrics server port.
MetricsPort string
// OTelEndpoint is the OpenTelemetry collector gRPC endpoint.
OTelEndpoint string
// Environment is the deployment environment (development, production).
Environment string
// ServiceName is the service identifier for telemetry.
ServiceName string
// LogLevel is the minimum log level (debug, info, warn, error).
LogLevel string
// LogFormat is the log output format (json, text).
LogFormat string
// TLSEnabled controls whether mTLS is enforced.
TLSEnabled bool
// TLSCACert is the path to the CA certificate for mTLS.
TLSCACert string
// TLSServerCert is the path to the server certificate for mTLS.
TLSServerCert string
// TLSServerKey is the path to the server private key for mTLS.
TLSServerKey string
// RedpandaBrokers is a list of Redpanda broker addresses.
RedpandaBrokers []string
}

// Load reads configuration from environment variables.
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

cfg := &Config{
GRPCAddr:        ":" + grpcPort,
HealthPort:      ":" + healthPort,
MetricsPort:     ":" + metricsPort,
OTelEndpoint:    getEnv("RTSA_OTEL_ENDPOINT", "localhost:4317"),
Environment:     getEnv("RTSA_ENVIRONMENT", "development"),
ServiceName:     getEnv("RTSA_SERVICE_NAME", "svc-nato-adapter"),
LogLevel:        getEnv("RTSA_LOG_LEVEL", "info"),
LogFormat:       getEnv("RTSA_LOG_FORMAT", "json"),
TLSEnabled:      tlsEnabled,
TLSCACert:       getEnv("RTSA_TLS_CA_CERT", "./certs/dev/ca.crt"),
TLSServerCert:   getEnv("RTSA_TLS_SERVER_CERT", "./certs/dev/server.crt"),
TLSServerKey:    getEnv("RTSA_TLS_SERVER_KEY", "./certs/dev/server.key"),
RedpandaBrokers: brokers,
}

if err := cfg.validate(); err != nil {
return nil, fmt.Errorf("config.Load: %w", err)
}

return cfg, nil
}

// validate checks that configuration values are self-consistent.
func (c *Config) validate() error {
logLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
if !logLevels[c.LogLevel] {
return fmt.Errorf("RTSA_LOG_LEVEL must be one of debug/info/warn/error, got %q", c.LogLevel)
}
if c.ServiceName == "" {
return fmt.Errorf("RTSA_SERVICE_NAME must not be empty")
}
return nil
}

// getEnv returns the value of the environment variable or the default.
func getEnv(key, defaultVal string) string {
if val := os.Getenv(key); val != "" {
return val
}
return defaultVal
}
