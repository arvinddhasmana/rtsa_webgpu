// CLASSIFICATION: UNCLASSIFIED
package security_test

import (
	"context"
	"strings"
	"testing"

	commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
	"github.com/arvinddhasmana/rtsa_webgpu/svc-query/internal/security"
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

func TestInjectFilter_Levels(t *testing.T) {
	f := &security.ClassificationFilter{}
	levels := []commonv1.ClassificationLevel{
		commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
		commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_A,
		commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B,
		commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_C,
		commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
	}
	for _, l := range levels {
		t.Run(l.String(), func(t *testing.T) {
			query := "SELECT * FROM tracks_fused"
			modified, params := f.InjectFilter(query, l)
			if !strings.Contains(modified, "classification_level") {
				t.Errorf("expected classification filter for level %v", l)
			}
			if len(params) == 0 {
				t.Errorf("expected params for level %v", l)
			}
		})
	}
}
