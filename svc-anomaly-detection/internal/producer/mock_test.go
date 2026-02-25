// CLASSIFICATION: UNCLASSIFIED
package producer

import (
"context"
)

// mockProducer is a test double for redpanda.MessageProducer.
type mockProducer struct {
topic   string
key     []byte
value   []byte
headers map[string]string
closed  bool
err     error
}

func (m *mockProducer) Produce(_ context.Context, topic string, key, value []byte, headers map[string]string) error {
if m.err != nil {
return m.err
}
m.topic = topic
m.key = key
m.value = value
m.headers = headers
return nil
}

func (m *mockProducer) Close() error {
m.closed = true
return nil
}
