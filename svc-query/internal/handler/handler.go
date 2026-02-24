// CLASSIFICATION: UNCLASSIFIED

// Package handler implements the gRPC QueryService server handlers.
// Handlers are thin orchestration layers: validate -> guardrail -> query -> audit -> respond.
package handler

import (
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/audit"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/domain"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/repository"
queryv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/query/v1"
)

// Handler implements the QueryServiceServer gRPC interface.
// It orchestrates validation, guardrail enforcement, repository calls,
// and audit event emission for every query operation.
type Handler struct {
queryv1.UnimplementedQueryServiceServer

tracksRepo   repository.TracksQuerier
anomalyRepo  repository.AnomaliesQuerier
auditRepo    repository.AuditQuerier
auditEmitter audit.Emitter
guard        *domain.QueryGuardrail
serviceID    string
}

// New creates a Handler with all required dependencies.
func New(
tracksRepo repository.TracksQuerier,
anomalyRepo repository.AnomaliesQuerier,
auditRepo repository.AuditQuerier,
auditEmitter audit.Emitter,
guard *domain.QueryGuardrail,
serviceID string,
) *Handler {
return &Handler{
tracksRepo:   tracksRepo,
anomalyRepo:  anomalyRepo,
auditRepo:    auditRepo,
auditEmitter: auditEmitter,
guard:        guard,
serviceID:    serviceID,
}
}
