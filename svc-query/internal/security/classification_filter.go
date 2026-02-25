// CLASSIFICATION: UNCLASSIFIED
package security

import (
"context"
"fmt"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
"google.golang.org/grpc/metadata"
)

const metadataKeyClassification = "x-rtsa-clearance"

// ClassificationFilter injects server-side classification filtering into queries.
// NEVER trust client-supplied classification level.
type ClassificationFilter struct{}

// InjectFilter appends a classification WHERE clause to the query.
// Returns the modified query and the classification ordinal parameter.
func (f *ClassificationFilter) InjectFilter(
query string,
callerClearance commonv1.ClassificationLevel,
) (string, interface{}) {
ordinal := classificationOrdinal(callerClearance)
return fmt.Sprintf("%s AND classification_level <= ?", query), ordinal
}

// ExtractClearance retrieves the caller's classification clearance from gRPC
// context metadata. Defaults to UNCLASSIFIED (deny by default — least privilege).
func ExtractClearance(ctx context.Context) commonv1.ClassificationLevel {
md, ok := metadata.FromIncomingContext(ctx)
if !ok {
return commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED
}
vals := md.Get(metadataKeyClassification)
if len(vals) == 0 {
return commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED
}
switch vals[0] {
case "PROTECTED_A":
return commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_A
case "PROTECTED_B":
return commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B
case "PROTECTED_C":
return commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_C
case "SECRET":
return commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET
default:
return commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED
}
}

// classificationOrdinal maps ClassificationLevel to the ClickHouse Enum8 ordinal.
func classificationOrdinal(level commonv1.ClassificationLevel) int32 {
switch level {
case commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_A:
return 2
case commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B:
return 3
case commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_C:
return 4
case commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET:
return 5
default:
return 1 // UNCLASSIFIED
}
}
