package cache

import (
	"fmt"
	"sync"

	pb "github.com/weaveworks/weave-gitops/pkg/api/profiles"
)

// Cache defines an interface to work with the profile data cacher.
type Cache interface {
	Add(key string, value Data)
	Remove(key string)
	Get(key string) *Data
}

// Data is explicit data for a specific profile including values.
type Data struct {
	Profile *pb.Profile
	Values  []byte
}

// HelmCache is used to cache profiles data from scanner helm repositories.
type HelmCache struct {
	cache map[string]*Data
	mut   *sync.RWMutex
}

// NewCache initialises the cache and returns it.
func NewCache() *HelmCache {
	return &HelmCache{
		cache: make(map[string]*Data),
		mut:   &sync.RWMutex{},
	}
}

// Add adds a new entry to the cache. The key should be generated using MakeKey function.
func (c *HelmCache) Add(key string, value Data) {
	c.mut.Lock()
	defer c.mut.Unlock()
	c.cache[key] = &value
}

func (c *HelmCache) Remove(key string) {
	c.mut.Lock()
	defer c.mut.Unlock()
	delete(c.cache, key)
}

// Get will return nil in case the data cannot be found but no error.
func (c *HelmCache) Get(key string) *Data {
	c.mut.RLock()
	defer c.mut.RUnlock()

	return c.cache[key]
}

// MakeKey defines the way a key for a specific data is generated.
// format: Repo/Chart/Version
func (c *HelmCache) MakeKey(repo, chart, version string) string {
	return fmt.Sprintf("%s/%s/%s", repo, chart, version)
}
