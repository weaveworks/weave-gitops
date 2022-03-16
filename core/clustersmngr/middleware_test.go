package clustersmngr_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/clustersmngrfakes"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

func TestWithClustersClientsMiddleware(t *testing.T) {
	cluster := makeLeafCluster(t)
	clustersFetcher := &clustersmngrfakes.FakeClusterFetcher{}
	clustersFetcher.FetchReturns([]clustersmngr.Cluster{cluster}, nil)

	g := NewGomegaWithT(t)

	defaultHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientsPool := clustersmngr.ClientsPoolFromCtx(r.Context())

			g.Expect(clientsPool.Clients()).To(HaveKey(cluster.Name))

			next.ServeHTTP(w, r)
		})
	}(defaultHandler)

	middleware = clustersmngr.WithClustersClients(clustersFetcher, middleware)
	middleware = authMiddleware(middleware)

	req := httptest.NewRequest(http.MethodGet, "http://www.foo.com/", nil)
	res := httptest.NewRecorder()
	middleware.ServeHTTP(res, req)

	g.Expect(res).To(HaveHTTPStatus(http.StatusOK))
}

func TestWithClustersClientsMiddlewareFailsToFetchCluster(t *testing.T) {
	defaultHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	clustersFetcher := &clustersmngrfakes.FakeClusterFetcher{}
	clustersFetcher.FetchReturns(nil, errors.New("error"))

	middleware := clustersmngr.WithClustersClients(clustersFetcher, defaultHandler)
	middleware = authMiddleware(middleware)

	req := httptest.NewRequest(http.MethodGet, "http://www.foo.com/", nil)
	res := httptest.NewRecorder()
	middleware.ServeHTTP(res, req)

	g := NewGomegaWithT(t)

	g.Expect(res).To(HaveHTTPStatus(http.StatusInternalServerError))
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(auth.WithPrincipal(r.Context(), &auth.UserPrincipal{ID: "user@weave.gitops", Groups: []string{"developers"}})))
	})
}
