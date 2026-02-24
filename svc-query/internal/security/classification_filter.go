// CLASSIFICATION: UNCLASSIFIED

// Package security provides server-side classification enforcement for the query service.
// Classification filtering is ALWAYS injected server-side; client-supplied levels are NEVER trusted.
package security

import (
"context"
"log/slog"
"strconv"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
"google.golang.org/grpc/metadata"
)

// MetadataKeyClassification is the gRPC metadata key carrying the caller's clearance level.
// This is set by the mTLS interceptor from the client certificate.
// In the development environment it is forwarded by the API gateway as a header.
const MetadataKeyClassification = "x-rtsa-clearance-level"

// ClassificationFilter injects server-side classification filtering into SQL queries.
//
// Security guarantee: the caller's clearance is extracted exclusively from gRPC
// context metadata (set by the mTLS/auth interceptor). It is NEVER read from the
// request payload. If no clearance is present, the filter defaults to UNCLASSIFIED
// (least-privilege deny-by-default).
type ClassificationFilter struct{}

// NewClassificationFilter returns a new ClassificationFilter.
func NewClassificationFilter() *ClassificationFilter {
return &ClassificationFilter{}
}

// ExtractClearance reads the caller's clearance level from gRPC metadata.
// Returns CLASSIFICATION_LEVEL_UNCLASSIFIED if the metadata key is absent or invalid.
// This implements deny-by-default: an unknown caller sees only UNCLASSIFIED data.
func ExtractClearance(ctx context.Context) commonv1.ClassificationLevel {
md, ok := metadata.FromIncomingContext(ctx)
if !ok {
return commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED
}
vals := md.Get(MetadataKeyClassification)
if len(vals) == 0 {
return commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED
}
// Try numeric ordinal first
n, err := strconv.Atoi(vals[0])
if err == nil {
level := commonv1.ClassificationLevel(n)
if isValidClearance(level) {
return level
}
slog.Warn("security.ExtractClearance: unrecognised numeric clearance, defaulting to UNCLASSIFIED",
"raw_value", vals[0])
return commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED
}
// Try string name
if v, ok2 := commonv1.ClassificationLevel_value[vals[0]]; ok2 {
return commonv1.ClassificationLevel(v)
}
slog.Warn("security.ExtractClearance: unrecognised string clearance, defaulting to UNCLASSIFIED",
"raw_value", vals[0])
return commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED
}

// InjectFilter appends a classification_level <= ? predicate to a SQL WHERE clause.
//
// The ClickHouse Enum8 ordinals map directly to the proto enum values:
//
//UNCLASSIFIED=1, PROTECTED_A=2, PROTECTED_B=3, PROTECTED_C=4, SECRET=5
//
// Using toInt8(classification_level) compares against the integer ordinal,
// regardless of enum name lexicographic ordering.
//
// Returns the modified query and the ordinal int8 to bind as the final positional parameter.
// The classification parameter is ALWAYS the last parameter appended by this function.
func (f *ClassificationFilter) InjectFilter(
query string,
callerClearance commonv1.ClassificationLevel,
) (string, int8) {
ordinal := classificationOrdinal(callerClearance)
query += " AND toInt8(classification_level) <= ?"
return query, ordinal
}

// classificationOrdinal maps a ClassificationLevel to its Enum8 ordinal.
// Defaults to 1 (UNCLASSIFIED) for any unrecognised or UNSPECIFIED value.
func classificationOrdinal(level commonv1.ClassificationLevel) int8 {
switch level {
case commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED:
return 1
case commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_A:
return 2
case commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B:
return 3
case commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_C:
return 4
case commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET:
return 5
default:
// UNSPECIFIED or unknown -> deny-by-default -> UNCLASSIFIED
return 1
}
}

// isValidClearance reports whether level is a recognised, non-zero clearance.
func isValidClearance(level commonv1.ClassificationLevel) bool {
switch level {
case commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_A,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_C,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET:
return true
default:
return false
}
}
