// CLASSIFICATION: UNCLASSIFIED
package repository

import (
	"context"
	"fmt"
	"time"

	auditv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/audit/v1"
	commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
	"github.com/arvinddhasmana/rtsa_webgpu/svc-audit/internal/domain"
	"github.com/arvinddhasmana/rtsa_webgpu/svc-audit/internal/mapper"
	"github.com/arvinddhasmana/rtsa_webgpu/svc-audit/internal/security"
	"github.com/ClickHouse/clickhouse-go/v2"
	chdriver "github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

// AuditRepository provides append-only write and read access to the audit_log table.
//
// IMMUTABILITY CONTRACT:
//   - The only write operation is BatchInsert (INSERT INTO)
//   - No UPDATE, DELETE, ALTER, or TRUNCATE operations exist in this package
type AuditRepository struct {
	conn   clickhouse.Conn
	filter *security.ClassificationFilter
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
	return &AuditRepository{conn: conn, filter: &security.ClassificationFilter{}}, nil
}

// NewAuditRepositoryFromConn creates a new AuditRepository from an existing connection.
func NewAuditRepositoryFromConn(conn clickhouse.Conn) *AuditRepository {
	return &AuditRepository{conn: conn, filter: &security.ClassificationFilter{}}
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
		resource_type, resource_id, action, detail_json, classification_level, event_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
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
			int8(event.ClassificationLevel),
			eventTime,
		); err != nil {
			return fmt.Errorf("audit_repo: append batch row: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("audit_repo: send batch: %w", err)
	}
	return nil
}

// GetEntry retrieves a single audit event by audit_id.
// Classification filter is always applied.
func (r *AuditRepository) GetEntry(
	ctx context.Context,
	auditID string,
	callerClearance commonv1.ClassificationLevel,
) (*auditv1.AuditEvent, error) {
	baseQuery := `SELECT audit_id, service_id, event_type, actor_id,
		actor_type, resource_type, resource_id, action, detail_json,
		classification_level, event_time
		FROM audit_log
		WHERE audit_id = ?`

	params := []interface{}{auditID}

	query, classParam := r.filter.InjectFilter(baseQuery, callerClearance)
	params = append(params, classParam)
	query += " LIMIT 1"

	rows, err := r.conn.Query(ctx, query, params...)
	if err != nil {
		return nil, fmt.Errorf("audit_repo: get entry query: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("audit_repo: get entry rows error: %w", err)
		}
		return nil, nil // not found
	}

	event, err := scanRow(rows)
	if err != nil {
		return nil, err
	}
	return event, nil
}

// QueryAuditLog performs a filtered, paginated query against the audit_log table.
// All queries are parameterized. Classification filter is always injected.
//
// Sort: event_time ASC, audit_id ASC (cursor-based pagination).
func (r *AuditRepository) QueryAuditLog(
	ctx context.Context,
	req *auditv1.StreamAuditLogRequest,
	callerClearance commonv1.ClassificationLevel,
	pageToken *domain.PaginationToken,
	pageSize int,
) ([]*auditv1.AuditEvent, *domain.PaginationToken, error) {
	baseQuery := `SELECT audit_id, service_id, event_type, actor_id,
		actor_type, resource_type, resource_id, action, detail_json,
		classification_level, event_time
		FROM audit_log
		WHERE event_time >= toDateTime64(?, 3, 'UTC') AND event_time <= toDateTime64(?, 3, 'UTC')`

	params := []interface{}{
		req.GetTimeRange().GetStartTime().AsTime(),
		req.GetTimeRange().GetEndTime().AsTime(),
	}

	// Classification filter (always applied)
	query, classParam := r.filter.InjectFilter(baseQuery, callerClearance)
	params = append(params, classParam)

	// Optional filters
	if len(req.GetServiceIds()) > 0 {
		query += " AND service_id IN (?)"
		params = append(params, req.GetServiceIds())
	}
	if len(req.GetEventTypes()) > 0 {
		query += " AND event_type IN (?)"
		params = append(params, req.GetEventTypes())
	}
	if len(req.GetActorIds()) > 0 {
		query += " AND actor_id IN (?)"
		params = append(params, req.GetActorIds())
	}
	if req.GetActorType() != auditv1.ActorType_ACTOR_TYPE_UNSPECIFIED {
		query += " AND actor_type = ?"
		params = append(params, req.GetActorType().String())
	}
	if len(req.GetResourceTypes()) > 0 {
		query += " AND resource_type IN (?)"
		params = append(params, req.GetResourceTypes())
	}
	if len(req.GetActions()) > 0 {
		query += " AND action IN (?)"
		params = append(params, req.GetActions())
	}

	// Cursor-based pagination
	if pageToken != nil {
		query += " AND (event_time > toDateTime64(?, 3, 'UTC') OR (event_time = toDateTime64(?, 3, 'UTC') AND audit_id > ?))"
		tsStr := pageToken.LastTimestamp.Format("2006-01-02 15:04:05.000")
		params = append(params, tsStr, tsStr, pageToken.LastID)
	}
	query += fmt.Sprintf(" ORDER BY event_time ASC, audit_id ASC LIMIT %d", pageSize)

	rows, err := r.conn.Query(ctx, query, params...)
	if err != nil {
		return nil, nil, fmt.Errorf("audit_repo: query audit log: %w", err)
	}
	defer rows.Close()

	var events []*auditv1.AuditEvent
	var lastID string
	var lastTS time.Time

	for rows.Next() {
		event, err := scanRow(rows)
		if err != nil {
			return nil, nil, err
		}
		events = append(events, event)
		lastID = event.AuditId
		if event.EventTime != nil {
			lastTS = event.EventTime.AsTime()
		}
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("audit_repo: rows error: %w", err)
	}

	var nextToken *domain.PaginationToken
	if len(events) == pageSize {
		nextToken = &domain.PaginationToken{
			LastID:        lastID,
			LastTimestamp: lastTS,
			PageSize:      pageSize,
		}
	}

	return events, nextToken, nil
}

// scanRow scans a single ClickHouse row into an AuditEvent.
func scanRow(rows chdriver.Rows) (*auditv1.AuditEvent, error) {
	var (
		auditID      string
		serviceID    string
		eventType    string
		actorID      string
		actorTypeStr string
		resourceType string
		resourceID   string
		action       string
		detailJSON   string
		classLevelStr string
		eventTime    time.Time
	)
	if err := rows.Scan(
		&auditID, &serviceID, &eventType, &actorID,
		&actorTypeStr, &resourceType, &resourceID, &action, &detailJSON,
		&classLevelStr, &eventTime,
	); err != nil {
		return nil, fmt.Errorf("audit_repo: scan row: %w", err)
	}

	classLevelInt, ok := commonv1.ClassificationLevel_value["CLASSIFICATION_LEVEL_" + classLevelStr]
	if !ok {
		// Fallback for safety if Enum prefix was included in clickhouse
		classLevelInt, ok = commonv1.ClassificationLevel_value[classLevelStr]
		if !ok {
			classLevelInt = int32(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNSPECIFIED)
		}
	}

	return mapper.RowToAuditEvent(
		auditID, serviceID, eventType, actorID, actorTypeStr,
		resourceType, resourceID, action, detailJSON,
		classLevelInt, eventTime,
	), nil
}
