// CLASSIFICATION: UNCLASSIFIED

package security

import (
"context"
"testing"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
"google.golang.org/grpc/metadata"
)

func TestExtractClearance_noMetadata(t *testing.T) {
ctx := context.Background()
got := ExtractClearance(ctx)
want := commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED
if got != want {
t.Errorf("ExtractClearance() = %v, want %v", got, want)
}
}

func TestExtractClearance_emptyMetadata(t *testing.T) {
md := metadata.New(nil)
ctx := metadata.NewIncomingContext(context.Background(), md)
got := ExtractClearance(ctx)
if got != commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED {
t.Errorf("expected UNCLASSIFIED for empty metadata, got %v", got)
}
}

func TestExtractClearance_numericOrdinals(t *testing.T) {
tests := []struct {
raw  string
want commonv1.ClassificationLevel
}{
{"1", commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED},
{"2", commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_A},
{"3", commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B},
{"4", commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_C},
{"5", commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET},
{"99", commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED},
{"0", commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED},
}
for _, tc := range tests {
t.Run(tc.raw, func(t *testing.T) {
md := metadata.Pairs(MetadataKeyClassification, tc.raw)
ctx := metadata.NewIncomingContext(context.Background(), md)
got := ExtractClearance(ctx)
if got != tc.want {
t.Errorf("ExtractClearance(%q) = %v, want %v", tc.raw, got, tc.want)
}
})
}
}

func TestExtractClearance_stringNames(t *testing.T) {
tests := []struct {
raw  string
want commonv1.ClassificationLevel
}{
{"CLASSIFICATION_LEVEL_UNCLASSIFIED", commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED},
{"CLASSIFICATION_LEVEL_SECRET", commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET},
{"TOP_SECRET", commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED},
}
for _, tc := range tests {
t.Run(tc.raw, func(t *testing.T) {
md := metadata.Pairs(MetadataKeyClassification, tc.raw)
ctx := metadata.NewIncomingContext(context.Background(), md)
got := ExtractClearance(ctx)
if got != tc.want {
t.Errorf("ExtractClearance(%q) = %v, want %v", tc.raw, got, tc.want)
}
})
}
}

func TestInjectFilter(t *testing.T) {
f := NewClassificationFilter()

tests := []struct {
name         string
query        string
clearance    commonv1.ClassificationLevel
wantOrdinal  int8
wantContains string
}{
{
name:         "UNCLASSIFIED caller",
query:        "SELECT * FROM tracks_fused WHERE event_time >= ?",
clearance:    commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
wantOrdinal:  1,
wantContains: "toInt8(classification_level) <= ?",
},
{
name:         "PROTECTED_B caller",
query:        "SELECT * FROM tracks_fused WHERE event_time >= ?",
clearance:    commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B,
wantOrdinal:  3,
wantContains: "toInt8(classification_level) <= ?",
},
{
name:         "SECRET caller",
query:        "SELECT * FROM tracks_fused WHERE event_time >= ?",
clearance:    commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
wantOrdinal:  5,
wantContains: "toInt8(classification_level) <= ?",
},
{
name:         "UNSPECIFIED deny-by-default",
query:        "SELECT * FROM tracks_fused WHERE event_time >= ?",
clearance:    commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNSPECIFIED,
wantOrdinal:  1,
wantContains: "toInt8(classification_level) <= ?",
},
}

for _, tc := range tests {
t.Run(tc.name, func(t *testing.T) {
gotQuery, gotOrdinal := f.InjectFilter(tc.query, tc.clearance)

if gotOrdinal != tc.wantOrdinal {
t.Errorf("ordinal = %d, want %d", gotOrdinal, tc.wantOrdinal)
}
found := false
for i := 0; i <= len(gotQuery)-len(tc.wantContains); i++ {
if gotQuery[i:i+len(tc.wantContains)] == tc.wantContains {
found = true
break
}
}
if !found {
t.Errorf("query does not contain %q:\n  got: %q", tc.wantContains, gotQuery)
}
})
}
}

// TestClassificationHierarchy verifies that higher clearance gives higher ordinal.
func TestClassificationHierarchy(t *testing.T) {
f := NewClassificationFilter()

_, secretOrdinal := f.InjectFilter("", commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET)
_, protCOrdinal := f.InjectFilter("", commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_C)
_, protBOrdinal := f.InjectFilter("", commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B)
_, protAOrdinal := f.InjectFilter("", commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_A)
_, unclassOrdinal := f.InjectFilter("", commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED)

if !(secretOrdinal > protCOrdinal && protCOrdinal > protBOrdinal &&
protBOrdinal > protAOrdinal && protAOrdinal > unclassOrdinal) {
t.Errorf("hierarchy violated: SECRET=%d PROT_C=%d PROT_B=%d PROT_A=%d UNCLASS=%d",
secretOrdinal, protCOrdinal, protBOrdinal, protAOrdinal, unclassOrdinal)
}
}
