// CLASSIFICATION: UNCLASSIFIED
//go:build integration

package testutil

import (
"context"
"testing"

"github.com/testcontainers/testcontainers-go"
"github.com/testcontainers/testcontainers-go/wait"
)

// StartClickHouse starts a ClickHouse container for integration tests.
// Returns DSN and cleanup function.
func StartClickHouse(t *testing.T) (dsn string, cleanup func()) {
t.Helper()
ctx := context.Background()

req := testcontainers.ContainerRequest{
Image:        "clickhouse/clickhouse-server:24.3",
ExposedPorts: []string{"9000/tcp"},
WaitingFor:   wait.ForLog("Ready for connections"),
}

container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
ContainerRequest: req,
Started:          true,
})
if err != nil {
t.Fatalf("testutil: start clickhouse container: %v", err)
}

port, err := container.MappedPort(ctx, "9000")
if err != nil {
container.Terminate(ctx)
t.Fatalf("testutil: get clickhouse port: %v", err)
}

dsn = "clickhouse://default:@localhost:" + port.Port() + "/default"
return dsn, func() {
container.Terminate(ctx)
}
}

// ApplySchema applies a SQL schema to a ClickHouse instance.
func ApplySchema(t *testing.T, dsn string) {
t.Helper()
// Schema application would be done via the ClickHouse driver.
// Implementation depends on the schema files in deploy/clickhouse/
t.Log("testutil: ApplySchema called (implementation pending schema files)")
}
