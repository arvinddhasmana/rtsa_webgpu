// CLASSIFICATION: UNCLASSIFIED

package domain

import (
"encoding/base64"
"encoding/json"
"fmt"
"time"
)

// PaginationToken encodes cursor state for efficient keyset pagination.
// It is serialized as base64(JSON) in the page_token field of responses.
// The cursor is the (LastTimestamp, LastID) pair — the last row seen.
type PaginationToken struct {
LastID        string    `json:"lid"`
LastTimestamp time.Time `json:"lt"`
PageSize      int       `json:"ps"`
}

// EncodePaginationToken serializes a PaginationToken to an opaque base64 string
// suitable for inclusion in a gRPC response page_token field.
// Returns an empty string if token is nil.
func EncodePaginationToken(token *PaginationToken) string {
if token == nil {
return ""
}
b, err := json.Marshal(token)
if err != nil {
// json.Marshal on a struct with primitive types cannot fail.
return ""
}
return base64.StdEncoding.EncodeToString(b)
}

// DecodePaginationToken deserializes a page_token string.
// Returns nil if s is empty (first page).
// Returns an error if s is non-empty but malformed.
func DecodePaginationToken(s string) (*PaginationToken, error) {
if s == "" {
return nil, nil
}
raw, err := base64.StdEncoding.DecodeString(s)
if err != nil {
return nil, fmt.Errorf("domain.DecodePaginationToken: base64 decode: %w", err)
}
var token PaginationToken
if err := json.Unmarshal(raw, &token); err != nil {
return nil, fmt.Errorf("domain.DecodePaginationToken: json unmarshal: %w", err)
}
if token.PageSize <= 0 {
return nil, fmt.Errorf("domain.DecodePaginationToken: invalid page size %d", token.PageSize)
}
return &token, nil
}

// ApplyPagination appends cursor-based WHERE/ORDER/LIMIT clauses to a query.
//
// When token is nil (first page):
//
//ORDER BY event_time ASC, <idCol> ASC LIMIT ?
//
// When token is set (subsequent page), it adds:
//
//AND (event_time > ? OR (event_time = ? AND <idCol> > ?))
//ORDER BY event_time ASC, <idCol> ASC LIMIT ?
//
// idCol is the column name used as a secondary sort key (e.g. "track_id").
// params receives the additional bind parameters.
// The effective limit is always taken from the guardrail-enforced pageSize.
func ApplyPagination(
query string,
params []interface{},
token *PaginationToken,
idCol string,
pageSize int,
) (string, []interface{}) {
if token != nil {
query += fmt.Sprintf(
" AND (event_time > ? OR (event_time = ? AND %s > ?))",
idCol,
)
params = append(params, token.LastTimestamp, token.LastTimestamp, token.LastID)
}
query += fmt.Sprintf(" ORDER BY event_time ASC, %s ASC LIMIT ?", idCol)
params = append(params, pageSize)
return query, params
}
