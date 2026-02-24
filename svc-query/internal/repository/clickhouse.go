// CLASSIFICATION: UNCLASSIFIED

// Package repository provides ClickHouse query implementations for the query service.
// All queries are parameterized — no user-supplied values are interpolated into SQL strings.
package repository

import (
"context"
"crypto/tls"
"crypto/x509"
"fmt"
"net/url"
"os"
"time"

"github.com/ClickHouse/clickhouse-go/v2"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/config"
)

// ClickHouseClient wraps a ClickHouse connection with lifecycle management.
type ClickHouseClient struct {
conn     clickhouse.Conn
database string
}

// NewClickHouseClient opens and verifies a ClickHouse connection.
// Applies mTLS when certificate files are provided in cfg.
// Falls back to a plain connection when TLS files are absent (dev environment).
func NewClickHouseClient(cfg *config.Config) (*ClickHouseClient, error) {
opts, err := buildOptions(cfg)
if err != nil {
return nil, fmt.Errorf("repository.NewClickHouseClient: build options: %w", err)
}

conn, err := clickhouse.Open(opts)
if err != nil {
return nil, fmt.Errorf("repository.NewClickHouseClient: open connection: %w", err)
}

pingCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

if err := conn.Ping(pingCtx); err != nil {
return nil, fmt.Errorf("repository.NewClickHouseClient: ping: %w", err)
}

return &ClickHouseClient{
conn:     conn,
database: cfg.ClickHouseDatabase,
}, nil
}

// Close releases the ClickHouse connection.
func (c *ClickHouseClient) Close() error {
if err := c.conn.Close(); err != nil {
return fmt.Errorf("repository.ClickHouseClient.Close: %w", err)
}
return nil
}

// buildOptions constructs clickhouse.Options from the service config.
func buildOptions(cfg *config.Config) (*clickhouse.Options, error) {
addr, auth, err := parseDSN(cfg.ClickHouseDSN)
if err != nil {
return nil, fmt.Errorf("buildOptions: parse DSN: %w", err)
}

opts := &clickhouse.Options{
Addr: []string{addr},
Auth: auth,
Settings: clickhouse.Settings{
"max_execution_time": 30,
},
DialTimeout:      15 * time.Second,
MaxOpenConns:     10,
MaxIdleConns:     5,
ConnMaxLifetime:  time.Hour,
ConnOpenStrategy: clickhouse.ConnOpenInOrder,
Compression: &clickhouse.Compression{
Method: clickhouse.CompressionLZ4,
},
}

// Apply mTLS if cert files are configured
if cfg.ClickHouseTLSCertFile != "" && cfg.ClickHouseTLSKeyFile != "" {
tlsCfg, err := buildClickHouseTLS(cfg)
if err != nil {
return nil, fmt.Errorf("buildOptions: TLS config: %w", err)
}
opts.TLS = tlsCfg
}

return opts, nil
}

// parseDSN extracts addr and auth from a clickhouse:// DSN.
func parseDSN(dsn string) (string, clickhouse.Auth, error) {
u, err := url.Parse(dsn)
if err != nil {
return "", clickhouse.Auth{}, fmt.Errorf("parseDSN: invalid URL: %w", err)
}

host := u.Hostname()
port := u.Port()
if port == "" {
port = "9000"
}
addr := host + ":" + port

database := u.Path
if len(database) > 1 {
database = database[1:] // strip leading /
}
if database == "" {
database = "default"
}

user := ""
password := ""
if u.User != nil {
user = u.User.Username()
password, _ = u.User.Password()
}

return addr, clickhouse.Auth{
Database: database,
Username: user,
Password: password,
}, nil
}

// buildClickHouseTLS loads mTLS certificates for ClickHouse.
func buildClickHouseTLS(cfg *config.Config) (*tls.Config, error) {
cert, err := tls.LoadX509KeyPair(cfg.ClickHouseTLSCertFile, cfg.ClickHouseTLSKeyFile)
if err != nil {
return nil, fmt.Errorf("buildClickHouseTLS: load key pair: %w", err)
}

tlsCfg := &tls.Config{
Certificates: []tls.Certificate{cert},
MinVersion:   tls.VersionTLS12,
}

if cfg.ClickHouseTLSCAFile != "" {
caCert, err := os.ReadFile(cfg.ClickHouseTLSCAFile)
if err != nil {
return nil, fmt.Errorf("buildClickHouseTLS: read CA file: %w", err)
}
pool := x509.NewCertPool()
if !pool.AppendCertsFromPEM(caCert) {
return nil, fmt.Errorf("buildClickHouseTLS: failed to parse CA cert")
}
tlsCfg.RootCAs = pool
}

return tlsCfg, nil
}
