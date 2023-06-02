package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-kit/log/level"
	"github.com/google/go-querystring/query"
	"golang.org/x/oauth2"
)

const (
	InstanceStatusActive  = "active"
	InstanceStatusDeleted = "deleted"
)

var (
	ErrInvalidInstanceID = errors.New("error: specified instance ID is invalid")
	ErrInstanceNotFound  = errors.New("error: specified instance ID does not exist")
)

// Instance represents a Grafana Cloud hosted metrics instance
type Instance struct {
	ID                     int          `json:"id"`
	OrgID                  int          `json:"orgId"`
	OrgSlug                string       `json:"orgSlug,omitempty"`
	OrgName                string       `json:"orgName"`
	Type                   InstanceType `json:"type"`
	ClusterID              int          `json:"clusterId"`
	ClusterSlug            string       `json:"clusterSlug,omitempty"`
	ClusterName            string       `json:"clusterName,omitempty"`
	Name                   string       `json:"name"`
	URL                    string       `json:"url,omitempty"`
	Description            string       `json:"description,omitempty"`
	CreatedAt              time.Time    `json:"createdAt,omitempty"`
	CreatedBy              string       `json:"createdBy,omitempty"`
	UpdatedAt              *time.Time   `json:"updatedAt,omitempty"`
	UpdatedBy              string       `json:"updatedBy,omitempty"`
	Plan                   string       `json:"plan,omitempty"`
	PlanName               string       `json:"planName,omitempty"`
	AlertsInstanceID       int          `json:"amInstanceId,omitempty"`
	LogsInstanceID         int          `json:"hlInstanceId,omitempty"`
	LogsInstanceURL        string       `json:"hlInstanceUrl,omitempty"`
	PromInstanceID         int          `json:"hmInstancePromId,omitempty"`
	PromInstanceURL        string       `json:"hmInstancePromUrl,omitempty"`
	GraphiteInstanceID     int          `json:"hmInstanceGraphiteId,omitempty"`
	TracesInstanceID       int          `json:"htInstanceId,omitempty"`
	TracesInstanceURL      string       `json:"htInstanceUrl,omitempty"`
	GrafanaInstanceID      int          `json:"grafanaInstanceId,omitempty"`
	GrafanaInstanceURL     string       `json:"grafanaInstanceUrl,omitempty"`
	ProfilesInstanceID     int          `json:"hpInstanceId,omitempty"`
	ProfilesInstanceURL    string       `json:"hpInstanceUrl,omitempty"`
	GeneratorURL           string       `json:"amInstanceGeneratorUrl,omitempty"`
	GeneratorURLDatasource string       `json:"amInstanceGeneratorUrlDatasource,omitempty"`
	Status                 string       `json:"status,omitempty"`

	Links []struct {
		Rel  string `json:"rel,omitempty"`
		Href string `json:"href,omitempty"`
	} `json:"links,omitempty"`
}

func (instance *Instance) GetStackID() int {
	if instance.Type == Grafana {
		return instance.ID
	}

	return instance.GrafanaInstanceID
}

// InstanceRequestOptions contains query params used when
// requesting grafana cloud instances
type InstanceRequestOptions struct {
	ID              int          `url:"id,omitempty"`
	OrgID           int          `url:"orgId,omitempty"`
	OrgSlug         string       `url:"orgSlug,omitempty"`
	Type            InstanceType `url:"type,omitempty"`
	Cluster         string       `url:"cluster,omitempty"`
	ClusterSlug     string       `url:"clusterSlug,omitempty"`
	ClusterIDs      []int        `url:"clusterIdIn,comma,omitempty"`
	Name            string       `url:"name,omitempty"`
	IncludeDeleted  bool         `url:"includeDeleted,omitempty"`
	MachineLearning bool         `url:"machineLearning,omitempty"`
	Incident        bool         `url:"incident,omitempty"`

	// TODO (Ryan Melendez): All instance endpoints have been updated to use cursor-based pagination as
	// a necessary performance improvement. All uses of this Page field should be replaced with the new
	// approach.
	Page     int `url:"page,omitempty"`
	Cursor   int `url:"cursor,omitempty"`
	PageSize int `url:"pageSize,omitempty"`

	// Only get instances that have been created, updated, or deleted since this UpdatedOrCreatedAtMin.
	UpdatedOrCreatedAtMin time.Time `url:"updatedOrCreatedAtMin,omitempty"`

	// Key is used to override the default client api key
	Key   string        `url:"-"`
	Token *oauth2.Token `url:"-"`
}

// InstanceResponse is the `/api/hosted-(metrics|logs)` response
// from grafana.com
type InstanceResponse struct {
	Items []Instance `json:"items"`
}

// ListInstances returns a list of instances from the Grafana Cloud API
func (c *client) ListInstances(ctx context.Context, options InstanceRequestOptions) ([]Instance, error) {
	var (
		req      *http.Request
		err      error
		noType   bool // if the endpoint does not supports the type parameter
		endpoint string
	)

	cfgEndpoint := strings.TrimRight(c.endpoint.String(), "/")

	switch options.Type {
	case Logs:
		noType = true
		endpoint = cfgEndpoint + "/hosted-logs"
	case Alerts:
		noType = true
		endpoint = cfgEndpoint + "/hosted-alerts"
	case Traces:
		noType = true
		endpoint = cfgEndpoint + "/hosted-traces"
	case Grafana:
		noType = true
		endpoint = cfgEndpoint + "/instances"
		options.ClusterIDs = nil // clusterIdIn is not supported for /instances
	case Metrics:
		noType = true
		endpoint = cfgEndpoint + "/hosted-metrics"
	case Profiles:
		noType = true
		endpoint = cfgEndpoint + "/hosted-profiles"
	default:
		endpoint = cfgEndpoint + "/hosted-metrics"
	}

	if noType {
		options.Type = ""
	}

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

	var instances InstanceResponse
	if err := json.NewDecoder(res.Body).Decode(&instances); err != nil {
		return nil, fmt.Errorf("error decoding instance response, %v", err)
	}

	return instances.Items, nil
}

// GetInstance returns a single instance from the Grafana Cloud API.
//
// Currently it's just a convenience wrapper around ListInstances,
// which checks that provided options contain the ID and that the API returns exactly 1 instance.
//
// If one of those conditions is not satisfied, an error is returned.
//
// TODO: implement this on top of GET /instances/:id instead.
func (c *client) GetInstance(ctx context.Context, options InstanceRequestOptions) (Instance, error) {
	var i Instance

	if options.ID == 0 {
		return i, ErrInvalidInstanceID
	}

	res, err := c.ListInstances(ctx, options)
	if err != nil {
		return i, err
	}

	if len(res) < 1 {
		level.Debug(c.logger).Log(
			"msg", "no grafana-com instances found with specified ID",
			"instanceID", options.ID,
		)

		return i, ErrInstanceNotFound
	}

	if len(res) > 1 {
		level.Warn(c.logger).Log(
			"msg", "multiple instances returned with same ID. Using first instance",
			"instanceID", options.ID,
			"instances", fmt.Sprint(res),
		)
	}

	return res[0], nil
}

func (c *client) ListInstancesWithPagination(ctx context.Context, options InstanceRequestOptions) ([]Instance, error) {
	var instances []Instance

	cursor := 1
	for {
		level.Debug(c.logger).Log(
			"msg", "fetching page of instances",
			"cursor", cursor,
		)

		options.Cursor = cursor
		pageInstances, err := c.ListInstances(ctx, options)
		if err != nil {
			return instances, err
		}

		instances = append(instances, pageInstances...)

		// No more instances to fetch
		if len(pageInstances) < options.PageSize {
			break
		}

		cursor = instances[len(instances)-1].ID + 1
	}

	return instances, nil
}
