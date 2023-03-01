package faroreceiver

import (
	"context"

	"github.com/grafana/opentelemetry-collector-components/internal/sharedcomponent"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
)

const (
	typeStr   = "faro"
	stability = component.StabilityLevelAlpha

	defaultHTTPEndpoint = "0.0.0.0:8886"
)

func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		typeStr,
		createDefaultConfig,
		receiver.WithLogs(createLogsReceiver, stability),
		receiver.WithTraces(createTracesReceiver, stability),
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		HTTP: &confighttp.HTTPServerSettings{
			Endpoint: defaultHTTPEndpoint,
		},
	}
}

// createLogReceiver creates a log receiver based on provided config.
func createLogsReceiver(
	_ context.Context,
	set receiver.CreateSettings,
	cfg component.Config,
	consumer consumer.Logs,
) (receiver.Logs, error) {
	c := cfg.(*Config)
	recv, err := receivers.GetOrAdd(c, func() (*faroReceiver, error) {
		return newFaroReceiver(*c, set, withLogs(consumer))
	})
	if err != nil {
		return nil, err
	}

	r := recv.Unwrap()
	if r.logConsumer == nil {
		r.logConsumer = consumer
	}

	return r, nil
}

// createTracesReceiver creates a trace receiver based on provided config.
func createTracesReceiver(
	_ context.Context,
	set receiver.CreateSettings,
	cfg component.Config,
	consumer consumer.Traces,
) (receiver.Traces, error) {
	c := cfg.(*Config)
	recv, err := receivers.GetOrAdd(c, func() (*faroReceiver, error) {
		return newFaroReceiver(*c, set, withTraces(consumer))
	})
	if err != nil {
		return nil, err
	}
	r := recv.Unwrap()
	if r.traceConsumer == nil {
		r.traceConsumer = consumer
	}

	return r, nil
}

// This is the map of already created OpenCensus receivers for particular configurations.
// We maintain this map because the Factory is asked trace and metric receivers separately
// when it gets CreateTracesReceiver() and CreateMetricsReceiver() but they must not
// create separate objects, they must use one ocReceiver object per configuration.
var receivers = sharedcomponent.NewSharedComponents[*Config, *faroReceiver]()
