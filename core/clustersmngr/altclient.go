package clustersmngr

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewAltClient() (Client, error) {

	return &altClient{}, nil
}

type altClient struct {
	Client
}

func (ac *altClient) ClusteredList(ctx context.Context, clist ClusteredObjectList, namespaced bool, opts ...client.ListOption) error {
	// build a client from values extracted from ctx or read from cache
	// cache the resulting client (?)
	// Q: should we do this on login?
	// Q: can we take the user to some special loading page while we build their client?
	// - imagining a special button they click on the login screen that navigates to a page to trigger this work.
	// - show progress
	// - bust the client cache
	// - guard against retries that will saturate the api server
	// - mutex to lock the go routine that gets cleaned up
	// - endpoint to hit to generate client
	// - endpoint to give client generation status; SSE?

	// get a list of clusters

	// for each cluster
	// get a list of namespaces
	// do selfsubjectacccessreviews and store the available namespaces in the cached client

	// for each namespace
	// list objects
	// our cached client might get a 403 if the user's namespace access has been revoked,
	// but we can just ignore the 403 and update the available namespaces in the stored client.

	return nil
}
