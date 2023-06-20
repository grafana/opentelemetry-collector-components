package multiexporterexporter

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	collectorclient "go.opentelemetry.io/collector/client"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

func TestExporter(t *testing.T) {
	t.Parallel()

	mappingKey := "mkey"
	mappingKeyValue := "value"

	logsExporter, _ := exporterhelper.NewLogsExporter(
		context.Background(),
		exporter.CreateSettings{TelemetrySettings: componenttest.NewNopTelemetrySettings()},
		Config{},
		func(ctx context.Context, ld plog.Logs) error { return nil },
	)

	metricsExporter, _ := exporterhelper.NewMetricsExporter(
		context.Background(),
		exporter.CreateSettings{TelemetrySettings: componenttest.NewNopTelemetrySettings()},
		Config{},
		func(ctx context.Context, md pmetric.Metrics) error { return nil },
	)

	tracesExporter, _ := exporterhelper.NewTracesExporter(
		context.Background(),
		exporter.CreateSettings{TelemetrySettings: componenttest.NewNopTelemetrySettings()},
		Config{},
		func(ctx context.Context, td ptrace.Traces) error { return nil },
	)

	tests := []struct {
		name     string
		context  func() context.Context
		data     interface{}
		exporter interface{}
		error    string
	}{
		{
			name: "export logs",
			context: func() context.Context {
				return collectorclient.NewContext(context.Background(), collectorclient.Info{
					Metadata: collectorclient.NewMetadata(
						map[string][]string{mappingKey: {mappingKeyValue}},
					),
				})
			},
			exporter: logsExporter,
			data:     plog.NewLogs(),
		},
		{
			name: "invalid payload for the logs exporter",
			context: func() context.Context {
				return collectorclient.NewContext(context.Background(), collectorclient.Info{
					Metadata: collectorclient.NewMetadata(
						map[string][]string{mappingKey: {mappingKeyValue}},
					),
				})
			},
			exporter: logsExporter,
			data:     ptrace.NewTraces(),
			error:    "invalid data type \"ptrace.Traces\" for the logs exporter",
		},
		{
			name: "export metrics",
			context: func() context.Context {
				return collectorclient.NewContext(context.Background(), collectorclient.Info{
					Metadata: collectorclient.NewMetadata(
						map[string][]string{mappingKey: {mappingKeyValue}},
					),
				})
			},
			exporter: metricsExporter,
			data:     pmetric.NewMetrics(),
		},
		{
			name: "invalid payload for the metrics exporter",
			context: func() context.Context {
				return collectorclient.NewContext(context.Background(), collectorclient.Info{
					Metadata: collectorclient.NewMetadata(
						map[string][]string{mappingKey: {mappingKeyValue}},
					),
				})
			},
			exporter: metricsExporter,
			data:     ptrace.NewTraces(),
			error:    "invalid data type \"ptrace.Traces\" for the metrics exporter",
		},
		{
			name: "export traces",
			context: func() context.Context {
				return collectorclient.NewContext(context.Background(), collectorclient.Info{
					Metadata: collectorclient.NewMetadata(
						map[string][]string{mappingKey: {mappingKeyValue}},
					),
				})
			},
			exporter: tracesExporter,
			data:     ptrace.NewTraces(),
		},
		{
			name: "invalid payload for the traces exporter",
			context: func() context.Context {
				return collectorclient.NewContext(context.Background(), collectorclient.Info{
					Metadata: collectorclient.NewMetadata(
						map[string][]string{mappingKey: {mappingKeyValue}},
					),
				})
			},
			exporter: tracesExporter,
			data:     plog.NewLogs(),
			error:    "invalid data type \"plog.Logs\" for the traces traces",
		},
		{
			name: "missing exporter",
			context: func() context.Context {
				return collectorclient.NewContext(context.Background(), collectorclient.Info{
					Metadata: collectorclient.NewMetadata(
						map[string][]string{mappingKey: {"value1"}},
					),
				})
			},
			exporter: logsExporter,
			data:     plog.NewLogs(),
			error:    "exporter is not found for the mapping key: \"value1\"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newMultiExporter(
				map[string]interface{}{
					mappingKeyValue: tt.exporter,
				},
				mappingKey,
				zap.NewNop(),
			)

			err := e.consume(tt.context(), tt.data)
			if tt.error != "" {
				assert.ErrorContains(t, err, tt.error)
				return
			}
		})
	}
}

func TestRetrieveFromContext(t *testing.T) {
	md := collectorclient.NewMetadata(map[string][]string{
		"X-Scope-Orgid": {"123"},
	})
	info := collectorclient.Info{
		Metadata: md,
	}
	ctx := collectorclient.NewContext(context.Background(), info)
	orgID, err := retrieveFromContext(ctx, "X-Scope-Orgid")
	assert.NoError(t, err)
	assert.Equal(t, "123", orgID)
}
