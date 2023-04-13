package gcomapiprocessor

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	collectorclient "go.opentelemetry.io/collector/client"
	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"

	"github.com/grafana/opentelemetry-collector-components/processor/gcomapiprocessor/internal/gcom"
	"github.com/grafana/opentelemetry-collector-components/processor/gcomapiprocessor/internal/gcom/client"
)

func TestProcessor_EnrichContextWithSignalInstanceURL(t *testing.T) {
	t.Parallel()

	cfg := createDefaultConfig().(*Config)
	cfg.Client.Endpoint = "mock://fake.com"
	p, err := newAPIProcessor(cfg, component.TelemetrySettings{Logger: zap.NewNop()})
	assert.NoError(t, err)

	tests := []struct {
		name    string
		context func() context.Context
		want    string
		error   string
	}{
		{
			name: "traces instance url is set",
			context: func() context.Context {
				return collectorclient.NewContext(context.Background(), collectorclient.Info{
					Metadata: collectorclient.NewMetadata(
						map[string][]string{orgID: {strconv.Itoa(gcom.GrafanaInstanceOne.ID)}},
					),
				})
			},
			want: gcom.GrafanaInstanceOne.TracesInstanceURL,
		},
		{
			name: "missing org id header",
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
			name: "empty org id header",
			context: func() context.Context {
				return collectorclient.NewContext(context.Background(), collectorclient.Info{
					Metadata: collectorclient.NewMetadata(
						map[string][]string{orgID: {""}},
					),
				})
			},
			error: "invalid \"X-Scope-OrgID\" header: ",
		},
		{
			name: "invalid org id header",
			context: func() context.Context {
				return collectorclient.NewContext(context.Background(), collectorclient.Info{
					Metadata: collectorclient.NewMetadata(
						map[string][]string{orgID: {"1a"}},
					),
				})
			},
			error: "invalid \"X-Scope-OrgID\" header: 1a",
		},
		{
			name: "org id header has more than one value ",
			context: func() context.Context {
				return collectorclient.NewContext(context.Background(), collectorclient.Info{
					Metadata: collectorclient.NewMetadata(
						map[string][]string{
							orgID: {strconv.Itoa(gcom.GrafanaInstanceOne.ID), "test"}},
					),
				})
			},
			error: "2 source keys found in the context, can't determine which one to use",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, err := p.enrichContextWithSignalInstanceURL(
				tt.context(),
				func(i client.Instance) string {
					return i.TracesInstanceURL
				},
			)
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
