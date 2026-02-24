// CLASSIFICATION: UNCLASSIFIED
package main

import (
"fmt"
"net"
"os"
"os/signal"
"syscall"

ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/ingestion"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-elint-ingestion/internal/domain"
"go.uber.org/zap"
"google.golang.org/grpc"
)

func main() {
cfg := ingestion.MustLoad("svc-elint-ingestion", "sensors.elint.detections", "dlq.sensors.elint", 50053)

logger, err := zap.NewProduction()
if err != nil {
panic(fmt.Sprintf("failed to init logger: %v", err))
}
defer logger.Sync() //nolint:errcheck

validator := domain.NewValidator()
normalizer := domain.NewNormalizer()
producer := ingestion.NewLogProducer(cfg.OutputTopic)
dlqProducer := ingestion.NewLogProducer(cfg.DLQTopic)

handler := ingestion.NewHandler(validator, normalizer, producer, dlqProducer, logger, cfg)

lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
if err != nil {
logger.Fatal("failed to listen", zap.Error(err))
}

srv := grpc.NewServer()
ingestionv1.RegisterIngestionServiceServer(srv, handler)

logger.Info("starting svc-elint-ingestion", zap.Int("port", cfg.GRPCPort))

go func() {
if err := srv.Serve(lis); err != nil {
logger.Error("server error", zap.Error(err))
}
}()

quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

srv.GracefulStop()
if err := producer.Close(); err != nil {
logger.Error("producer close error", zap.Error(err))
}
if err := dlqProducer.Close(); err != nil {
logger.Error("dlq producer close error", zap.Error(err))
}
logger.Info("svc-elint-ingestion stopped")
}
