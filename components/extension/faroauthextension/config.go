package faroauthextension // import "github.com/grafana/opentelemetry-collector-components/component/extension/faroauthextension"

import (
	"errors"
)

var (
	errNoFaroAPIEndpoint = errors.New("no faro api endpoint provided")
)

type Config struct {
	FaroAPI *FaroAPI `mapstructure:"faro_api"`
}

type FaroAPI struct {
	Endpoint string `mapstructure:"endpoint"`
}

func (cfg *Config) Validate() error {
	if cfg.FaroAPI == nil || cfg.FaroAPI.Endpoint == "" {
		return errNoFaroAPIEndpoint
	}
	return nil
}
