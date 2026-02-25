// CLASSIFICATION: UNCLASSIFIED
package domain

import (
"encoding/base64"
"encoding/json"
"fmt"
"time"
)

// PaginationToken encodes cursor state for efficient pagination.
// Encoded as base64(JSON) in the page_token field.
type PaginationToken struct {
LastID        string    `json:"lid"`
LastTimestamp time.Time `json:"lt"`
PageSize      int       `json:"ps"`
}

// EncodePaginationToken serializes the token to a base64 string.
func EncodePaginationToken(token *PaginationToken) string {
if token == nil {
return ""
}
b, err := json.Marshal(token)
if err != nil {
return ""
}
return base64.StdEncoding.EncodeToString(b)
}

// DecodePaginationToken deserializes from a base64 string.
// Returns nil if s is empty (first page).
func DecodePaginationToken(s string) (*PaginationToken, error) {
if s == "" {
return nil, nil
}
b, err := base64.StdEncoding.DecodeString(s)
if err != nil {
return nil, fmt.Errorf("paginator: invalid page_token encoding: %w", err)
}
var token PaginationToken
if err := json.Unmarshal(b, &token); err != nil {
return nil, fmt.Errorf("paginator: invalid page_token content: %w", err)
}
return &token, nil
}
