package gcomapiprocessor

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	collectorclient "go.opentelemetry.io/collector/client"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/otel/metric/noop"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"

	"github.com/grafana/opentelemetry-collector-components/processor/gcomapiprocessor/internal/gcom"
)

func TestProcessor_EnrichContextWithSignalInstanceURL(t *testing.T) {
	t.Parallel()

	cfg := createDefaultConfig().(*Config)
	cfg.Client.Endpoint = "mock://fake.com"
	cfg.GrafanaClusterFilters = "1"

	tests := []struct {
		name    string
		signal  signal
		context func() context.Context
		want    string
		error   string
	}{
		{
			name:   "traces instance url is set",
			signal: traces,
			context: func() context.Context {
				return collectorclient.NewContext(context.Background(), collectorclient.Info{
					Metadata: collectorclient.NewMetadata(
						map[string][]string{headerOrgID: {strconv.Itoa(gcom.GrafanaInstanceOne.ID)}},
					),
				})
			},
			want: gcom.GrafanaInstanceOne.TracesInstanceURL,
		},
		{
			name:   "logs instance url is set",
			signal: logs,
			context: func() context.Context {
				return collectorclient.NewContext(context.Background(), collectorclient.Info{
					Metadata: collectorclient.NewMetadata(
						map[string][]string{headerOrgID: {strconv.Itoa(gcom.GrafanaInstanceOne.ID)}},
					),
				})
			},
			want: gcom.GrafanaInstanceOne.LogsInstanceURL,
		},
		{
			name:   "metrics instance url is set",
			signal: metrics,
			context: func() context.Context {
				return collectorclient.NewContext(context.Background(), collectorclient.Info{
					Metadata: collectorclient.NewMetadata(
						map[string][]string{headerOrgID: {strconv.Itoa(gcom.GrafanaInstanceOne.ID)}},
					),
				})
			},
			want: gcom.GrafanaInstanceOne.PromInstanceURL,
		},

		{
			name:   "canonical id header",
			signal: traces,
			context: func() context.Context {
				canonicalOrgID := http.CanonicalHeaderKey(headerOrgID)
				return collectorclient.NewContext(context.Background(), collectorclient.Info{
					Metadata: collectorclient.NewMetadata(
						map[string][]string{canonicalOrgID: {strconv.Itoa(gcom.GrafanaInstanceOne.ID)}},
					),
				})
			},
			want: gcom.GrafanaInstanceOne.TracesInstanceURL,
		},
		{
			name:   "missing org id header",
			signal: traces,
			context: func() context.Context {
				return collectorclient.NewContext(context.Background(), collectorclient.Info{
					Metadata: collectorclient.NewMetadata(
						map[string][]string{},
					),
				})
			},
			error: "missing \"X-Scope-OrgID\" header",
		},
		{
			name:   "empty org id header",
			signal: traces,
			context: func() context.Context {
				return collectorclient.NewContext(context.Background(), collectorclient.Info{
					Metadata: collectorclient.NewMetadata(
						map[string][]string{headerOrgID: {""}},
					),
				})
			},
			error: "invalid \"X-Scope-OrgID\" header: ",
		},
		{
			name:   "invalid org id header",
			signal: traces,
			context: func() context.Context {
				return collectorclient.NewContext(context.Background(), collectorclient.Info{
					Metadata: collectorclient.NewMetadata(
						map[string][]string{headerOrgID: {"1a"}},
					),
				})
			},
			error: "invalid \"X-Scope-OrgID\" header: 1a",
		},
		{
			name:   "org id header has more than one value ",
			signal: traces,
			context: func() context.Context {
				return collectorclient.NewContext(context.Background(), collectorclient.Info{
					Metadata: collectorclient.NewMetadata(
						map[string][]string{
							headerOrgID: {strconv.Itoa(gcom.GrafanaInstanceOne.ID), "test"}},
					),
				})
			},
			error: "2 source keys found in the context, can't determine which one to use",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := newAPIProcessor(
				cfg,
				component.TelemetrySettings{
					Logger:        zap.NewNop(),
					MeterProvider: noop.NewMeterProvider(),
				},
				tt.signal,
			)
			require.NoError(t, err)

			ctx, err := p.enrichContextWithSignalInstanceURL(tt.context())
			if tt.error != "" {
				assert.ErrorContains(t, err, tt.error)
				return
			}

			md, ok := metadata.FromIncomingContext(ctx)
			assert.True(t, ok)
			got := md.Get(instanceURL)
			if len(got) > 1 {
				assert.Fail(t, fmt.Sprintf("too many arguments are set: %d", len(got)))
			}
			assert.Equal(t, tt.want, got[0])
		})
	}
}

func TestRetrieveOrgIdFromContext(t *testing.T) {
	md := collectorclient.NewMetadata(map[string][]string{
		"X-Scope-Orgid": {"123"},
	})
	info := collectorclient.Info{
		Metadata: md,
	}
	ctx := collectorclient.NewContext(context.Background(), info)
	orgID, err := retrieveOrgIdFromContext(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "123", orgID)
}
