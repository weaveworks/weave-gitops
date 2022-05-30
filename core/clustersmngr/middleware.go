package clustersmngr

import (
	"context"
	"fmt"
	"net/http"

	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

// WithClustersClient creates clusters client for provided user in the context
func WithClustersClient(clientsFactory ClientsFactory, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := auth.Principal(r.Context())
		if user == nil {
			next.ServeHTTP(w, r)
			return
		}

		client, err := clientsFactory.GetImpersonatedClient(r.Context(), user)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "failed getting impersonated client:", err)
			return
		}

		ctx := context.WithValue(r.Context(), ClustersClientCtxKey, client)

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
