package client

import (
	"context"
	"flag"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/grafana/opentelemetry-collector-components/processor/gcomapiprocessor/internal/gcom/common"
)

var (
	grafanaComReqs = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grafanacom_api_request_duration_seconds",
			Help:    "A histogram of request latencies by status_code and path.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"code", "path"},
	)
)

// Config contains the configuration required to
// create a grafana.com Client
type Config struct {
	Key      string
	Endpoint string
	Timeout  time.Duration
	Client   *http.Client

	// Mock allows for a mock client to be substituted
	// and returned by the New function
	Mock Client
}

// RegisterFlags adds the flags required to config this to the given FlagSet
func (cfg *Config) RegisterFlags(f *flag.FlagSet) {
	f.StringVar(&cfg.Key, "grafanacloud.client.key", "", "API key for accessing grafana.com")
	f.StringVar(&cfg.Endpoint, "grafanacloud.client.endpoint", "https://www.grafana.com/api", "URL for grafana.com api")
	f.DurationVar(&cfg.Timeout, "grafanacloud.client.timeout", 1*time.Minute, "Timeout for requests to grafana.com api")
}

// Client is used to interact with the GrafanaCloud api
type Client interface {
	// CheckAPIKey checks a Grafana Cloud API key and returns its parsed representation.
	CheckAPIKey(ctx context.Context, key string) (*APIKey, error)

	// ListInstances returns a list of Grafana Cloud instances from API.
	ListInstances(ctx context.Context, options InstanceRequestOptions) ([]Instance, error)

	// ListInstancesWithPagination returns a list of Grafana Cloud instances from API,
	// but it requests them one page at a time to avoid a timeout.
	ListInstancesWithPagination(ctx context.Context, options InstanceRequestOptions) ([]Instance, error)

	// GetInstance returns a single Grafana Cloud instance with the ID provided in the options.
	// If the ID is not provided, an error is returned.
	GetInstance(ctx context.Context, options InstanceRequestOptions) (Instance, error)

	// ListOrgs returns a list of Grafana Cloud organizations from API.
	ListOrgs(ctx context.Context, options *OrgRequestOptions) ([]Org, error)

	// ListClusters returns a list of clusters.
	ListClusters(ctx context.Context, options ClusterRequestOptions) ([]Cluster, error)
}

type client struct {
	key      string
	endpoint *url.URL
	client   *http.Client
	logger   log.Logger
}

// New creates a grafana.com Client
func New(cfg Config, name string, logger log.Logger) (Client, error) {
	if cfg.Mock != nil {
		return cfg.Mock, nil
	}
	endpoint, err := url.Parse(cfg.Endpoint)
	if err != nil {
		return nil, err
	}

	var httpCli *http.Client

	if cfg.Client != nil {
		httpCli = cfg.Client
	} else {
		httpCli = &http.Client{
			Timeout: cfg.Timeout,
			Transport: newInstrumentedRoundtripper(
				common.NewConntrackRoundTripper(common.NewDefaultHTTPTransport(), name)),
		}
	}

	client := &client{
		key:      cfg.Key,
		endpoint: endpoint,
		client:   httpCli,
		logger:   logger,
	}

	level.Info(logger).Log("msg", "grafana.com client configured", "endpoint", endpoint.String())
	return client, nil
}

// Using a custom roundtripper and not reusing the promhttp one because I also want to record the paths.
// The advantage is that gcom API looks endpoint/{hosted-logs, hosted-metrics, orgs, etc.}/.* and we can
// capture just the first "directory".
type instrumentedRoundtripper struct {
	rt http.RoundTripper
}

func newInstrumentedRoundtripper(rt http.RoundTripper) http.RoundTripper {
	return instrumentedRoundtripper{
		rt: rt,
	}
}

func (rt instrumentedRoundtripper) RoundTrip(r *http.Request) (*http.Response, error) {
	// Figure out the path from the known ones.
	path := "<unrecognised>"
	urlStr := r.URL.String()

	knownPaths := []string{
		"hosted-metrics",
		"hosted-alerts",
		"hosted-logs",
		"hosted-traces",
		"orgs",
		"api-keys",
	}

	for _, kp := range knownPaths {
		if strings.Contains(urlStr, kp) {
			path = kp
			break
		}
	}

	start := time.Now()
	resp, err := rt.rt.RoundTrip(r)
	dur := time.Since(start)
	code := 500
	if err == nil && resp != nil {
		code = resp.StatusCode
	}
	grafanaComReqs.WithLabelValues(strconv.Itoa(code), path).Observe(dur.Seconds())

	return resp, err
}
