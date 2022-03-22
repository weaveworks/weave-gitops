package cache

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const pollInteralSeconds = 300

type namespaceStore struct {
	client     client.Client
	namespaces []v1.Namespace
}

func newNamespaceStore(c client.Client) *namespaceStore {
	return &namespaceStore{
		client:     c,
		namespaces: []v1.Namespace{},
	}
}

func (n *namespaceStore) Namespaces() []v1.Namespace {
	return n.namespaces
}

func (n *namespaceStore) Start(ctx context.Context) {
	go func() {
		for ; true; <-time.Tick(pollInteralSeconds * time.Second) {
			list := &v1.NamespaceList{}

			err := n.client.List(ctx, list)
			if err != nil {
				logrus.Infof("poll error: %s", err.Error())
			}

			newList := []v1.Namespace{}
			newList = append(newList, list.Items...)

			n.namespaces = newList
		}
	}()
}
