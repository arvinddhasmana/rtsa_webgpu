// CLASSIFICATION: UNCLASSIFIED
package config_test

import (
"os"
"testing"

"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/config"
)

func TestLoad_AllVarsSet(t *testing.T) {
t.Setenv("RTSA_SERVICE_NAME", "test-service")
t.Setenv("RTSA_SERVICE_VERSION", "1.0.0")
t.Setenv("RTSA_ENVIRONMENT", "test")
t.Setenv("RTSA_GRPC_PORT", "50052")

cfg, err := config.Load()
if err != nil {
t.Fatalf("expected no error, got: %v", err)
}
if cfg.ServiceName != "test-service" {
t.Errorf("expected ServiceName=test-service, got %s", cfg.ServiceName)
}
if cfg.GRPCPort != 50052 {
t.Errorf("expected GRPCPort=50052, got %d", cfg.GRPCPort)
}
}

func TestLoad_MissingRequired(t *testing.T) {
os.Unsetenv("RTSA_SERVICE_NAME")
_, err := config.Load()
if err == nil {
t.Fatal("expected error for missing RTSA_SERVICE_NAME")
}
}

func TestLoad_Defaults(t *testing.T) {
t.Setenv("RTSA_SERVICE_NAME", "svc")
os.Unsetenv("RTSA_GRPC_PORT")
os.Unsetenv("RTSA_LOG_LEVEL")

cfg, err := config.Load()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if cfg.GRPCPort != 50051 {
t.Errorf("expected default port 50051, got %d", cfg.GRPCPort)
}
if cfg.LogLevel != "info" {
t.Errorf("expected default log level 'info', got %s", cfg.LogLevel)
}
}

func TestLoad_CommaSeparatedBrokers(t *testing.T) {
t.Setenv("RTSA_SERVICE_NAME", "svc")
t.Setenv("RTSA_REDPANDA_BROKERS", "a:9092,b:9092")

cfg, err := config.Load()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if len(cfg.RedpandaBrokers) != 2 {
t.Errorf("expected 2 brokers, got %d", len(cfg.RedpandaBrokers))
}
}

func TestLoad_InvalidPort(t *testing.T) {
t.Setenv("RTSA_SERVICE_NAME", "svc")
t.Setenv("RTSA_GRPC_PORT", "abc")

_, err := config.Load()
if err == nil {
t.Fatal("expected error for invalid port")
}
}

func TestLoadInto_CustomStruct(t *testing.T) {
type MyConfig struct {
Name    string `env:"RTSA_TEST_NAME" envDefault:"default-name"`
Count   int    `env:"RTSA_TEST_COUNT" envDefault:"5"`
Enabled bool   `env:"RTSA_TEST_ENABLED" envDefault:"true"`
}

t.Setenv("RTSA_TEST_NAME", "my-service")
t.Setenv("RTSA_TEST_COUNT", "42")

var cfg MyConfig
if err := config.LoadInto(&cfg); err != nil {
t.Fatalf("unexpected error: %v", err)
}
if cfg.Name != "my-service" {
t.Errorf("expected Name=my-service, got %s", cfg.Name)
}
if cfg.Count != 42 {
t.Errorf("expected Count=42, got %d", cfg.Count)
}
}

func TestLoad_TLSEnabled(t *testing.T) {
t.Setenv("RTSA_SERVICE_NAME", "svc")
t.Setenv("RTSA_REDPANDA_TLS_ENABLED", "false")

cfg, err := config.Load()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if cfg.RedpandaTLSEnabled {
t.Error("expected TLS disabled")
}
}

func TestLoad_MaxClassification(t *testing.T) {
t.Setenv("RTSA_SERVICE_NAME", "svc")
t.Setenv("RTSA_MAX_CLASSIFICATION", "SECRET")

cfg, err := config.Load()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if cfg.MaxClassification != "SECRET" {
t.Errorf("expected SECRET, got %s", cfg.MaxClassification)
}
}

func TestLoadInto_RequiredMissing(t *testing.T) {
type MyConfig struct {
Name string `env:"RTSA_REQUIRED_FIELD" envRequired:"true"`
}
os.Unsetenv("RTSA_REQUIRED_FIELD")
var cfg MyConfig
err := config.LoadInto(&cfg)
if err == nil {
t.Fatal("expected error for missing required field")
}
}

func TestLoadInto_NotPointerToStruct(t *testing.T) {
err := config.LoadInto("not-a-struct")
if err == nil {
t.Fatal("expected error for non-pointer-to-struct")
}
}

func TestLoadInto_BoolField(t *testing.T) {
type MyConfig struct {
Enabled bool `env:"RTSA_TEST_BOOL" envDefault:"true"`
}
t.Setenv("RTSA_TEST_BOOL", "true")
var cfg MyConfig
if err := config.LoadInto(&cfg); err != nil {
t.Fatalf("unexpected error: %v", err)
}
if !cfg.Enabled {
t.Error("expected Enabled=true")
}
}

func TestLoadInto_Float64Field(t *testing.T) {
type MyConfig struct {
Speed float64 `env:"RTSA_TEST_SPEED" envDefault:"100.5"`
}
t.Setenv("RTSA_TEST_SPEED", "99.9")
var cfg MyConfig
if err := config.LoadInto(&cfg); err != nil {
t.Fatalf("unexpected error: %v", err)
}
if cfg.Speed != 99.9 {
t.Errorf("expected 99.9, got %f", cfg.Speed)
}
}

func TestLoadInto_Int64Field(t *testing.T) {
type MyConfig struct {
Count int64 `env:"RTSA_TEST_INT64" envDefault:"42"`
}
t.Setenv("RTSA_TEST_INT64", "1000")
var cfg MyConfig
if err := config.LoadInto(&cfg); err != nil {
t.Fatalf("unexpected error: %v", err)
}
if cfg.Count != 1000 {
t.Errorf("expected 1000, got %d", cfg.Count)
}
}

func TestLoadInto_SliceField(t *testing.T) {
type MyConfig struct {
Brokers []string `env:"RTSA_TEST_BROKERS" envDefault:"a:9092,b:9092"`
}
t.Setenv("RTSA_TEST_BROKERS", "x:9092,y:9092,z:9092")
var cfg MyConfig
if err := config.LoadInto(&cfg); err != nil {
t.Fatalf("unexpected error: %v", err)
}
if len(cfg.Brokers) != 3 {
t.Errorf("expected 3 brokers, got %d", len(cfg.Brokers))
}
}

func TestLoadInto_InvalidBool(t *testing.T) {
type MyConfig struct {
Enabled bool `env:"RTSA_TEST_INVALID_BOOL"`
}
t.Setenv("RTSA_TEST_INVALID_BOOL", "notabool")
var cfg MyConfig
err := config.LoadInto(&cfg)
if err == nil {
t.Fatal("expected error for invalid bool")
}
}
