// CLASSIFICATION: UNCLASSIFIED
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
	queryv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/query/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TimelineRepository struct {
	db clickhouse.Conn
}

func NewTimelineRepository(conn clickhouse.Conn) *TimelineRepository {
	return &TimelineRepository{db: conn}
}

// GetEventTimeline executes a UNION ALL over tracks_fused, anomaly_detections, operator_feedback, and audit_log
func (r *TimelineRepository) GetEventTimeline(
	ctx context.Context,
	req *queryv1.GetEventTimelineRequest,
	clearance commonv1.ClassificationLevel,
) ([]*queryv1.TimelineEvent, error) {

	query := `
		SELECT
			event_time,
			event_type_enum,
			summary,
			detail_type,
			prev_status,
			new_status,
			lat,
			lon,
			conf_score,
			alert_id,
			anomaly_type,
			severity,
			explanation,
			feedback_id,
			feedback_type,
			trust_score,
			audit_id,
			audit_action,
			actor_id
		FROM (
			-- 1. Track State Changes
			SELECT
				event_time,
				if(track_status = 'NEW', 1, if(track_status = 'DROPPED', 4, 2)) AS event_type_enum,
				concat('Track status changed to ', track_status) AS summary,
				'track_change' AS detail_type,
				'' AS prev_status,
				track_status AS new_status,
				latitude AS lat,
				longitude AS lon,
				confidence_score AS conf_score,
				'' AS alert_id,
				0 AS anomaly_type,
				0 AS severity,
				'' AS explanation,
				'' AS feedback_id,
				0 AS feedback_type,
				0.0 AS trust_score,
				'' AS audit_id,
				'' AS audit_action,
				'' AS actor_id
			FROM tracks_fused
			WHERE track_id = $1 AND classification_level <= $2 AND event_time BETWEEN $3 AND $4

			UNION ALL

			-- 2. Anomaly Detections
			SELECT
				event_time,
				5 AS event_type_enum,
				concat('Anomaly detected: ', toString(anomaly_type)) AS summary,
				'anomaly' AS detail_type,
				'' AS prev_status,
				'' AS new_status,
				0.0 AS lat,
				0.0 AS lon,
				confidence_score AS conf_score,
				alert_id,
				CAST(anomaly_type, 'Int32') AS anomaly_type,
				CAST(severity, 'Int32') AS severity,
				explanation,
				'' AS feedback_id,
				0 AS feedback_type,
				0.0 AS trust_score,
				'' AS audit_id,
				'' AS audit_action,
				'' AS actor_id
			FROM anomaly_detections
			WHERE track_id = $1 AND classification_level <= $2 AND event_time BETWEEN $3 AND $4

			UNION ALL

			-- 3. Operator Feedback
			SELECT
				event_time,
				7 AS event_type_enum,
				concat('Feedback submitted: ', toString(feedback_type)) AS summary,
				'feedback' AS detail_type,
				'' AS prev_status,
				'' AS new_status,
				0.0 AS lat,
				0.0 AS lon,
				0.0 AS conf_score,
				'' AS alert_id,
				0 AS anomaly_type,
				0 AS severity,
				'' AS explanation,
				feedback_id,
				CAST(feedback_type, 'Int32') AS feedback_type,
				trust_score,
				'' AS audit_id,
				'' AS audit_action,
				'' AS actor_id
			FROM operator_feedback
			WHERE track_id = $1 AND classification_level <= $2 AND event_time BETWEEN $3 AND $4

			UNION ALL

			-- 4. Audit Log
			SELECT
				event_time,
				0 AS event_type_enum,
				concat('Audit event: ', action) AS summary,
				'audit' AS detail_type,
				'' AS prev_status,
				'' AS new_status,
				0.0 AS lat,
				0.0 AS lon,
				0.0 AS conf_score,
				'' AS alert_id,
				0 AS anomaly_type,
				0 AS severity,
				'' AS explanation,
				'' AS feedback_id,
				0 AS feedback_type,
				0.0 AS trust_score,
				audit_id,
				action AS audit_action,
				actor_id
			FROM audit_log
			WHERE resource_id = $1 AND resource_type = 'track' AND classification_level <= $2 AND event_time BETWEEN $3 AND $4
		)
		ORDER BY event_time ASC
		LIMIT $5
	`

	limit := req.MaxEvents
	if limit == 0 {
		limit = 200
	}

	rows, err := r.db.Query(ctx, query,
		req.TrackId,
		int32(clearance),
		req.TimeRange.StartTime.AsTime(),
		req.TimeRange.EndTime.AsTime(),
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("TimelineRepository.GetEventTimeline: query error: %w", err)
	}
	defer rows.Close()

	var events []*queryv1.TimelineEvent
	for rows.Next() {
		var (
			eventTime    time.Time
			eventType    int32
			summary      string
			detailType   string
			prevStatus   string
			newStatus    string
			lat, lon     float64
			confScore    float64
			alertID      string
			anomalyType  int32
			severity     int32
			explanation  string
			feedbackID   string
			feedbackType int32
			trustScore   float64
			auditID      string
			auditAction  string
			actorID      string
		)

		if err := rows.Scan(
			&eventTime, &eventType, &summary, &detailType,
			&prevStatus, &newStatus, &lat, &lon, &confScore,
			&alertID, &anomalyType, &severity, &explanation,
			&feedbackID, &feedbackType, &trustScore,
			&auditID, &auditAction, &actorID,
		); err != nil {
			return nil, fmt.Errorf("TimelineRepository.GetEventTimeline: scan error: %w", err)
		}

		event := &queryv1.TimelineEvent{
			EventTime: timestamppb.New(eventTime),
			EventType: queryv1.TimelineEventType(eventType),
			Summary:   summary,
		}

		switch detailType {
		case "track_change":
			event.Detail = &queryv1.TimelineEvent_TrackChange{
				TrackChange: &queryv1.TrackStateChange{
					PreviousStatus: prevStatus,
					NewStatus:      newStatus,
					Position:       &commonv1.Position{Latitude: lat, Longitude: lon},
					ConfidenceScore: confScore,
				},
			}
		case "anomaly":
			event.Detail = &queryv1.TimelineEvent_Anomaly{
				Anomaly: &queryv1.AnomalyEventDetail{
					AlertId:         alertID,
					AnomalyType:     commonv1.AnomalyType(anomalyType),
					Severity:        commonv1.AlertSeverity(severity),
					ConfidenceScore: confScore,
					Explanation:     explanation,
				},
			}
		case "feedback":
			event.Detail = &queryv1.TimelineEvent_Feedback{
				Feedback: &queryv1.FeedbackEventDetail{
					FeedbackId:   feedbackID,
					FeedbackType: commonv1.FeedbackType(feedbackType),
					TrustScore:   trustScore,
				},
			}
		case "audit":
			event.Detail = &queryv1.TimelineEvent_Audit{
				Audit: &queryv1.AuditEventDetail{
					AuditId: auditID,
					Action:  auditAction,
					ActorId: actorID,
				},
			}
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("TimelineRepository.GetEventTimeline: rows iteration error: %w", err)
	}

	return events, nil
}
