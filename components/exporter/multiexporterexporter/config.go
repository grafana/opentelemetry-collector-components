package multiexporterexporter

import (
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
)

type exporterType string

const (
	otlp     exporterType = "otlp"
	otlphttp              = "otlphttp"
	loki                  = "loki"
)

type Config struct {
	Metrics SignalExportConfig `mapstructure:"metrics"`
	Logs    SignalExportConfig `mapstructure:"logs"`
	Traces  SignalExportConfig `mapstructure:"traces"`

	MappingKey string `mapstructure:"mapping_key"`
}

type SignalExportConfig struct {
	HTTP confighttp.HTTPClientSettings `mapstructure:"http_client"`
	GRPC configgrpc.GRPCClientSettings `mapstructure:"grpc_client"`

	exporterhelper.TimeoutSettings `mapstructure:"timeout"`
	exporterhelper.QueueSettings   `mapstructure:"sending_queue"`
	exporterhelper.RetrySettings   `mapstructure:"retry_on_failure"`

	Exporter exporterType      `mapstructure:"exporter"`
	Mapping  map[string]string `mapstructure:"mapping"`
}

var _ component.Config = (*Config)(nil)

// Validate checks if the exporter configuration is valid
func (cfg *Config) Validate() error {
	return nil
}
