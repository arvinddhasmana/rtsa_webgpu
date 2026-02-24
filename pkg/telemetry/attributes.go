// CLASSIFICATION: UNCLASSIFIED
package telemetry

import "go.opentelemetry.io/otel/attribute"

// Standard metric attribute keys used across all services.
var (
AttrServiceName    = attribute.Key("service.name")
AttrSensorType     = attribute.Key("sensor.type")
AttrEntityType     = attribute.Key("entity.type")
AttrClassification = attribute.Key("classification.level")
AttrAnomalyType    = attribute.Key("anomaly.type")
AttrAlertSeverity  = attribute.Key("alert.severity")
AttrFeedbackType   = attribute.Key("feedback.type")
AttrOperatorID     = attribute.Key("operator.id")
AttrTrackStatus    = attribute.Key("track.status")
AttrHostileClass   = attribute.Key("hostile.classification")
AttrTopicName      = attribute.Key("redpanda.topic")
AttrConsumerGroup  = attribute.Key("redpanda.consumer_group")
AttrGRPCMethod     = attribute.Key("grpc.method")
AttrGRPCStatusCode = attribute.Key("grpc.status_code")
)
