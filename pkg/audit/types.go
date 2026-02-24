// CLASSIFICATION: UNCLASSIFIED
package audit

// Standard event type constants.
const (
EventTrackCreated          = "track.created"
EventTrackUpdated          = "track.updated"
EventTrackMerged           = "track.merged"
EventTrackDropped          = "track.dropped"
EventAlertGenerated        = "alert.generated"
EventAlertAcknowledged     = "alert.acknowledged"
EventFeedbackSubmitted     = "feedback.submitted"
EventFeedbackValidated     = "feedback.validated"
EventFeedbackRejected      = "feedback.rejected"
EventQueryExecuted         = "query.executed"
EventModelPublished        = "model.published"
EventModelRolledBack       = "model.rolled_back"
EventSensorConnected       = "sensor.connected"
EventSensorDisconnected    = "sensor.disconnected"
EventNATOExport            = "nato.export"
EventNATOImport            = "nato.import"
EventClassificationViolation = "classification.violation"
EventObservationIngested   = "observation.ingested"
EventObservationRejected   = "observation.rejected"
)

// AuditTopic is the Redpanda topic for audit events.
const AuditTopic = "audit.events"
