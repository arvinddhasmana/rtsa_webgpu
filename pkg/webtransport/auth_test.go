// CLASSIFICATION: UNCLASSIFIED
// pkg/webtransport/auth_test.go — Unit tests for JWT authentication

package webtransport_test

import (
"net/http"
"net/url"
"testing"
"time"

"github.com/arvinddhasmana/rtsa_webgpu/pkg/webtransport"
)

var testSecret = []byte("test-secret-key-for-unit-tests-only")

func makeRequest(token string) *http.Request {
u := &url.URL{
Scheme:   "https",
Host:     "localhost:4443",
Path:     "/wt",
RawQuery: "token=" + token,
}
return &http.Request{
Method: http.MethodConnect,
URL:    u,
Header: http.Header{},
}
}

func TestValidateRequest_ValidToken(t *testing.T) {
validator := webtransport.NewTokenValidator(testSecret)
raw, err := webtransport.CreateToken(testSecret, "operator-42", 3, 5*time.Minute)
if err != nil {
t.Fatalf("CreateToken: %v", err)
}

claims, err := validator.ValidateRequest(makeRequest(raw))
if err != nil {
t.Fatalf("ValidateRequest returned error: %v", err)
}
if claims.OperatorID != "operator-42" {
t.Errorf("OperatorID: want operator-42, got %s", claims.OperatorID)
}
if claims.ClearanceLevel != 3 {
t.Errorf("ClearanceLevel: want 3, got %d", claims.ClearanceLevel)
}
}

func TestValidateRequest_MissingToken(t *testing.T) {
validator := webtransport.NewTokenValidator(testSecret)
r := &http.Request{URL: &url.URL{}, Header: http.Header{}}
_, err := validator.ValidateRequest(r)
if err == nil {
t.Error("expected error for missing token")
}
}

func TestValidateRequest_ExpiredToken(t *testing.T) {
validator := webtransport.NewTokenValidator(testSecret)
// Create a token that is already expired (negative TTL simulation using leeway gap)
raw, err := webtransport.CreateToken(testSecret, "op-1", 1, -1*time.Minute)
if err != nil {
t.Fatalf("CreateToken: %v", err)
}
_, err = validator.ValidateRequest(makeRequest(raw))
if err == nil {
t.Error("expected error for expired token")
}
}

func TestValidateRequest_WrongSecret(t *testing.T) {
validator := webtransport.NewTokenValidator(testSecret)
otherSecret := []byte("wrong-secret")
raw, err := webtransport.CreateToken(otherSecret, "op-2", 1, 5*time.Minute)
if err != nil {
t.Fatalf("CreateToken: %v", err)
}
_, err = validator.ValidateRequest(makeRequest(raw))
if err == nil {
t.Error("expected error for wrong secret")
}
}

func TestValidateRequest_MissingOperatorID(t *testing.T) {
// Create a valid token but without operator_id via direct API
// (CreateToken always sets operator_id so we test ValidateToken directly)
validator := webtransport.NewTokenValidator(testSecret)
raw, err := webtransport.CreateToken(testSecret, "", 1, 5*time.Minute)
if err != nil {
// CreateToken may itself reject empty operator_id — either is acceptable
return
}
_, err = validator.ValidateToken(raw)
if err == nil {
t.Error("expected error for missing operator_id")
}
}

func TestValidateRequest_BearerHeader(t *testing.T) {
validator := webtransport.NewTokenValidator(testSecret)
raw, err := webtransport.CreateToken(testSecret, "op-bearer", 2, 5*time.Minute)
if err != nil {
t.Fatalf("CreateToken: %v", err)
}
r := &http.Request{
URL: &url.URL{},
Header: http.Header{
"Authorization": []string{"Bearer " + raw},
},
}
claims, err := validator.ValidateRequest(r)
if err != nil {
t.Fatalf("ValidateRequest via Authorization header: %v", err)
}
if claims.OperatorID != "op-bearer" {
t.Errorf("OperatorID: want op-bearer, got %s", claims.OperatorID)
}
}

