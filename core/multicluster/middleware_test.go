package multicluster_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/weaveworks/weave-gitops/core/multicluster"
	"github.com/weaveworks/weave-gitops/core/multicluster/multiclusterfakes"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

func TestWithClustersClientsMiddleware(t *testing.T) {
	cluster := makeLeafCluster(t)
	clustersFetcher := &multiclusterfakes.FakeClustersFetcher{}
	clustersFetcher.FetchReturns([]multicluster.Cluster{cluster}, nil)

	g := NewGomegaWithT(t)

	defaultHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientsPool := multicluster.ClientsPoolFromCtx(r.Context())

			clients := clientsPool.Clients()
			if _, ok := clients[cluster.Name]; !ok {
				g.Fail("leaf cluster client not present")
			}

			next.ServeHTTP(w, r)
		})
	}(defaultHandler)

	middleware = multicluster.WithClustersClients(clustersFetcher, middleware)
	middleware = authMiddleware(middleware)

	req := httptest.NewRequest(http.MethodGet, "http://www.foo.com/", nil)
	res := httptest.NewRecorder()
	middleware.ServeHTTP(res, req)

	g.Expect(res).To(HaveHTTPStatus(http.StatusOK))
}

func TestWithClustersClientsMiddlewareFailsToFetchCluster(t *testing.T) {
	defaultHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	clustersFetcher := &multiclusterfakes.FakeClustersFetcher{}
	clustersFetcher.FetchReturns(nil, errors.New("error"))

	middleware := multicluster.WithClustersClients(clustersFetcher, defaultHandler)
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
