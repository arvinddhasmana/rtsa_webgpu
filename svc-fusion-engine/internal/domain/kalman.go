// CLASSIFICATION: UNCLASSIFIED
package domain

import (
	"math"
	"time"
)

const (
	// metersPerDegLat converts meters to degrees of latitude.
	metersPerDegLat = 111_320.0
)

// KalmanState holds the 4-state kinematic estimate [lat, lon, vN, vE].
type KalmanState struct {
	Latitude   float64 // degrees
	Longitude  float64 // degrees
	VelocityN  float64 // m/s north
	VelocityE  float64 // m/s east
	P          [16]float64 // 4x4 covariance matrix, row-major
	LastUpdate time.Time
}

// Measurement represents a sensor position/velocity input for Kalman update.
type Measurement struct {
	Latitude  float64
	Longitude float64
	VelocityN *float64 // nil if not available
	VelocityE *float64 // nil if not available
	Time      time.Time
}

// KalmanFilter implements a 4-state constant-velocity Kalman filter.
type KalmanFilter struct {
	processNoise        float64 // spectral density (m/s²/√Hz)
	measurementNoisePos float64 // position noise (deg²)
	measurementNoiseVel float64 // velocity noise (m²/s²)
}

// NewKalmanFilter creates a Kalman filter with default noise parameters.
func NewKalmanFilter() *KalmanFilter {
	return &KalmanFilter{
		processNoise:        0.1,
		measurementNoisePos: 1e-6,
		measurementNoiseVel: 1.0,
	}
}

// InitState initialises a KalmanState from a position measurement.
func (kf *KalmanFilter) InitState(m *Measurement) *KalmanState {
	ks := &KalmanState{
		Latitude:   m.Latitude,
		Longitude:  m.Longitude,
		LastUpdate: m.Time,
	}
	if m.VelocityN != nil {
		ks.VelocityN = *m.VelocityN
	}
	if m.VelocityE != nil {
		ks.VelocityE = *m.VelocityE
	}
	// Initial covariance: large position uncertainty, moderate velocity uncertainty.
	ks.P = [16]float64{
		1e-4, 0, 0, 0,
		0, 1e-4, 0, 0,
		0, 0, 10, 0,
		0, 0, 0, 10,
	}
	return ks
}

// Predict extrapolates the Kalman state forward by dt seconds using a constant-velocity model.
// F = [[1,0,dt/dLat,0],[0,1,0,dt/dLon],[0,0,1,0],[0,0,0,1]]
func (kf *KalmanFilter) Predict(state *KalmanState, dt float64) {
	if dt <= 0 {
		return
	}
	lat := state.Latitude
	dLat := metersPerDegLat
	dLon := metersPerDegLat * math.Cos(degreesToRadians(lat))
	if dLon == 0 {
		dLon = 1 // avoid division by zero at poles
	}

	// Advance position
	state.Latitude += state.VelocityN * dt / dLat
	state.Longitude += state.VelocityE * dt / dLon

	// F matrix (4x4):
	// [1  0  dt/dLat   0     ]
	// [0  1  0         dt/dLon]
	// [0  0  1         0     ]
	// [0  0  0         1     ]
	dtLat := dt / dLat
	dtLon := dt / dLon
	F := [16]float64{
		1, 0, dtLat, 0,
		0, 1, 0, dtLon,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}

	// P = F * P * F' + Q
	state.P = kf.predictCovariance(state.P, F, dt)
	state.LastUpdate = state.LastUpdate.Add(time.Duration(dt * float64(time.Second)))
}

// predictCovariance computes F*P*F' + Q.
func (kf *KalmanFilter) predictCovariance(P [16]float64, F [16]float64, dt float64) [16]float64 {
	// P' = F * P
	FP := mat4Mul(F, P)
	// P'' = (F*P) * F' = F * P * F^T
	var FT [16]float64
	for r := 0; r < 4; r++ {
		for c := 0; c < 4; c++ {
			FT[r*4+c] = F[c*4+r]
		}
	}
	FPFT := mat4Mul(FP, FT)

	// Process noise Q: diagonal, scaled by dt
	q := kf.processNoise * kf.processNoise * dt
	FPFT[0] += q * 0.25 * dt * dt / (metersPerDegLat * metersPerDegLat)
	FPFT[5] += q * 0.25 * dt * dt / (metersPerDegLat * metersPerDegLat)
	FPFT[10] += q
	FPFT[15] += q

	return FPFT
}

// Update incorporates a new measurement into the Kalman state.
// Uses standard Kalman update: K = P*H'*(H*P*H'+R)^-1, x += K*(z-Hx), P = (I-KH)*P.
func (kf *KalmanFilter) Update(state *KalmanState, m *Measurement) {
	if m.VelocityN != nil && m.VelocityE != nil {
		kf.update4(state, m)
	} else {
		kf.update2(state, m)
	}
	state.LastUpdate = m.Time
}

// update2 performs a 2-measurement update (position only).
func (kf *KalmanFilter) update2(state *KalmanState, m *Measurement) {
	// z = [lat, lon]
	// H = [[1,0,0,0],[0,1,0,0]]
	// R = diag(rPos, rPos)
	rPos := kf.measurementNoisePos

	// Innovation: y = z - H*x
	yLat := m.Latitude - state.Latitude
	yLon := m.Longitude - state.Longitude

	// S = H*P*H' + R  (2x2)
	// H*P selects first two rows: p[0..3], p[4..7]
	// (H*P*H')_ij = sum_k H_ik * P_km * H_jm = P[i*4+j] for i,j in {0,1}
	s00 := state.P[0] + rPos
	s01 := state.P[1]
	s10 := state.P[4]
	s11 := state.P[5] + rPos

	// S^-1 (2x2)
	det := s00*s11 - s01*s10
	if math.Abs(det) < 1e-30 {
		return
	}
	inv00 := s11 / det
	inv01 := -s01 / det
	inv10 := -s10 / det
	inv11 := s00 / det

	// K = P * H' * S^-1  (4x2)
	// P*H' selects columns 0 and 1 of P.
	ph := [8]float64{
		state.P[0], state.P[1],
		state.P[4], state.P[5],
		state.P[8], state.P[9],
		state.P[12], state.P[13],
	}
	// K[i][j] = sum_k ph[i][k] * inv[k][j]
	K := [8]float64{}
	for i := 0; i < 4; i++ {
		K[i*2+0] = ph[i*2+0]*inv00 + ph[i*2+1]*inv10
		K[i*2+1] = ph[i*2+0]*inv01 + ph[i*2+1]*inv11
	}

	// x = x + K * y
	state.Latitude += K[0]*yLat + K[1]*yLon
	state.Longitude += K[2]*yLat + K[3]*yLon
	state.VelocityN += K[4]*yLat + K[5]*yLon
	state.VelocityE += K[6]*yLat + K[7]*yLon

	// P = (I - K*H) * P  using Joseph form for numerical stability
	// K*H (4x4): first two columns from K, rest zero
	var KH [16]float64
	for i := 0; i < 4; i++ {
		KH[i*4+0] = K[i*2+0]
		KH[i*4+1] = K[i*2+1]
	}
	state.P = applyJosephUpdate(state.P, KH)
}

// update4 performs a 4-measurement update (position + velocity).
func (kf *KalmanFilter) update4(state *KalmanState, m *Measurement) {
	rPos := kf.measurementNoisePos
	rVel := kf.measurementNoiseVel

	// z = [lat, lon, vN, vE]
	// H = I4
	// S = P + R
	S := state.P
	S[0] += rPos
	S[5] += rPos
	S[10] += rVel
	S[15] += rVel

	Sinv, ok := invert4x4(S)
	if !ok {
		return
	}

	// K = P * S^-1
	K := mat4Mul(state.P, Sinv)

	// innovation y = z - x
	y := [4]float64{
		m.Latitude - state.Latitude,
		m.Longitude - state.Longitude,
		*m.VelocityN - state.VelocityN,
		*m.VelocityE - state.VelocityE,
	}

	// x = x + K * y
	Ky := mat4Vec(K, y)
	state.Latitude += Ky[0]
	state.Longitude += Ky[1]
	state.VelocityN += Ky[2]
	state.VelocityE += Ky[3]

	// P = (I - K) * P
	var KH [16]float64
	copy(KH[:], K[:])
	state.P = applyJosephUpdate(state.P, KH)
}

// applyJosephUpdate computes (I - KH) * P using the Joseph stabilised form.
func applyJosephUpdate(P, KH [16]float64) [16]float64 {
	// IminusKH = I - KH
	IminusKH := KH
	for i := 0; i < 4; i++ {
		IminusKH[i*4+i] = 1.0 - KH[i*4+i]
		for j := 0; j < 4; j++ {
			if i != j {
				IminusKH[i*4+j] = -KH[i*4+j]
			}
		}
	}
	return mat4Mul(IminusKH, P)
}

// mat4Mul multiplies two 4x4 matrices stored in row-major order.
func mat4Mul(A, B [16]float64) [16]float64 {
	var C [16]float64
	for r := 0; r < 4; r++ {
		for c := 0; c < 4; c++ {
			for k := 0; k < 4; k++ {
				C[r*4+c] += A[r*4+k] * B[k*4+c]
			}
		}
	}
	return C
}

// mat4Vec multiplies a 4x4 matrix by a 4-vector.
func mat4Vec(A [16]float64, v [4]float64) [4]float64 {
	var out [4]float64
	for r := 0; r < 4; r++ {
		for c := 0; c < 4; c++ {
			out[r] += A[r*4+c] * v[c]
		}
	}
	return out
}

// invert4x4 computes the inverse of a 4x4 matrix using cofactor expansion.
// Returns (inverse, true) on success or (zero, false) if singular.
func invert4x4(m [16]float64) ([16]float64, bool) {
	var inv [16]float64
	inv[0] = m[5]*m[10]*m[15] - m[5]*m[11]*m[14] - m[9]*m[6]*m[15] + m[9]*m[7]*m[14] + m[13]*m[6]*m[11] - m[13]*m[7]*m[10]
	inv[4] = -m[4]*m[10]*m[15] + m[4]*m[11]*m[14] + m[8]*m[6]*m[15] - m[8]*m[7]*m[14] - m[12]*m[6]*m[11] + m[12]*m[7]*m[10]
	inv[8] = m[4]*m[9]*m[15] - m[4]*m[11]*m[13] - m[8]*m[5]*m[15] + m[8]*m[7]*m[13] + m[12]*m[5]*m[11] - m[12]*m[7]*m[9]
	inv[12] = -m[4]*m[9]*m[14] + m[4]*m[10]*m[13] + m[8]*m[5]*m[14] - m[8]*m[6]*m[13] - m[12]*m[5]*m[10] + m[12]*m[6]*m[9]
	inv[1] = -m[1]*m[10]*m[15] + m[1]*m[11]*m[14] + m[9]*m[2]*m[15] - m[9]*m[3]*m[14] - m[13]*m[2]*m[11] + m[13]*m[3]*m[10]
	inv[5] = m[0]*m[10]*m[15] - m[0]*m[11]*m[14] - m[8]*m[2]*m[15] + m[8]*m[3]*m[14] + m[12]*m[2]*m[11] - m[12]*m[3]*m[10]
	inv[9] = -m[0]*m[9]*m[15] + m[0]*m[11]*m[13] + m[8]*m[1]*m[15] - m[8]*m[3]*m[13] - m[12]*m[1]*m[11] + m[12]*m[3]*m[9]
	inv[13] = m[0]*m[9]*m[14] - m[0]*m[10]*m[13] - m[8]*m[1]*m[14] + m[8]*m[2]*m[13] + m[12]*m[1]*m[10] - m[12]*m[2]*m[9]
	inv[2] = m[1]*m[6]*m[15] - m[1]*m[7]*m[14] - m[5]*m[2]*m[15] + m[5]*m[3]*m[14] + m[13]*m[2]*m[7] - m[13]*m[3]*m[6]
	inv[6] = -m[0]*m[6]*m[15] + m[0]*m[7]*m[14] + m[4]*m[2]*m[15] - m[4]*m[3]*m[14] - m[12]*m[2]*m[7] + m[12]*m[3]*m[6]
	inv[10] = m[0]*m[5]*m[15] - m[0]*m[7]*m[13] - m[4]*m[1]*m[15] + m[4]*m[3]*m[13] + m[12]*m[1]*m[7] - m[12]*m[3]*m[5]
	inv[14] = -m[0]*m[5]*m[14] + m[0]*m[6]*m[13] + m[4]*m[1]*m[14] - m[4]*m[2]*m[13] - m[12]*m[1]*m[6] + m[12]*m[2]*m[5]
	inv[3] = -m[1]*m[6]*m[11] + m[1]*m[7]*m[10] + m[5]*m[2]*m[11] - m[5]*m[3]*m[10] - m[9]*m[2]*m[7] + m[9]*m[3]*m[6]
	inv[7] = m[0]*m[6]*m[11] - m[0]*m[7]*m[10] - m[4]*m[2]*m[11] + m[4]*m[3]*m[10] + m[8]*m[2]*m[7] - m[8]*m[3]*m[6]
	inv[11] = -m[0]*m[5]*m[11] + m[0]*m[7]*m[9] + m[4]*m[1]*m[11] - m[4]*m[3]*m[9] - m[8]*m[1]*m[7] + m[8]*m[3]*m[5]
	inv[15] = m[0]*m[5]*m[10] - m[0]*m[6]*m[9] - m[4]*m[1]*m[10] + m[4]*m[2]*m[9] + m[8]*m[1]*m[6] - m[8]*m[2]*m[5]

	det := m[0]*inv[0] + m[1]*inv[4] + m[2]*inv[8] + m[3]*inv[12]
	if math.Abs(det) < 1e-30 {
		return [16]float64{}, false
	}
	invDet := 1.0 / det
	for i := range inv {
		inv[i] *= invDet
	}
	return inv, true
}
