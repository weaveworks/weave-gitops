package server_test

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"

	"github.com/go-logr/logr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/api/applications"
	appspb "github.com/weaveworks/weave-gitops/pkg/api/applications"
	"github.com/weaveworks/weave-gitops/pkg/api/commits"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"
	"github.com/weaveworks/weave-gitops/pkg/middleware"
	"github.com/weaveworks/weave-gitops/pkg/server"
	fakelogr "github.com/weaveworks/weave-gitops/pkg/vendorfakes/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

var _ = Describe("ApplicationsServer", func() {
	var (
		namespace *corev1.Namespace
		err       error
	)

	BeforeEach(func() {
		namespace = &corev1.Namespace{}
		namespace.Name = "kube-test-" + rand.String(5)
		err = k8sClient.Create(context.Background(), namespace)
		Expect(err).NotTo(HaveOccurred(), "failed to create test namespace")

		k = &kube.KubeHTTP{Client: k8sClient, ClusterName: testClustername}
	})
	It("ListApplication", func() {
		ctx := context.Background()
		name := "my-app"
		app := &wego.Application{ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace.Name,
		}}

		Expect(k8sClient.Create(ctx, app)).Should(Succeed())

		res, err := appsClient.ListApplications(context.Background(), &applications.ListApplicationsRequest{})

		Expect(err).NotTo(HaveOccurred())

		Expect(len(res.Applications)).To(Equal(1))
	})
	It("GetApplication", func() {
		ctx := context.Background()
		name := "my-app"
		app := &wego.Application{ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace.Name,
		}}

		Expect(k8sClient.Create(ctx, app)).Should(Succeed())
		res, err := appsClient.GetApplication(context.Background(), &applications.GetApplicationRequest{
			Name:      name,
			Namespace: namespace.Name,
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(res.Application.Name).To(Equal(name))
	})

	Describe("middleware", func() {
		Describe("logging", func() {
			var log *fakelogr.FakeLogger
			var kubeClient *kubefakes.FakeKube
			var appsSrv appspb.ApplicationsServer
			var mux *runtime.ServeMux
			var httpHandler http.Handler
			var err error

			BeforeEach(func() {
				log = makeFakeLogr()
				kubeClient = &kubefakes.FakeKube{}
				appsSrv = server.NewApplicationsServer(&server.ApplicationsConfig{KubeClient: kubeClient})
				mux = runtime.NewServeMux(middleware.WithGrpcErrorLogging(log))
				httpHandler = middleware.WithLogging(log, mux)
				err = appspb.RegisterApplicationsHandlerServer(context.Background(), mux, appsSrv)
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

var _ = Describe("Applications handler", func() {
	It("works as a standalone handler", func() {
		log := makeFakeLogr()
		k := &kubefakes.FakeKube{}
		k.GetApplicationsStub = func(c context.Context, s string) ([]wego.Application, error) {
			return []wego.Application{{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-app",
					Namespace: "wego-system",
				},
				Spec: wego.ApplicationSpec{
					Branch: "main",
					Path:   "./k8s",
				},
			}}, nil
		}

		cfg := server.ApplicationsConfig{
			KubeClient: k,
			Logger:     log,
		}

		handler, err := server.NewApplicationsHandler(context.Background(), &cfg)
		Expect(err).NotTo(HaveOccurred())

		ts := httptest.NewServer(handler)
		defer ts.Close()

		path := "/v1/applications"
		url := ts.URL + path

		res, err := http.Get(url)
		Expect(err).NotTo(HaveOccurred())

		Expect(res.StatusCode).To(Equal(http.StatusOK))

		b, err := ioutil.ReadAll(res.Body)
		Expect(err).NotTo(HaveOccurred())

		r := &appspb.ListApplicationsResponse{}
		err = json.Unmarshal(b, r)
		Expect(err).NotTo(HaveOccurred())

		Expect(r.Applications).To(HaveLen(1))
	})
})

var _ = Describe("CommitsServer", func() {
	var (
		namespace *corev1.Namespace
		err       error
	)

	BeforeEach(func() {
		namespace = &corev1.Namespace{}
		namespace.Name = "kube-test-" + rand.String(5)
		err = k8sClient.Create(context.Background(), namespace)
		Expect(err).NotTo(HaveOccurred(), "failed to create test namespace")

		k = &kube.KubeHTTP{Client: k8sClient, ClusterName: testClustername}
	})
	It("ListCommits", func() {
		ctx := context.Background()
		name := "my-app"
		app := &wego.Application{ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace.Name,
		}}

		Expect(k8sClient.Create(ctx, app)).Should(Succeed())

		res, err := commitClient.ListCommits(context.Background(), &commits.ListCommitsRequest{})

		Expect(err).NotTo(HaveOccurred())

		Expect(len(res.Commits)).To(Equal(1))
	})
})

func makeFakeLogr() *fakelogr.FakeLogger {
	log := &fakelogr.FakeLogger{}
	log.WithValuesStub = func(i ...interface{}) logr.Logger {
		return log
	}
	log.VStub = func(i int) logr.Logger {
		return log
	}
	return log
}

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
