// CLASSIFICATION: UNCLASSIFIED
// Package config provides environment-variable-driven configuration for svc-training.
//
// Feature: FEAT-12 Training Pipeline
// UC: UC014, UC015
// Requirements: CR-FB-003, CR-FB-004
package config

import (
	"fmt"
	"os"
	"strings"
)

// Config holds all runtime configuration for svc-training.
type Config struct {
// HealthPort is the HTTP health check port.
HealthPort string
// InputTopic is the Redpanda topic to consume validated feedback from.
InputTopic string
// OutputTopic is the Redpanda topic to produce noop model candidates to.
OutputTopic string
// Brokers is the list of Redpanda broker addresses.
Brokers []string
// ConsumerGroup is the Kafka consumer group for this service.
ConsumerGroup string
// Environment is the deployment environment.
Environment string
// OTelEndpoint is the OpenTelemetry collector gRPC endpoint.
OTelEndpoint string
// ServiceName is the service identifier for telemetry.
ServiceName string
// LogLevel is the minimum log level.
LogLevel string
// LogFormat is the log output format.
LogFormat string
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
brokerStr := getEnv("RTSA_REDPANDA_BROKERS", "localhost:19092")
brokers := strings.Split(brokerStr, ",")
for i, b := range brokers {
brokers[i] = strings.TrimSpace(b)
}

healthPort := getEnv("RTSA_HEALTH_PORT", "8081")

cfg := &Config{
HealthPort:    ":" + healthPort,
InputTopic:    getEnv("INPUT_TOPIC", "feedback.operator.validated"),
OutputTopic:   getEnv("OUTPUT_TOPIC", "models.anomaly.candidates"),
Brokers:       brokers,
ConsumerGroup: getEnv("CONSUMER_GROUP", "svc-training"),
Environment:   getEnv("RTSA_ENVIRONMENT", "development"),
OTelEndpoint:  getEnv("RTSA_OTEL_ENDPOINT", "localhost:4317"),
ServiceName:   getEnv("RTSA_SERVICE_NAME", "svc-training"),
LogLevel:      getEnv("RTSA_LOG_LEVEL", "info"),
LogFormat:     getEnv("RTSA_LOG_FORMAT", "json"),
}

if err := cfg.validate(); err != nil {
return nil, fmt.Errorf("config.Load: %w", err)
}

return cfg, nil
}

// validate checks that configuration values are self-consistent.
func (c *Config) validate() error {
if len(c.Brokers) == 0 {
return fmt.Errorf("RTSA_REDPANDA_BROKERS must not be empty")
}
if c.InputTopic == "" {
return fmt.Errorf("INPUT_TOPIC must not be empty")
}
if c.OutputTopic == "" {
return fmt.Errorf("OUTPUT_TOPIC must not be empty")
}
if c.ConsumerGroup == "" {
return fmt.Errorf("CONSUMER_GROUP must not be empty")
}
logLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
if !logLevels[c.LogLevel] {
return fmt.Errorf("RTSA_LOG_LEVEL must be one of debug/info/warn/error, got %q", c.LogLevel)
}
return nil
}

// getEnv returns the environment variable value or the default.
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
