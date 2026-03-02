// CLASSIFICATION: UNCLASSIFIED
package config_test

import (
	"os"
	"testing"

	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-fusion-engine/internal/config"
)

func TestConfig_Load(t *testing.T) {
	// Clear relevant env vars before testing
	os.Unsetenv("RTSA_FUSION_GATE_SURFACE_DISTANCE")
	os.Unsetenv("RTSA_FUSION_WEIGHT_POSITION")
	os.Unsetenv("RTSA_FUSION_CONSUMER_GROUP")

	tests := []struct {
		name    string
		envVars map[string]string
		verify  func(t *testing.T, cfg *config.Config)
	}{
		{
			name:    "defaults",
			envVars: map[string]string{},
			verify: func(t *testing.T, cfg *config.Config) {
				if cfg.ConsumerGroup != "fusion-engine" {
					t.Errorf("expected default consumer group, got %s", cfg.ConsumerGroup)
				}
				if cfg.GateSurfaceDistanceNM != 5.0 {
					t.Errorf("expected default gate surface distance 5.0, got %f", cfg.GateSurfaceDistanceNM)
				}
				if cfg.WeightPosition != 0.35 {
					t.Errorf("expected default weight position 0.35, got %f", cfg.WeightPosition)
				}
				if cfg.StaleTimeoutSec != 60 {
					t.Errorf("expected default stale timeout 60, got %d", cfg.StaleTimeoutSec)
				}
			},
		},
		{
			name: "custom env vars",
			envVars: map[string]string{
				"RTSA_FUSION_CONSUMER_GROUP":          "test-fusion-group",
				"RTSA_FUSION_GATE_SURFACE_DISTANCE": "10.5",
				"RTSA_FUSION_WEIGHT_POSITION":       "0.5",
			},
			verify: func(t *testing.T, cfg *config.Config) {
				if cfg.ConsumerGroup != "test-fusion-group" {
					t.Errorf("expected test-fusion-group, got %s", cfg.ConsumerGroup)
				}
				if cfg.GateSurfaceDistanceNM != 10.5 {
					t.Errorf("expected gate surface distance 10.5, got %f", cfg.GateSurfaceDistanceNM)
				}
				if cfg.WeightPosition != 0.5 {
					t.Errorf("expected weight position 0.5, got %f", cfg.WeightPosition)
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

			cfg := config.Load()
			tt.verify(t, cfg)

			// Clean up env vars
			for k := range tt.envVars {
				os.Unsetenv(k)
			}
		})
	}
}
