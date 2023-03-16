package gcpricingestimatorconnector

import (
	"context"
	"sync"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/connector"
	"go.opentelemetry.io/collector/consumer"
	"go.uber.org/zap"
)

const (
	// Grafana Cloud Pricing Estimator
	typeStr   = "gcpricingestimator"
	stability = component.StabilityLevelDevelopment
)

var processorCapabilities = consumer.Capabilities{MutatesData: false}

type factory struct {
	lock sync.Mutex
}

func NewFactory() connector.Factory {
	return connector.NewFactory(
		typeStr,
		createDefaultConfig,
		connector.WithMetricsToMetrics(createMetricsToMetricsConnector, stability),
		connector.WithTracesToMetrics(createTracesToMetricsConnector, stability),
		connector.WithLogsToMetrics(createLogsToMetricsConnector, stability),
	)
}

func createDefaultConfig() component.Config {
	return &Config{}
}

func createMetricsToMetricsConnector(
	_ context.Context,
	set connector.CreateSettings,
	cfg component.Config,
	next consumer.Metrics,
) (connector.Metrics, error) {
	return toMetricsConnector(set.Logger, cfg.(*Config), next)
}

func createTracesToMetricsConnector(
	_ context.Context,
	set connector.CreateSettings,
	cfg component.Config,
	next consumer.Metrics,
) (connector.Traces, error) {
	return toMetricsConnector(set.Logger, cfg.(*Config), next)
}

func createLogsToMetricsConnector(
	_ context.Context,
	set connector.CreateSettings,
	cfg component.Config,
	next consumer.Metrics,
) (connector.Logs, error) {
	return toMetricsConnector(set.Logger, cfg.(*Config), next)
}

func toMetricsConnector(logger *zap.Logger, cfg *Config, next consumer.Metrics) (*estimator, error) {
	e, err := newEstimator(logger, cfg)
	if err != nil {
		return nil, err
	}
	e.mc = next
	return e, nil
}
