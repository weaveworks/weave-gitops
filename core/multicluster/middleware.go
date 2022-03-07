package multicluster

import (
	"context"
	"fmt"
	"net/http"

	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"k8s.io/client-go/rest"
)

// WithClustersClients creates clusters client for provided user in the context
func WithClustersClients(hubRestConfig *rest.Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := auth.Principal(r.Context())
		if user == nil {
			http.Error(w, "Failed creating clusters clients. No user authenticated", http.StatusUnauthorized)
			return
		}

		clustersFetcher, err := NewConfigMapClustersFetcher(hubRestConfig, "flux-system")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "failed getting hub cluster client: %s", err)
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
			if err := clientsPool.Add(user, c); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "failed adding cluster client to the pool: %s", err)
				return
			}
		}

		ctx := context.WithValue(r.Context(), ClustersClientsPoolCtxKey, clientsPool)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func WithPrincipal(ctx context.Context, cp ClientsPool) context.Context {
	return context.WithValue(ctx, ClustersClientsPoolCtxKey, cp)
}

func ClientsPoolFromCtx(ctx context.Context) ClientsPool {
	pool, ok := ctx.Value(ClustersClientsPoolCtxKey).(*clientsPool)
	if ok {
		return pool
	}

	return nil
}
