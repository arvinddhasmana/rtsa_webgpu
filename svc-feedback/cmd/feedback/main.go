// CLASSIFICATION: UNCLASSIFIED

// Command feedback is the entrypoint for the RTSA Feedback & Trust Scoring Service.
// It starts a gRPC server exposing FeedbackService on the configured port.
package main

import (
"context"
"fmt"
"net"
"net/http"
"os"
"os/signal"
"syscall"
"time"

feedbackv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/feedback/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/audit"
auditv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/audit/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-feedback/internal/config"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-feedback/internal/domain"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-feedback/internal/handler"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-feedback/internal/producer"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-feedback/internal/state"
"go.uber.org/zap"
"google.golang.org/grpc"
"google.golang.org/grpc/reflection"
)

// auditEmitterAdapter adapts pkg/audit.Emitter to the handler.AuditEmitter interface.
// It maps a fully-formed *auditv1.AuditEvent to the structured audit.AuditParams
// so that events are produced via the real pkg/audit infrastructure.
type auditEmitterAdapter struct{ emitter *audit.Emitter }

func (a *auditEmitterAdapter) EmitAudit(ctx context.Context, event *auditv1.AuditEvent) error {
a.emitter.Emit(ctx, audit.AuditParams{
EventType:    event.GetEventType(),
ActorID:      event.GetActorId(),
ActorType:    event.GetActorType(),
ResourceType: event.GetResourceType(),
ResourceID:   event.GetResourceId(),
Action:       event.GetAction(),
})
return nil
}

func main() {
if err := run(); err != nil {
fmt.Fprintf(os.Stderr, "svc-feedback: fatal: %v\n", err)
os.Exit(1)
}
}

func run() error {
// Load configuration from environment.
cfg, err := config.Load()
if err != nil {
return fmt.Errorf("[main.run]: config: %w", err)
}

// Initialise structured logger.
var logger *zap.Logger
if cfg.LogLevel == "debug" {
logger, err = zap.NewDevelopment()
} else {
logger, err = zap.NewProduction()
}
if err != nil {
return fmt.Errorf("[main.run]: logger: %w", err)
}
defer func() { _ = logger.Sync() }()

logger.Info("starting svc-feedback",
zap.String("grpc_port", cfg.GRPCPort),
zap.String("health_port", cfg.HealthPort),
zap.Strings("brokers", cfg.RedpandaBrokers),
)

// Build shared in-memory state.
history := state.NewOperatorHistory()

// Build domain components.
trustScorer := domain.NewTrustScorer(history)
antiPoison := domain.NewAntiPoisonGuard(history, logger)
rateLimiter := domain.NewRateLimiter(cfg.RateLimitPerMin)

// Build Redpanda producers.
rawProducer, err := producer.NewFeedbackProducer(
cfg.RedpandaBrokers, "feedback.operator.submissions")
if err != nil {
return fmt.Errorf("[main.run]: submissions producer: %w", err)
}
defer func() { _ = rawProducer.Close() }()

validProducer, err := producer.NewFeedbackProducer(
cfg.RedpandaBrokers, "feedback.operator.validated")
if err != nil {
return fmt.Errorf("[main.run]: validated producer: %w", err)
}
defer func() { _ = validProducer.Close() }()

// Build handler.
auditEmitter := &auditEmitterAdapter{emitter: audit.NewLogEmitter(logger)}
fbHandler := handler.NewFeedbackHandler(
trustScorer,
antiPoison,
rateLimiter,
rawProducer,
validProducer,
auditEmitter,
history,
cfg.ServiceName,
logger,
)

// Initialise gRPC server.
// NOTE: In production, replace grpc.NewServer() with an mTLS-configured
// server using grpc.Creds(credentials.NewTLS(tlsConfig)).
grpcServer := grpc.NewServer()
feedbackv1.RegisterFeedbackServiceServer(grpcServer, fbHandler)
reflection.Register(grpcServer)

// Start health endpoint.
healthMux := http.NewServeMux()
healthMux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
w.WriteHeader(http.StatusOK)
_, _ = w.Write([]byte("ok"))
})
healthSrv := &http.Server{
Addr:         ":" + cfg.HealthPort,
Handler:      healthMux,
ReadTimeout:  5 * time.Second,
WriteTimeout: 5 * time.Second,
}
go func() {
if hErr := healthSrv.ListenAndServe(); hErr != nil && hErr != http.ErrServerClosed {
logger.Error("health server error", zap.Error(hErr))
}
}()

// Start gRPC listener.
lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
if err != nil {
return fmt.Errorf("[main.run]: listen: %w", err)
}

// Graceful shutdown on SIGTERM/SIGINT.
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

go func() {
<-quit
logger.Info("shutting down svc-feedback")
grpcServer.GracefulStop()
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
if shutErr := healthSrv.Shutdown(ctx); shutErr != nil {
logger.Error("health server shutdown error", zap.Error(shutErr))
}
}()

logger.Info("gRPC server listening", zap.String("addr", lis.Addr().String()))
if err := grpcServer.Serve(lis); err != nil {
return fmt.Errorf("[main.run]: grpc serve: %w", err)
}

return nil
}
