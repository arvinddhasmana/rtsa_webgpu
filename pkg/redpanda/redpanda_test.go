// CLASSIFICATION: UNCLASSIFIED
package redpanda_test

import (
	"testing"

	"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/redpanda"
)

func TestNewProducer_NoBrokers(t *testing.T) {
	_, err := redpanda.NewProducer(nil)
	if err == nil {
		t.Error("expected error for empty brokers, got nil")
	}
}

func TestNewConsumer_NoBrokers(t *testing.T) {
	_, err := redpanda.NewConsumer(nil, "grp", []string{"topic"})
	if err == nil {
		t.Error("expected error for empty brokers, got nil")
	}
}
