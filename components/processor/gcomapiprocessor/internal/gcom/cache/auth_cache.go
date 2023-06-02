package cache

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/opentracing/opentracing-go"

	"github.com/grafana/opentelemetry-collector-components/processor/gcomapiprocessor/internal/gcom/client"
)

var errBadKey = errors.New("API Key Not Valid")

type AuthCache struct {
	client   client.Client
	logger   log.Logger
	ttl      time.Duration
	items    map[string]cacheItem
	itemsMtx sync.RWMutex
}

func NewAuthCache(client client.Client, logger log.Logger, ttl time.Duration) *AuthCache {
	return &AuthCache{
		client: client,
		logger: logger,
		ttl:    ttl,
		items:  map[string]cacheItem{},
	}
}

type cacheItem struct {
	Key        *client.APIKey
	ExpireTime time.Time
}

func (a *AuthCache) get(key string) (*client.APIKey, bool) {
	a.itemsMtx.RLock()
	defer a.itemsMtx.RUnlock()
	if c, ok := a.items[key]; ok {
		return c.Key, c.ExpireTime.After(time.Now())
	}
	return nil, false
}

func (a *AuthCache) Set(key string, u *client.APIKey) {
	a.itemsMtx.Lock()
	defer a.itemsMtx.Unlock()

	a.items[key] = cacheItem{
		Key:        u,
		ExpireTime: time.Now().Add(a.ttl),
	}
}

func (a *AuthCache) Clear() {
	a.itemsMtx.Lock()
	defer a.itemsMtx.Unlock()

	a.items = make(map[string]cacheItem)
}

// CheckKey checks if the provided key is valid
func (a *AuthCache) CheckKey(ctx context.Context, key string) (*client.APIKey, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "Gateway.CheckKey")
	defer span.Finish()

	k, ok := a.get(key)
	if ok {
		if k == nil {
			level.Debug(a.logger).Log("msg", "cache hit for invalid key")
			return k, errBadKey
		}
		level.Debug(a.logger).Log("msg", "cache hit for valid key")
		return k, nil
	}

	keyResponse, err := a.client.CheckAPIKey(ctx, key)
	if err != nil {
		if errors.Is(err, client.ErrAccess) && k != nil {
			level.Warn(a.logger).Log("msg", "unable to access grafana.com, using cached value")
			a.Set(key, k)
			return k, nil
		}

		level.Debug(a.logger).Log("msg", "cache miss set for key")
		a.Set(key, nil)
		return nil, err
	}

	level.Debug(a.logger).Log("msg", "cache hit set for key")
	a.Set(key, keyResponse)
	return keyResponse, nil
}
