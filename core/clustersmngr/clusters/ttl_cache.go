package clusters

import (
	"fmt"
	"strings"

	"github.com/cheshir/ttlcache"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ttlCache struct {
	cluster       Cluster
	usersClients  *usersClients
	serverClients *usersClients
}

func NewTTLCacheFetcher(cluster Cluster) Cluster {
	return &ttlCache{
		cluster:       cluster,
		usersClients:  &usersClients{Cache: ttlcache.New(usersClientResolution)},
		serverClients: &usersClients{Cache: ttlcache.New(usersClientResolution)},
	}
}

func (c *ttlCache) GetName() string {
	return c.cluster.GetName()
}

func (c *ttlCache) GetHost() string {
	return c.cluster.GetHost()
}

func (c *ttlCache) GetUserClient(user *auth.UserPrincipal) (client.Client, error) {
	if client, found := c.usersClients.Get(user, c.GetName()); found {
		return client, nil
	}
	client, err := c.cluster.GetUserClient(user)
	if err != nil {
		return nil, fmt.Errorf("failed creating client for cluster=%s: %w", c.GetName(), err)
	}

	c.usersClients.Set(user, c.GetName(), client)

	return client, nil
}

func (c *ttlCache) GetServerClient() (client.Client, error) {
	key := &auth.UserPrincipal{}
	if client, found := c.serverClients.Get(key, c.GetName()); found {
		return client, nil
	}
	client, err := c.cluster.GetServerClient()
	if err != nil {
		return nil, fmt.Errorf("failed creating client for cluster=%s: %w", c.GetName(), err)
	}

	c.serverClients.Set(key, c.GetName(), client)

	return client, nil
}

func (c *ttlCache) GetUserClientset(user *auth.UserPrincipal) (kubernetes.Interface, error) {
	return c.cluster.GetUserClientset(user)
}

func (c *ttlCache) GetServerClientset() (kubernetes.Interface, error) {
	return c.cluster.GetServerClientset()
}

type usersClients struct {
	Cache *ttlcache.Cache
}

func (uc *usersClients) cacheKey(user *auth.UserPrincipal, clusterName string) uint64 {
	return ttlcache.StringKey(fmt.Sprintf("%s:%s-%s", user.ID, strings.Join(user.Groups, "/"), clusterName))
}

func (uc *usersClients) Set(user *auth.UserPrincipal, clusterName string, client client.Client) {
	uc.Cache.Set(uc.cacheKey(user, clusterName), client, usersClientsTTL)
}

func (uc *usersClients) Get(user *auth.UserPrincipal, clusterName string) (client.Client, bool) {
	if val, found := uc.Cache.Get(uc.cacheKey(user, clusterName)); found {
		return val.(client.Client), true
	}

	return nil, false
}

func (uc *usersClients) Clear() {
	uc.Cache.Clear()
}
