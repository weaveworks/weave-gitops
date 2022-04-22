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
	v1 "k8s.io/api/core/v1"
)

func TestWithClustersClientMiddleware(t *testing.T) {
	g := NewGomegaWithT(t)

	cluster := makeLeafCluster(t)
	clustersFetcher := &clustersmngrfakes.FakeClusterFetcher{}
	clustersFetcher.FetchReturns([]clustersmngr.Cluster{cluster}, nil)

	clientsFactory := &clustersmngrfakes.FakeClientsFactory{}

	clientsPool := clustersmngr.NewClustersClientsPool()
	g.Expect(clientsPool.Add(clustersmngr.ClientConfigWithUser(&auth.UserPrincipal{}), cluster)).To(Succeed())

	client := clustersmngr.NewClient(clientsPool, map[string][]v1.Namespace{})
	clientsFactory.GetUserClientReturns(client, nil)

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

func TestWithClustersClientsMiddlewareFailsToGetUserClient(t *testing.T) {
	defaultHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	clustersFetcher := &clustersmngrfakes.FakeClusterFetcher{}
	clientsFactory := &clustersmngrfakes.FakeClientsFactory{}
	clientsFactory.GetUserClientReturns(nil, errors.New("error"))

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
