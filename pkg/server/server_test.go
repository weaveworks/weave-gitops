package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/weaveworks/weave-gitops/pkg/services/auth/authfakes"

	"google.golang.org/grpc/codes"

	"github.com/weaveworks/weave-gitops/pkg/services/auth"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/go-logr/logr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	pb "github.com/weaveworks/weave-gitops/pkg/api/applications"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders/gitprovidersfakes"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/middleware"
	"github.com/weaveworks/weave-gitops/pkg/services/app"
	fakelogr "github.com/weaveworks/weave-gitops/pkg/vendorfakes/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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

		res, err := appsClient.ListApplications(context.Background(), &pb.ListApplicationsRequest{})

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
		res, err := appsClient.GetApplication(context.Background(), &pb.GetApplicationRequest{
			Name:      name,
			Namespace: namespace.Name,
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(res.Application.Name).To(Equal(name))
	})
	It("Authorize", func() {
		ctx := context.Background()
		provider := "github"
		token := "token"

		jwtClient := auth.NewJwtClient()
		expectedToken, err := jwtClient.GenerateJWT(auth.SecretKey, auth.ExpirationTime, gitproviders.GitProviderGitHub, token)
		Expect(err).NotTo(HaveOccurred())

		res, err := appsClient.Authenticate(ctx, &pb.AuthenticateRequest{
			ProviderName: provider,
			AccessToken:  token,
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(res.Token).To(Equal(expectedToken))
	})
	It("Authorize fails on wrong provider", func() {
		ctx := context.Background()
		provider := "wrong_provider"
		token := "token"

		_, err := appsClient.Authenticate(ctx, &pb.AuthenticateRequest{
			ProviderName: provider,
			AccessToken:  token,
		})

		Expect(err.Error()).To(ContainSubstring(BadErrProvider.Error()))
		Expect(err.Error()).To(ContainSubstring(codes.InvalidArgument.String()))

	})

	It("Authorize fails on empty provider token", func() {
		ctx := context.Background()
		provider := "github"

		_, err := appsClient.Authenticate(ctx, &pb.AuthenticateRequest{
			ProviderName: provider,
			AccessToken:  "",
		})

		Expect(err.Error()).To(ContainSubstring(ErrEmptyAccessToken.Error()))
		Expect(err.Error()).To(ContainSubstring(codes.InvalidArgument.String()))

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
				log = makeFakeLogr()
				kubeClient = &kubefakes.FakeKube{}
				appsSrv = NewApplicationsServer(&ApplicationConfig{App: app.New(nil, nil, nil, kubeClient, nil), JwtClient: auth.NewJwtClient()})
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
			It("Authorize fails generating jwt token", func() {

				fakeJWTToken := &authfakes.FakeJWTClient{}
				fakeJWTToken.GenerateJWTStub = func(s2 string, duration time.Duration, name gitproviders.GitProviderName, s22 string) (string, error) {
					return "", fmt.Errorf("some error")
				}

				appsSrv = NewApplicationsServer(&ApplicationConfig{App: app.New(nil, nil, nil, kubeClient, nil), JwtClient: fakeJWTToken})
				mux = runtime.NewServeMux(middleware.WithGrpcErrorLogging(log))
				httpHandler = middleware.WithLogging(log, mux)
				err = pb.RegisterApplicationsHandlerServer(context.Background(), mux, appsSrv)
				Expect(err).NotTo(HaveOccurred())

				ts := httptest.NewServer(httpHandler)
				defer ts.Close()

				// A valid URL for our server
				path := "/v1/authenticate/github"
				url := ts.URL + path

				res, err := http.Post(url, "application/json", strings.NewReader("{\"accessToken\":\"sometoken\"}"))
				Expect(err).NotTo(HaveOccurred())
				Expect(res.StatusCode).To(Equal(http.StatusInternalServerError))

				response := make(map[string]interface{}, 0)
				err = json.NewDecoder(res.Body).Decode(&response)
				Expect(err).NotTo(HaveOccurred())

				Expect(response["message"]).To(Equal("error generating jwt token. some error"))

				Expect(log.InfoCallCount()).To(BeNumerically(">", 0))
				msg, _ := log.InfoArgsForCall(0)
				Expect(msg).To(ContainSubstring(middleware.ServerErrorText))

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

		cfg := ApplicationConfig{
			App:    app.New(nil, nil, nil, k, nil),
			Logger: log,
		}

		handler, err := NewApplicationsHandler(context.Background(), &cfg)
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

		r := &pb.ListApplicationsResponse{}
		err = json.Unmarshal(b, r)
		Expect(err).NotTo(HaveOccurred())

		Expect(r.Applications).To(HaveLen(1))
	})

	It("get commits", func() {
		log := makeFakeLogr()
		kubeClient := &kubefakes.FakeKube{}
		kubeClient.GetApplicationStub = func(context.Context, types.NamespacedName) (*wego.Application, error) {
			return &wego.Application{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-app",
					Namespace: "wego-system",
				},
				Spec: wego.ApplicationSpec{
					Branch: "main",
					Path:   "./k8s",
					URL:    "https://github.com/owner/repo1",
				},
			}, nil
		}

		appSrv := app.New(logger.NewCLILogger(os.Stderr), nil, nil, kubeClient, nil)
		gitProviders := &gitprovidersfakes.FakeGitProvider{}
		commits := []gitprovider.Commit{&fakeCommit{}}

		gitProviders.GetCommitsFromUserRepoStub = func(gitprovider.UserRepositoryRef, string, int, int) ([]gitprovider.Commit, error) {
			return commits, nil
		}

		gitProviders.GetAccountTypeStub = func(string) (gitproviders.ProviderAccountType, error) {
			return gitproviders.AccountTypeUser, nil
		}

		appSrv.GitProviderFactory = func(token string) (gitproviders.GitProvider, error) {
			return gitProviders, nil
		}

		cfg := ApplicationConfig{
			Logger: log,
			App:    appSrv,
		}

		handler, err := NewApplicationsHandler(context.Background(), &cfg)
		Expect(err).NotTo(HaveOccurred())

		ts := httptest.NewServer(handler)
		defer ts.Close()

		path := "/v1/applications/testapp/commits"
		url := ts.URL + path

		res, err := http.Get(url)
		Expect(err).NotTo(HaveOccurred())

		Expect(res.StatusCode).To(Equal(http.StatusOK))

		b, err := ioutil.ReadAll(res.Body)
		Expect(err).NotTo(HaveOccurred())

		r := &pb.ListCommitsResponse{}
		err = json.Unmarshal(b, r)
		Expect(err).NotTo(HaveOccurred())

		Expect(r.Commits).To(HaveLen(1))
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

type fakeCommit struct {
	commitInfo gitprovider.CommitInfo
}

func (fc *fakeCommit) APIObject() interface{} {
	return &fc.commitInfo
}

func (fc *fakeCommit) Get() gitprovider.CommitInfo {
	return testCommit()
}

func testCommit() gitprovider.CommitInfo {
	return gitprovider.CommitInfo{
		Sha:       "testsha",
		Author:    "testauthor",
		Message:   "some awesome commit",
		CreatedAt: time.Now(),
	}
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
