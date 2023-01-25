package gcomapiprocessor

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap/confmaptest"
)

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	duration := time.Duration(1) * time.Second
	tests := []struct {
		id           component.ID
		expected     component.Config
		errorMessage string
	}{
		{
			id: component.NewIDWithName(typeStr, ""),
			expected: &Config{
				ServiceName: "otlp-gateway",
				Client: clientConfig{
					Key:      "test_key",
					Endpoint: "http://localhost:3000",
					Timeout:  duration,
				},
				Cache: cacheConfig{
					CompleteRefreshDuration:    duration,
					IncrementalRefreshDuration: duration,
				},
			},
		},
		{
			id:           component.NewIDWithName(typeStr, "missing_endpoint"),
			errorMessage: "grafana API endpoint is missing",
		},
		{
			id:           component.NewIDWithName(typeStr, "missing_key"),
			errorMessage: "grafana API key is missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.id.String(), func(t *testing.T) {
			cm, err := confmaptest.LoadConf(filepath.Join("testdata", "config.yaml"))
			assert.NoError(t, err)

			factory := NewFactory()
			cfg := factory.CreateDefaultConfig()

			sub, err := cm.Sub(tt.id.String())
			assert.NoError(t, err)
			assert.NoError(t, component.UnmarshalConfig(sub, cfg))

			if tt.expected == nil {
				assert.EqualError(t, component.ValidateConfig(cfg), tt.errorMessage)
				return
			}
			assert.NoError(t, component.ValidateConfig(cfg))
			assert.Equal(t, tt.expected, cfg)
		})
	}
}
