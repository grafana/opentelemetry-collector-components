package gcomapiprocessor

import (
	"errors"
	"strings"
	"time"

	"go.opentelemetry.io/collector/component"
)

type Config struct {
	ServiceName string       `mapstructure:"service_name"`
	Client      clientConfig `mapstructure:"client"`
	Cache       cacheConfig  `mapstructure:"cache"`
}

type clientConfig struct {
	Key      string        `mapstructure:"key"`
	Endpoint string        `mapstructure:"endpoint"`
	Timeout  time.Duration `mapstructure:"timeout"`
}

type cacheConfig struct {
	CompleteRefreshDuration    time.Duration `mapstructure:"complete_refresh_duration"`
	IncrementalRefreshDuration time.Duration `mapstructure:"incremental_refresh_duration"`
	GrafanaClusterFilters      string        `mapstructure:"grafana_cluster_filters"`
}

var _ component.Config = (*Config)(nil)

func (c *Config) Validate() error {
	if c.Client.Endpoint == "" {
		return errors.New("grafana API endpoint is missing")
	}
	if c.Client.Key == "" {
		return errors.New("grafana API key is missing")
	}
	if c.Cache.GrafanaClusterFilters == "" {
		return errors.New("grafana cluster filters is missing")
	}
	return nil
}

func (c *Config) isDryRun() bool {
	return strings.HasPrefix(c.Client.Endpoint, "mock://")
}
