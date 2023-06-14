package gcom

import (
	"github.com/grafana/opentelemetry-collector-components/processor/gcomapiprocessor/internal/gcom/client"
	gcom_mock "github.com/grafana/opentelemetry-collector-components/processor/gcomapiprocessor/internal/gcom/client/mock"
)

var (
	ClusterOne = 1
	StackOne   = 1

	OrgOne = 10

	MetricsInstanceOne = client.Instance{
		ID:                11,
		OrgID:             OrgOne,
		Type:              "prometheus",
		ClusterID:         ClusterOne,
		ClusterSlug:       "dev-01-dev-us-central-0",
		URL:               "https://prometheus-dev-01-dev-us-central-0.grafana.net",
		GrafanaInstanceID: 1,
	}

	LogsInstanceOne = client.Instance{
		ID:                1111,
		OrgID:             OrgOne,
		Type:              "logs",
		ClusterID:         ClusterOne,
		ClusterSlug:       "dev-01-dev-us-central-0",
		URL:               "https://loki-dev-01-dev-us-central-0.grafana.net",
		GrafanaInstanceID: 1,
	}

	TracesInstanceOne = client.Instance{
		ID:                111111,
		OrgID:             OrgOne,
		Type:              "traces",
		ClusterID:         ClusterOne,
		ClusterSlug:       "dev-01-dev-us-central-0",
		URL:               "https://tempo-dev-01-dev-us-central-0.grafana.net",
		GrafanaInstanceID: 1,
	}

	// Grafana instance is 1. Prometheus: 11, Logs: 1111, Traces: 111111.
	GrafanaInstanceOne = client.Instance{
		ID:                 StackOne,
		OrgID:              OrgOne,
		ClusterName:        "1",
		LogsInstanceID:     1111,
		LogsInstanceURL:    "https://loki-dev-01-dev-us-central-0.grafana.net",
		PromInstanceID:     11,
		PromInstanceURL:    "https://prometheus-dev-01-dev-us-central-0.grafana.net",
		GraphiteInstanceID: 11111,
		TracesInstanceID:   111111,
		TracesInstanceURL:  "https://tempo-dev-01-dev-us-central-0.grafana.net",
		GrafanaInstanceID:  StackOne,
	}
)

func NewMockGcomClient() client.Client {
	return &gcom_mock.Client{
		Instances: map[client.InstanceType][]client.Instance{
			client.Metrics: {
				MetricsInstanceOne,
			},
			client.Logs: {
				LogsInstanceOne,
			},
			client.Traces: {
				TracesInstanceOne,
			},
			client.Grafana: {
				GrafanaInstanceOne,
			},
		},
	}
}
