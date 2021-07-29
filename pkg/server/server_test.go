package server_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"

	"github.com/go-logr/logr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/api/applications"
	pb "github.com/weaveworks/weave-gitops/pkg/api/applications"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"
	"github.com/weaveworks/weave-gitops/pkg/middleware"
	"github.com/weaveworks/weave-gitops/pkg/server"
	fakelogr "github.com/weaveworks/weave-gitops/pkg/vendorfakes/logr"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("ApplicationsServer", func() {
	It("AddApplication", func() {
		kubeClient.GetApplicationsStub = func(ctx context.Context, ns string) ([]wego.Application, error) {
			return []wego.Application{
				{
					ObjectMeta: v1.ObjectMeta{Name: "my-app"},
					Spec:       wego.ApplicationSpec{Path: "bar"},
				},
				{
					ObjectMeta: v1.ObjectMeta{Name: "my-app1"},
					Spec:       wego.ApplicationSpec{Path: "bar2"},
				},
			}, nil
		}

		res, err := client.ListApplications(context.Background(), &applications.ListApplicationsRequest{})

		Expect(err).NotTo(HaveOccurred())

		Expect(len(res.Applications)).To(Equal(2))
	})
	It("GetApplication", func() {
		kubeClient.GetApplicationStub = func(ctx context.Context, name types.NamespacedName) (*wego.Application, error) {
			return &wego.Application{
				ObjectMeta: v1.ObjectMeta{Name: "my-app"},
				Spec:       wego.ApplicationSpec{Path: "bar"},
			}, nil
		}

		res, err := client.GetApplication(context.Background(), &applications.GetApplicationRequest{Name: "my-app"})
		Expect(err).NotTo(HaveOccurred())

		Expect(res.Application.Name).To(Equal("my-app"))
	})
	Describe("middleware", func() {
		Describe("logging", func() {
			var log *fakelogr.FakeLogger
			var kubeClient *kubefakes.FakeKube
			var appsSrv pb.ApplicationsServer
			var mux *runtime.ServeMux
			var httpHandler http.Handler
			var err error

			BeforeEach(func() {
				log = &fakelogr.FakeLogger{}
				log.WithValuesStub = func(i ...interface{}) logr.Logger {
					return log
				}

				log.VStub = func(i int) logr.Logger {
					return log
				}

				kubeClient = &kubefakes.FakeKube{}
				appsSrv = server.NewApplicationsServer(kubeClient)
				mux = runtime.NewServeMux(middleware.WithGrpcErrorLogging(log))
				httpHandler = middleware.WithLogging(log, mux)
				err = pb.RegisterApplicationsHandlerServer(context.Background(), mux, appsSrv)
				Expect(err).NotTo(HaveOccurred())
			})
			It("logs invalid requests", func() {
				ts := httptest.NewServer(httpHandler)
				defer ts.Close()

				// Test a 404 here
				path := "/foo"
				url := ts.URL + path

				res, err := http.Get(url)
				Expect(res.StatusCode).To(Equal(http.StatusNotFound))

				Expect(err).NotTo(HaveOccurred())
				Expect(log.InfoCallCount()).To(BeNumerically(">", 0))
				vals := log.WithValuesArgsForCall(0)

				expectedStatus := strconv.Itoa(res.StatusCode)

				list := formatLogVals(vals)
				Expect(list).To(ConsistOf("uri", path, "status", expectedStatus))

			})
			It("logs server errors", func() {
				ts := httptest.NewServer(httpHandler)
				defer ts.Close()

				errMsg := "there was a big problem"

				// Pretend something went horribly wrong
				kubeClient.GetApplicationsStub = func(c context.Context, s string) ([]wego.Application, error) {
					return nil, errors.New(errMsg)
				}

				path := "/v1/applications"
				url := ts.URL + path

				res, err := http.Get(url)
				// err is still nil even if we get a 5XX.
				Expect(err).NotTo(HaveOccurred())
				Expect(res.StatusCode).To(Equal(http.StatusInternalServerError))

				Expect(log.ErrorCallCount()).To(BeNumerically(">", 0))
				vals := log.WithValuesArgsForCall(0)
				list := formatLogVals(vals)

				expectedStatus := strconv.Itoa(res.StatusCode)
				Expect(list).To(ConsistOf("uri", path, "status", expectedStatus))

				err, msg, _ := log.ErrorArgsForCall(0)
				// This is the meat of this test case.
				// Check that the same error passed by kubeClient is logged.
				Expect(err.Error()).To(Equal(errMsg))
				Expect(msg).To(Equal(middleware.ServerErrorText))

			})
			It("logs ok requests", func() {
				ts := httptest.NewServer(httpHandler)
				defer ts.Close()

				// A valid URL for our server
				path := "/v1/applications"
				url := ts.URL + path

				res, err := http.Get(url)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.StatusCode).To(Equal(http.StatusOK))

				Expect(log.InfoCallCount()).To(BeNumerically(">", 0))
				msg, _ := log.InfoArgsForCall(0)
				Expect(msg).To(ContainSubstring(middleware.RequestOkText))

				vals := log.WithValuesArgsForCall(0)
				list := formatLogVals(vals)

				expectedStatus := strconv.Itoa(res.StatusCode)
				Expect(list).To(ConsistOf("uri", path, "status", expectedStatus))
			})
		})

	})
})

func formatLogVals(vals []interface{}) []string {
	list := []string{}
	for _, v := range vals {
		// vals is a slice of empty interfaces. convert them.
		s, ok := v.(string)
		if !ok {
			// Last value is a status code represented as an int
			n := v.(int)
			s = strconv.Itoa(n)
		}
		list = append(list, s)
	}
	return list
}
