// CLASSIFICATION: UNCLASSIFIED
package redpanda

import (
"context"
"time"
)

// Message represents a Redpanda/Kafka message.
type Message struct {
Topic     string
Key       []byte
Value     []byte
Headers   map[string]string
Offset    int64
Partition int32
Timestamp time.Time
}

// MessageHandler is a function that processes a consumed message.
type MessageHandler func(ctx context.Context, msg *Message) error

// MessageProducer defines the interface for producing messages to Redpanda.
type MessageProducer interface {
// Produce sends a message to the specified topic.
Produce(ctx context.Context, topic string, key []byte, value []byte, headers map[string]string) error
// Close flushes and closes the producer.
Close() error
}

// MessageConsumer defines the interface for consuming messages from Redpanda.
type MessageConsumer interface {
// Consume starts consuming from the specified topics and calls handler for each message.
Consume(ctx context.Context, topics []string, handler MessageHandler) error
// Close stops the consumer.
Close() error
}
