// CLASSIFICATION: UNCLASSIFIED
package config

import (
"fmt"
"os"
"reflect"
"strconv"
"strings"
)

// BaseConfig contains configuration fields shared by all services.
// All fields are loaded from environment variables with the RTSA_ prefix.
type BaseConfig struct {
// ── Service Identity ──
ServiceName    string // RTSA_SERVICE_NAME (required)
ServiceVersion string // RTSA_SERVICE_VERSION (default: "dev")
Environment    string // RTSA_ENVIRONMENT (default: "development")

// ── gRPC Server ──
GRPCPort int    // RTSA_GRPC_PORT (default: 50051)
TLSCert  string // RTSA_TLS_CERT (default: "/certs/server.crt")
TLSKey   string // RTSA_TLS_KEY (default: "/certs/server.key")
TLSCA    string // RTSA_TLS_CA (default: "/certs/ca.crt")

// ── Redpanda ──
RedpandaBrokers    []string // RTSA_REDPANDA_BROKERS (comma-separated, default: "localhost:9092")
RedpandaTLSEnabled bool     // RTSA_REDPANDA_TLS_ENABLED (default: true)

// ── ClickHouse ──
ClickHouseDSN string // RTSA_CLICKHOUSE_DSN

// ── Observability ──
OTelEndpoint string // RTSA_OTEL_ENDPOINT (default: "localhost:4317")
LogLevel     string // RTSA_LOG_LEVEL (default: "info")
LogFormat    string // RTSA_LOG_FORMAT (default: "json")

// ── Classification ──
MaxClassification string // RTSA_MAX_CLASSIFICATION (default: "UNCLASSIFIED")
}

// Load reads configuration from environment variables.
// Required variables cause an error if missing.
func Load() (*BaseConfig, error) {
cfg := &BaseConfig{}

var errs []string

// Required fields
serviceName, ok := os.LookupEnv("RTSA_SERVICE_NAME")
if !ok || serviceName == "" {
errs = append(errs, "config: RTSA_SERVICE_NAME is required")
}
cfg.ServiceName = serviceName

if len(errs) > 0 {
return nil, fmt.Errorf("%s", strings.Join(errs, "; "))
}

// Optional fields with defaults
cfg.ServiceVersion = getEnvOrDefault("RTSA_SERVICE_VERSION", "dev")
cfg.Environment = getEnvOrDefault("RTSA_ENVIRONMENT", "development")

// gRPC port
portStr := getEnvOrDefault("RTSA_GRPC_PORT", "50051")
port, err := strconv.Atoi(portStr)
if err != nil {
return nil, fmt.Errorf("config: invalid port value for RTSA_GRPC_PORT: %w", err)
}
cfg.GRPCPort = port

cfg.TLSCert = getEnvOrDefault("RTSA_TLS_CERT", "/certs/server.crt")
cfg.TLSKey = getEnvOrDefault("RTSA_TLS_KEY", "/certs/server.key")
cfg.TLSCA = getEnvOrDefault("RTSA_TLS_CA", "/certs/ca.crt")

// Redpanda brokers (comma-separated)
brokersStr := getEnvOrDefault("RTSA_REDPANDA_BROKERS", "localhost:9092")
cfg.RedpandaBrokers = splitCSV(brokersStr)

tlsEnabled := getEnvOrDefault("RTSA_REDPANDA_TLS_ENABLED", "true")
cfg.RedpandaTLSEnabled = strings.EqualFold(tlsEnabled, "true")

cfg.ClickHouseDSN = getEnvOrDefault("RTSA_CLICKHOUSE_DSN",
"clickhouse://default:@localhost:9440/rtsa?secure=true")

cfg.OTelEndpoint = getEnvOrDefault("RTSA_OTEL_ENDPOINT", "localhost:4317")
cfg.LogLevel = getEnvOrDefault("RTSA_LOG_LEVEL", "info")
cfg.LogFormat = getEnvOrDefault("RTSA_LOG_FORMAT", "json")
cfg.MaxClassification = getEnvOrDefault("RTSA_MAX_CLASSIFICATION", "UNCLASSIFIED")

return cfg, nil
}

// MustLoad calls Load and panics on error. Use only in main.go.
func MustLoad() *BaseConfig {
cfg, err := Load()
if err != nil {
panic(fmt.Sprintf("config: failed to load configuration: %v", err))
}
return cfg
}

// LoadInto loads environment variables into a user-provided struct using reflection.
// Struct fields must be tagged with `env:"RTSA_VAR_NAME"` and optionally `envDefault:"value"`.
// Use `envRequired:"true"` for mandatory fields.
func LoadInto(cfg interface{}) error {
v := reflect.ValueOf(cfg)
if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
return fmt.Errorf("config: LoadInto requires a pointer to a struct")
}
v = v.Elem()
t := v.Type()

var errs []string
for i := 0; i < t.NumField(); i++ {
field := t.Field(i)
fieldVal := v.Field(i)

envKey, ok := field.Tag.Lookup("env")
if !ok {
continue
}

required := field.Tag.Get("envRequired") == "true"
defaultVal := field.Tag.Get("envDefault")

raw, found := os.LookupEnv(envKey)
if !found || raw == "" {
if required {
errs = append(errs, fmt.Sprintf("config: %s is required", envKey))
continue
}
raw = defaultVal
}

if err := setField(fieldVal, raw); err != nil {
errs = append(errs, fmt.Sprintf("config: field %s: %v", envKey, err))
}
}

if len(errs) > 0 {
return fmt.Errorf("%s", strings.Join(errs, "; "))
}
return nil
}

func setField(v reflect.Value, raw string) error {
switch v.Kind() {
case reflect.String:
v.SetString(raw)
case reflect.Int, reflect.Int64:
n, err := strconv.ParseInt(raw, 10, 64)
if err != nil {
return fmt.Errorf("invalid integer: %w", err)
}
v.SetInt(n)
case reflect.Bool:
b, err := strconv.ParseBool(raw)
if err != nil {
return fmt.Errorf("invalid boolean: %w", err)
}
v.SetBool(b)
case reflect.Float64:
f, err := strconv.ParseFloat(raw, 64)
if err != nil {
return fmt.Errorf("invalid float: %w", err)
}
v.SetFloat(f)
case reflect.Slice:
parts := splitCSV(raw)
slice := reflect.MakeSlice(v.Type(), len(parts), len(parts))
for i, p := range parts {
slice.Index(i).SetString(p)
}
v.Set(slice)
default:
return fmt.Errorf("unsupported field type: %s", v.Kind())
}
return nil
}

func getEnvOrDefault(key, defaultVal string) string {
if val, ok := os.LookupEnv(key); ok && val != "" {
return val
}
return defaultVal
}

func splitCSV(s string) []string {
parts := strings.Split(s, ",")
result := make([]string, 0, len(parts))
for _, p := range parts {
p = strings.TrimSpace(p)
if p != "" {
result = append(result, p)
}
}
return result
}

// LoadBase loads a BaseConfig from environment variables using safe defaults.
// RTSA_SERVICE_NAME is optional here (defaults to "unknown").
func LoadBase() BaseConfig {
return BaseConfig{
ServiceName:    GetEnv("RTSA_SERVICE_NAME", "unknown"),
ServiceVersion: GetEnv("RTSA_SERVICE_VERSION", "dev"),
Environment:    GetEnv("RTSA_ENVIRONMENT", "development"),
GRPCPort:       GetEnvInt("RTSA_GRPC_PORT", 50051),
TLSCert:        GetEnv("RTSA_TLS_CERT", "/certs/server.crt"),
TLSKey:         GetEnv("RTSA_TLS_KEY", "/certs/server.key"),
TLSCA:          GetEnv("RTSA_TLS_CA", "/certs/ca.crt"),
RedpandaBrokers: GetEnvStrSlice("RTSA_REDPANDA_BROKERS", []string{"localhost:9092"}),
RedpandaTLSEnabled: strings.EqualFold(GetEnv("RTSA_REDPANDA_TLS_ENABLED", "true"), "true"),
OTelEndpoint:      GetEnv("RTSA_OTEL_ENDPOINT", "localhost:4317"),
LogLevel:          GetEnv("RTSA_LOG_LEVEL", "info"),
LogFormat:         GetEnv("RTSA_LOG_FORMAT", "json"),
MaxClassification: GetEnv("RTSA_MAX_CLASSIFICATION", "UNCLASSIFIED"),
}
}

// GetEnv returns the value of the environment variable or defaultVal if not set.
func GetEnv(key, defaultVal string) string {
return getEnvOrDefault(key, defaultVal)
}

// GetEnvInt returns the integer value of the environment variable or defaultVal if not set or invalid.
func GetEnvInt(key string, defaultVal int) int {
raw := getEnvOrDefault(key, "")
if raw == "" {
return defaultVal
}
n, err := strconv.Atoi(raw)
if err != nil {
return defaultVal
}
return n
}

// GetEnvFloat returns the float64 value of the environment variable or defaultVal if not set or invalid.
func GetEnvFloat(key string, defaultVal float64) float64 {
raw := getEnvOrDefault(key, "")
if raw == "" {
return defaultVal
}
f, err := strconv.ParseFloat(raw, 64)
if err != nil {
return defaultVal
}
return f
}

// GetEnvStrSlice returns a comma-separated list of strings from the environment variable,
// or defaultVal if not set.
func GetEnvStrSlice(key string, defaultVal []string) []string {
raw := getEnvOrDefault(key, "")
if raw == "" {
return defaultVal
}
return splitCSV(raw)
}
