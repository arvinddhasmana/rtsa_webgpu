// CLASSIFICATION: UNCLASSIFIED
package handler

import (
	"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/audit"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/domain"
	"go.uber.org/zap"
)

// NewQueryServerForTest creates a QueryServer suitable for unit tests
// that only exercise validation logic (before any repository calls).
func NewQueryServerForTest(guardrail *domain.QueryGuardrail, logger *zap.Logger) *QueryServer {
	return &QueryServer{
		guardrail:    guardrail,
		auditEmitter: &logAuditEmitter{emitter: audit.NewLogEmitter(logger)},
		pageSize:     100,
		logger:       logger,
	}
}
