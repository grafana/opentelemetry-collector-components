package multiexporterexporter

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"go.opentelemetry.io/collector/client"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

type multiExporter struct {
	exporters  map[string]interface{}
	mappingKey string

	log *zap.Logger
}

func newMultiExporter(
	exporters map[string]interface{},
	mappingKey string,
	log *zap.Logger,
) *multiExporter {
	return &multiExporter{
		exporters:  exporters,
		mappingKey: mappingKey,
		log:        log,
	}
}

func (m multiExporter) consume(ctx context.Context, data interface{}) error {
	mappingKey, err := retrieveFromContext(ctx, m.mappingKey)
	if err != nil {
		return err
	}

	e, ok := m.exporters[mappingKey]
	if !ok {
		return errors.New(fmt.Sprintf("exporter is not found for the mapping key: %q", mappingKey))
	}

	switch v := e.(type) {
	case exporter.Logs:
		ld, ok := data.(plog.Logs)
		if !ok {
			return fmt.Errorf("invalid data type %q for the logs exporter", reflect.TypeOf(data))
		}
		return v.ConsumeLogs(ctx, ld)
	case exporter.Traces:
		td, ok := data.(ptrace.Traces)
		if !ok {
			return fmt.Errorf("invalid data type %q for the traces traces", reflect.TypeOf(data))
		}
		return v.ConsumeTraces(ctx, td)
	case exporter.Metrics:
		md, ok := data.(pmetric.Metrics)
		if !ok {
			return fmt.Errorf("invalid data type %q for the metrics exporter", reflect.TypeOf(data))
		}
		return v.ConsumeMetrics(ctx, md)
	default:
		return errors.New("")
	}
}

func (m *multiExporter) start(ctx context.Context, host component.Host) (err error) {
	for _, e := range m.exporters {
		exp, ok := e.(component.Component)
		if !ok {
			return fmt.Errorf("registered exporter is not a component")
		}
		err = exp.Start(ctx, host)
		if err != nil {
			return err
		}
	}
	return nil
}

// retrieveFromContext extract headers from context
func retrieveFromContext(ctx context.Context, header string) (string, error) {
	info := client.FromContext(ctx)
	v := info.Metadata.Get(header)

	if len(v) == 0 {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return "", fmt.Errorf("missing %q header, is include_metadata enabled?", header)
		}

		v = md.Get(header)
	}

	if len(v) > 1 {
		return "", fmt.Errorf("%d source keys found in the context, can't determine which one to use", len(v))
	}

	return v[0], nil
}
