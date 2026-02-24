// CLASSIFICATION: UNCLASSIFIED
package consumer

import (
"context"

"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/redpanda"
)

// mockMessageConsumer is a test double for redpanda.MessageConsumer.
type mockMessageConsumer struct {
messageData []byte
closed      bool
}

func (m *mockMessageConsumer) Consume(_ context.Context, topics []string, handler redpanda.MessageHandler) error {
if m.messageData != nil {
msg := &redpanda.Message{
Topic: topics[0],
Value: m.messageData,
}
return handler(context.Background(), msg)
}
return nil
}

func (m *mockMessageConsumer) Close() error {
m.closed = true
return nil
}
