// CLASSIFICATION: UNCLASSIFIED
package security_test

import (
	"context"
	"strings"
	"testing"

	commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/security"
	"google.golang.org/grpc/metadata"
)

func TestExtractClearance_NoMetadata(t *testing.T) {
ctx := context.Background()
level := security.ExtractClearance(ctx)
if level != commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED {
t.Errorf("expected UNCLASSIFIED, got %v", level)
}
}

func TestExtractClearance_WithClearance(t *testing.T) {
md := metadata.Pairs("x-rtsa-clearance", "SECRET")
ctx := metadata.NewIncomingContext(context.Background(), md)
level := security.ExtractClearance(ctx)
if level != commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET {
t.Errorf("expected SECRET, got %v", level)
}
}

func TestExtractClearance_Unknown(t *testing.T) {
md := metadata.Pairs("x-rtsa-clearance", "UNKNOWN_LEVEL")
ctx := metadata.NewIncomingContext(context.Background(), md)
level := security.ExtractClearance(ctx)
if level != commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED {
t.Errorf("expected UNCLASSIFIED for unknown, got %v", level)
}
}

func TestInjectFilter(t *testing.T) {
f := &security.ClassificationFilter{}
query := "SELECT * FROM tracks_fused WHERE entity_type = ?"
modified, params := f.InjectFilter(query, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED)
if modified == query {
t.Error("expected query to be modified with classification filter")
}
if len(params) == 0 {
t.Error("expected non-empty params for classification filter")
}
// Verify IN clause is used (not <=) to match LowCardinality(String) column type.
if !strings.Contains(modified, "IN") {
t.Error("expected IN clause in classification filter, got: " + modified)
}
}
