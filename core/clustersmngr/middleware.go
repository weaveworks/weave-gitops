package clustersmngr

import (
	"context"
	"fmt"
	"net/http"

	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

// WithClustersClient creates clusters client for provided user in the context
func WithClustersClient(clustersFetcher ClusterFetcher, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := auth.Principal(r.Context())
		if user == nil {
			next.ServeHTTP(w, r)
			return
		}

		clusters, err := clustersFetcher.Fetch(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "failed fetching clusters list: %w", err)
			return
		}

		clientsPool := NewClustersClientsPool()
		for _, c := range clusters {
			if err := clientsPool.Add(ClientConfigWithUser(user), c); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "failed adding cluster client to the pool: %s", err)
				return
			}
		}

		clustersClient := NewClient(clientsPool)

		ctx := context.WithValue(r.Context(), ClustersClientCtxKey, clustersClient)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ClientFromCtx returns the ClusterClient stored in the context
func ClientFromCtx(ctx context.Context) Client {
	client, ok := ctx.Value(ClustersClientCtxKey).(*clustersClient)
	if ok {
		return client
	}

	return nil
}
