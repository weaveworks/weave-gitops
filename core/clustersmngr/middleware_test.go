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

func TestWithClustersClientMiddleware(t *testing.T) {
	cluster := makeLeafCluster(t)
	clustersFetcher := &clustersmngrfakes.FakeClusterFetcher{}
	clustersFetcher.FetchReturns([]clustersmngr.Cluster{cluster}, nil)
	clientsFactory := &clustersmngrfakes.FakeClientsFactory{}

	g := NewGomegaWithT(t)

	defaultHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clustersClient := clustersmngr.ClientFromCtx(r.Context())

			g.Expect(clustersClient.ClientsPool().Clients()).To(HaveKey(cluster.Name))

			next.ServeHTTP(w, r)
		})
	}(defaultHandler)

	middleware = clustersmngr.WithClustersClient(clientsFactory, clustersFetcher, middleware)
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
	clientsFactory := &clustersmngrfakes.FakeClientsFactory{}

	middleware := clustersmngr.WithClustersClient(clientsFactory, clustersFetcher, defaultHandler)
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
