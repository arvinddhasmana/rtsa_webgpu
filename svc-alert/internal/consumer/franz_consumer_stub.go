// CLASSIFICATION: UNCLASSIFIED

//go:build !integration

// Package consumer provides a stub FranzConsumerClient for non-integration builds.
// The real implementation (requiring a live Redpanda broker) is in franz_consumer.go
// and is only compiled with the "integration" build tag.
package consumer

import (
"context"
"fmt"
"log/slog"
)

// FranzConsumerClient is a stub that satisfies the ConsumerClient interface
// in non-integration builds. It returns an error if Consume is called,
// directing operators to run with the "integration" build tag for production use.
type FranzConsumerClient struct {
brokers []string
group   string
logger  *slog.Logger
}

// NewFranzConsumerClient creates a stub client. No real connection is attempted.
func NewFranzConsumerClient(brokers []string, group string, logger *slog.Logger) (*FranzConsumerClient, error) {
logger.Warn("using stub FranzConsumerClient; rebuild with -tags integration for production",
"brokers", brokers,
"group", group,
)
return &FranzConsumerClient{brokers: brokers, group: group, logger: logger}, nil
}

// Consume blocks until ctx is cancelled. No messages are delivered.
// In production builds, replace with the real franz-go implementation.
func (f *FranzConsumerClient) Consume(ctx context.Context, topics []string, _ MessageHandler) error {
f.logger.WarnContext(ctx, "stub consumer: no messages will be delivered",
"topics", topics,
"hint", "rebuild with -tags integration")
<-ctx.Done()
return fmt.Errorf("[consumer].[FranzConsumerClient.Consume]: %w (stub; use -tags integration)", ctx.Err())
}

// Close is a no-op for the stub.
func (f *FranzConsumerClient) Close() error {
return nil
}
