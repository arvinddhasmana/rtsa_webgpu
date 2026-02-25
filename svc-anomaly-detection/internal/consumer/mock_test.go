// CLASSIFICATION: UNCLASSIFIED
package consumer

import "context"

// mockMessageConsumer is a test double for MessageConsumer.
type mockMessageConsumer struct {
	messageData []byte
	closed      bool
}

func (m *mockMessageConsumer) Consume(_ context.Context, topics []string, handler MessageHandler) error {
	if m.messageData != nil {
		msg := &Message{
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
