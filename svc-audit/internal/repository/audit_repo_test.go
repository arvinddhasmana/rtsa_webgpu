// CLASSIFICATION: UNCLASSIFIED
package repository_test

import (
	"testing"

	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-audit/internal/repository"
)

func TestNewAuditRepository_InvalidDSN(t *testing.T) {
	_, err := repository.NewAuditRepository("not-a-valid-dsn")
	if err == nil {
		t.Error("expected error for invalid DSN, got nil")
	}
}

func TestBatchInsert_Empty(t *testing.T) {
	// With nil conn, BatchInsert on an empty slice returns nil before any DB call.
	// This tests the early-return path.
	repo := repository.NewAuditRepositoryFromConn(nil)
	err := repo.BatchInsert(nil, nil)
	if err != nil {
		t.Errorf("expected nil error for empty batch, got %v", err)
	}
}
