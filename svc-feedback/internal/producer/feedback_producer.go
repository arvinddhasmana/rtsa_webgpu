// CLASSIFICATION: UNCLASSIFIED

// Package producer provides message production to Redpanda (Kafka-protocol)
// topics using the franz-go client library.
package producer

import (
	"context"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// MessageProducer is the interface for producing protobuf-encoded messages
// to a named topic. Implementations must be safe for concurrent use.
type MessageProducer interface {
// Produce serialises msg and publishes it to the configured topic.
// Returns an error if serialisation or publishing fails.
Produce(ctx context.Context, key string, msg proto.Message) error
// Close flushes pending messages and releases resources.
Close() error
}

// FeedbackProducer is a MessageProducer backed by a franz-go Kafka client.
type FeedbackProducer struct {
client *kgo.Client
topic  string
}

// NewFeedbackProducer constructs a FeedbackProducer connected to the given
// brokers that publishes to the specified topic.
// No authentication credentials are accepted as parameters; TLS and SASL
// must be configured via the opts variadic parameter for production use.
func NewFeedbackProducer(brokers []string, topic string, opts ...kgo.Opt) (*FeedbackProducer, error) {
baseOpts := []kgo.Opt{
kgo.SeedBrokers(brokers...),
kgo.DefaultProduceTopic(topic),
}
baseOpts = append(baseOpts, opts...)

client, err := kgo.NewClient(baseOpts...)
if err != nil {
return nil, fmt.Errorf("[producer.NewFeedbackProducer(%s)]: %w", topic, err)
}

return &FeedbackProducer{client: client, topic: topic}, nil
}

// Produce serialises the protobuf message as JSON and publishes it synchronously.
// Uses protojson (UseProtoNames=true) so downstream wasm-transforms and
// Redpanda Connect Bloblang pipelines can parse field names without a schema registry.
// The key is used as the Kafka record key for partition routing.
func (p *FeedbackProducer) Produce(ctx context.Context, key string, msg proto.Message) error {
payload, err := protojson.MarshalOptions{UseProtoNames: true}.Marshal(msg)
if err != nil {
return fmt.Errorf("[producer.FeedbackProducer.Produce]: marshal: %w", err)
}

record := &kgo.Record{
Topic: p.topic,
Key:   []byte(key),
Value: payload,
}

results := p.client.ProduceSync(ctx, record)
if err := results.FirstErr(); err != nil {
return fmt.Errorf("[producer.FeedbackProducer.Produce(%s)]: %w", p.topic, err)
}

return nil
}

// Close flushes pending writes and closes the Kafka client.
func (p *FeedbackProducer) Close() error {
p.client.Close()
return nil
}

// MockProducer is a thread-safe in-memory MessageProducer for unit testing.
// It records every produced message so tests can assert on published events.
type MockProducer struct {
Messages []MockMessage
Err      error // if non-nil, Produce returns this error
}

// MockMessage records a single Produce invocation.
type MockMessage struct {
Key string
Msg proto.Message
}

// Produce records the message or returns MockProducer.Err.
func (m *MockProducer) Produce(_ context.Context, key string, msg proto.Message) error {
if m.Err != nil {
return m.Err
}
m.Messages = append(m.Messages, MockMessage{Key: key, Msg: msg})
return nil
}

// Close is a no-op for the mock.
func (m *MockProducer) Close() error { return nil }
