// CLASSIFICATION: UNCLASSIFIED
//go:build integration

package testutil

import (
"context"
"testing"

"github.com/testcontainers/testcontainers-go"
"github.com/testcontainers/testcontainers-go/wait"
"github.com/twmb/franz-go/pkg/kadm"
"github.com/twmb/franz-go/pkg/kgo"
)

// StartRedpanda starts a Redpanda container for integration tests.
// Returns broker address (e.g., "localhost:19092") and cleanup function.
func StartRedpanda(t *testing.T) (brokers string, cleanup func()) {
t.Helper()

ctx := context.Background()
req := testcontainers.ContainerRequest{
Image:        "redpandadata/redpanda:v24.1.1",
ExposedPorts: []string{"19092/tcp"},
Cmd: []string{
"redpanda", "start",
"--overprovisioned",
"--smp", "1",
"--memory", "512M",
"--reserve-memory", "0M",
"--node-id", "0",
"--check=false",
"--kafka-addr", "0.0.0.0:19092",
"--advertise-kafka-addr", "localhost:19092",
},
WaitingFor: wait.ForLog("Successfully started Redpanda!"),
}

container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
ContainerRequest: req,
Started:          true,
})
if err != nil {
t.Fatalf("testutil: start redpanda container: %v", err)
}

mappedPort, err := container.MappedPort(ctx, "19092")
if err != nil {
container.Terminate(ctx)
t.Fatalf("testutil: get redpanda port: %v", err)
}

broker := "localhost:" + mappedPort.Port()
return broker, func() {
container.Terminate(ctx)
}
}

// CreateTopics creates the specified topics in the test Redpanda instance.
func CreateTopics(t *testing.T, brokers string, topics ...string) {
t.Helper()
client, err := kgo.NewClient(kgo.SeedBrokers(brokers))
if err != nil {
t.Fatalf("testutil: create kgo client: %v", err)
}
defer client.Close()

admin := kadm.NewClient(client)
for _, topic := range topics {
_, err := admin.CreateTopic(context.Background(), 1, 1, nil, topic)
if err != nil {
t.Logf("testutil: create topic %s: %v (may already exist)", topic, err)
}
}
}
