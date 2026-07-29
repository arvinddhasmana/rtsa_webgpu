// CLASSIFICATION: UNCLASSIFIED
package redpanda_test

import (
"testing"

"github.com/arvinddhasmana/rtsa_webgpu/pkg/redpanda"
)

func TestConnectionOptions_BuildKgoOpts_InvalidSASL(t *testing.T) {
opts := &redpanda.ConnectionOptions{
Brokers:    []string{"localhost:9092"},
TLSEnabled: false,
SASL: &redpanda.SASLConfig{
Mechanism: "UNKNOWN",
Username:  "user",
Password:  "pass",
},
}
_, err := opts.BuildKgoOpts()
if err == nil {
t.Error("expected error for unknown SASL mechanism")
}
}

func TestConnectionOptions_BuildKgoOpts_NoTLS(t *testing.T) {
opts := &redpanda.ConnectionOptions{
Brokers:    []string{"localhost:9092"},
TLSEnabled: false,
ClientID:   "test-client",
}
kgoOpts, err := opts.BuildKgoOpts()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if len(kgoOpts) == 0 {
t.Error("expected at least one kgo option")
}
}
