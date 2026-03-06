// CLASSIFICATION: UNCLASSIFIED
// pkg/webtransport/testhelpers_test.go — Test helper utilities

package webtransport_test

import (
"go.opentelemetry.io/otel/metric"
"go.opentelemetry.io/otel/metric/noop"
)

// otelnoopMeter returns a no-op OTel meter for testing metric registration paths.
func otelnoopMeter() metric.Meter {
return noop.NewMeterProvider().Meter("test")
}
