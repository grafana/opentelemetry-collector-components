package cache

import (
	"context"
	"flag"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

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

// String implements a diagnostic for the flag Value interface
func (it *instanceTypes) String() string {
	return fmt.Sprint(*it)
}

// Set implements the flag Value interface
// This enables either `-flag val1,val2` OR `-flag val1 -flag val2`
func (it *instanceTypes) Set(value string) error {
	for _, strInstanceType := range strings.Split(value, ",") {
		instanceType, err := client.InstanceTypeFromString(strInstanceType)
		if err != nil {
			return err
		}
		*it = append(*it, instanceType)
	}
	return nil
}

// RegisterFlagsWithPrefix adds the flags required to config this to the given FlagSet
func (cfg *InstanceCacheConfig) RegisterFlagsWithPrefix(prefix string, f *flag.FlagSet) {
	f.DurationVar(
		&cfg.IncrementalCacheRefreshDuration,
		prefix+"incremental-cache-refresh-duration",
		5*time.Minute,
		"duration until instance cache is updated with changes.",
	)
	f.DurationVar(
		&cfg.CompleteCacheRefreshDuration,
		prefix+"complete-cache-refresh-duration",
		5*time.Hour,
		"duration until instance cache is completely refreshed.",
	)
	f.StringVar(&cfg.GrafanaClusterFilter,
		prefix+"instance-cache.grafana-cluster-filter",
		"",
		"load instances in grafana cache only from specified cluster")
	f.Var(&cfg.InstanceTypes,
		prefix+"instance-cache.instance-types",
		"comma-separated list of instance type caches")
}

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
