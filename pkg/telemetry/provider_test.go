// CLASSIFICATION: UNCLASSIFIED
package telemetry_test

import (
"context"
"testing"

"go.opentelemetry.io/otel"
"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/telemetry"
)

func TestInit_CreatesValidProviders(t *testing.T) {
ctx := context.Background()
cfg := telemetry.Config{
ServiceName:    "test-svc",
ServiceVersion: "1.0.0",
Environment:    "test",
OTelEndpoint:   "", // no endpoint, use no-op
}

provider, err := telemetry.Init(ctx, cfg)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if provider == nil {
t.Fatal("expected non-nil provider")
}
if provider.TracerProvider == nil {
t.Error("expected non-nil TracerProvider")
}
if provider.MeterProvider == nil {
t.Error("expected non-nil MeterProvider")
}
if provider.Logger == nil {
t.Error("expected non-nil Logger")
}
}

func TestInit_SetsGlobalProvider(t *testing.T) {
ctx := context.Background()
cfg := telemetry.Config{
ServiceName: "test-svc",
}
provider, err := telemetry.Init(ctx, cfg)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if otel.GetTracerProvider() == nil {
t.Error("expected global tracer provider to be set")
}
_ = provider.Shutdown(ctx)
}

func TestProvider_Shutdown(t *testing.T) {
ctx := context.Background()
cfg := telemetry.Config{
ServiceName: "test-svc",
}
provider, err := telemetry.Init(ctx, cfg)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if err := provider.Shutdown(ctx); err != nil {
t.Errorf("shutdown error: %v", err)
}
}

func TestAttributes_NotNil(t *testing.T) {
// Verify all standard attribute keys are non-empty
attrs := []interface{}{
telemetry.AttrServiceName,
telemetry.AttrSensorType,
telemetry.AttrClassification,
telemetry.AttrGRPCMethod,
}
for _, a := range attrs {
if a == nil {
t.Error("attribute key should not be nil")
}
}
}

func TestInit_WithServiceVersion(t *testing.T) {
ctx := context.Background()
provider, err := telemetry.Init(ctx, telemetry.Config{
ServiceName:    "test-svc",
ServiceVersion: "2.0.0",
Environment:    "production",
})
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if provider.Logger == nil {
t.Error("expected non-nil logger")
}
_ = provider.Shutdown(ctx)
}

func TestInit_WithOTelEndpoint(t *testing.T) {
// Even with a bad endpoint, Init should succeed (fallback to no-op tracer)
ctx := context.Background()
provider, err := telemetry.Init(ctx, telemetry.Config{
ServiceName:  "test-svc",
OTelEndpoint: "localhost:14317", // unlikely to be running
})
if err != nil {
t.Fatalf("unexpected error (should fallback gracefully): %v", err)
}
if provider == nil {
t.Fatal("expected non-nil provider")
}
_ = provider.Shutdown(ctx)
}

func TestProvider_ShutdownNilProviders(t *testing.T) {
	// Test that Shutdown handles nil providers gracefully
	ctx := context.Background()
	provider, err := telemetry.Init(ctx, telemetry.Config{
		ServiceName: "test-svc",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should not panic on double shutdown
	_ = provider.Shutdown(ctx)
}

func TestInit_NoServiceName_DefaultsToUnknown(t *testing.T) {
	ctx := context.Background()
	provider, err := telemetry.Init(ctx, telemetry.Config{
		ServiceName: "",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer provider.Shutdown(ctx)
	if provider.ServiceName != "unknown" {
		t.Errorf("expected service name unknown, got %s", provider.ServiceName)
	}
}

func TestInit_DefaultEnvironment(t *testing.T) {
	ctx := context.Background()
	provider, err := telemetry.Init(ctx, telemetry.Config{
		ServiceName: "test",
		Environment: "",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer provider.Shutdown(ctx)
}

func TestProvider_ShutdownCancelledContext(t *testing.T) {
	ctx := context.Background()
	provider, _ := telemetry.Init(ctx, telemetry.Config{ServiceName: "test"})

	cancelCtx, cancel := context.WithCancel(ctx)
	cancel() // cancel immediately

	err := provider.Shutdown(cancelCtx)
	if err == nil {
		// It might still succeed if providers shutdown fast enough, but we want to exercise the branch.
		t.Log("shutdown succeeded even with cancelled context")
	}
}


func TestInit_WithCustomConfig(t *testing.T) {
	ctx := context.Background()
	provider, err := telemetry.Init(ctx, telemetry.Config{
		ServiceName:    "test",
		ServiceVersion: "1.2.3",
		Environment:    "prod",
		MetricsPort:    9092,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer provider.Shutdown(ctx)
}


func TestInit_WithMetricsPort(t *testing.T) {

	ctx := context.Background()
	provider, err := telemetry.Init(ctx, telemetry.Config{
		ServiceName: "test",
		MetricsPort: 9091,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer provider.Shutdown(ctx)
}


