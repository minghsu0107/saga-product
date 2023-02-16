package cache

import (
	"encoding/json"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/minghsu0107/saga-product/config"
)

// LocalCache is the interface of local cache
type LocalCache interface {
	Get(key string, dst interface{}) (bool, error)
	Set(key string, val interface{}) error
	Delete(key string) error
}

// LocalCacheImpl implements the Cache interface
type LocalCacheImpl struct {
	cache *bigcache.BigCache
}

// NewLocalCache is the factory of local cache
func NewLocalCache(config *config.Config) (LocalCache, error) {
	cacheConfig := bigcache.DefaultConfig(time.Duration(config.LocalCacheConfig.ExpirationSeconds) * time.Second)
	cacheConfig.CleanWindow = 5 * time.Minute
	cache, err := bigcache.NewBigCache(cacheConfig)
	if err != nil {
		return nil, err
	}
	return &LocalCacheImpl{
		cache: cache,
	}, nil
}

// Get returns true if the key already exists and set dst to the corresponding value
func (lc *LocalCacheImpl) Get(key string, dst interface{}) (bool, error) {
	val, err := lc.cache.Get(key)
	if err == bigcache.ErrEntryNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	} else {
		if err := json.Unmarshal([]byte(val), dst); err != nil {
			return false, err
		}
	}
	return true, nil
}

// Set sets a value by key
func (lc *LocalCacheImpl) Set(key string, val interface{}) error {
	jsonVal, err := json.Marshal(val)
	if err != nil {
		return err
	}
	if err = lc.cache.Set(key, jsonVal); err != nil {
		return err
	}
	return nil
}

// Delete deletes a key
func (lc *LocalCacheImpl) Delete(key string) error {
	if err := lc.cache.Delete(key); err != nil {
		return err
	}
	return nil
}
