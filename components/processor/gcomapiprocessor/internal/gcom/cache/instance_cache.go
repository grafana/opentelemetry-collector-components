package cache

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/pkg/errors"

	"github.com/grafana/opentelemetry-collector-components/processor/gcomapiprocessor/internal/gcom/client"
)

type InstanceCache interface {
	GetInstanceInfo(instanceType client.InstanceType, instanceID int) (client.Instance, error)
	GetMetricsInstanceIDByOrgIDAndInstanceName(orgID, instanceName string) int
}

type InstanceCacheConfig struct {
	CompleteCacheRefreshDuration    time.Duration
	IncrementalCacheRefreshDuration time.Duration
	GrafanaClusterFilter            string
	InstanceTypes                   instanceTypes
}

type instanceTypes []client.InstanceType

type InstanceCachesConfig struct {
	CompleteCacheRefreshDuration    time.Duration
	IncrementalCacheRefreshDuration time.Duration
	GrafanaClusterFilters           clusterFilters
}

type clusterFilters []string

type instanceCache struct {
	cfg             InstanceCacheConfig
	logger          log.Logger
	gclient         client.Client
	instanceTypes   []client.InstanceType
	lastCacheUpdate time.Time

	cacheMutex       sync.RWMutex
	instances        map[client.InstanceType]map[int]client.Instance
	idByOrgIDAndName map[client.InstanceType]map[string]int
}

// NewStaticInstanceCache creates an instanceCache for tests with a static list of instances.
func NewStaticInstanceCache(
	logger log.Logger,
	instances map[client.InstanceType]map[int]client.Instance,
	idByOrgIDAndName map[client.InstanceType]map[string]int,
) InstanceCache {
	return &instanceCache{
		logger:           logger,
		instances:        instances,
		idByOrgIDAndName: idByOrgIDAndName,
	}
}

// NewInstanceCache creates an instanceCache which maintains a cache of instances and keeps it updated.
func NewInstanceCache(
	cfg InstanceCacheConfig,
	logger log.Logger,
	instanceTypes []client.InstanceType,
	gclient client.Client,
) (InstanceCache, error) {

	if len(instanceTypes) == 0 {
		return nil, fmt.Errorf("configure at least 1 instance cache type")
	}

	i := &instanceCache{
		cfg:              cfg,
		logger:           logger,
		gclient:          gclient,
		instanceTypes:    instanceTypes,
		instances:        map[client.InstanceType]map[int]client.Instance{},
		idByOrgIDAndName: map[client.InstanceType]map[string]int{},
	}

	for _, instanceType := range instanceTypes {
		err := i.completeCacheRefresh(instanceType)
		if err != nil {
			return nil, err
		}
	}

	go i.run()

	return i, nil
}

func (i *instanceCache) setLastCacheUpdate(instance client.Instance) {
	if instance.UpdatedAt != nil && instance.UpdatedAt.After(i.lastCacheUpdate) {
		i.lastCacheUpdate = *instance.UpdatedAt
	}

	if instance.CreatedAt.After(i.lastCacheUpdate) {
		i.lastCacheUpdate = instance.CreatedAt
	}
}

func (i *instanceCache) completeCacheRefresh(instanceType client.InstanceType) error {
	level.Info(i.logger).Log("msg", "attempting to build instance cache", "type", instanceType)

	instanceReq := client.InstanceRequestOptions{
		Type:     instanceType,
		PageSize: 1000,
	}

	if instanceType == client.Grafana {
		instanceReq.Cluster = i.cfg.GrafanaClusterFilter
	}

	instances, err := i.gclient.ListInstancesWithPagination(context.TODO(), instanceReq)
	if err != nil {
		return err
	}

	instanceCache := map[int]client.Instance{}
	idByOrgIDAndName := map[string]int{}

	for _, instance := range instances {
		i.setLastCacheUpdate(instance)
		instanceCache[instance.ID] = instance
		idByOrgIDAndName[strconv.Itoa(instance.OrgID)+instance.Name] = instance.ID
		level.Debug(i.logger).Log(
			"msg", "adding instance to cache",
			"instanceID", instance.ID,
			"orgID", instance.OrgID,
			"name", instance.Name)
	}

	i.cacheMutex.Lock()
	defer i.cacheMutex.Unlock()

	i.instances[instanceType] = instanceCache
	i.idByOrgIDAndName[instanceType] = idByOrgIDAndName
	return nil
}

func (i *instanceCache) incrementalCacheRefresh(instanceType client.InstanceType) error {
	level.Info(i.logger).Log("msg", "attempting to build instance cache", "type", instanceType)

	instanceReq := client.InstanceRequestOptions{
		Type:                  instanceType,
		PageSize:              1000,
		IncludeDeleted:        true,
		UpdatedOrCreatedAtMin: i.lastCacheUpdate,
	}

	if instanceType == client.Grafana {
		instanceReq.Cluster = i.cfg.GrafanaClusterFilter
	}

	instanceList, err := i.gclient.ListInstancesWithPagination(context.TODO(), instanceReq)
	if err != nil {
		return err
	}

	i.cacheMutex.Lock()
	defer i.cacheMutex.Unlock()
	for _, instance := range instanceList {
		i.setLastCacheUpdate(instance)

		_, found := i.instances[instanceType][instance.ID]

		if instance.Status == client.InstanceStatusDeleted {
			if found {
				level.Debug(i.logger).Log("msg", "removing deleted instance from cache", "instanceID", instance.ID)
				delete(i.instances[instanceType], instance.ID)
			}

			continue
		}

		level.Debug(i.logger).Log("msg", "adding/updating instance to cache", "instanceID", instance.ID)
		i.instances[instanceType][instance.ID] = instance
		i.idByOrgIDAndName[instanceType][strconv.Itoa(instance.OrgID)+instance.Name] = instance.ID
	}

	return nil
}

func (i *instanceCache) run() {
	completeRefreshTicker := time.NewTicker(i.cfg.CompleteCacheRefreshDuration)
	go func() {
		for range completeRefreshTicker.C {
			for _, instanceType := range i.instanceTypes {
				err := i.completeCacheRefresh(instanceType)
				if err != nil {
					level.Error(i.logger).Log("msg", "unable to rebuild instance cache", "type", instanceType, "err", err)
				}
			}
		}
	}()

	incrementalRefreshTicker := time.NewTicker(i.cfg.IncrementalCacheRefreshDuration)
	go func() {
		for range incrementalRefreshTicker.C {
			for _, instanceType := range i.instanceTypes {
				err := i.incrementalCacheRefresh(instanceType)
				if err != nil {
					level.Error(i.logger).Log("msg", "unable to rebuild instance cache", "type", instanceType, "err", err)
				}
			}
		}
	}()
}

func (i *instanceCache) GetInstanceInfo(instanceType client.InstanceType, instanceID int) (client.Instance, error) {
	i.cacheMutex.RLock()
	defer i.cacheMutex.RUnlock()

	instances, ok := i.instances[instanceType]
	if !ok {
		return client.Instance{}, fmt.Errorf("%s instance cache doesn't exist", instanceType)
	}

	instance, ok := instances[instanceID]
	if !ok {
		return client.Instance{}, fmt.Errorf("%s instance with ID %d does not exist", instanceType, instanceID)
	}

	return instance, nil
}

// GetMetricsInstanceIDByOrgIDAndInstanceName returns the instance id mapped by given orgID and instanceName.
// If matching instance id is not found it returns -1.
func (i *instanceCache) GetMetricsInstanceIDByOrgIDAndInstanceName(orgID, instanceName string) int {
	i.cacheMutex.RLock()
	defer i.cacheMutex.RUnlock()

	key := orgID + instanceName
	level.Debug(i.logger).Log("msg", "getting instance map", "key", key)

	metricsIDByOrgIDAndName, ok := i.idByOrgIDAndName[client.Metrics]
	if !ok {
		return -1
	}

	id, ok := metricsIDByOrgIDAndName[key]
	if !ok {
		return -1
	}

	return id
}

type multiInstanceCache struct {
	instanceCaches []InstanceCache
}

// NewMultiInstanceCache creates an multiInstanceCache which maintains a cache of instances and keeps it updated.
func NewMultiInstanceCache(
	cfg InstanceCachesConfig,
	logger log.Logger,
	gclient client.Client,
) (InstanceCache, error) {
	if len(cfg.GrafanaClusterFilters) == 0 {
		return nil, errors.New("failed to create multiInstanceCache. must include at least one InstanceCache")
	}

	var caches []InstanceCache
	for _, clusterFilter := range cfg.GrafanaClusterFilters {
		ic, err := NewInstanceCache(
			InstanceCacheConfig{
				CompleteCacheRefreshDuration:    cfg.CompleteCacheRefreshDuration,
				IncrementalCacheRefreshDuration: cfg.IncrementalCacheRefreshDuration,
				GrafanaClusterFilter:            clusterFilter,
				InstanceTypes:                   []client.InstanceType{client.Grafana},
			},
			logger,
			[]client.InstanceType{client.Grafana},
			gclient,
		)
		if err != nil {
			return nil, err
		}
		caches = append(caches, ic)
	}

	return &multiInstanceCache{
		instanceCaches: caches,
	}, nil
}

// GetInstanceInfo implements InstanceCache
func (c *multiInstanceCache) GetInstanceInfo(instanceType client.InstanceType, instanceID int) (client.Instance, error) {
	var err error
	var instance client.Instance
	for _, ch := range c.instanceCaches {
		instance, err = ch.GetInstanceInfo(instanceType, instanceID)
		if err == nil {
			return instance, nil
		}
	}
	return client.Instance{}, err
}

// GetMetricsInstanceIDByOrgIDAndInstanceName implements InstanceCache
func (c *multiInstanceCache) GetMetricsInstanceIDByOrgIDAndInstanceName(orgID string, instanceName string) int {
	for _, ch := range c.instanceCaches {
		id := ch.GetMetricsInstanceIDByOrgIDAndInstanceName(orgID, instanceName)
		if id != -1 {
			return id
		}
	}
	return -1
}
