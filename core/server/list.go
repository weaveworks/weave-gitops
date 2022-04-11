package server

import (
	"context"
	"fmt"

	"github.com/weaveworks/weave-gitops/core/clustersmngr"
)

// listFn is a function that contains the logic how to list objects
// from a single namespace.
type listFn func(context.Context, clustersmngr.Client, string) ([]interface{}, error)

// listObejcts lists objects across all namespaces.
//
// This function can be simplified with generics once we move to Go 1.18,
// but for now, it's a weird interface{} magic.
func (cs *coreServer) listObjects(
	ctx context.Context,
	namespace string,
	list listFn,
) ([]interface{}, error) {
	clustersClient := clustersmngr.ClientFromCtx(ctx)

	if namespace != "" {
		res, err := list(ctx, clustersClient, namespace)

		return res, err
	}

	var results []interface{}

	nsList, found := cs.cacheContainer.Namespaces()[clustersmngr.DefaultCluster]
	if !found {
		return nil, defaultClusterNotFound{}
	}

	for _, ns := range nsList {
		nsResult, err := list(ctx, clustersClient, ns.Name)
		if err != nil {
			cs.logger.Error(err, fmt.Sprintf("unable to list objects in namespace: %s", ns.Name))

			continue
		}

		results = append(results, nsResult...)
	}

	return results, nil
}
