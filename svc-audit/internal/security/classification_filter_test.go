// CLASSIFICATION: UNCLASSIFIED
package security_test

import (
"context"
"testing"

commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
"github.com/arvinddhasmana/rtsa_webgpu/svc-audit/internal/security"
"google.golang.org/grpc/metadata"
)

func TestExtractClearance_NoMetadata(t *testing.T) {
level := security.ExtractClearance(context.Background())
if level != commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED {
t.Errorf("expected UNCLASSIFIED for empty context, got %v", level)
}
}

func TestExtractClearance_ProtectedA(t *testing.T) {
md := metadata.New(map[string]string{"x-rtsa-clearance": "PROTECTED_A"})
ctx := metadata.NewIncomingContext(context.Background(), md)
level := security.ExtractClearance(ctx)
if level != commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_A {
t.Errorf("expected PROTECTED_A, got %v", level)
}
}

func TestExtractClearance_Secret(t *testing.T) {
md := metadata.New(map[string]string{"x-rtsa-clearance": "SECRET"})
ctx := metadata.NewIncomingContext(context.Background(), md)
level := security.ExtractClearance(ctx)
if level != commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET {
t.Errorf("expected SECRET, got %v", level)
}
}

func TestExtractClearance_UnknownValue(t *testing.T) {
md := metadata.New(map[string]string{"x-rtsa-clearance": "TOP_SECRET"})
ctx := metadata.NewIncomingContext(context.Background(), md)
level := security.ExtractClearance(ctx)
if level != commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED {
t.Errorf("expected UNCLASSIFIED for unknown value, got %v", level)
}
}

func TestInjectFilter_AppendsWhereClause(t *testing.T) {
f := &security.ClassificationFilter{}
base := "SELECT * FROM audit_log WHERE event_time >= ?"
query, param := f.InjectFilter(base, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B)
if query == base {
t.Error("expected filter to be injected into query")
}
if param == nil {
t.Error("expected non-nil parameter")
}
// PROTECTED_B ordinal is 3
if param.(int32) != 3 {
t.Errorf("expected ordinal 3 for PROTECTED_B, got %v", param)
}
}

func TestExtractClearance_ProtectedB(t *testing.T) {
md := metadata.New(map[string]string{"x-rtsa-clearance": "PROTECTED_B"})
ctx := metadata.NewIncomingContext(context.Background(), md)
level := security.ExtractClearance(ctx)
if level != commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B {
t.Errorf("expected PROTECTED_B, got %v", level)
}
}

func TestExtractClearance_ProtectedC(t *testing.T) {
md := metadata.New(map[string]string{"x-rtsa-clearance": "PROTECTED_C"})
ctx := metadata.NewIncomingContext(context.Background(), md)
level := security.ExtractClearance(ctx)
if level != commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_C {
t.Errorf("expected PROTECTED_C, got %v", level)
}
}

func TestInjectFilter_SecretOrdinal(t *testing.T) {
f := &security.ClassificationFilter{}
base := "SELECT * FROM audit_log WHERE event_time >= ?"
_, param := f.InjectFilter(base, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET)
if param.(int32) != 5 {
t.Errorf("expected ordinal 5 for SECRET, got %v", param)
}
}

func TestInjectFilter_UnclassifiedOrdinal(t *testing.T) {
f := &security.ClassificationFilter{}
base := "SELECT * FROM audit_log WHERE event_time >= ?"
_, param := f.InjectFilter(base, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED)
if param.(int32) != 1 {
t.Errorf("expected ordinal 1 for UNCLASSIFIED, got %v", param)
}
}

func TestInjectFilter_ProtectedAOrdinal(t *testing.T) {
f := &security.ClassificationFilter{}
base := "SELECT * FROM audit_log WHERE event_time >= ?"
_, param := f.InjectFilter(base, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_A)
if param.(int32) != 2 {
t.Errorf("expected ordinal 2 for PROTECTED_A, got %v", param)
}
}

func TestInjectFilter_ProtectedCOrdinal(t *testing.T) {
f := &security.ClassificationFilter{}
base := "SELECT * FROM audit_log WHERE event_time >= ?"
_, param := f.InjectFilter(base, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_C)
if param.(int32) != 4 {
t.Errorf("expected ordinal 4 for PROTECTED_C, got %v", param)
}
}
