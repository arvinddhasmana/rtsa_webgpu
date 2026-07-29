// CLASSIFICATION: UNCLASSIFIED
package consumer

import (
	"log/slog"
	"os"
	"testing"
	"time"

	commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
	ingestionv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/ingestion/v1"
)

func TestSensorObservationConsumer_SubscribeUnsubscribe(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	// We don't need a real Kafka client just to test subscribe/unsubscribe logic.
	// But since NewSensorObservationConsumer tries to connect, we manually construct
	// the struct for this test.
	c := &SensorObservationConsumer{
		logger:      logger,
		subscribers: make(map[uint64]chan *ingestionv1.SensorObservation),
	}

	id1, ch1 := c.Subscribe()
	id2, ch2 := c.Subscribe()

	c.mu.RLock()
	if len(c.subscribers) != 2 {
		t.Errorf("expected 2 subscribers, got %d", len(c.subscribers))
	}
	c.mu.RUnlock()

	if id1 == id2 {
		t.Errorf("expected different IDs, got %d and %d", id1, id2)
	}

	c.Unsubscribe(id1)

	c.mu.RLock()
	if len(c.subscribers) != 1 {
		t.Errorf("expected 1 subscriber after unsubscribe, got %d", len(c.subscribers))
	}
	if _, ok := c.subscribers[id1]; ok {
		t.Errorf("expected subscriber %d to be removed", id1)
	}
	c.mu.RUnlock()

	// Ensure channel was closed
	select {
	case _, ok := <-ch1:
		if ok {
			t.Error("expected channel to be closed, but it was open")
		}
	default:
		t.Error("expected channel to be closed (and yield immediately), but it blocked")
	}

	// Unsubscribe second client
	c.Unsubscribe(id2)

	// Ensure second channel was closed
	select {
	case _, ok := <-ch2:
		if ok {
			t.Error("expected ch2 to be closed, but it was open")
		}
	default:
		t.Error("expected ch2 to be closed, but it blocked")
	}
}

// simulatePoll Fetches simulates a Kafka poll loop by directly injecting a proto message to the subscribers.
func TestSensorObservationConsumer_PublishToSubscribers(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	c := &SensorObservationConsumer{
		logger:      logger,
		subscribers: make(map[uint64]chan *ingestionv1.SensorObservation),
	}

	id, ch := c.Subscribe()
	defer c.Unsubscribe(id)

	obs := &ingestionv1.SensorObservation{
		ObservationId: "obs-001",
		SensorId:      "radar-1",
		SensorType:    commonv1.SensorType_SENSOR_TYPE_RADAR,
	}

	// Manually inject the message (simulating the inner loop of Run)
	c.mu.RLock()
	for _, subCh := range c.subscribers {
		subCh <- obs
	}
	c.mu.RUnlock()

	select {
	case msg := <-ch:
		if msg.ObservationId != "obs-001" {
			t.Errorf("expected obs-001, got %s", msg.ObservationId)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for message")
	}
}

func TestSensorTopics(t *testing.T) {
	expected := []string{
		"sensors.radar.tracks",
		"sensors.ew.intercepts",
		"sensors.elint.detections",
		"sensors.isr.observations",
		"sensors.ais.positions",
		"sensors.cyber.iocs",
	}

	if len(SensorTopics) != len(expected) {
		t.Errorf("expected %d topics, got %d", len(expected), len(SensorTopics))
	}

	for i, topic := range expected {
		if SensorTopics[i] != topic {
			t.Errorf("topic[%d]: expected %q, got %q", i, topic, SensorTopics[i])
		}
	}
}
