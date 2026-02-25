// CLASSIFICATION: UNCLASSIFIED
package mapper_test

import (
"context"
"testing"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/classification"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-isr-ingestion/internal/mapper"
)

func TestEnricher_SetsObservationID(t *testing.T) {
guard := classification.NewGuard(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET)
e := mapper.NewEnricher("svc-isr-ingestion", guard)

obs := &ingestionv1.SensorObservation{
Classification: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
}

if err := e.Enrich(context.Background(), obs); err != nil {
t.Fatalf("unexpected error: %v", err)
}
if obs.ObservationId == "" {
t.Error("expected observation_id to be set")
}
}

func TestEnricher_DoesNotOverwriteExistingID(t *testing.T) {
guard := classification.NewGuard(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET)
e := mapper.NewEnricher("svc-isr-ingestion", guard)

obs := &ingestionv1.SensorObservation{
ObservationId:  "existing-id-123",
Classification: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
}

if err := e.Enrich(context.Background(), obs); err != nil {
t.Fatalf("unexpected error: %v", err)
}
if obs.ObservationId != "existing-id-123" {
t.Errorf("expected existing ID preserved, got: %s", obs.ObservationId)
}
}

func TestEnricher_SetsMetadata(t *testing.T) {
guard := classification.NewGuard(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET)
e := mapper.NewEnricher("svc-isr-ingestion", guard)

obs := &ingestionv1.SensorObservation{
Classification: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
}

if err := e.Enrich(context.Background(), obs); err != nil {
t.Fatalf("unexpected error: %v", err)
}
if obs.Metadata["rtsa.source_service"] != "svc-isr-ingestion" {
t.Errorf("expected rtsa.source_service to be set, got: %s", obs.Metadata["rtsa.source_service"])
}
if obs.Metadata["rtsa.ingestion_time"] == "" {
t.Error("expected rtsa.ingestion_time to be set")
}
}

func TestEnricher_ClassificationCeilingViolation(t *testing.T) {
guard := classification.NewGuard(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED)
e := mapper.NewEnricher("svc-isr-ingestion", guard)

obs := &ingestionv1.SensorObservation{
Classification: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
}

err := e.Enrich(context.Background(), obs)
if err == nil {
t.Fatal("expected error for classification ceiling violation")
}
}

func TestEnricher_NilGuard(t *testing.T) {
e := mapper.NewEnricher("svc-isr-ingestion", nil)

obs := &ingestionv1.SensorObservation{
Classification: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
}

if err := e.Enrich(context.Background(), obs); err != nil {
t.Errorf("unexpected error with nil guard: %v", err)
}
}

func TestEnricher_InitializesNilMetadata(t *testing.T) {
guard := classification.NewGuard(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET)
e := mapper.NewEnricher("svc-isr-ingestion", guard)

obs := &ingestionv1.SensorObservation{
Classification: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
Metadata:       nil,
}

if err := e.Enrich(context.Background(), obs); err != nil {
t.Fatalf("unexpected error: %v", err)
}
if obs.Metadata == nil {
t.Error("expected metadata to be initialized")
}
}

func TestClassificationLevel_Helper(t *testing.T) {
obs := &ingestionv1.SensorObservation{
Classification: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
}
level := mapper.ClassificationLevel(obs)
if level != commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED {
t.Errorf("expected UNCLASSIFIED, got %v", level)
}
}
