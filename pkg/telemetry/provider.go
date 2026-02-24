// CLASSIFICATION: UNCLASSIFIED
package telemetry

import (
"context"
"fmt"

"go.opentelemetry.io/otel"
"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
"go.opentelemetry.io/otel/exporters/prometheus"
"go.opentelemetry.io/otel/sdk/metric"
"go.opentelemetry.io/otel/sdk/resource"
sdktrace "go.opentelemetry.io/otel/sdk/trace"
semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
"go.uber.org/zap"
"go.uber.org/zap/zapcore"
)

// Config holds telemetry configuration.
type Config struct {
ServiceName    string
ServiceVersion string
Environment    string
OTelEndpoint   string
MetricsPort    int
}

// Provider holds initialized telemetry providers.
type Provider struct {
TracerProvider *sdktrace.TracerProvider
MeterProvider  *metric.MeterProvider
Logger         *zap.Logger
}

// Init initializes OpenTelemetry tracing, metrics, and structured logging.
func Init(ctx context.Context, cfg Config) (*Provider, error) {
// Build logger
loggerCfg := zap.NewProductionConfig()
loggerCfg.Encoding = "json"
loggerCfg.EncoderConfig.TimeKey = "ts"
loggerCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
logger, err := loggerCfg.Build(
zap.Fields(
zap.String("service", cfg.ServiceName),
zap.String("version", cfg.ServiceVersion),
zap.String("env", cfg.Environment),
),
)
if err != nil {
return nil, fmt.Errorf("telemetry: build logger: %w", err)
}

// Build resource
res, err := resource.New(ctx,
resource.WithAttributes(
semconv.ServiceNameKey.String(cfg.ServiceName),
semconv.ServiceVersionKey.String(cfg.ServiceVersion),
),
)
if err != nil {
return nil, fmt.Errorf("telemetry: build resource: %w", err)
}

// Build tracer provider
var tp *sdktrace.TracerProvider
if cfg.OTelEndpoint != "" {
exp, err := otlptracegrpc.New(ctx,
otlptracegrpc.WithEndpoint(cfg.OTelEndpoint),
otlptracegrpc.WithInsecure(),
)
if err != nil {
// Fall back to no-op tracer if OTEL endpoint unavailable
logger.Warn("telemetry: OTLP exporter unavailable, using no-op tracer",
zap.String("endpoint", cfg.OTelEndpoint),
zap.Error(err))
tp = sdktrace.NewTracerProvider()
} else {
tp = sdktrace.NewTracerProvider(
sdktrace.WithBatcher(exp),
sdktrace.WithResource(res),
sdktrace.WithSampler(sdktrace.AlwaysSample()),
)
}
} else {
tp = sdktrace.NewTracerProvider(sdktrace.WithResource(res))
}
otel.SetTracerProvider(tp)

// Build meter provider with Prometheus exporter
promExporter, err := prometheus.New()
if err != nil {
return nil, fmt.Errorf("telemetry: prometheus exporter: %w", err)
}
mp := metric.NewMeterProvider(
metric.WithReader(promExporter),
metric.WithResource(res),
)
otel.SetMeterProvider(mp)

return &Provider{
TracerProvider: tp,
MeterProvider:  mp,
Logger:         logger,
}, nil
}

// Shutdown gracefully shuts down all telemetry providers.
func (p *Provider) Shutdown(ctx context.Context) error {
var firstErr error
if p.TracerProvider != nil {
if err := p.TracerProvider.Shutdown(ctx); err != nil {
firstErr = fmt.Errorf("telemetry: tracer shutdown: %w", err)
}
}
if p.MeterProvider != nil {
if err := p.MeterProvider.Shutdown(ctx); err != nil {
if firstErr == nil {
firstErr = fmt.Errorf("telemetry: meter shutdown: %w", err)
}
}
}
if p.Logger != nil {
_ = p.Logger.Sync()
}
return firstErr
}
