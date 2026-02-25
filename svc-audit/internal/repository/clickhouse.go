// CLASSIFICATION: UNCLASSIFIED
package repository

import (
"context"
"fmt"

clickhouse "github.com/ClickHouse/clickhouse-go/v2"
)

// ClickHouseClient wraps a ClickHouse connection.
type ClickHouseClient struct {
conn clickhouse.Conn
}

// NewClickHouseClient creates a new ClickHouse client using the provided DSN.
func NewClickHouseClient(dsn string) (*ClickHouseClient, error) {
opts, err := clickhouse.ParseDSN(dsn)
if err != nil {
return nil, fmt.Errorf("repository: parse clickhouse DSN: %w", err)
}
conn, err := clickhouse.Open(opts)
if err != nil {
return nil, fmt.Errorf("repository: open clickhouse connection: %w", err)
}
return &ClickHouseClient{conn: conn}, nil
}

// Ping verifies the ClickHouse connection is alive.
func (c *ClickHouseClient) Ping(ctx context.Context) error {
return c.conn.Ping(ctx)
}

// Close closes the ClickHouse connection.
func (c *ClickHouseClient) Close() error {
return c.conn.Close()
}

// Conn returns the underlying ClickHouse connection.
func (c *ClickHouseClient) Conn() clickhouse.Conn {
return c.conn
}
