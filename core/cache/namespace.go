package cache

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	v1 "k8s.io/api/core/v1"
)

const pollIntervalSeconds = 120

type namespaceStore struct {
	client       clustersmngr.Client
	namespaces   map[string][]v1.Namespace
	logger       logr.Logger
	forceRefresh chan bool
	cancel       func()
}

func newNamespaceStore(c clustersmngr.Client, logger logr.Logger) namespaceStore {
	return namespaceStore{
		client:       c,
		namespaces:   map[string][]v1.Namespace{},
		logger:       logger,
		cancel:       nil,
		forceRefresh: make(chan bool),
	}
}

func (n *namespaceStore) Namespaces() map[string][]v1.Namespace {
	return n.namespaces
}

func (n *namespaceStore) ForceRefresh() {
	n.forceRefresh <- true
}

func (n *namespaceStore) Stop() {
	if n.cancel != nil {
		n.cancel()
	}
}

func (n *namespaceStore) Start(ctx context.Context) {
	var newCtx context.Context

	newCtx, n.cancel = context.WithCancel(ctx)

	go func() {
		ticker := time.NewTicker(pollIntervalSeconds * time.Second)

		defer ticker.Stop()

		for {
			for name, c := range n.client.ClientsPool().Clients() {
				list := &v1.NamespaceList{}

				err := c.List(newCtx, list)
				if err != nil {
					continue
				}

				newList := []v1.Namespace{}
				newList = append(newList, list.Items...)

				n.namespaces[name] = newList
			}

			select {
			case <-newCtx.Done():
				break
			case <-n.forceRefresh:
				continue
			case <-ticker.C:
				continue
			}
		}
	}()
}
