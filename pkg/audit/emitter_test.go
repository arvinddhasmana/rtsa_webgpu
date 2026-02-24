// CLASSIFICATION: UNCLASSIFIED
package audit_test

import (
	"context"
	"testing"

	"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/audit"
	"go.uber.org/zap"
)

func TestEmitter_Emit_NoError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	e := audit.NewEmitter(logger)
	// Should not panic or return an error
	e.Emit(context.Background(), audit.Event{
		EventType:    "TEST",
		ResourceType: "Track",
		ResourceID:   "track-123",
		Action:       "CREATE",
	})
}
