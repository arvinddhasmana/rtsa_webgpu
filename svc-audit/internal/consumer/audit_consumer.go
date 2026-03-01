// CLASSIFICATION: UNCLASSIFIED
package consumer

import (
	"context"
	"fmt"
	"sync"
	"time"

	auditv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/audit/v1"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-audit/internal/repository"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
)

const auditTopic = "audit.events"

// AuditConsumer consumes from the audit.events topic and writes
// each record to the audit_log ClickHouse table.
//
// IMMUTABILITY RULE: Only INSERT operations. Never UPDATE, DELETE, or ALTER.
type AuditConsumer struct {
	client      *kgo.Client
	repo        *repository.AuditRepository
	batchSize   int
	flushPeriod time.Duration
	logger      *zap.Logger

	mu     sync.Mutex
	buffer []*auditv1.AuditEvent
}

// NewAuditConsumer creates an AuditConsumer.
func NewAuditConsumer(
	brokers []string,
	consumerGroup string,
	repo *repository.AuditRepository,
	batchSize int,
	flushPeriodSec int,
	logger *zap.Logger,
) (*AuditConsumer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(consumerGroup),
		kgo.ConsumeTopics(auditTopic),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
	)
	if err != nil {
		return nil, fmt.Errorf("audit_consumer: create franz-go client: %w", err)
	}
	return &AuditConsumer{
		client:      client,
		repo:        repo,
		batchSize:   batchSize,
		flushPeriod: time.Duration(flushPeriodSec) * time.Second,
		logger:      logger,
	}, nil
}

// Start begins consuming audit events until ctx is cancelled.
// Flow per message:
//  1. Deserialize AuditEvent protobuf
//  2. Validate required fields
//  3. Buffer the event
//  4. Flush when batch is full or flush timer fires
func (c *AuditConsumer) Start(ctx context.Context) error {
	flushTicker := time.NewTicker(c.flushPeriod)
	defer flushTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Final flush before exit
			if err := c.flush(context.Background()); err != nil {
				c.logger.Error("audit_consumer: final flush failed", zap.Error(err))
			}
			return nil

		case <-flushTicker.C:
			if err := c.flush(ctx); err != nil {
				c.logger.Error("audit_consumer: periodic flush failed", zap.Error(err))
			}

		default:
			fetches := c.client.PollFetches(ctx)
			if fetches.IsClientClosed() {
				return nil
			}
			fetches.EachError(func(t string, p int32, err error) {
				c.logger.Error("audit_consumer: fetch error",
					zap.String("topic", t), zap.Int32("partition", p), zap.Error(err))
			})

			var toFlush bool
			fetches.EachRecord(func(r *kgo.Record) {
				if err := c.processRecord(r); err != nil {
					c.logger.Error("audit_consumer: process record failed", zap.Error(err))
					return
				}
				if len(c.buffer) >= c.batchSize {
					toFlush = true
				}
			})

			if toFlush {
				if err := c.flush(ctx); err != nil {
					c.logger.Error("audit_consumer: batch flush failed", zap.Error(err))
				}
			}
		}
	}
}

// processRecord deserializes and validates a Redpanda record.
func (c *AuditConsumer) processRecord(r *kgo.Record) error {
	var event auditv1.AuditEvent
	// protojson: matches pkg/audit emitter which uses protojson.MarshalOptions{UseProtoNames: true}.
	if err := protojson.Unmarshal(r.Value, &event); err != nil {
		c.logger.Warn("audit_consumer: skipping malformed record",
			zap.String("topic", r.Topic),
			zap.Int64("offset", r.Offset),
			zap.Error(err))
		return nil // skip malformed records — do NOT DLQ audit events
	}

	if err := validateEvent(&event); err != nil {
		c.logger.Warn("audit_consumer: skipping invalid record",
			zap.String("audit_id", event.AuditId), zap.Error(err))
		return nil // skip invalid records
	}

	c.mu.Lock()
	c.buffer = append(c.buffer, &event)
	c.mu.Unlock()
	return nil
}

// flush writes the current buffer to ClickHouse.
func (c *AuditConsumer) flush(ctx context.Context) error {
	c.mu.Lock()
	if len(c.buffer) == 0 {
		c.mu.Unlock()
		return nil
	}
	batch := c.buffer
	c.buffer = nil
	c.mu.Unlock()

	if err := c.repo.BatchInsert(ctx, batch); err != nil {
		return fmt.Errorf("audit_consumer: batch insert: %w", err)
	}
	c.logger.Info("audit_consumer: flushed batch", zap.Int("count", len(batch)))
	return nil
}

// Close shuts down the Kafka client.
func (c *AuditConsumer) Close() {
	c.client.Close()
}

// validateEvent checks required fields on an AuditEvent.
func validateEvent(event *auditv1.AuditEvent) error {
	if event.AuditId == "" {
		return fmt.Errorf("audit_id is required")
	}
	if event.ServiceId == "" {
		return fmt.Errorf("service_id is required")
	}
	if event.EventType == "" {
		return fmt.Errorf("event_type is required")
	}
	if event.EventTime == nil {
		return fmt.Errorf("event_time is required")
	}
	return nil
}
