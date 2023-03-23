package faroauthextension

import (
	"context"

	pb "github.com/grafana/opentelemetry-collector-components/components/extension/faroauthextension/lookup/proto"
	"go.opentelemetry.io/collector/client"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	orgID      = "X-Scope-OrgID"
	faroAppKey = "X-Scope-FaroAppKey"
)

type faroAuth struct {
	cfg *Config

	conn   *grpc.ClientConn
	client pb.AppLookupClient

	logger *zap.Logger
}

func newFaroAuthExtension(logger *zap.Logger, cfg *Config) auth.Server {
	fa := &faroAuth{
		cfg:    cfg,
		logger: logger,
	}
	return auth.NewServer(
		auth.WithServerStart(fa.serverStart),
		auth.WithServerAuthenticate(fa.authenticate),
	)
}

func (fa *faroAuth) serverStart(ctx context.Context, host component.Host) error {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	var err error
	fa.conn, err = grpc.Dial(fa.cfg.FaroAPI.Endpoint, opts...)
	fa.client = pb.NewAppLookupClient(fa.conn)
	return err
}

func (fa *faroAuth) serverShutdown(ctx context.Context) error {
	if fa.conn != nil {
		return fa.conn.Close()
	}
	return nil
}

func (fa *faroAuth) authenticate(ctx context.Context, headers map[string][]string) (context.Context, error) {
	// get app key from context
	appKey := ctx.Value(faroAppKey)
	if appKey == nil {
		// if there is no app key, by-pass this authenticator
		// any receivers that use this can call it directly.
		return ctx, nil
	}

	req := &pb.LookupRequest{
		AppKey: appKey.(string),
	}
	res, err := fa.client.Lookup(ctx, req)
	if err != nil {
		fa.logger.Error("lookup request error", zap.Error(err))
		return ctx, err
	}

	cl := client.FromContext(ctx)

	md := make(map[string][]string)
	for k, v := range headers {
		md[k] = v
	}

	// set stackID as X-Org-ScopeID in client context metadata for the gcomapiprocessor
	stackId := res.GetStackID()
	md[orgID] = []string{stackId}

	// TODO: set extra log labels in Metadata? AuthData?

	// TODO: set cors allowed origins in ""

	cl.Metadata = client.NewMetadata(md)
	return client.NewContext(ctx, cl), nil
}
