package cache

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const pollIntervalSeconds = 120

type namespaceStore struct {
	client       client.Client
	namespaces   []v1.Namespace
	logger       logr.Logger
	forceRefresh chan bool
	cancel       func()
}

func newNamespaceStore(c client.Client, logger logr.Logger) namespaceStore {
	return namespaceStore{
		client:       c,
		namespaces:   []v1.Namespace{},
		logger:       logger,
		cancel:       nil,
		forceRefresh: make(chan bool),
	}
}

func (n *namespaceStore) Namespaces() []v1.Namespace {
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
			list := &v1.NamespaceList{}

			err := n.client.List(newCtx, list)
			if err != nil {
				n.logger.Error(err, "poll error")
			}

			newList := []v1.Namespace{}
			newList = append(newList, list.Items...)

			n.namespaces = newList

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
