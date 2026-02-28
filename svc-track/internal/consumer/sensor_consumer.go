// CLASSIFICATION: UNCLASSIFIED
package consumer

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/encoding/protojson"
)

var SensorTopics = []string{
	"sensors.radar.tracks",
	"sensors.ew.intercepts",
	"sensors.elint.detections",
	"sensors.isr.observations",
	"sensors.ais.positions",
	"sensors.cyber.iocs",
}

type SensorObservationConsumer struct {
	client *kgo.Client
	logger *slog.Logger

	mu          sync.RWMutex
	subscribers map[uint64]chan *ingestionv1.SensorObservation
	subCounter  uint64
}

func NewSensorObservationConsumer(brokers []string, groupID string, logger *slog.Logger) (*SensorObservationConsumer, error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(groupID),
		kgo.ConsumeTopics(SensorTopics...),
		kgo.WithLogger(newSlogKgoLogger(logger, kgo.LogLevelWarn)),
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("consumer.NewSensorObservationConsumer: kgo.NewClient: %w", err)
	}

	return &SensorObservationConsumer{
		client:      client,
		logger:      logger,
		subscribers: make(map[uint64]chan *ingestionv1.SensorObservation),
	}, nil
}

func (c *SensorObservationConsumer) Subscribe() (uint64, <-chan *ingestionv1.SensorObservation) {
	c.mu.Lock()
	defer c.mu.Unlock()

	id := c.subCounter
	c.subCounter++
	ch := make(chan *ingestionv1.SensorObservation, 1000)
	c.subscribers[id] = ch
	return id, ch
}

func (c *SensorObservationConsumer) Unsubscribe(id uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ch, ok := c.subscribers[id]; ok {
		close(ch)
		delete(c.subscribers, id)
	}
}

func (c *SensorObservationConsumer) Run(ctx context.Context) error {
	c.logger.InfoContext(ctx, "sensor observation consumer starting", slog.Any("topics", SensorTopics))

	for {
		fetches := c.client.PollFetches(ctx)
		if fetches.IsClientClosed() {
			return nil
		}

		if err := ctx.Err(); err != nil {
			return nil
		}

		fetches.EachRecord(func(record *kgo.Record) {
			var obs ingestionv1.SensorObservation
			if err := protojson.Unmarshal(record.Value, &obs); err != nil {
				return
			}

			c.mu.RLock()
			for _, ch := range c.subscribers {
				select {
				case ch <- &obs:
				default:
				}
			}
			c.mu.RUnlock()
		})
	}
}

func (c *SensorObservationConsumer) Close() {
	c.client.Close()
	c.mu.Lock()
	defer c.mu.Unlock()
	for id, ch := range c.subscribers {
		close(ch)
		delete(c.subscribers, id)
	}
}
