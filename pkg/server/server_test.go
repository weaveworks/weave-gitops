package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"time"

	"github.com/weaveworks/weave-gitops/pkg/services/auth/authfakes"
	"github.com/weaveworks/weave-gitops/pkg/testutils"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/weaveworks/weave-gitops/pkg/services/auth"

	"github.com/fluxcd/go-git-providers/gitprovider"
	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	pb "github.com/weaveworks/weave-gitops/pkg/api/applications"
	"github.com/weaveworks/weave-gitops/pkg/apputils/apputilsfakes"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders/gitprovidersfakes"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"
	"github.com/weaveworks/weave-gitops/pkg/middleware"
	"github.com/weaveworks/weave-gitops/pkg/services/app"
	fakelogr "github.com/weaveworks/weave-gitops/pkg/vendorfakes/logr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	appsv1 "k8s.io/api/apps/v1"
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

	Describe("GetApplication", func() {
		var (
			ctx  context.Context
			name string
			app  *wego.Application
		)

		BeforeEach(func() {
			ctx = context.Background()
			name = "my-app-" + rand.String(5)
			app = &wego.Application{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace.Name,
				},
				Spec: wego.ApplicationSpec{
					SourceType: wego.SourceTypeGit,
				},
			}

			Expect(k8sClient.Create(ctx, app)).Should(Succeed())
		})

		AfterEach(func() {
			deletePolicy := metav1.DeletePropagationForeground
			Expect(k8sClient.Delete(ctx, app, &client.DeleteOptions{PropagationPolicy: &deletePolicy})).Should(Succeed())
		})

		It("fetches an application", func() {
			resp, err := appsClient.GetApplication(context.Background(), &pb.GetApplicationRequest{
				Name:      name,
				Namespace: namespace.Name,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(resp.Application.Name).To(Equal(name))
		})

		Describe("fetches the application source", func() {
			It("fetches a git repository", func() {
				git := &sourcev1.GitRepository{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: namespace.Name,
					},
					Spec: sourcev1.GitRepositorySpec{
						URL:       "ssh://my-repo",
						Interval:  metav1.Duration{Duration: 1 * time.Second},
						Timeout:   &metav1.Duration{Duration: 1 * time.Second},
						Reference: &sourcev1.GitRepositoryRef{Branch: "master"},
					},
				}
				Expect(k8sClient.Create(ctx, git)).Should(Succeed())

				resp, err := appsClient.GetApplication(context.Background(), &pb.GetApplicationRequest{
					Name:      name,
					Namespace: namespace.Name,
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(resp.Application.Source.Type).To(Equal(pb.Source_Git))
				Expect(resp.Application.Source.Url).To(Equal("ssh://my-repo"))
				Expect(resp.Application.Source.Interval).To(Equal("1s"))
				Expect(resp.Application.Source.Timeout).To(Equal("1s"))
				Expect(resp.Application.Source.Reference).To(Equal("master"))

				Expect(k8sClient.Delete(ctx, git)).Should(Succeed())
			})

			It("fetches a helm repository", func() {
				name = "my-app-" + rand.String(5)
				app = &wego.Application{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: namespace.Name,
					},
					Spec: wego.ApplicationSpec{
						SourceType: wego.SourceTypeHelm,
					},
				}
				Expect(k8sClient.Create(ctx, app)).Should(Succeed())

				helm := &sourcev1.HelmRepository{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: namespace.Name,
					},
					Spec: sourcev1.HelmRepositorySpec{
						URL:      "http://my-chart",
						Interval: metav1.Duration{Duration: 10 * time.Second},
						Timeout:  &metav1.Duration{Duration: 10 * time.Second},
					},
				}
				Expect(k8sClient.Create(ctx, helm)).Should(Succeed())

				resp, err := appsClient.GetApplication(context.Background(), &pb.GetApplicationRequest{
					Name:      name,
					Namespace: namespace.Name,
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(resp.Application.Source.Name).To(Equal(name))
				Expect(resp.Application.Source.Url).To(Equal("http://my-chart"))
				Expect(resp.Application.Source.Type).To(Equal(pb.Source_Helm))
				Expect(resp.Application.Source.Interval).To(Equal("10s"))
				Expect(resp.Application.Source.Timeout).To(Equal("10s"))

				Expect(k8sClient.Delete(ctx, helm)).Should(Succeed())
			})
		})

		Describe("fetches the application deployment", func() {
			It("fetches a kustomization", func() {
				kust := &kustomizev1.Kustomization{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: namespace.Name,
					},
					Spec: kustomizev1.KustomizationSpec{
						TargetNamespace: "target-namespace",
						Path:            "/path",
						Interval:        metav1.Duration{Duration: 1 * time.Second},
						Prune:           true,
						SourceRef: kustomizev1.CrossNamespaceSourceReference{
							Kind: "GitRepository",
							Name: name,
						},
					},
				}
				Expect(k8sClient.Create(ctx, kust)).Should(Succeed())

				resp, err := appsClient.GetApplication(context.Background(), &pb.GetApplicationRequest{
					Name:      name,
					Namespace: namespace.Name,
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(resp.Application.Kustomization.TargetNamespace).To(Equal("target-namespace"))
				Expect(resp.Application.Kustomization.Path).To(Equal("/path"))
				Expect(resp.Application.Kustomization.Interval).To(Equal("1s"))

				Expect(k8sClient.Delete(ctx, kust)).Should(Succeed())
			})

			It("fetches a helm release", func() {
				name = "my-app-" + rand.String(5)
				app = &wego.Application{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: namespace.Name,
					},
					Spec: wego.ApplicationSpec{
						DeploymentType: wego.DeploymentTypeHelm,
					},
				}
				Expect(k8sClient.Create(ctx, app)).Should(Succeed())

				release := &helmv2.HelmRelease{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: namespace.Name,
					},
					Spec: helmv2.HelmReleaseSpec{
						TargetNamespace: "target-namespace",
						Chart: helmv2.HelmChartTemplate{
							Spec: helmv2.HelmChartTemplateSpec{
								Chart:       "https://my-chart",
								Version:     "v1.2.3",
								ValuesFiles: []string{"file-1.yaml"},
								SourceRef: helmv2.CrossNamespaceObjectReference{
									Kind: "GitRepository",
									Name: name,
								},
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, release)).Should(Succeed())

				resp, err := appsClient.GetApplication(context.Background(), &pb.GetApplicationRequest{
					Name:      name,
					Namespace: namespace.Name,
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(resp.Application.HelmRelease.TargetNamespace).To(Equal("target-namespace"))
				Expect(resp.Application.HelmRelease.Chart.Chart).To(Equal("https://my-chart"))
				Expect(resp.Application.HelmRelease.Chart.Version).To(Equal("v1.2.3"))
				Expect(resp.Application.HelmRelease.Chart.ValuesFiles).To(Equal([]string{"file-1.yaml"}))

				Expect(k8sClient.Delete(ctx, release)).Should(Succeed())
			})
		})

	})

	It("Authorize", func() {
		ctx := context.Background()
		provider := "github"
		token := "token"

		jwtClient := auth.NewJwtClient(secretKey)
		expectedToken, err := jwtClient.GenerateJWT(auth.ExpirationTime, gitproviders.GitProviderGitHub, token)
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

		Expect(err.Error()).To(ContainSubstring(ErrBadProvider.Error()))
		Expect(err.Error()).To(ContainSubstring(codes.InvalidArgument.String()))

	})
	It("Authorize fails on empty provider token", func() {
		ctx := context.Background()
		provider := "github"

		_, err := appsClient.Authenticate(ctx, &pb.AuthenticateRequest{
			ProviderName: provider,
			AccessToken:  "",
		})

		Expect(err).Should(MatchGRPCError(codes.InvalidArgument, ErrEmptyAccessToken))
	})
	Describe("GetReconciledObjects", func() {
		It("gets object with a kustomization + git repo configuration", func() {
			ctx := context.Background()
			name := "my-app"
			kustomization := kustomizev1.Kustomization{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace.Name,
				},
				Spec: kustomizev1.KustomizationSpec{
					SourceRef: kustomizev1.CrossNamespaceSourceReference{
						Kind: sourcev1.GitRepositoryKind,
					},
				},
				Status: kustomizev1.KustomizationStatus{
					Snapshot: &kustomizev1.Snapshot{
						Entries: []kustomizev1.SnapshotEntry{
							{
								Namespace: namespace.Name,
								Kinds: map[string]string{
									namespace.Name: "Deployment",
								},
							},
						},
					},
				},
			}
			reconciledObj := appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-deployment",
					Namespace: namespace.Name,
					Labels: map[string]string{
						KustomizeNameKey:      name,
						KustomizeNamespaceKey: namespace.Name,
					},
				},
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": name,
						},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{"app": name},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{
								Name:  "nginx",
								Image: "nginx",
							}},
						},
					},
				},
			}
			app := &wego.Application{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace.Name,
				},
				Spec: wego.ApplicationSpec{
					DeploymentType: wego.DeploymentTypeKustomize,
				},
			}
			Expect(k8sClient.Create(ctx, &kustomization)).Should(Succeed())
			Expect(k8sClient.Create(ctx, &reconciledObj)).Should(Succeed())
			Expect(k8sClient.Create(ctx, app)).Should(Succeed())
			res, err := appsClient.GetReconciledObjects(ctx, &pb.GetReconciledObjectsReq{
				AutomationName:      name,
				AutomationNamespace: namespace.Name,
				AutomationKind:      pb.AutomationKind_Kustomize,
				Kinds:               []*pb.GroupVersionKind{{Group: "apps", Version: "v1", Kind: "Deployment"}},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(res.Objects).To(HaveLen(1))

			first := res.Objects[0]
			Expect(first.GroupVersionKind.Kind).To(Equal("Deployment"))
			Expect(first.Name).To(Equal(reconciledObj.Name))
		})
		It("returns an error when helm is specified as an automation type", func() {
			ctx := context.Background()
			name := "my-app"
			_, err := appsClient.GetReconciledObjects(ctx, &pb.GetReconciledObjectsReq{
				AutomationName:      name,
				AutomationNamespace: namespace.Name,
				AutomationKind:      pb.AutomationKind_Helm,
				Kinds:               []*pb.GroupVersionKind{{Group: "apps", Version: "v1", Kind: "Deployment"}},
			})

			s, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(s.Code()).To(Equal(codes.Unimplemented))
		})
	})
	Describe("GetChildObjects", func() {
		It("returns child objects for a parent", func() {
			ctx := context.Background()
			name := "my-app"
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-deployment",
					Namespace: namespace.Name,
				},
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": name,
						},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{"app": name},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{
								Name:  "nginx",
								Image: "nginx",
							}},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, deployment)).Should(Succeed())
			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, deployment))
			Expect(deployment.UID).NotTo(Equal(""))
			rs := &appsv1.ReplicaSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-123abcd", name),
					Namespace: namespace.Name,
				},
				Spec: appsv1.ReplicaSetSpec{
					Template: deployment.Spec.Template,
					Selector: deployment.Spec.Selector,
				},
			}
			rs.SetOwnerReferences([]metav1.OwnerReference{{
				UID:        deployment.UID,
				APIVersion: deployment.APIVersion,
				Kind:       deployment.Kind,
				Name:       deployment.Name,
			}})

			Expect(k8sClient.Create(ctx, rs)).Should(Succeed())

			res, err := appsClient.GetChildObjects(ctx, &pb.GetChildObjectsReq{
				ParentUid:        string(deployment.UID),
				GroupVersionKind: &pb.GroupVersionKind{Group: "apps", Version: "v1", Kind: "ReplicaSet"},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(res.Objects).To(HaveLen(1))

			first := res.Objects[0]
			Expect(first.GroupVersionKind.Kind).To(Equal("ReplicaSet"))
			Expect(first.Name).To(Equal(rs.Name))
		})
	})

	Describe("GetGithubDeviceCode", func() {
		It("returns a device code", func() {
			ctx := context.Background()
			code := "123-456"
			ghAuthClient.GetDeviceCodeStub = func() (*auth.GithubDeviceCodeResponse, error) {
				return &auth.GithubDeviceCodeResponse{DeviceCode: code}, nil
			}

			res, err := appsClient.GetGithubDeviceCode(ctx, &pb.GetGithubDeviceCodeRequest{})
			Expect(err).NotTo(HaveOccurred())

			Expect(res.DeviceCode).To(Equal(code))
		})
		// Possibly drop this
		It("returns an error when github returns an error", func() {
			ctx := context.Background()
			someError := errors.New("some gh error")
			ghAuthClient.GetDeviceCodeStub = func() (*auth.GithubDeviceCodeResponse, error) {
				return nil, someError
			}
			_, err := appsClient.GetGithubDeviceCode(ctx, &pb.GetGithubDeviceCodeRequest{})
			Expect(err).To(HaveOccurred())
			st, ok := status.FromError(err)
			Expect(ok).To(BeTrue(), "could not get grpc status from err")
			Expect(st.Message()).To(ContainSubstring(someError.Error()))
		})
	})

	Describe("GetGithubAuthStatus", func() {
		It("returns an ErrAuthPending when the user is not yet authenticated", func() {
			ctx := context.Background()
			ghAuthClient.GetDeviceCodeAuthStatusStub = func(s string) (string, error) {
				return "", auth.ErrAuthPending
			}
			res, err := appsClient.GetGithubAuthStatus(ctx, &pb.GetGithubAuthStatusRequest{DeviceCode: "somedevicecode"})
			Expect(err).To(HaveOccurred())
			st, ok := status.FromError(err)
			Expect(ok).To(BeTrue(), "could not get status from err")
			Expect(st.Message()).To(ContainSubstring(auth.ErrAuthPending.Error()))
			Expect(res).To(BeNil())
		})
		It("retuns a jwt if the user has authenticated", func() {
			ctx := context.Background()
			token := "abc123def456"
			ghAuthClient.GetDeviceCodeAuthStatusStub = func(s string) (string, error) {
				return token, nil
			}
			res, err := appsClient.GetGithubAuthStatus(ctx, &pb.GetGithubAuthStatusRequest{DeviceCode: "somedevicecode"})
			Expect(err).NotTo(HaveOccurred())

			verified, err := auth.NewJwtClient(secretKey).VerifyJWT(res.AccessToken)
			Expect(err).NotTo(HaveOccurred())
			Expect(verified.ProviderToken).To(Equal(token))
		})
		It("returns an error other than ErrAuthPending", func() {
			ctx := context.Background()
			someErr := errors.New("some other err")
			ghAuthClient.GetDeviceCodeAuthStatusStub = func(s string) (string, error) {
				return "", someErr
			}
			res, err := appsClient.GetGithubAuthStatus(ctx, &pb.GetGithubAuthStatusRequest{DeviceCode: "somedevicecode"})
			Expect(err).To(HaveOccurred())
			st, ok := status.FromError(err)
			Expect(ok).To(BeTrue(), "could not get status from err")
			Expect(st.Message()).To(ContainSubstring(someErr.Error()))
			Expect(res).To(BeNil())
		})
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
				log = testutils.MakeFakeLogr()
				kubeClient = &kubefakes.FakeKube{}

				rand.Seed(time.Now().UnixNano())
				secretKey := rand.String(20)

				appFactory := &apputilsfakes.FakeAppFactory{}
				appFactory.GetAppServiceStub = func(ctx context.Context, name, namespace string) (app.AppService, error) {
					return app.New(ctx, nil, nil, nil, nil, nil, kubeClient, nil), nil
				}
				appFactory.GetKubeServiceStub = func() (kube.Kube, error) {
					return kubeClient, nil
				}
				appsSrv = NewApplicationsServer(&ApplicationsConfig{AppFactory: appFactory, JwtClient: auth.NewJwtClient(secretKey)})
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
				fakeJWTToken.GenerateJWTStub = func(duration time.Duration, name gitproviders.GitProviderName, s22 string) (string, error) {
					return "", fmt.Errorf("some error")
				}

				appFactory := &apputilsfakes.FakeAppFactory{}
				appFactory.GetAppServiceStub = func(ctx context.Context, name, namespace string) (app.AppService, error) {
					return app.New(ctx, nil, nil, nil, nil, nil, kubeClient, nil), nil
				}
				appFactory.GetKubeServiceStub = func() (kube.Kube, error) {
					return kubeClient, nil
				}
				appsSrv = NewApplicationsServer(&ApplicationsConfig{AppFactory: appFactory, JwtClient: fakeJWTToken})
				mux = runtime.NewServeMux(middleware.WithGrpcErrorLogging(log))
				httpHandler = middleware.WithLogging(log, mux)
				err = pb.RegisterApplicationsHandlerServer(context.Background(), mux, appsSrv)
				Expect(err).NotTo(HaveOccurred())

				ts := httptest.NewServer(httpHandler)
				defer ts.Close()

				// A valid URL for our server
				path := "/v1/authenticate/github"
				url := ts.URL + path

				res, err := http.Post(url, "application/json", strings.NewReader(`{"accessToken":"sometoken"}`))
				Expect(err).NotTo(HaveOccurred())
				Expect(res.StatusCode).To(Equal(http.StatusInternalServerError))

				bts, err := ioutil.ReadAll(res.Body)
				Expect(err).NotTo(HaveOccurred())

				Expect(bts).To(MatchJSON(`{"code": 13,"message": "error generating jwt token. some error","details": []}`))

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
		log := testutils.MakeFakeLogr()
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

		appFactory := &apputilsfakes.FakeAppFactory{}
		appFactory.GetAppServiceStub = func(ctx context.Context, name, namespace string) (app.AppService, error) {
			return app.New(ctx, nil, nil, nil, nil, nil, k, nil), nil
		}
		appFactory.GetKubeServiceStub = func() (kube.Kube, error) {
			return k, nil
		}

		cfg := ApplicationsConfig{
			AppFactory: appFactory,
			Logger:     log,
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
		log := testutils.MakeFakeLogr()
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

		gitProviders := &gitprovidersfakes.FakeGitProvider{}
		appFactory := &apputilsfakes.FakeAppFactory{}
		commits := []gitprovider.Commit{&fakeCommit{}}
		jwtClient := &authfakes.FakeJWTClient{
			VerifyJWTStub: func(s string) (*auth.Claims, error) {
				return &auth.Claims{
					ProviderToken: "provider-token",
				}, nil
			},
		}

		appFactory.GetAppServiceStub = func(ctx context.Context, name, namespace string) (app.AppService, error) {
			return app.New(ctx, nil, nil, nil, gitProviders, nil, kubeClient, nil), nil
		}

		appFactory.GetKubeServiceStub = func() (kube.Kube, error) {
			return kubeClient, nil
		}

		gitProviders.GetCommitsFromUserRepoStub = func(gitprovider.UserRepositoryRef, string, int, int) ([]gitprovider.Commit, error) {
			return commits, nil
		}

		gitProviders.GetAccountTypeStub = func(string) (gitproviders.ProviderAccountType, error) {
			return gitproviders.AccountTypeUser, nil
		}

		cfg := ApplicationsConfig{
			Logger:     log,
			AppFactory: appFactory,
			JwtClient:  jwtClient,
		}

		handler, err := NewApplicationsHandler(context.Background(), &cfg)
		Expect(err).NotTo(HaveOccurred())

		ts := httptest.NewServer(handler)
		defer ts.Close()

		path := "/v1/applications/testapp/commits"
		url := ts.URL + path

		req, err := http.NewRequest(http.MethodGet, url, nil)
		Expect(err).NotTo(HaveOccurred())
		req.Header.Add("Authorization", "token my-jwt-token")

		res, err := ts.Client().Do(req)
		Expect(err).NotTo(HaveOccurred())
		Expect(res.StatusCode).To(Equal(http.StatusOK))

		b, err := ioutil.ReadAll(res.Body)
		Expect(err).NotTo(HaveOccurred())

		r := &pb.ListCommitsResponse{}
		err = json.Unmarshal(b, r)
		Expect(err).NotTo(HaveOccurred())

		Expect(r.Commits).To(HaveLen(1))
		Expect(r.Commits[0].Url).To(Equal("http://github.com/testrepo/commit/2349898"))
		Expect(r.Commits[0].Message).To(Equal("if a message is above fifty characters then it wi..."))
		Expect(r.Commits[0].Hash).To(Equal("2349898"))
	})
})

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
		Sha:       "23498987239879892348768",
		Author:    "testauthor",
		Message:   "if a message is above fifty characters then it will be truncated",
		CreatedAt: time.Now(),
		URL:       "http://github.com/testrepo/commit/2349898723987989234",
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
