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
	Add(key string, value Data)
	// Get always returns everything from the cache.
	Get(key string) *Data
	// I will need a GetValues as well, because that works differently.
	Key(namespace, name string) string
}

// Data is explicit data for a specific profile including values.
type Data struct {
	Profiles []*pb.Profile
	// TODO: This should be a map based on name and version of the profiles that this data contains for easy access.
	// How will I store this? ( path revision something something ).
	// ProfileName / ProfileVersion -> values
	Values map[string]map[string][]byte // TODO store on disk because it can be rather ram intensive
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

// Key defines the way a key for a specific data is generated.
// format: Repo/Chart/Version
func (c *HelmCache) Key(namespace, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}
