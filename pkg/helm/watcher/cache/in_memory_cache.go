package cache

import (
	"fmt"
	"sync"

	pb "github.com/weaveworks/weave-gitops/pkg/api/profiles"
)

// Cache defines an interface to work with the profile data cacher.
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . Cache
type Cache interface {
	Add(key string, value Data) error
	// Get always returns everything from the cache.
	Get(key string) (*Data, error)
	Key(helmNamespace, helmName string) string
}

// Data is explicit data for a specific profile including values.
type Data struct {
	Profiles []*pb.Profile
	Values   map[string]map[string][]byte
}

// HelmCache is used to cache profiles data from scanner helm repositories.
// TODO: Turn this into a file based cache.
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
func (c *HelmCache) Add(key string, value Data) error {
	c.mut.Lock()
	defer c.mut.Unlock()
	c.cache[key] = &value

	return nil
}

// Get will return nil in case the data cannot be found but no error.
func (c *HelmCache) Get(key string) (*Data, error) {
	c.mut.RLock()
	defer c.mut.RUnlock()

	return c.cache[key], nil
}

// Key defines the way a key for a specific data is generated.
func (c *HelmCache) Key(helmNamespace, helmName string) string {
	return fmt.Sprintf("%s/%s", helmNamespace, helmName)
}
