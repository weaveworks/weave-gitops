package server_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"time"

	"github.com/go-logr/logr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/common/pkg/testutils"
	"github.com/weaveworks/weave-gitops/common/pkg/vendorfakes/fakelogr"
	"github.com/weaveworks/weave-gitops/gitops-server/pkg/server/middleware"
	"k8s.io/apimachinery/pkg/util/rand"
)

var _ = Describe("ApplicationsServer", func() {
	Describe("middleware", func() {
		Describe("logging", func() {
			var (
				sink *fakelogr.LogSink
				mux  *runtime.ServeMux
				ts   *httptest.Server
			)

			BeforeEach(func() {
				rand.Seed(time.Now().UnixNano())
			})

			JustBeforeEach(func() {
				var log logr.Logger

				log, sink = testutils.MakeFakeLogr()

				mux = runtime.NewServeMux(middleware.WithGrpcErrorLogging(log))
				httpHandler := middleware.WithLogging(log, mux)

				ts = httptest.NewServer(httpHandler)
			})

			AfterEach(func() {
				ts.Close()
			})

			It("logs invalid requests", func() {
				// Test a 404 here
				path := "/foo"
				url := ts.URL + path

				res, err := http.Get(url)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.StatusCode).To(Equal(http.StatusNotFound))

				Expect(sink.InfoCallCount()).To(BeNumerically(">", 0))
				vals := sink.WithValuesArgsForCall(0)

				expectedStatus := strconv.Itoa(res.StatusCode)

				list := formatLogVals(vals)
				Expect(list).To(ConsistOf("uri", path, "status", expectedStatus))
			})

			It("logs server errors", func() {
				path := "/v1/applications/parse_repo_url"
				url := ts.URL + path

				res, err := http.Get(url)
				// err is still nil even if we get a 5XX.
				Expect(err).NotTo(HaveOccurred())
				Expect(res.StatusCode).To(Equal(http.StatusNotImplemented))

				Expect(sink.ErrorCallCount()).To(BeNumerically(">", 0))
				vals := sink.WithValuesArgsForCall(0)
				list := formatLogVals(vals)

				expectedStatus := strconv.Itoa(res.StatusCode)
				Expect(list).To(ConsistOf("uri", path, "status", expectedStatus))

				err, msg, _ := sink.ErrorArgsForCall(0)
				// This is the meat of this test case.
				// Check that the same error passed by kubeClient is logged.
				Expect(msg).To(Equal(middleware.ServerErrorText))
				Expect(err.Error()).To(ContainSubstring("ParseRepoURL not implemented"))
			})

			It("logs ok requests", func() {
				// A valid URL for our server
				path := "/v1/applications/parse_repo_url?url=https://github.com/user/repo.git"
				url := ts.URL + path

				res, err := http.Get(url)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.StatusCode).To(Equal(http.StatusOK))

				Expect(sink.InfoCallCount()).To(BeNumerically(">", 0))
				_, msg, _ := sink.InfoArgsForCall(0)
				Expect(msg).To(ContainSubstring(middleware.RequestOkText))

				vals := sink.WithValuesArgsForCall(0)
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
