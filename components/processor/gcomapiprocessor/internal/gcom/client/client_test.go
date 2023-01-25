package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
)

// setup sets up a test HTTP server along with grafana.com client configured to
// use the test server
func setup() (client Client, mux *http.ServeMux, teardown func()) {
	// mux is the HTTP request multiplexer used with the test server.
	mux = http.NewServeMux()

	// We want to ensure that tests catch mistakes where the endpoint URL is
	// specified as absolute rather than relative. It only makes a difference
	// when there's a non-empty base URL path. So, use that. See issue #752.
	apiHandler := http.NewServeMux()
	apiHandler.Handle("/api/", http.StripPrefix("/api", mux))

	// server is a test HTTP server used to provide mock API responses.
	server := httptest.NewServer(apiHandler)

	// client is the GitHub client being tested and is
	// configured to use test server.

	url, _ := url.Parse(server.URL + "/api/")
	c, _ := New(Config{
		Key:      "",
		Endpoint: url.String(),
	}, "test", log.NewNopLogger())

	return c, mux, server.Close
}

func TestClientMetrics(t *testing.T) {
	// So I am trying to test that we're seeing different status codes.
	// The test ideally has a CompareAndGather which compares things
	// but with histograms, its not possible because the _sum will be different.
	// To work around this, I am looking to create 4 different possibilites of labels
	// and then count that.

	ctx := context.Background()

	c, mux, teardown := setup()
	defer teardown()

	alreadySeen := false

	mux.HandleFunc("/hosted-metrics", func(w http.ResponseWriter, r *http.Request) {
		if alreadySeen {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		alreadySeen = true
		fmt.Fprint(w, `{"items": [`+hostedPrometheusInstance+`], "orderBy": "name", "direction": "asc", "links": [{ "rel": "self", "href": "/hosted-metrics" }]}`)
	})

	mux.HandleFunc("/hosted-logs", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprint(w, sampleLogResponse)
	})

	mux.HandleFunc("/hosted-alerts", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, sampleAlertsResponse)
	})

	_, err := c.ListInstances(ctx, InstanceRequestOptions{})
	require.NoError(t, err)
	_, err = c.ListInstances(ctx, InstanceRequestOptions{
		Type: Prometheus,
	})
	require.Error(t, err) // Should see a 500 now.
	_, err = c.ListInstances(ctx, InstanceRequestOptions{
		Type: Logs,
	})
	require.Error(t, err) // Should see a 429
	_, err = c.ListInstances(ctx, InstanceRequestOptions{
		Type: Alerts,
	})
	require.Error(t, err) // Should see a 404

	numChildren, err := testutil.GatherAndCount(prometheus.DefaultGatherer, "grafanacom_api_request_duration_seconds")
	require.NoError(t, err)
	require.Equal(t, 4, numChildren)
}
