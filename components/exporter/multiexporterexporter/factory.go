package multiexporterexporter

import (
	"context"
	"fmt"
	"reflect"

	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/lokiexporter"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/exporter/otlphttpexporter"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

const (
	typeStr   = "multiexporter"
	stability = component.StabilityLevelBeta
)

func NewFactory() exporter.Factory {
	return exporter.NewFactory(
		typeStr,
		createDefaultConfig,
		exporter.WithTraces(newTracesExporter, stability),
		exporter.WithMetrics(newMetricsExporter, stability),
		exporter.WithLogs(newLogsExporter, stability),
	)
}

var exporterFactories = map[exporterType]exporter.Factory{
	loki:     lokiexporter.NewFactory(),
	otlphttp: otlphttpexporter.NewFactory(),
	otlp:     otlpexporter.NewFactory(),
}

func createDefaultConfig() component.Config {
	return &Config{}
}

func newTracesExporter(ctx context.Context, params exporter.CreateSettings, conf component.Config) (exporter.Traces, error) {
	c := conf.(*Config)
	cfg := c.Traces
	factory, ok := exporterFactories[cfg.Exporter]
	if !ok {
		return nil, fmt.Errorf("unknown exporter type %s", cfg.Exporter)
	}

	exporters := make(map[string]interface{})
	for mappingKey, url := range cfg.Mapping {
		c, err := createExporterConfig(factory, url, cfg)
		if err != nil {
			return nil, err
		}
		e, err := factory.CreateTracesExporter(ctx, params, c)
		if err != nil {
			return nil, err
		}
		exporters[mappingKey] = e
	}

	exp := newMultiExporter(exporters, c.MappingKey, params.Logger)
	return exporterhelper.NewTracesExporter(
		ctx,
		params,
		cfg,
		func(ctx context.Context, td ptrace.Traces) error {
			return exp.consume(ctx, td)
		},
		exporterhelper.WithStart(exp.start),
	)
}

func newMetricsExporter(ctx context.Context, params exporter.CreateSettings, conf component.Config) (exporter.Metrics, error) {
	c := conf.(*Config)
	cfg := c.Metrics
	factory, ok := exporterFactories[cfg.Exporter]
	if !ok {
		return nil, fmt.Errorf("unknown exporter type %s", cfg.Exporter)
	}

	exporters := make(map[string]interface{})
	for mappingKey, url := range cfg.Mapping {
		c, err := createExporterConfig(factory, url, cfg)
		if err != nil {
			return nil, err
		}
		e, err := factory.CreateMetricsExporter(ctx, params, c)
		if err != nil {
			return nil, err
		}
		exporters[mappingKey] = e
	}

	exp := newMultiExporter(exporters, c.MappingKey, params.Logger)
	return exporterhelper.NewMetricsExporter(
		ctx,
		params,
		cfg,
		func(ctx context.Context, md pmetric.Metrics) error {
			return exp.consume(ctx, md)
		},
		exporterhelper.WithStart(exp.start),
	)
}

func newLogsExporter(ctx context.Context, params exporter.CreateSettings, conf component.Config) (exporter.Logs, error) {
	c := conf.(*Config)
	cfg := c.Logs
	factory, ok := exporterFactories[cfg.Exporter]
	if !ok {
		return nil, fmt.Errorf("unknown exporter type %s", cfg.Exporter)
	}

	exporters := make(map[string]interface{})
	for mappingKey, url := range cfg.Mapping {
		c, err := createExporterConfig(factory, url, cfg)
		if err != nil {
			return nil, err
		}
		e, err := factory.CreateLogsExporter(ctx, params, c)
		if err != nil {
			return nil, err
		}
		exporters[mappingKey] = e
	}

	exp := newMultiExporter(exporters, c.MappingKey, params.Logger)
	return exporterhelper.NewLogsExporter(
		ctx,
		params,
		cfg,
		func(ctx context.Context, ld plog.Logs) error {
			return exp.consume(ctx, ld)
		},
		exporterhelper.WithStart(exp.start),
	)
}

func createExporterConfig(factory exporter.Factory, url string, conf SignalExportConfig) (component.Config, error) {
	switch exporterType(factory.Type()) {
	case loki:
		cfg := factory.CreateDefaultConfig().(*lokiexporter.Config)
		cfg.RetrySettings = conf.RetrySettings
		cfg.Endpoint = url
		cfg.QueueSettings = conf.QueueSettings

		return cfg, nil
	case otlphttp:
		cfg := factory.CreateDefaultConfig().(*otlphttpexporter.Config)
		cfg.HTTPClientSettings = conf.HTTP
		cfg.Endpoint = url
		cfg.RetrySettings = conf.RetrySettings
		cfg.QueueSettings = conf.QueueSettings

		return cfg, nil
	case otlp:
		cfg := factory.CreateDefaultConfig().(*otlpexporter.Config)
		cfg.GRPCClientSettings = conf.GRPC
		cfg.Endpoint = url
		cfg.RetrySettings = conf.RetrySettings
		cfg.QueueSettings = conf.QueueSettings

		return cfg, nil
	}

	return nil, fmt.Errorf("unknown exporter factory %q", reflect.TypeOf(factory))
}
