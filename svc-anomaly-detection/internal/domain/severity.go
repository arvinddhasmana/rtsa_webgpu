// CLASSIFICATION: UNCLASSIFIED
package domain

import (
commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
)

// MapSeverity converts an anomaly confidence score to an AlertSeverity.
//   - < 0.50        → NORMAL  (no alert produced)
//   - 0.50 – 0.69   → WATCH
//   - 0.70 – 0.89   → ELEVATED
//   - ≥ 0.90        → CRITICAL
func MapSeverity(confidence float64) commonv1.AlertSeverity {
switch {
case confidence >= 0.90:
return commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL
case confidence >= 0.70:
return commonv1.AlertSeverity_ALERT_SEVERITY_ELEVATED
case confidence >= 0.50:
return commonv1.AlertSeverity_ALERT_SEVERITY_WATCH
default:
return commonv1.AlertSeverity_ALERT_SEVERITY_NORMAL
}
}

// SeverityTopic returns the output Redpanda topic for a given alert severity.
// Returns "" for NORMAL severity (no alert should be produced).
func SeverityTopic(severity commonv1.AlertSeverity) string {
switch severity {
case commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL:
return "alerts.anomaly.critical"
case commonv1.AlertSeverity_ALERT_SEVERITY_ELEVATED:
return "alerts.anomaly.elevated"
case commonv1.AlertSeverity_ALERT_SEVERITY_WATCH:
return "alerts.anomaly.watch"
default:
return ""
}
}
