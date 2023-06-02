package cache

import (
	"testing"
	"time"

	"github.com/go-kit/log"
	"github.com/stretchr/testify/require"

	"github.com/grafana/opentelemetry-collector-components/processor/gcomapiprocessor/internal/gcom/client"
	"github.com/grafana/opentelemetry-collector-components/processor/gcomapiprocessor/internal/gcom/client/mock"
)

const someCluster = "cluster"
const otherCluster = "cluster_other"

var mockInstances = map[client.InstanceType][]client.Instance{
	client.Metrics: {
		{
			ID:          1,
			OrgID:       1,
			ClusterName: otherCluster,
		},
		{
			ID:          2,
			OrgID:       2,
			ClusterName: otherCluster,
		},
	},
	client.Logs: {
		{
			ID:          3,
			OrgID:       3,
			ClusterName: otherCluster,
		},
		{
			ID:          4,
			OrgID:       4,
			ClusterName: otherCluster,
		},
	},
	client.Traces: {
		{
			ID:          5,
			OrgID:       5,
			ClusterName: otherCluster,
		},
		{
			ID:          6,
			OrgID:       6,
			ClusterName: otherCluster,
		},
	},
	client.Alerts: {
		{
			ID:          7,
			OrgID:       7,
			ClusterName: otherCluster,
		},
		{
			ID:          8,
			OrgID:       8,
			ClusterName: otherCluster,
		},
	},
	client.Grafana: {
		{
			ID:          9,
			OrgID:       9,
			ClusterName: someCluster,
		},
		{
			ID:          10,
			OrgID:       10,
			ClusterName: otherCluster,
		},
	},
}

func TestInstanceCache_NewInstanceCache_ClusterFilter(t *testing.T) {
	mockClient := &mock.Client{Instances: mockInstances}

	tests := []struct {
		name          string
		instanceTypes []client.InstanceType
	}{
		{
			name: "Select Instance Types",
			instanceTypes: []client.InstanceType{
				client.Alerts,
				client.Grafana,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cacheCfg := InstanceCacheConfig{
				IncrementalCacheRefreshDuration: 10 * time.Minute,
				CompleteCacheRefreshDuration:    10 * time.Minute,
				GrafanaClusterFilter:            someCluster,
			}
			cache, err := NewInstanceCache(
				cacheCfg,
				log.NewNopLogger(),
				tt.instanceTypes,
				mockClient,
			)

			if err != nil {
				t.Errorf("error building cache: %s", err)
				return
			}

			instanceTypes := []client.InstanceType{
				client.Metrics,
				client.Logs,
				client.Alerts,
				client.Traces,
				client.Grafana,
			}

			for _, instanceType := range instanceTypes {
				enabled := false

				if tt.instanceTypes == nil {
					enabled = true
				}

				for _, curType := range tt.instanceTypes {
					if curType == instanceType {
						enabled = true
					}
				}

				mockInstances := mockInstances[instanceType]

				for _, instance := range mockInstances {
					if instanceType == client.Grafana && instance.ClusterName != cacheCfg.GrafanaClusterFilter {
						// If the cluster was filtered out, do not report this as an error
						continue
					}

					cachedInstance, err := cache.GetInstanceInfo(instanceType, instance.ID)
					if err != nil && !enabled {
						continue
					} else if err != nil && enabled {
						t.Errorf("Expected instance cache to contain instance %d but it doesn't", instance.ID)
					} else if err == nil && !enabled {
						t.Errorf("Expected instance cache to not contain instance %d but it does", instance.ID)
					} else {
						if cachedInstance.ID != instance.ID {
							t.Errorf("Expected instance id %d to to equal %d", cachedInstance.ID, instance.ID)
						}
						if cachedInstance.OrgID != instance.OrgID {
							t.Errorf("Expected instance org id %d to to equal %d", cachedInstance.OrgID, instance.OrgID)
						}
					}
				}
			}
		})
	}
}

func TestInstanceCache_NewInstanceCache(t *testing.T) {
	mockClient := &mock.Client{Instances: mockInstances}

	tests := []struct {
		name          string
		instanceTypes []client.InstanceType
	}{
		{
			name: "Select Instance Types",
			instanceTypes: []client.InstanceType{
				client.Alerts,
				client.Grafana,
			},
		},
		{
			name:          "Metrics Only",
			instanceTypes: []client.InstanceType{client.Metrics},
		},
		{
			name:          "Logs Only",
			instanceTypes: []client.InstanceType{client.Logs},
		},
		{
			name:          "Traces Only",
			instanceTypes: []client.InstanceType{client.Traces},
		},
		{
			name:          "Alerts Only",
			instanceTypes: []client.InstanceType{client.Alerts},
		},
		{
			name:          "Grafana Only",
			instanceTypes: []client.InstanceType{client.Grafana},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache, err := NewInstanceCache(
				InstanceCacheConfig{
					IncrementalCacheRefreshDuration: 10 * time.Minute,
					CompleteCacheRefreshDuration:    10 * time.Minute,
				},
				log.NewNopLogger(),
				tt.instanceTypes,
				mockClient,
			)

			if err != nil {
				t.Errorf("error building cache: %s", err)
				return
			}

			instanceTypes := []client.InstanceType{
				client.Metrics,
				client.Logs,
				client.Alerts,
				client.Traces,
				client.Grafana,
			}

			for _, instanceType := range instanceTypes {
				enabled := false

				if tt.instanceTypes == nil {
					enabled = true
				}

				for _, curType := range tt.instanceTypes {
					if curType == instanceType {
						enabled = true
					}
				}

				mockInstances := mockInstances[instanceType]

				for _, instance := range mockInstances {
					cachedInstance, err := cache.GetInstanceInfo(instanceType, instance.ID)
					if err != nil && !enabled {
						continue
					} else if err != nil && enabled {
						t.Errorf("Expected instance cache to contain instance %d but it doesn't", instance.ID)
					} else if err == nil && !enabled {
						t.Errorf("Expected instance cache to not contain instance %d but it does", instance.ID)
					} else {
						if cachedInstance.ID != instance.ID {
							t.Errorf("Expected instance id %d to to equal %d", cachedInstance.ID, instance.ID)
						}
						if cachedInstance.OrgID != instance.OrgID {
							t.Errorf("Expected instance org id %d to to equal %d", cachedInstance.OrgID, instance.OrgID)
						}
					}
				}
			}
		})
	}
}

func TestInstanceCache_completeCacheRefresh(t *testing.T) {
	initialInstances := map[client.InstanceType][]client.Instance{
		client.Logs: {{ID: 1}, {ID: 2}},
	}

	mockClient := &mock.Client{Instances: initialInstances}

	cache, err := NewInstanceCache(
		InstanceCacheConfig{
			IncrementalCacheRefreshDuration: 10 * time.Minute,
			CompleteCacheRefreshDuration:    20 * time.Millisecond,
		},
		log.NewNopLogger(),
		[]client.InstanceType{client.Logs},
		mockClient,
	)
	require.NoError(t, err)

	newInstances := map[client.InstanceType][]client.Instance{
		client.Logs: {{ID: 3}, {ID: 4}},
	}

	mockClient.ResetInstances(newInstances)

	// Verify cache is completely refreshed after TTL and contains new instances.
	require.Eventually(t, func() bool {
		_, err = cache.GetInstanceInfo(client.Logs, initialInstances[client.Logs][0].ID)
		_, err = cache.GetInstanceInfo(client.Logs, initialInstances[client.Logs][1].ID)
		if err == nil {
			return false
		}

		_, err = cache.GetInstanceInfo(client.Logs, newInstances[client.Logs][0].ID)
		_, err = cache.GetInstanceInfo(client.Logs, newInstances[client.Logs][1].ID)
		return err == nil
	}, time.Second, 5*time.Millisecond)
}

func TestInstanceCache_incrementalCacheRefresh(t *testing.T) {
	initialInstances := map[client.InstanceType][]client.Instance{
		client.Logs: {{ID: 1}, {ID: 2}},
	}

	mockClient := &mock.Client{Instances: initialInstances}

	cache, err := NewInstanceCache(
		InstanceCacheConfig{
			IncrementalCacheRefreshDuration: 20 * time.Millisecond,
			CompleteCacheRefreshDuration:    10 * time.Minute,
		},
		log.NewNopLogger(),
		[]client.InstanceType{client.Logs},
		mockClient,
	)
	require.NoError(t, err)

	nextInstances := map[client.InstanceType][]client.Instance{
		client.Logs: {{ID: 3}},
	}

	mockClient.ResetInstances(nextInstances)

	// Verify cache is incrementally refreshed after TTL and contains the new instance.
	require.Eventually(t, func() bool {
		_, err = cache.GetInstanceInfo(client.Logs, initialInstances[client.Logs][0].ID)
		_, err = cache.GetInstanceInfo(client.Logs, initialInstances[client.Logs][1].ID)
		_, err = cache.GetInstanceInfo(client.Logs, nextInstances[client.Logs][0].ID)
		return err == nil
	}, time.Second, 5*time.Millisecond)

	deletedNonCachedInstance := client.Instance{ID: 5, Status: client.InstanceStatusDeleted}

	nextInstances = map[client.InstanceType][]client.Instance{
		client.Logs: {
			{ID: 1, Status: client.InstanceStatusDeleted},
			deletedNonCachedInstance,
		},
	}

	mockClient.ResetInstances(nextInstances)

	// Verify cache is incrementally refreshed after TTL and the deleted instances aren't cached.
	require.Eventually(t, func() bool {
		_, err = cache.GetInstanceInfo(client.Logs, initialInstances[client.Logs][0].ID)
		if err == nil {
			return false
		}

		_, err = cache.GetInstanceInfo(client.Logs, deletedNonCachedInstance.ID)
		return err != nil
	}, time.Second, 5*time.Millisecond)
}
