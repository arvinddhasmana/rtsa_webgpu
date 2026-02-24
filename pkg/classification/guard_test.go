// CLASSIFICATION: UNCLASSIFIED
package classification_test

import (
"testing"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/classification"
)

func TestGuard_CheckWithinCeiling(t *testing.T) {
g := classification.NewGuard(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET)
err := g.Check(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED)
if err != nil {
t.Errorf("expected nil, got: %v", err)
}
}

func TestGuard_CheckAboveCeiling(t *testing.T) {
g := classification.NewGuard(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED)
err := g.Check(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET)
if err == nil {
t.Error("expected error for SECRET above UNCLASSIFIED ceiling")
}
}

func TestCanAccess_Sufficient(t *testing.T) {
result := classification.CanAccess(
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
)
if !result {
t.Error("expected true: SECRET clearance >= UNCLASSIFIED data")
}
}

func TestCanAccess_Insufficient(t *testing.T) {
result := classification.CanAccess(
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
)
if result {
t.Error("expected false: UNCLASSIFIED clearance < SECRET data")
}
}

func TestMax_ProtectedBAndSecret(t *testing.T) {
result := classification.Max(
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
)
if result != commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET {
t.Errorf("expected SECRET, got %v", result)
}
}

func TestMaxAll_Mixed(t *testing.T) {
result := classification.MaxAll(
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_A,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B,
)
if result != commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B {
t.Errorf("expected PROTECTED_B, got %v", result)
}
}

func TestLevelToString_Roundtrip(t *testing.T) {
levels := []commonv1.ClassificationLevel{
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_A,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
}
for _, l := range levels {
s := classification.LevelToString(l)
back := classification.StringToLevel(s)
if back != l {
t.Errorf("roundtrip failed for %v: got %v", l, back)
}
}
}

func TestStringToLevel_Unknown(t *testing.T) {
result := classification.StringToLevel("INVALID")
if result != commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNSPECIFIED {
t.Errorf("expected UNSPECIFIED, got %v", result)
}
}

func TestGuard_EqualCeiling(t *testing.T) {
g := classification.NewGuard(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B)
err := g.Check(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B)
if err != nil {
t.Errorf("expected nil for equal level, got: %v", err)
}
}

func TestMax_Equal(t *testing.T) {
result := classification.Max(
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_A,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_A,
)
if result != commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_A {
t.Errorf("expected PROTECTED_A, got %v", result)
}
}

func TestMaxAll_Empty(t *testing.T) {
result := classification.MaxAll()
if result != commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNSPECIFIED {
t.Errorf("expected UNSPECIFIED for empty input, got %v", result)
}
}

func TestMaxAll_Single(t *testing.T) {
result := classification.MaxAll(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET)
if result != commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET {
t.Errorf("expected SECRET, got %v", result)
}
}

func TestLevelToString_AllLevels(t *testing.T) {
tests := []struct {
level    commonv1.ClassificationLevel
expected string
}{
{commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNSPECIFIED, "UNSPECIFIED"},
{commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, "UNCLASSIFIED"},
{commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_A, "PROTECTED_A"},
{commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B, "PROTECTED_B"},
{commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_C, "PROTECTED_C"},
{commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET, "SECRET"},
}
for _, tt := range tests {
got := classification.LevelToString(tt.level)
if got != tt.expected {
t.Errorf("LevelToString(%v) = %q, want %q", tt.level, got, tt.expected)
}
}
}

func TestIsValid(t *testing.T) {
if classification.IsValid(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNSPECIFIED) {
t.Error("UNSPECIFIED should not be valid")
}
if !classification.IsValid(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED) {
t.Error("UNCLASSIFIED should be valid")
}
}

func TestPropagateMax(t *testing.T) {
levels := []commonv1.ClassificationLevel{
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_A,
}
result := classification.PropagateMax(levels)
if result != commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET {
t.Errorf("expected SECRET, got %v", result)
}
}
