// CLASSIFICATION: UNCLASSIFIED
package repository

import (
	"context"
	"fmt"
	"time"

	auditv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/audit/v1"
	"github.com/ClickHouse/clickhouse-go/v2"
)

// AuditRepository provides append-only write access to the audit_log table.
//
// IMMUTABILITY CONTRACT:
//   - The only write operation is BatchInsert (INSERT INTO)
//   - No UPDATE, DELETE, ALTER, or TRUNCATE operations exist in this package
type AuditRepository struct {
	conn clickhouse.Conn
}

// NewAuditRepository creates a new AuditRepository using the provided DSN.
func NewAuditRepository(dsn string) (*AuditRepository, error) {
	opts, err := clickhouse.ParseDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("audit_repo: parse DSN: %w", err)
	}
	conn, err := clickhouse.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("audit_repo: open connection: %w", err)
	}
	return &AuditRepository{conn: conn}, nil
}

// NewAuditRepositoryFromConn creates a new AuditRepository from an existing connection.
func NewAuditRepositoryFromConn(conn clickhouse.Conn) *AuditRepository {
	return &AuditRepository{conn: conn}
}

// Ping checks the ClickHouse connection.
func (r *AuditRepository) Ping(ctx context.Context) error {
	return r.conn.Ping(ctx)
}

// Close closes the ClickHouse connection.
func (r *AuditRepository) Close() error {
	return r.conn.Close()
}

// BatchInsert inserts a batch of audit events into the audit_log table.
// Only INSERT operations are used — immutability is enforced.
// Duplicate audit_ids are silently ignored (idempotent via ClickHouse deduplication).
func (r *AuditRepository) BatchInsert(ctx context.Context, events []*auditv1.AuditEvent) error {
	if len(events) == 0 {
		return nil
	}

	batch, err := r.conn.PrepareBatch(ctx,
		`INSERT INTO audit_log (audit_id, service_id, event_type, actor_id, actor_type,
			resource_type, resource_id, action, detail_json, classification_level, event_time, trace_id)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("audit_repo: prepare batch: %w", err)
	}

	for _, event := range events {
		var eventTime time.Time
		if event.EventTime != nil {
			eventTime = event.EventTime.AsTime()
		}
		if err := batch.Append(
			event.AuditId,
			event.ServiceId,
			event.EventType,
			event.ActorId,
			event.ActorType.String(),
			event.ResourceType,
			event.ResourceId,
			event.Action.String(),
			event.DetailJson,
			int32(event.ClassificationLevel),
			eventTime,
			event.TraceId,
		); err != nil {
			return fmt.Errorf("audit_repo: append batch row: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("audit_repo: send batch: %w", err)
	}
	return nil
}
