// CLASSIFICATION: UNCLASSIFIED
package redpanda

import (
"context"
"fmt"

"github.com/twmb/franz-go/pkg/kgo"
)

// Producer is a Redpanda message producer backed by franz-go.
type Producer struct {
client *kgo.Client
}

// NewProducer creates a new Producer connected to the given brokers.
func NewProducer(brokers []string, opts ...kgo.Opt) (*Producer, error) {
defaultOpts := []kgo.Opt{
kgo.SeedBrokers(brokers...),
kgo.ProducerBatchCompression(kgo.SnappyCompression()),
}
defaultOpts = append(defaultOpts, opts...)

client, err := kgo.NewClient(defaultOpts...)
if err != nil {
return nil, fmt.Errorf("[redpanda.NewProducer]: %w", err)
}
return &Producer{client: client}, nil
}

// Produce sends a message to the specified topic.
func (p *Producer) Produce(ctx context.Context, topic string, key []byte, value []byte, headers map[string]string) error {
rec := &kgo.Record{
Topic: topic,
Key:   key,
Value: value,
}
for k, v := range headers {
rec.Headers = append(rec.Headers, kgo.RecordHeader{Key: k, Value: []byte(v)})
}
if err := p.client.ProduceSync(ctx, rec).FirstErr(); err != nil {
return fmt.Errorf("[redpanda.Producer.Produce](%s): %w", topic, err)
}
return nil
}

// Close flushes and closes the producer.
func (p *Producer) Close() error {
p.client.Close()
return nil
}
