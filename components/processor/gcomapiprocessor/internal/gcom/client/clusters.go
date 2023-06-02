package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/go-querystring/query"
)

type ClusterRequestOptions struct {
	Slugs []string `url:"slugIn,omitempty"`

	// Type is only used to select the correct API endpoint.
	Type InstanceType `url:"-"`
}

// This is purposefully minimal, as the only use case right now is
// looking up the cluster ID. Other fields can be parsed as necessary.
type Cluster struct {
	ID   int    `json:"id"`
	Slug string `json:"slug"`
}

// ClusterResponse is the `/api/hm-clusters` response from grafana.com
type ClusterResponse struct {
	Items []Cluster `json:"items"`
}

// ListClusters returns a list of clusters from the Grafana Cloud API.
func (c *client) ListClusters(ctx context.Context, options ClusterRequestOptions) ([]Cluster, error) {
	var (
		req      *http.Request
		err      error
		endpoint string
	)

	cfgEndpoint := strings.TrimRight(c.endpoint.String(), "/")

	switch options.Type {
	case Grafana:
		endpoint = cfgEndpoint + "/hg-clusters"
	default:
		// All other hosted services share the hosted-metrics endpoint.
		endpoint = cfgEndpoint + "/hm-clusters"
	}

	req, err = http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", c.key))

	q, err := query.Values(options)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = q.Encode()

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		msg, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("invalid response, url=%v status=%v error=%v", req.URL.String(), res.StatusCode, err)
		}
		return nil, fmt.Errorf("invalid response, url=%v status=%v, msg=%v",
			req.URL.String(),
			res.StatusCode,
			string(msg))
	}

	var clusters ClusterResponse
	if err := json.NewDecoder(res.Body).Decode(&clusters); err != nil {
		return nil, fmt.Errorf("error decoding hm-clusters response, %v", err)
	}

	return clusters.Items, nil
}
