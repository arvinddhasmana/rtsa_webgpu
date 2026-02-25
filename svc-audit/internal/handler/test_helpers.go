// CLASSIFICATION: UNCLASSIFIED
package handler

import (
auditv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/audit/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-audit/internal/domain"
"go.uber.org/zap"
)

// NewAuditServerForTest creates an AuditServer suitable for unit tests
// that only exercise validation logic (before any repository calls).
func NewAuditServerForTest(guardrail *domain.QueryGuardrail, logger *zap.Logger) *AuditServer {
return &AuditServer{
guardrail: guardrail,
pageSize:  100,
logger:    logger,
}
}

// StreamAuditLogValidateOnly validates the StreamAuditLog request without
// requiring a live repository. Used in unit tests to exercise validation only.
func (s *AuditServer) StreamAuditLogValidateOnly(req *auditv1.StreamAuditLogRequest) error {
return s.guardrail.ValidateTimeRange(
req.GetTimeRange().GetStartTime(),
req.GetTimeRange().GetEndTime(),
)
}
