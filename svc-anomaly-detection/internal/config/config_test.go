// CLASSIFICATION: UNCLASSIFIED
package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-anomaly-detection/internal/config"
)

func TestConfig_Load(t *testing.T) {
	// Clear relevant env vars before testing
	os.Unsetenv("RTSA_SERVICE_NAME")
	os.Unsetenv("RTSA_DETECTOR_SPEED_ENABLED")
	os.Unsetenv("RTSA_EXCLUSION_ZONES_JSON")

	tests := []struct {
		name    string
		envVars map[string]string
		verify  func(t *testing.T, cfg *config.Config, err error)
	}{
		{
			name:    "defaults",
			envVars: map[string]string{},
			verify: func(t *testing.T, cfg *config.Config, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if cfg.ServiceName != "svc-anomaly-detection" {
					t.Errorf("expected default service name, got %s", cfg.ServiceName)
				}
				if !cfg.Detectors.SpeedEnabled {
					t.Errorf("expected speed detector to be enabled by default")
				}
				if cfg.TrackHistoryMaxAge != 2*time.Hour {
					t.Errorf("expected track history max age 2h, got %v", cfg.TrackHistoryMaxAge)
				}
				if len(cfg.ExclusionZones) != 0 {
					t.Errorf("expected no exclusion zones by default, got %d", len(cfg.ExclusionZones))
				}
			},
		},
		{
			name: "custom env vars",
			envVars: map[string]string{
				"RTSA_SERVICE_NAME":           "test-anomaly-service",
				"RTSA_DETECTOR_SPEED_ENABLED": "false",
				"RTSA_TRACK_HISTORY_MAX_ENTRIES": "500",
			},
			verify: func(t *testing.T, cfg *config.Config, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if cfg.ServiceName != "test-anomaly-service" {
					t.Errorf("expected test-anomaly-service, got %s", cfg.ServiceName)
				}
				if cfg.Detectors.SpeedEnabled {
					t.Errorf("expected speed detector to be disabled")
				}
				if cfg.TrackHistoryMaxEntries != 500 {
					t.Errorf("expected 500 max entries, got %d", cfg.TrackHistoryMaxEntries)
				}
			},
		},
		{
			name: "valid exclusion zones json",
			envVars: map[string]string{
				"RTSA_EXCLUSION_ZONES_JSON": `[{"name":"Zone 1","center_lat":45.0,"center_lon":-30.0,"radius_nm":10.5}]`,
			},
			verify: func(t *testing.T, cfg *config.Config, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(cfg.ExclusionZones) != 1 {
					t.Fatalf("expected 1 exclusion zone, got %d", len(cfg.ExclusionZones))
				}
				zone := cfg.ExclusionZones[0]
				if zone.Name != "Zone 1" || zone.CenterLat != 45.0 || zone.RadiusNM != 10.5 {
					t.Errorf("parsed zone incorrect: %+v", zone)
				}
			},
		},
		{
			name: "invalid exclusion zones json",
			envVars: map[string]string{
				"RTSA_EXCLUSION_ZONES_JSON": `[{"name":"Zone 1"`, // broken JSON
			},
			verify: func(t *testing.T, cfg *config.Config, err error) {
				if err == nil {
					t.Fatal("expected error for invalid JSON, got nil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set env vars
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			cfg, err := config.Load()

			tt.verify(t, cfg, err)

			// Clean up env vars
			for k := range tt.envVars {
				os.Unsetenv(k)
			}
		})
	}
}
