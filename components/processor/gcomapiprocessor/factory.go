package gcomapiprocessor

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/processor"

	"github.com/grafana/opentelemetry-collector-components/processor/gcomapiprocessor/internal/gcom/client"
)

const (
	typeStr   = "gcomapi"
	stability = component.StabilityLevelAlpha
)

func NewFactory() processor.Factory {
	return processor.NewFactory(
		typeStr,
		createDefaultConfig,
		processor.WithLogs(createLogsProcessor, stability),
		processor.WithTraces(createTracesProcessor, stability),
		processor.WithMetrics(createMetricsProcessor, stability),
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		ServiceName: "otlp-gateway",
		Client: clientConfig{
			Endpoint: "https://www.grafana.com/api",
			Timeout:  1 * time.Minute,
		},
		Cache: cacheConfig{
			CompleteRefreshDuration:    5 * time.Hour,
			IncrementalRefreshDuration: 5 * time.Minute,
		},
	}
}

func createLogsProcessor(
	_ context.Context,
	set processor.CreateSettings,
	cfg component.Config,
	nextConsumer consumer.Logs,
) (processor.Logs, error) {
	proc, err := newAPIProcessor(cfg.(*Config), set.TelemetrySettings)
	if err != nil {
		return nil, err
	}
	proc.Logs, err = consumer.NewLogs(func(ctx context.Context, td plog.Logs) error {
		newCtx, err := proc.enrichContextWithSignalInstanceURL(
			ctx,
			func(i client.Instance) string { return i.LogsInstanceURL },
		)
		if err != nil {
			return err
		}
		return nextConsumer.ConsumeLogs(newCtx, td)
	})
	if err != nil {
		return nil, err
	}

	return proc, nil
}

func createTracesProcessor(
	_ context.Context,
	set processor.CreateSettings,
	cfg component.Config,
	nextConsumer consumer.Traces,
) (processor.Traces, error) {
	proc, err := newAPIProcessor(cfg.(*Config), set.TelemetrySettings)
	if err != nil {
		return nil, err
	}
	proc.Traces, err = consumer.NewTraces(func(ctx context.Context, td ptrace.Traces) error {
		newCtx, err := proc.enrichContextWithSignalInstanceURL(
			ctx,
			func(i client.Instance) string { return i.TracesInstanceURL },
		)
		if err != nil {
			return err
		}
		return nextConsumer.ConsumeTraces(newCtx, td)
	})
	if err != nil {
		return nil, err
	}

	return proc, nil
}

func createMetricsProcessor(
	_ context.Context,
	set processor.CreateSettings,
	cfg component.Config,
	nextConsumer consumer.Metrics,
) (processor.Metrics, error) {
	proc, err := newAPIProcessor(cfg.(*Config), set.TelemetrySettings)
	if err != nil {
		return nil, err
	}

	proc.Metrics, err = consumer.NewMetrics(func(ctx context.Context, md pmetric.Metrics) error {
		newCtx, err := proc.enrichContextWithSignalInstanceURL(
			ctx,
			func(i client.Instance) string { return i.PromInstanceURL },
		)
		if err != nil {
			return err
		}
		return nextConsumer.ConsumeMetrics(newCtx, md)
	})
	if err != nil {
		return nil, err
	}

	return proc, nil
}
