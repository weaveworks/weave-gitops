package clustersmngr

import (
	"context"
	"fmt"
	"net/http"

	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

type clientFactoryAndUser struct {
	clientsFactory ClientsFactory
	user           *auth.UserPrincipal
}

// WithClustersClient creates clusters client for provided user in the context
func WithClustersClient(clientsFactory ClientsFactory, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := auth.Principal(r.Context())
		if user == nil {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), ClustersClientCtxKey, &clientFactoryAndUser{
			clientsFactory: clientsFactory,
			user:           user,
		})

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ClientFromCtx returns the ClusterClient stored in the context
func ClientFromCtx(ctx context.Context) Client {
	cu, ok := ctx.Value(ClustersClientCtxKey).(*clientFactoryAndUser)
	if !ok {
		fmt.Println("not ok!")
		return nil
	}

	client, err := cu.clientsFactory.GetImpersonatedClient(ctx, cu.user)
	if err != nil {
		fmt.Printf("error getting client, %v\n", err)
		return nil
	}

	return client

}
