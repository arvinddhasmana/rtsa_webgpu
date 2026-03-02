// CLASSIFICATION: UNCLASSIFIED
package redpanda

import (
"context"
"fmt"
"time"

"github.com/twmb/franz-go/pkg/kgo"
)

// RedpandaProducer defines the interface for producing messages.
type RedpandaProducer interface {
	Produce(ctx context.Context, topic string, key []byte, value []byte, classification string, traceID string, extraHeaders ...kgo.RecordHeader) error
	Close() error
	Healthy(ctx context.Context) bool
}

// Producer wraps franz-go client for producing messages to Redpanda.
type Producer struct {
	client      *kgo.Client
	serviceName string
	schemaVer   string
}

// ProducerConfig configures a new Producer.
type ProducerConfig struct {
Connection   ConnectionOptions
ServiceName  string
SchemaVersion string
MaxRetries   int
Compression  string
RequiredAcks string
}

// NewProducer creates a Redpanda producer with standard configuration.
func NewProducer(ctx context.Context, cfg ProducerConfig) (*Producer, error) {
kgoOpts, err := cfg.Connection.BuildKgoOpts()
if err != nil {
return nil, fmt.Errorf("redpanda: build connection opts: %w", err)
}

// Set compression
switch cfg.Compression {
case "snappy":
kgoOpts = append(kgoOpts, kgo.ProducerBatchCompression(kgo.SnappyCompression()))
case "lz4":
kgoOpts = append(kgoOpts, kgo.ProducerBatchCompression(kgo.Lz4Compression()))
default:
kgoOpts = append(kgoOpts, kgo.ProducerBatchCompression(kgo.ZstdCompression()))
}

// Required acks
switch cfg.RequiredAcks {
case "leader":
kgoOpts = append(kgoOpts, kgo.RequiredAcks(kgo.LeaderAck()))
case "none":
kgoOpts = append(kgoOpts, kgo.RequiredAcks(kgo.NoAck()))
default:
kgoOpts = append(kgoOpts, kgo.RequiredAcks(kgo.AllISRAcks()))
}

client, err := kgo.NewClient(kgoOpts...)
if err != nil {
return nil, fmt.Errorf("redpanda: create producer client: %w", err)
}

// Verify connectivity
if err := client.Ping(ctx); err != nil {
client.Close()
return nil, fmt.Errorf("redpanda: producer ping failed: %w", err)
}

return &Producer{
client:      client,
serviceName: cfg.ServiceName,
schemaVer:   cfg.SchemaVersion,
}, nil
}

// Produce sends a single message to the specified topic.
func (p *Producer) Produce(ctx context.Context, topic string, key []byte, value []byte,
classification string, traceID string, extraHeaders ...kgo.RecordHeader) error {

headers := StandardHeaders(classification, p.serviceName, traceID, p.schemaVer)
headers = append(headers, extraHeaders...)

record := &kgo.Record{
Topic:   topic,
Key:     key,
Value:   value,
Headers: headers,
}

results := p.client.ProduceSync(ctx, record)
if err := results.FirstErr(); err != nil {
return fmt.Errorf("redpanda: produce to topic %s: %w", topic, err)
}
return nil
}

// ProduceBatch sends multiple messages.
func (p *Producer) ProduceBatch(ctx context.Context, records []*kgo.Record) error {
results := p.client.ProduceSync(ctx, records...)
if err := results.FirstErr(); err != nil {
return fmt.Errorf("redpanda: produce batch: %w", err)
}
return nil
}

// Close flushes pending messages and closes the producer.
func (p *Producer) Close() error {
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

if err := p.client.Flush(ctx); err != nil {
return fmt.Errorf("redpanda: flush on close: %w", err)
}
p.client.Close()
return nil
}

// Healthy returns true if the producer can reach brokers.
func (p *Producer) Healthy(ctx context.Context) bool {
return p.client.Ping(ctx) == nil
}
