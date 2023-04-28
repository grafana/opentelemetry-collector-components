package gcomapiprocessor

import (
	"context"
	"fmt"
	"strconv"

	collectorclient "go.opentelemetry.io/collector/client"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/instrument"
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

	instrumentScopeName = "github.com/grafana/opentelemetry-collector-components/components/gcomapiprocessor"
	nameSep             = "/"
)

type grafanaAPIProcessor struct {
	logger *zap.Logger
	cache  cache.InstanceCache

	consumer.Metrics
	consumer.Logs
	consumer.Traces

	component.StartFunc
	component.ShutdownFunc

	reqsCounter instrument.Int64Counter

	signal signal
}

func newAPIProcessor(cfg *Config, settings component.TelemetrySettings, s signal) (*grafanaAPIProcessor, error) {
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

	meter := settings.MeterProvider.Meter(instrumentScopeName + nameSep + string(s))
	reqsCounter, err := meter.Int64Counter(
		fmt.Sprintf("otlp_gateway_gcom_api_%s_requests_total", s),
		instrument.WithDescription(fmt.Sprintf("The number of authenticated %s requests.", s)),
	)
	if err != nil {
		return nil, err
	}

	return &grafanaAPIProcessor{
		cache:       ic,
		logger:      settings.Logger,
		signal:      s,
		reqsCounter: reqsCounter,
	}, nil

}

// enrichContextWithSignalInstanceURL resolves signal instance URL from StackID that
// is set via `X-Scope-OrgID` header, and wraps the incoming context in a new
// context that has the signal instance URL in `X-Scope-InstanceURL` metadata field.
func (p *grafanaAPIProcessor) enrichContextWithSignalInstanceURL(ctx context.Context) (context.Context, error) {
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

	tenant, clusterURL := extractTenantIDAndURL(p.signal, instance)
	p.reqsCounter.Add(
		ctx,
		1,
		attribute.Int("org_id", instance.OrgID),
		attribute.String("tenant_id", strconv.Itoa(tenant)),
		attribute.String("cluster_url", clusterURL),
	)

	// Set X-Scope-InstanceURL
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = map[string][]string{instanceURL: {clusterURL}}
	} else {
		md.Set(instanceURL, clusterURL)
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

func extractTenantIDAndURL(s signal, i client.Instance) (int, string) {
	switch s {
	case traces:
		return i.TracesInstanceID, i.TracesInstanceURL
	case metrics:
		return i.PromInstanceID, i.PromInstanceURL
	case logs:
		return i.LogsInstanceID, i.LogsInstanceURL
	}
	// should not happen
	return 0, ""
}
