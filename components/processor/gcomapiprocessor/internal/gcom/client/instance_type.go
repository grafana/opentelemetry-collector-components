package client

import (
	"fmt"
)

// InstanceType enumerates the valid instance types for grafana.com
type InstanceType string

// Valid GrafanaCloud instance types
const (
	Prometheus     InstanceType = "prometheus"
	Graphite       InstanceType = "graphite"
	GraphiteShared InstanceType = "graphite-shared"
	Metrics        InstanceType = "metrics" // Instance type that covers any Hosted-Metrics instance, without checking the type.
	Logs           InstanceType = "logs"
	Alerts         InstanceType = "alerts"
	Traces         InstanceType = "traces"
	Grafana        InstanceType = "grafana"
	OnCall         InstanceType = "oncall"
	Profiles       InstanceType = "profiles"
)

func (i InstanceType) String() string {
	return string(i)
}

func (i InstanceType) IsMetrics() bool {
	if i == Prometheus || i == Graphite || i == GraphiteShared {
		return true
	}

	return false
}

func (i InstanceType) IsPlugin() bool {
	return i == OnCall
}

func InstanceTypeFromString(value string) (InstanceType, error) {
	switch value {
	case Prometheus.String():
		return Prometheus, nil
	case Graphite.String():
		return Graphite, nil
	case GraphiteShared.String():
		return GraphiteShared, nil
	case Logs.String():
		return Logs, nil
	case Alerts.String():
		return Alerts, nil
	case Traces.String():
		return Traces, nil
	case Grafana.String():
		return Grafana, nil
	case Metrics.String():
		return Metrics, nil
	case OnCall.String():
		return OnCall, nil
	case Profiles.String():
		return Profiles, nil
	}
	return "", fmt.Errorf("instance type '%v' is not valid", value)
}
