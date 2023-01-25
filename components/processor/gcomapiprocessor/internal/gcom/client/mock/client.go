package mock

import (
	"context"
	"errors"
	"sync"

	"github.com/grafana/opentelemetry-collector-components/processor/gcomapiprocessor/internal/gcom/client"
)

// Client implements the functions to serve as a grafana.com API client.
type Client struct {
	// TODO(56quarters): These should all be private to enforce they're only
	//  accessed while holding the mutex since they're both read and written
	//  from different goroutines during various unit tests.
	mtx       sync.Mutex
	Keys      map[string]*client.APIKey
	Instances map[client.InstanceType][]client.Instance
	Orgs      []client.Org
}

func (m *Client) AddInstance(instanceType client.InstanceType, i client.Instance) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.Instances[instanceType] = append(m.Instances[instanceType], i)
}

func (m *Client) ResetInstances(instances map[client.InstanceType][]client.Instance) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.Instances = instances
}

// CheckAPIKey returns the configured api key if found.
func (m *Client) CheckAPIKey(ctx context.Context, key string) (*client.APIKey, error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	k, found := m.Keys[key]
	if !found {
		return nil, errors.New("key not found")
	}
	return k, nil
}

// ListInstances returns the configured instances if found.
func (m *Client) ListInstances(ctx context.Context, options client.InstanceRequestOptions) ([]client.Instance, error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	var instances, foundInstances []client.Instance

	instanceType, _ := client.InstanceTypeFromString(options.Type.String())

	if instanceType == client.Prometheus || instanceType == client.Graphite || instanceType == client.GraphiteShared {
		instances = m.Instances[client.Metrics]
	} else if instanceType == "" {
		// Note that the real ListInstances function defaults to the metrics endpoint.
		instances = m.Instances[client.Metrics]
	} else {
		instances = m.Instances[instanceType]
	}

	for i := range instances {
		if options.ID == 0 || instances[i].ID == options.ID {
			if options.Cluster != "" && instances[i].ClusterName != options.Cluster {
				continue
			}
			foundInstances = append(foundInstances, instances[i])
		}
	}

	return foundInstances, nil
}

func (m *Client) ListInstancesWithPagination(ctx context.Context, options client.InstanceRequestOptions) ([]client.Instance, error) {
	return m.ListInstances(ctx, options)
}

// GetInstance returns a single instance by its ID, if found.
func (m *Client) GetInstance(ctx context.Context, options client.InstanceRequestOptions) (client.Instance, error) {
	// Note that we don't acquire the mutex in this method since ListInstances does
	i, err := m.ListInstances(ctx, options)
	if err != nil {
		return client.Instance{}, err
	}

	if len(i) < 1 {
		return client.Instance{}, client.ErrInstanceNotFound
	}

	return i[0], nil
}

// ListOrgs returns the configured orgs if found.
func (m *Client) ListOrgs(ctx context.Context, options *client.OrgRequestOptions) ([]client.Org, error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	orgs := m.Orgs
	foundOrgs := []client.Org{}

	for i := range orgs {
		if options.ID == 0 || orgs[i].ID == options.ID {
			foundOrgs = append(foundOrgs, orgs[i])
		}
	}

	return foundOrgs, nil
}
