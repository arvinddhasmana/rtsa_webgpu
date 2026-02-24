// CLASSIFICATION: UNCLASSIFIED
package testutil

import (
"strings"
"testing"
)

// AssertNoError fails the test if err is not nil.
func AssertNoError(t *testing.T, err error) {
t.Helper()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
}

// AssertError fails the test if err is nil.
func AssertError(t *testing.T, err error) {
t.Helper()
if err == nil {
t.Fatal("expected error, got nil")
}
}

// AssertErrorContains fails if err is nil or does not contain the substring.
func AssertErrorContains(t *testing.T, err error, substr string) {
t.Helper()
if err == nil {
t.Fatalf("expected error containing %q, got nil", substr)
}
if !strings.Contains(err.Error(), substr) {
t.Fatalf("expected error to contain %q, got: %v", substr, err)
}
}

// AssertEqual fails if a != b.
func AssertEqual[T comparable](t *testing.T, expected, actual T) {
t.Helper()
if expected != actual {
t.Fatalf("expected %v, got %v", expected, actual)
}
}

// AssertTrue fails if cond is false.
func AssertTrue(t *testing.T, cond bool, msg string) {
t.Helper()
if !cond {
t.Fatalf("assertion failed: %s", msg)
}
}
