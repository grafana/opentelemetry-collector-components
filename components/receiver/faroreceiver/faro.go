package faroreceiver // import "github.com/grafana/opentelemetry-collector-components/components/receiver/faroreceiver"

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
)

type faroReceiver struct {
	cfg           *Config
	server        *http.Server
	traceConsumer consumer.Traces
	logConsumer   consumer.Logs

	shutdownWG sync.WaitGroup

	//TODO
	//obsrepHTTP *obsreport.Receiver

	settings receiver.CreateSettings
}

type option func(r *faroReceiver)

func withLogs(c consumer.Logs) option {
	return func(r *faroReceiver) {
		r.logConsumer = c
	}
}

func withTraces(c consumer.Traces) option {
	return func(r *faroReceiver) {
		r.traceConsumer = c
	}
}

// newFaroReceiver TODO...
func newFaroReceiver(cfg Config, set receiver.CreateSettings, opts ...option) (*faroReceiver, error) {
	r := &faroReceiver{
		cfg:      &cfg,
		settings: set,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r, nil
}

// Start will start the receiver http server
func (r *faroReceiver) Start(ctx context.Context, host component.Host) error {
	var err error
	if r.cfg.HTTP != nil {
		if r.server, err = r.cfg.HTTP.ToServer(
			host,
			r.settings.TelemetrySettings,
			Handler{
				logConsumer:   r.logConsumer,
				traceConsumer: r.traceConsumer,
			},
		); err != nil {
			return err
		}

		if err = r.startHTTPServer(r.cfg.HTTP, host); err != nil {
			return err
		}
	}
	return nil
}

// Shutdown stops the receiver
func (r *faroReceiver) Shutdown(ctx context.Context) error {
	if r.server == nil {
		return nil
	}
	err := r.server.Shutdown(ctx)
	r.shutdownWG.Wait()
	return err
}

func (r *faroReceiver) startHTTPServer(cfg *confighttp.HTTPServerSettings, host component.Host) error {
	r.settings.Logger.Info("Starting HTTP server", zap.String("endpoint", cfg.Endpoint))
	var hln net.Listener
	hln, err := cfg.ToListener()
	if err != nil {
		return err
	}
	r.shutdownWG.Add(1)
	go func() {
		defer r.shutdownWG.Done()

		if errHTTP := r.server.Serve(hln); errHTTP != nil && !errors.Is(errHTTP, http.ErrServerClosed) {
			host.ReportFatalError(errHTTP)
		}
	}()
	return nil
}
