// CLASSIFICATION: UNCLASSIFIED
package security

import (
	"context"
	"fmt"
	"strings"

	commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
	"google.golang.org/grpc/metadata"
)

const metadataKeyClassification = "x-rtsa-clearance"

// ClassificationFilter injects server-side classification filtering into queries.
// NEVER trust client-supplied classification level.
type ClassificationFilter struct{}

// InjectFilter appends a classification WHERE clause to the query using an
// IN (...) list of string values that matches the LowCardinality(String) column
// type used in all RTSA ClickHouse tables.
// Returns the modified query and a slice of allowed classification level string params.
func (f *ClassificationFilter) InjectFilter(
query string,
callerClearance commonv1.ClassificationLevel,
) (string, []interface{}) {
allowed := allowedClassificationStrings(callerClearance)
placeholders := make([]string, len(allowed))
for i := range allowed {
placeholders[i] = "?"
}
params := make([]interface{}, len(allowed))
for i, v := range allowed {
params[i] = v
}
clause := fmt.Sprintf("AND classification_level IN (%s)", strings.Join(placeholders, ","))
return fmt.Sprintf("%s %s", query, clause), params
}

// allowedClassificationStrings returns the ordered list of ClickHouse string
// values for every classification level the caller is permitted to see.
// All levels with ordinal ≤ callerClearance are included.
func allowedClassificationStrings(level commonv1.ClassificationLevel) []string {
// Ordered least-sensitive → most-sensitive, matching classificationOrdinal.
all := []string{
"CLASSIFICATION_LEVEL_UNCLASSIFIED",
"CLASSIFICATION_LEVEL_PROTECTED_A",
"CLASSIFICATION_LEVEL_PROTECTED_B",
"CLASSIFICATION_LEVEL_PROTECTED_C",
"CLASSIFICATION_LEVEL_SECRET",
}
ord := int(classificationOrdinal(level)) // 1=UNCLASSIFIED … 5=SECRET
if ord > len(all) {
ord = len(all)
}
return all[:ord]
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
