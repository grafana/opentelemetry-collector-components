package gcomapiprocessor

import (
	"context"
	"fmt"
	"strconv"

	collectorclient "go.opentelemetry.io/collector/client"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"

	"github.com/grafana/opentelemetry-collector-components/processor/gcomapiprocessor/internal"
	"github.com/grafana/opentelemetry-collector-components/processor/gcomapiprocessor/internal/gcom"
	"github.com/grafana/opentelemetry-collector-components/processor/gcomapiprocessor/internal/gcom/cache"
	"github.com/grafana/opentelemetry-collector-components/processor/gcomapiprocessor/internal/gcom/client"
)

const (
	headerOrgID = "X-Scope-OrgID"
	instanceURL = "X-Scope-InstanceURL"
)

type grafanaAPIProcessor struct {
	logger *zap.Logger
	cache  cache.InstanceCache

	consumer.Metrics
	consumer.Logs
	consumer.Traces

	component.StartFunc
	component.ShutdownFunc
}

func newAPIProcessor(cfg *Config, settings component.TelemetrySettings) (*grafanaAPIProcessor, error) {
	logger := internal.NewZapToGokitLogAdapter(settings.Logger)

	cl, err := client.New(
		client.Config{
			Endpoint: cfg.Client.Endpoint,
			Key:      cfg.Client.Key,
			Timeout:  cfg.Client.Timeout,
		},
		cfg.ServiceName,
		logger,
	)
	if err != nil {
		return nil, err
	}
	if cfg.isDryRun() {
		cl = gcom.NewMockGcomClient()
	}

	ic, err := cache.NewInstanceCache(
		cache.InstanceCacheConfig{
			CompleteCacheRefreshDuration:    cfg.Cache.CompleteRefreshDuration,
			IncrementalCacheRefreshDuration: cfg.Cache.IncrementalRefreshDuration,
			InstanceTypes:                   []client.InstanceType{client.Grafana},
		},
		logger,
		[]client.InstanceType{client.Grafana},
		cl,
	)
	if err != nil {
		return nil, err
	}

	return &grafanaAPIProcessor{
		cache:  ic,
		logger: settings.Logger,
	}, nil

}

// enrichContextWithSignalInstanceURL resolves signal instance URL from StackID that
// is set via `X-Scope-OrgID` header, and wraps the incoming context in a new
// context that has the signal instance URL in `X-Scope-InstanceURL` metadata field.
func (p *grafanaAPIProcessor) enrichContextWithSignalInstanceURL(
	ctx context.Context,
	extractURL func(i client.Instance) string,
) (context.Context, error) {

	orgID, err := retrieveOrgIdFromContext(ctx)
	if err != nil {
		return nil, err
	}

	stackID, err := strconv.Atoi(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid %q header: %s", headerOrgID, orgID)
	}

	// Get Grafana instance by ID. X-Scope-OrgId here contains StackID, not the
	// metrics, logs, or traces instance ID.
	instance, err := p.cache.GetInstanceInfo(client.Grafana, stackID)
	if err != nil {
		return nil, fmt.Errorf("failure looking up by stack id: '%d', %s", stackID, err.Error())
	}

	// Set X-Scope-InstanceURL
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = map[string][]string{instanceURL: {extractURL(instance)}}
	} else {
		md.Set(instanceURL, extractURL(instance))
	}
	return metadata.NewIncomingContext(ctx, md), nil
}

func retrieveOrgIdFromContext(ctx context.Context) (string, error) {
	// Extract X-Scope-OrgID
	info := collectorclient.FromContext(ctx)
	v := info.Metadata.Get(headerOrgID)

	if len(v) == 0 {
		return "", fmt.Errorf("missing %q header, is include_metadata enabled?", headerOrgID)
	}

	if len(v) > 1 {
		return "", fmt.Errorf("%d source keys found in the context, can't determine which one to use", len(v))
	}

	return v[0], nil
}

func (p *grafanaAPIProcessor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}
