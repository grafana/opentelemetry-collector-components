package multiexporterexporter

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configauth"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/config/configtls"
	"go.opentelemetry.io/collector/confmap/confmaptest"
)

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	cm, err := confmaptest.LoadConf(filepath.Join("testdata", "config.yaml"))
	require.NoError(t, err)

	tests := []struct {
		id       component.ID
		expected component.Config
	}{
		{
			id: component.NewIDWithName(typeStr, ""),
			expected: &Config{
				MappingKey: "X-Scope-InstanceURL",
				Logs: SignalExportConfig{
					Exporter: loki,
					Mapping: map[string]string{
						"url1": "http",
					},
				},
				Traces: SignalExportConfig{
					Exporter: otlp,
					Mapping: map[string]string{
						"url2": "http",
						"url3": "http",
					},
					GRPC: configgrpc.GRPCClientSettings{
						TLSSetting: configtls.TLSClientSetting{
							Insecure: true,
						},
						Auth: &configauth.Authentication{
							AuthenticatorID: component.NewID("headers_setter"),
						},
					},
				},
				Metrics: SignalExportConfig{
					Exporter: otlphttp,
					Mapping: map[string]string{
						"url": "http",
					},
					HTTP: confighttp.HTTPClientSettings{
						TLSSetting: configtls.TLSClientSetting{
							Insecure: true,
						},
						Auth: &configauth.Authentication{
							AuthenticatorID: component.NewID("headers_setter"),
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.id.String(), func(t *testing.T) {
			factory := NewFactory()
			cfg := factory.CreateDefaultConfig()

			sub, err := cm.Sub(tt.id.String())
			require.NoError(t, err)
			require.NoError(t, component.UnmarshalConfig(sub, cfg))

			assert.NoError(t, component.ValidateConfig(cfg))
			assert.Equal(t, tt.expected, cfg)
		})
	}
}
