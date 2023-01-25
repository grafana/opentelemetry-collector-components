package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/go-querystring/query"
	"golang.org/x/oauth2"
)

// Org represents a Grafana Cloud org
type Org struct {
	ID                    int                 `json:"id"`
	Slug                  string              `json:"slug"`
	Name                  string              `json:"name"`
	GrafanaCloudType      OrgGrafanaCloudType `json:"grafanaCloud"`
	ContractType          OrgContractType     `json:"contractType"`
	MetricsUsage          int                 `json:"hmUsage"`
	MetricsIncludedSeries int                 `json:"hmIncludedSeries"`
	MetricsOverageAmount  float64             `json:"hmOverageAmount"`
	LogsUsage             float64             `json:"hlUsage"`
	LogsOverageAmount     float64             `json:"hlOverageAmount"`
	LogsIncludedUsage     float64             `json:"hlIncludedUsage"`

	Links []struct {
		Rel  string `json:"rel,omitempty"`
		Href string `json:"href,omitempty"`
	} `json:"links,omitempty"`
}

// OrgRequestOptions contains query params used when requesting orgs
type OrgRequestOptions struct {
	ID                int                   `url:"id,omitempty"`
	Slug              string                `url:"slug,omitempty"`
	Type              OrgType               `url:"type,omitempty"`
	GrafanaCloudType  OrgGrafanaCloudType   `url:"grafanaCloud,omitempty"`
	GrafanaCloudTypes []OrgGrafanaCloudType `url:"grafanaCloudIn,omitempty"`
	ContractType      OrgContractType       `url:"contractType,omitempty"`

	// Key is used to override the default client api key
	Key   string        `url:"-"`
	Token *oauth2.Token `url:"-"`
}

// OrgResponse is the `/api/orgs` response from grafana.com
type OrgResponse struct {
	Items []Org `json:"items"`
}

// ListOrgs returns a list of instances from the Grafana Cloud API
func (c *client) ListOrgs(ctx context.Context, options *OrgRequestOptions) ([]Org, error) {
	var (
		req *http.Request
		err error
	)

	endpoint := strings.TrimRight(c.endpoint.String(), "/") + "/orgs"

	if options != nil {
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, err
		}

		// Allow for the api key used to be passed to the request
		if options.Key != "" {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", options.Key))
		} else if options.Token != nil {
			options.Token.SetAuthHeader(req)
		} else if c.key != "" {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", c.key))
		}

		q, err := query.Values(options)
		if err != nil {
			return nil, err
		}
		req.URL.RawQuery = q.Encode()
	} else {
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", c.key))
	}

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

	var orgs OrgResponse
	if err := json.NewDecoder(res.Body).Decode(&orgs); err != nil {
		return nil, fmt.Errorf("error decoding org response, %v", err)
	}

	return orgs.Items, nil
}
