package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"

	"github.com/go-logr/logr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	. "github.com/onsi/ginkgo"
	"github.com/weaveworks/weave-gitops/pkg/api/ping"
	"github.com/weaveworks/weave-gitops/pkg/middleware"
	fakelogr "github.com/weaveworks/weave-gitops/pkg/vendorfakes/logr"

	. "github.com/onsi/gomega"
)

var _ = Describe("middleware", func() {
	Describe("logging", func() {
		var log *fakelogr.FakeLogger
		var pingSrv ping.TestServiceServer

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

			pingSrv = &TestPingService{}
			mux = runtime.NewServeMux(middleware.WithGrpcErrorLogging(log))
			httpHandler = middleware.WithLogging(log, mux)
			err = ping.RegisterTestServiceHandlerServer(context.Background(), mux, pingSrv)
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

			path := "/v1/ping/error"
			url := ts.URL + path

			res, err := http.Post(url, "application/json", nil)
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
			Expect(err.Error()).To(Equal("fooo"))
			Expect(msg).To(Equal(middleware.ServerErrorText))
		})
		It("logs ok requests", func() {
			ts := httptest.NewServer(httpHandler)
			defer ts.Close()

			// A valid URL for our server
			path := "/v1/ping/empty"
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
