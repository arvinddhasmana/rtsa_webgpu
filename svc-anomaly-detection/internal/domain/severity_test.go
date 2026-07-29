// CLASSIFICATION: UNCLASSIFIED
package domain

import (
"testing"

commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
)

func TestMapSeverity(t *testing.T) {
tests := []struct {
name       string
confidence float64
want       commonv1.AlertSeverity
}{
// T17
{"T17_critical", 0.95, commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL},
// T18
{"T18_elevated", 0.75, commonv1.AlertSeverity_ALERT_SEVERITY_ELEVATED},
// T19
{"T19_watch", 0.55, commonv1.AlertSeverity_ALERT_SEVERITY_WATCH},
// T20
{"T20_normal", 0.30, commonv1.AlertSeverity_ALERT_SEVERITY_NORMAL},
// Boundary conditions.
{"boundary_critical", 0.90, commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL},
{"boundary_elevated", 0.70, commonv1.AlertSeverity_ALERT_SEVERITY_ELEVATED},
{"boundary_watch", 0.50, commonv1.AlertSeverity_ALERT_SEVERITY_WATCH},
{"boundary_below_watch", 0.499, commonv1.AlertSeverity_ALERT_SEVERITY_NORMAL},
{"zero", 0.0, commonv1.AlertSeverity_ALERT_SEVERITY_NORMAL},
{"one", 1.0, commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
got := MapSeverity(tt.confidence)
if got != tt.want {
t.Errorf("MapSeverity(%v) = %v, want %v", tt.confidence, got, tt.want)
}
})
}
}

func TestSeverityTopic(t *testing.T) {
tests := []struct {
severity commonv1.AlertSeverity
want     string
}{
{commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL, "alerts.anomaly.critical"},
{commonv1.AlertSeverity_ALERT_SEVERITY_ELEVATED, "alerts.anomaly.elevated"},
{commonv1.AlertSeverity_ALERT_SEVERITY_WATCH, "alerts.anomaly.watch"},
{commonv1.AlertSeverity_ALERT_SEVERITY_NORMAL, ""},
{commonv1.AlertSeverity_ALERT_SEVERITY_UNSPECIFIED, ""},
}

for _, tt := range tests {
got := SeverityTopic(tt.severity)
if got != tt.want {
t.Errorf("SeverityTopic(%v) = %q, want %q", tt.severity, got, tt.want)
}
}
}
