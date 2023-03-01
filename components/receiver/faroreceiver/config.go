package faroreceiver

import (
	"go.opentelemetry.io/collector/config/confighttp"
)

type Config struct {
	HTTP *confighttp.HTTPServerSettings `mapstructure:"http"`
}
