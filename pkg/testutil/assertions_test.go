// CLASSIFICATION: UNCLASSIFIED
package testutil_test

import (
	"errors"
	"testing"

	"github.com/arvinddhasmana/rtsa_webgpu/pkg/testutil"
	"google.golang.org/grpc"
)

func TestAssertions(t *testing.T) {
	testutil.AssertEqual(t, 1, 1)
	testutil.AssertTrue(t, true, "true")
	testutil.AssertNoError(t, nil)
	testutil.AssertError(t, errors.New("fail"))
	testutil.AssertErrorContains(t, errors.New("hello world"), "hello")
}

func TestGRPCUtil(t *testing.T) {
	addr, cleanup := testutil.StartTestGRPCServer(t, func(s *grpc.Server) {
		// Mock registration
	})
	defer cleanup()

	conn := testutil.DialTestGRPC(t, addr)
	if conn == nil {
		t.Error("expected gRPC connection")
	}
}

func TestProtoUtil(t *testing.T) {
	obs := testutil.NewTestRadarObservation("test")
	if obs.SensorId != "test" {
		t.Error("expected sensor ID test")
	}
	track := testutil.NewTestFusedTrack("trk")
	if track.TrackId != "trk" {
		t.Error("expected track ID trk")
	}
	alert := testutil.NewTestAnomalyAlert("alt", "trk")
	if alert.AlertId != "alt" {
		t.Error("expected alert ID alt")
	}
	pos := testutil.NewTestPosition(10, 20)
	if pos.Latitude != 10 {
		t.Error("expected lat 10")
	}
}
