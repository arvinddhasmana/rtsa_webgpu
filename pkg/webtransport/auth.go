// CLASSIFICATION: UNCLASSIFIED
// pkg/webtransport/auth.go — JWT session authentication for WebTransport
//
// Validates short-lived JWT tokens presented by WebTransport clients.
// Tokens are issued by the gRPC cold-path auth service and carry the
// operator's clearance level for server-side classification filtering.
//
// Reference: docs/sdlc_guidelines/08_tech_specific/webtransport_guidelines.md §7.1

package webtransport

import (
"errors"
"fmt"
"net/http"
"time"

"github.com/golang-jwt/jwt/v5"
)

// SessionClaims holds operator identity and clearance from a validated JWT.
type SessionClaims struct {
OperatorID     string `json:"operator_id"`
ClearanceLevel uint32 `json:"clearance_level"` // mirrors ClassificationLevel enum
jwt.RegisteredClaims
}

// TokenValidator validates JWT tokens for WebTransport sessions.
type TokenValidator struct {
keyFunc jwt.Keyfunc
}

// NewTokenValidator creates a validator using an HMAC-SHA256 shared secret.
// In production, replace with an asymmetric key (RS256/ES256).
func NewTokenValidator(secret []byte) *TokenValidator {
return &TokenValidator{
keyFunc: func(t *jwt.Token) (interface{}, error) {
if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
return nil, fmt.Errorf("auth: unexpected signing method %v", t.Header["alg"])
}
return secret, nil
},
}
}

// ValidateRequest extracts and validates the JWT token from the HTTP request.
// The token is expected in the "token" query parameter (for WebTransport URL).
// Returns validated claims or an error.
//
// Reference: webtransport_guidelines.md §7.1
func (v *TokenValidator) ValidateRequest(r *http.Request) (*SessionClaims, error) {
raw := r.URL.Query().Get("token")
if raw == "" {
// Also check Authorization header as a fallback
authHeader := r.Header.Get("Authorization")
const prefix = "Bearer "
if len(authHeader) > len(prefix) {
raw = authHeader[len(prefix):]
}
}
if raw == "" {
return nil, errors.New("auth: missing token")
}
return v.ValidateToken(raw)
}

// ValidateToken parses and validates a raw JWT string.
func (v *TokenValidator) ValidateToken(raw string) (*SessionClaims, error) {
claims := &SessionClaims{}
token, err := jwt.ParseWithClaims(raw, claims, v.keyFunc,
jwt.WithExpirationRequired(),
jwt.WithIssuedAt(),
jwt.WithLeeway(30*time.Second),
)
if err != nil {
return nil, fmt.Errorf("auth: token validation failed: %w", err)
}
if !token.Valid {
return nil, errors.New("auth: token is invalid")
}
if claims.OperatorID == "" {
return nil, errors.New("auth: missing operator_id claim")
}
return claims, nil
}

// CreateToken mints a short-lived JWT for testing and development.
// In production, tokens are issued by the gRPC auth service.
func CreateToken(secret []byte, operatorID string, clearanceLevel uint32, ttl time.Duration) (string, error) {
now := time.Now()
claims := SessionClaims{
OperatorID:     operatorID,
ClearanceLevel: clearanceLevel,
RegisteredClaims: jwt.RegisteredClaims{
IssuedAt:  jwt.NewNumericDate(now),
ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
},
}
token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
return token.SignedString(secret)
}
