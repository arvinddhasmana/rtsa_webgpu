// CLASSIFICATION: UNCLASSIFIED
package domain_test

import (
	"math"
	"testing"
	"time"

	"github.com/arvinddhasmana/rtsa_webgpu/svc-fusion-engine/internal/domain"
)

func TestKalmanFilter_Predict_ConstantVelocity(t *testing.T) {
	kf := domain.NewKalmanFilter()
	now := time.Now()
	// Initial state: position 45°N, -60°E, vN=10 m/s, vE=0
	m := &domain.Measurement{
		Latitude:  45.0,
		Longitude: -60.0,
		Time:      now,
	}
	vN := 10.0
	m.VelocityN = &vN

	state := kf.InitState(m)
	state.VelocityN = 10.0

	before := state.Latitude
	kf.Predict(state, 10.0) // predict 10 seconds forward

	// Expected latitude advance: 10 m/s * 10 s / 111320 m/deg ≈ 0.000898 deg
	expectedDLat := 10.0 * 10.0 / 111_320.0
	gotDLat := state.Latitude - before

	if math.Abs(gotDLat-expectedDLat) > 1e-7 {
		t.Errorf("expected Δlat ≈ %.8f, got %.8f", expectedDLat, gotDLat)
	}
}

// T11 — Kalman update: reduces uncertainty (P diagonal decreases)
func TestKalmanFilter_Update_ReducesUncertainty(t *testing.T) {
	kf := domain.NewKalmanFilter()
	now := time.Now()
	m := &domain.Measurement{
		Latitude:  45.0,
		Longitude: -60.0,
		Time:      now,
	}
	state := kf.InitState(m)
	P_before := state.P

	// Update with same position
	meas := &domain.Measurement{
		Latitude:  45.0,
		Longitude: -60.0,
		Time:      now.Add(time.Second),
	}
	kf.Update(state, meas)

	// Diagonal of P should decrease
	if state.P[0] >= P_before[0] {
		t.Errorf("P[0,0] should decrease after update: before=%.6f after=%.6f", P_before[0], state.P[0])
	}
	if state.P[5] >= P_before[5] {
		t.Errorf("P[1,1] should decrease after update: before=%.6f after=%.6f", P_before[5], state.P[5])
	}
}

// Predict with zero dt should be a no-op
func TestKalmanFilter_Predict_ZeroDt(t *testing.T) {
	kf := domain.NewKalmanFilter()
	now := time.Now()
	m := &domain.Measurement{Latitude: 45.0, Longitude: -60.0, Time: now}
	state := kf.InitState(m)
	state.VelocityN = 100.0

	before := state.Latitude
	kf.Predict(state, 0)
	if state.Latitude != before {
		t.Errorf("zero-dt predict changed latitude: %.8f vs %.8f", before, state.Latitude)
	}
}

// Update with velocity measurement should refine state
func TestKalmanFilter_Update4_ConvergesPosition(t *testing.T) {
	kf := domain.NewKalmanFilter()
	now := time.Now()
	m := &domain.Measurement{Latitude: 45.0, Longitude: -60.0, Time: now}
	state := kf.InitState(m)

	vN := 5.0
	vE := 3.0
	meas := &domain.Measurement{
		Latitude:  45.001,
		Longitude: -60.001,
		VelocityN: &vN,
		VelocityE: &vE,
		Time:      now.Add(time.Second),
	}
	kf.Update(state, meas)

	// Position should move towards measurement
	if state.Latitude <= 45.0 {
		t.Errorf("latitude should move towards measurement 45.001, got %.6f", state.Latitude)
	}
}
