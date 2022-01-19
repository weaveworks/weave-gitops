package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/weaveworks/weave-gitops/pkg/services/auth/authfakes"
	"github.com/weaveworks/weave-gitops/pkg/testutils"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/weaveworks/weave-gitops/pkg/services/auth"
	authtypes "github.com/weaveworks/weave-gitops/pkg/services/auth/types"

	"github.com/fluxcd/go-git-providers/gitprovider"
	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev2 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	pb "github.com/weaveworks/weave-gitops/pkg/api/applications"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"
	"github.com/weaveworks/weave-gitops/pkg/logger/loggerfakes"
	"github.com/weaveworks/weave-gitops/pkg/models"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/osys/osysfakes"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/server/middleware"
	"github.com/weaveworks/weave-gitops/pkg/services/app"
	"github.com/weaveworks/weave-gitops/pkg/services/applicationv2"
	"github.com/weaveworks/weave-gitops/pkg/services/applicationv2/applicationv2fakes"
	"github.com/weaveworks/weave-gitops/pkg/services/servicesfakes"
	fakelogr "github.com/weaveworks/weave-gitops/pkg/vendorfakes/logr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
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
				kust := &kustomizev2.Kustomization{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: namespace.Name,
					},
					Spec: kustomizev2.KustomizationSpec{
						TargetNamespace: "target-namespace",
						Path:            "/path",
						Interval:        metav1.Duration{Duration: 1 * time.Second},
						Prune:           true,
						SourceRef: kustomizev2.CrossNamespaceSourceReference{
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
			kustomization := kustomizev2.Kustomization{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace.Name,
				},
				Spec: kustomizev2.KustomizationSpec{
					SourceRef: kustomizev2.CrossNamespaceSourceReference{
						Kind: sourcev1.GitRepositoryKind,
					},
				},
				Status: kustomizev2.KustomizationStatus{
					Inventory: &kustomizev2.ResourceInventory{
						Entries: []kustomizev2.ResourceRef{
							{
								Version: "v1",
								ID:      namespace.Name + "_my-deployment_apps_Deployment",
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
				APIVersion: appsv1.SchemeGroupVersion.String(),
				Kind:       "Deployment",
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

	Describe("AddApplication", func() {
		It("adds an app with an unspecified config repo", func() {
			ctx := context.Background()
			name := "my-app"
			appRequest := &pb.AddApplicationRequest{
				Name:      name,
				Namespace: namespace.Name,
				Url:       "ssh://git@github.com/some-org/somerepo.git",
				Path:      "./k8s/mydir",
				Branch:    "main",
			}
			gitProvider.GetRepoVisibilityReturns(gitprovider.RepositoryVisibilityVar(gitprovider.RepositoryVisibilityInternal), nil)

			gitProvider.CreatePullRequestReturns(testutils.DummyPullRequest{}, nil)

			res, err := appsClient.AddApplication(contextWithAuth(ctx), appRequest)
			Expect(err).NotTo(HaveOccurred())
			Expect(res.Success).To(BeTrue())

			Expect(gitProvider.CreatePullRequestCallCount()).To(Equal(1), "should have made a PR")
		})
		It("adds an app with a config repo url specified", func() {
			ctx := context.Background()
			name := "my-app"
			appRequest := &pb.AddApplicationRequest{
				Name:       name,
				Namespace:  namespace.Name,
				Url:        "ssh://git@github.com/some-org/somerepo.git",
				Path:       "./k8s/mydir",
				Branch:     "main",
				ConfigRepo: "ssh://git@github.com/some-org/my-config-url.git",
			}

			gitProvider.GetRepoVisibilityReturns(gitprovider.RepositoryVisibilityVar(gitprovider.RepositoryVisibilityInternal), nil)
			gitProvider.CreatePullRequestReturns(testutils.DummyPullRequest{}, nil)

			res, err := appsClient.AddApplication(contextWithAuth(ctx), appRequest)
			Expect(err).NotTo(HaveOccurred())
			Expect(res.Success).To(BeTrue())

			Expect(configGit.CommitCallCount()).To(Equal(0), "should have committed to config git repo")
			Expect(gitProvider.CreatePullRequestCallCount()).To(Equal(1), "should have made a PR")
		})
		It("adds an app with automerge and no config repo defined", func() {
			ctx := context.Background()
			name := "my-app"
			appRequest := &pb.AddApplicationRequest{
				Name:      name,
				Namespace: namespace.Name,
				Url:       "ssh://git@github.com/some-org/somerepo.git",
				Path:      "./k8s/mydir",
				Branch:    "main",
				AutoMerge: true,
			}
			gitProvider.GetRepoVisibilityReturns(gitprovider.RepositoryVisibilityVar(gitprovider.RepositoryVisibilityInternal), nil)
			gitProvider.CreatePullRequestReturns(testutils.DummyPullRequest{}, nil)

			res, err := appsClient.AddApplication(contextWithAuth(ctx), appRequest)
			Expect(err).NotTo((HaveOccurred()))
			Expect(res.Success).To(BeTrue())

			Expect(configGit.CommitCallCount()).To(Equal(1), "should have committed to the config git repo")
			Expect(gitProvider.CreatePullRequestCallCount()).To(Equal(0), "should NOT have made a PR")
		})
	})

	Context("RemoveApplication Tests", func() {
		var ctx context.Context
		var fakeKube *kubefakes.FakeKube
		var name string
		var manifestPaths map[string]bool
		var appDir string

		storeManifestPath := func(path string) {
			if strings.HasSuffix(path, "user/kustomization.yaml") {
				return
			}

			appDir = filepath.Dir(path)
			manifestPaths[path] = true
		}

		removeManifestPath := func(basepath string) error {
			path := filepath.Join(appDir, basepath)

			if !manifestPaths[path] {
				return fmt.Errorf("manifest path: %s not found in repository", path)
			}

			delete(manifestPaths, path)

			return nil
		}

		BeforeEach(func() {
			ctx = context.Background()
			fakeKube = &kubefakes.FakeKube{}
			name = "my-app"
			manifestPaths = map[string]bool{}

			fakeOsys := &osysfakes.FakeOsys{}
			fluxClient := flux.New(osys.New(), &testutils.LocalFluxRunner{Runner: &runner.CLIRunner{}})
			log := &loggerfakes.FakeLogger{}

			fakeFactory.GetAppServiceReturns(&app.AppSvc{
				Context: ctx,
				Flux:    fluxClient,
				Kube:    fakeKube,
				Logger:  log,
				Osys:    fakeOsys,
			}, nil)

			fakeFactory.GetGitClientsReturns(configGit, gitProvider, nil)
			fakeFactory.GetKubeServiceReturns(fakeKube, nil)
			gitProvider.CreatePullRequestReturns(testutils.DummyPullRequest{}, nil)

			configGit.WriteStub = func(path string, manifest []byte) error {
				storeManifestPath(path)
				return nil
			}

			configGit.RemoveStub = func(path string) error {
				return removeManifestPath(path)
			}

			fakeOsys.ReadDirStub = func(dirName string) ([]os.DirEntry, error) {
				return makeDirEntries(manifestPaths), nil
			}
		})

		DescribeTable(
			"Remove applications",
			func(
				url,
				configRepo string,
				sourceType wego.SourceType,
				deploymentType wego.DeploymentType,
				autoMerge bool,
				commitCount, prCount int,
			) {
				application := wego.Application{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: wego.DefaultNamespace,
					},
					Spec: wego.ApplicationSpec{
						Branch:         "main",
						Path:           "./k8s",
						URL:            url,
						ConfigRepo:     configRepo,
						SourceType:     sourceType,
						DeploymentType: deploymentType,
					},
				}

				fakeKube.GetApplicationReturns(&application, nil)

				appRequest := &pb.RemoveApplicationRequest{
					Name:      name,
					Namespace: namespace.Name,
					AutoMerge: autoMerge,
				}
				res, err := appsClient.RemoveApplication(contextWithAuth(ctx), appRequest)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.Success).To(BeTrue())

				Expect(configGit.CommitCallCount()).To(Equal(commitCount))
				Expect(gitProvider.CreatePullRequestCallCount()).To(Equal(prCount))
			},
			Entry(
				"kustomize, app repo config, auto merge",
				"ssh://git@github.com/foo/bar",
				"ssh://git@github.com/foo/bar",
				wego.SourceTypeGit,
				wego.DeploymentTypeKustomize,
				true,
				1,
				0),
			Entry(
				"kustomize, external repo config, auto merge",
				"ssh://git@github.com/foo/bar",
				"ssh://git@github.com/foo/baz",
				wego.SourceTypeGit,
				wego.DeploymentTypeKustomize,
				true,
				1,
				0))
	})

	Describe("ListCommits", func() {
		It("gets commits for an app", func() {
			testApp := &wego.Application{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testapp",
					Namespace: namespace.Name,
				},
				Spec: wego.ApplicationSpec{
					Branch: "main",
					Path:   "./k8s",
					URL:    "https://github.com/owner/repo1",
				},
			}
			Expect(k8sClient.Create(context.Background(), testApp)).To(Succeed())

			c := newTestcommit(gitprovider.CommitInfo{
				URL:     "http://github.com/testrepo/commit/2349898",
				Message: "my message",
				Sha:     "2349898",
			})
			commits := []gitprovider.Commit{c}

			gitProvider.GetCommitsReturns(commits, nil)

			res, err := appsClient.ListCommits(contextWithAuth(context.Background()), &pb.ListCommitsRequest{
				Name:      testApp.Name,
				Namespace: testApp.Namespace,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(res.Commits).To(HaveLen(1))
			desired := c.Get()
			Expect(res.Commits[0].Url).To(Equal(desired.URL))
			Expect(res.Commits[0].Message).To(Equal(desired.Message))
			Expect(res.Commits[0].Hash).To(Equal(desired.Sha))
		})
	})

	Describe("SyncApplication", func() {
		var (
			ctx    context.Context
			name   string
			app    *wego.Application
			kust   *kustomizev2.Kustomization
			source *sourcev1.GitRepository
		)

		BeforeEach(func() {
			ctx = context.Background()
			name = "my-app"
			app = &wego.Application{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace.Name,
				},
				Spec: wego.ApplicationSpec{
					SourceType:     wego.SourceTypeGit,
					DeploymentType: wego.DeploymentTypeKustomize,
				},
			}

			kust = &kustomizev2.Kustomization{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace.Name,
				},
				Spec: kustomizev2.KustomizationSpec{
					SourceRef: kustomizev2.CrossNamespaceSourceReference{
						Kind: "GitRepository",
					},
				},
				Status: kustomizev2.KustomizationStatus{
					ReconcileRequestStatus: meta.ReconcileRequestStatus{
						LastHandledReconcileAt: time.Now().Format(time.RFC3339Nano),
					},
				},
			}

			source = &sourcev1.GitRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace.Name,
				},
				Spec: sourcev1.GitRepositorySpec{
					URL: "https://github.com/owner/repo",
				},
				Status: sourcev1.GitRepositoryStatus{
					ReconcileRequestStatus: meta.ReconcileRequestStatus{
						LastHandledReconcileAt: time.Now().Format(time.RFC3339Nano),
					},
				},
			}

			Expect(k8sClient.Create(ctx, app)).Should(Succeed())
			Expect(k8sClient.Create(ctx, source)).Should(Succeed())
			Expect(k8sClient.Create(ctx, kust)).Should(Succeed())
		})

		// TODO: Issue 981 fix flaky test
		XIt("trigger the reconcile loop for an application", func() {
			appRequest := &pb.SyncApplicationRequest{
				Name:      name,
				Namespace: namespace.Name,
			}

			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace.Name}, source)).Should(Succeed())
			source.Status.SetLastHandledReconcileRequest(time.Now().Format(time.RFC3339Nano))
			Expect(k8sClient.Status().Update(ctx, source)).Should(Succeed())

			done := make(chan bool)
			defer close(done)

			go func() {
				defer GinkgoRecover()

				res, err := appsClient.SyncApplication(contextWithAuth(ctx), appRequest)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.Success).To(BeTrue())
				done <- true
			}()

			ticker := time.NewTicker(500 * time.Millisecond)
			for {
				select {
				case <-ticker.C:
					Expect(k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace.Name}, source)).Should(Succeed())
					source.Status.SetLastHandledReconcileRequest(time.Now().Format(time.RFC3339Nano))
					Expect(k8sClient.Status().Update(ctx, source)).Should(Succeed())
					Expect(k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace.Name}, kust)).Should(Succeed())
					kust.Status.SetLastHandledReconcileRequest(time.Now().Format(time.RFC3339Nano))
					Expect(k8sClient.Status().Update(ctx, kust)).Should(Succeed())
				case <-done:
					return
				case <-time.After(3 * time.Second):
					Fail("SyncApplication test timed out")
				}
			}
		})
	})
	Describe("ListCommits", func() {
		It("gets commits for an app", func() {
			testApp := &wego.Application{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testapp",
					Namespace: namespace.Name,
				},
				Spec: wego.ApplicationSpec{
					Branch: "main",
					Path:   "./k8s",
					URL:    "https://github.com/owner/repo1",
				},
			}
			Expect(k8sClient.Create(context.Background(), testApp)).To(Succeed())

			c := newTestcommit(gitprovider.CommitInfo{
				URL:     "http://github.com/testrepo/commit/2349898",
				Message: "my message",
				Sha:     "2349898",
			})
			commits := []gitprovider.Commit{c}
			gitProvider.GetCommitsReturns(commits, nil)
			gitProvider.GetCommitsReturns(commits, nil)

			res, err := appsClient.ListCommits(contextWithAuth(context.Background()), &pb.ListCommitsRequest{
				Name:      testApp.Name,
				Namespace: testApp.Namespace,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(res.Commits).To(HaveLen(1))
			desired := c.Get()
			Expect(res.Commits[0].Url).To(Equal(desired.URL))
			Expect(res.Commits[0].Message).To(Equal(desired.Message))
			Expect(res.Commits[0].Hash).To(Equal(desired.Sha))
		})
	})

	Describe("ParseRepoURL", func() {
		type expected struct {
			provider pb.GitProvider
			owner    string
			name     string
		}
		DescribeTable("parses a repo url", func(uri string, e expected) {
			res, err := appsClient.ParseRepoURL(context.Background(), &pb.ParseRepoURLRequest{
				Url: uri,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(res.Provider).To(Equal(e.provider))
			Expect(res.Owner).To(Equal(e.owner))
			Expect(res.Name).To(Equal(e.name))
		},
			Entry("github+ssh", "git@github.com:some-org/my-repo.git", expected{
				provider: pb.GitProvider_GitHub,
				owner:    "some-org",
				name:     "my-repo",
			}),
			Entry("gitlab+ssh", "git@gitlab.com:other-org/cool-repo.git", expected{
				provider: pb.GitProvider_GitLab,
				owner:    "other-org",
				name:     "cool-repo",
			}),
		)

		It("returns an error on an invalid URL", func() {
			_, err := appsClient.ParseRepoURL(context.Background(), &pb.ParseRepoURLRequest{
				Url: "not-a  -valid-url",
			})
			Expect(err).To(HaveOccurred(), "should have gotten an invalid arg error")
			s, ok := status.FromError(err)
			Expect(ok).To(BeTrue(), "could not get status from error")
			Expect(s.Code()).To(Equal(codes.InvalidArgument))
		})
	})
	Describe("GetGitlabAuthURL", func() {
		It("returns the gitlab url", func() {
			urlString := "http://gitlab.com/oauth/authorize"
			authUrl, err := url.Parse(urlString)
			Expect(err).NotTo(HaveOccurred())

			glAuthClient.AuthURLReturns(*authUrl, nil)

			res, err := appsClient.GetGitlabAuthURL(context.Background(), &pb.GetGitlabAuthURLRequest{
				RedirectUri: "http://example.com/oauth/fake",
			})
			Expect(err).NotTo(HaveOccurred())

			u, err := url.Parse(res.Url)
			Expect(err).NotTo(HaveOccurred())
			Expect(u.String()).To(Equal(urlString))
		})
	})

	Describe("AuthorizeGitlab", func() {
		It("exchanges a token", func() {
			token := "some-token"
			glAuthClient.ExchangeCodeReturns(&authtypes.TokenResponseState{AccessToken: token}, nil)

			res, err := appsClient.AuthorizeGitlab(context.Background(), &pb.AuthorizeGitlabRequest{
				RedirectUri: "http://example.com/oauth/callback",
				Code:        "some-challenge-code",
			})
			Expect(err).NotTo(HaveOccurred())

			claims, err := jwtClient.VerifyJWT(res.Token)
			Expect(err).NotTo(HaveOccurred())

			Expect(claims.ProviderToken).To(Equal(token))
		})
		It("returns an error if the exchange fails", func() {
			e := errors.New("some code exchange error")
			glAuthClient.ExchangeCodeReturns(nil, e)

			_, err := appsClient.AuthorizeGitlab(context.Background(), &pb.AuthorizeGitlabRequest{
				RedirectUri: "http://example.com/oauth/callback",
				Code:        "some-challenge-code",
			})
			status, ok := status.FromError(err)
			Expect(ok).To(BeTrue(), "could not get status from error response")
			Expect(status.Code()).To(Equal(codes.Unknown))
			Expect(err.Error()).To(ContainSubstring(e.Error()))
		})
	})
	DescribeTable("ValidateProviderToken", func(provider pb.GitProvider, ctx context.Context, errResponse error, expectedCode codes.Code, valid bool) {
		glAuthClient.ValidateTokenReturns(errResponse)
		ghAuthClient.ValidateTokenReturns(errResponse)

		res, err := appsClient.ValidateProviderToken(ctx, &pb.ValidateProviderTokenRequest{
			Provider: provider,
		})

		if errResponse != nil {
			Expect(err).To(HaveOccurred())
			s, ok := status.FromError(err)
			Expect(ok).To(BeTrue(), "could not get status from error")
			Expect(s.Code()).To(Equal(expectedCode))
			return
		}

		Expect(err).NotTo(HaveOccurred())

		if valid {
			// Note that res is nil when there is an error.
			// Only check a field in `res` when valid, else this will panic
			Expect(res.Valid).To(BeTrue())
		}
	},
		Entry("bad gitlab token", pb.GitProvider_GitLab, contextWithAuth(context.Background()), errors.New("this token is bad"), codes.InvalidArgument, false),
		Entry("good gitlab token", pb.GitProvider_GitLab, contextWithAuth(context.Background()), nil, codes.OK, true),
		Entry("bad github token", pb.GitProvider_GitHub, contextWithAuth(context.Background()), errors.New("this token is bad"), codes.InvalidArgument, false),
		Entry("good github token", pb.GitProvider_GitHub, contextWithAuth(context.Background()), nil, codes.OK, true),
		Entry("no gitops jwt", pb.GitProvider_GitHub, context.Background(), errors.New("unauth error"), codes.Unauthenticated, false),
	)

	Describe("middleware", func() {
		Describe("logging", func() {
			var log *fakelogr.FakeLogger
			var appsSrv pb.ApplicationsServer
			var mux *runtime.ServeMux
			var httpHandler http.Handler
			var err error

			BeforeEach(func() {
				log = testutils.MakeFakeLogr()
				k8s := fake.NewClientBuilder().WithScheme(kube.CreateScheme()).Build()

				rand.Seed(time.Now().UnixNano())
				secretKey := rand.String(20)

				fakeFactory := &servicesfakes.FakeFactory{}

				cfg := ApplicationsConfig{
					Logger:         log,
					JwtClient:      auth.NewJwtClient(secretKey),
					FetcherFactory: NewFakeFetcherFactory(applicationv2.NewFetcher(k8s)),
					Factory:        fakeFactory,
				}

				appsSrv = NewApplicationsServer(&cfg, WithClientGetter(NewFakeClientGetter(k8s)))
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
				errMsg := "there was a big problem"
				fakeFetcher := &applicationv2fakes.FakeFetcher{}
				// Pretend something went horribly wrong
				fakeFetcher.ListReturns([]models.Application{}, errors.New(errMsg))

				cfg := ApplicationsConfig{
					Logger:         log,
					FetcherFactory: NewFakeFetcherFactory(fakeFetcher),
				}

				k8s := fake.NewClientBuilder().WithScheme(kube.CreateScheme()).Build()
				appSrv := NewApplicationsServer(&cfg, WithClientGetter(NewFakeClientGetter(k8s)))
				err = pb.RegisterApplicationsHandlerServer(context.Background(), mux, appSrv)
				Expect(err).NotTo(HaveOccurred())

				ts := httptest.NewServer(httpHandler)
				defer ts.Close()

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

				kubeClient := &kubefakes.FakeKube{}
				factory := &servicesfakes.FakeFactory{}

				factory.GetKubeServiceStub = func() (kube.Kube, error) {
					return kubeClient, nil
				}
				appsSrv = NewApplicationsServer(&ApplicationsConfig{Factory: factory, JwtClient: fakeJWTToken})
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
		ctx := context.Background()
		k8s := fake.NewClientBuilder().WithScheme(kube.CreateScheme()).Build()

		app := &wego.Application{}
		app.Name = "my-app"
		app.Namespace = "some-ns"

		Expect(k8s.Create(ctx, app)).To(Succeed())

		app2 := &wego.Application{}
		app2.Name = "my-app2"
		app2.Namespace = "some-ns"

		Expect(k8s.Create(ctx, app2)).To(Succeed())

		factory := &servicesfakes.FakeFactory{}

		factory.GetKubeServiceStub = func() (kube.Kube, error) {
			return k, nil
		}

		cfg := ApplicationsConfig{
			Logger:         log,
			FetcherFactory: NewFakeFetcherFactory(applicationv2.NewFetcher(k8s)),
			Factory:        factory,
			ClusterConfig:  ClusterConfig{},
		}

		handler, err := NewHandlers(context.Background(), &Config{
			AppConfig:  &cfg,
			AppOptions: []ApplicationsOption{WithClientGetter(NewFakeClientGetter(k8s))},
		})
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

		Expect(r.Applications).To(HaveLen(2))
	})
})

type fakeCommit struct {
	commitInfo gitprovider.CommitInfo
}

func (fc *fakeCommit) APIObject() interface{} {
	return &fc.commitInfo
}

func (fc *fakeCommit) Get() gitprovider.CommitInfo {
	return fc.commitInfo
}

func newTestcommit(c gitprovider.CommitInfo) gitprovider.Commit {
	return &fakeCommit{commitInfo: c}
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

func contextWithAuth(ctx context.Context) context.Context {
	md := metadata.New(map[string]string{middleware.GRPCAuthMetadataKey: "mytoken"})
	ctx = metadata.NewOutgoingContext(ctx, md)

	return ctx
}

// Support for remove tests

type dummyDirEntry struct {
	name string
}

func (d dummyDirEntry) Name() string {
	return d.name
}

func (d dummyDirEntry) IsDir() bool {
	return false
}

func (d dummyDirEntry) Type() fs.FileMode {
	return fs.ModeDir
}

func (d dummyDirEntry) Info() (fs.FileInfo, error) {
	return nil, nil
}

func makeDirEntries(paths map[string]bool) []os.DirEntry {
	results := []os.DirEntry{}

	for path := range paths {
		results = append(results, dummyDirEntry{name: filepath.Base(path)})
	}

	return results
}

type FakeFetcherFactory struct {
	fake applicationv2.Fetcher
}

func NewFakeFetcherFactory(fake applicationv2.Fetcher) FetcherFactory {
	return &FakeFetcherFactory{
		fake: fake,
	}
}

func (f *FakeFetcherFactory) Create(client client.Client) applicationv2.Fetcher {
	return f.fake
}

type FakeClientGetter struct {
	client client.Client
}

func NewFakeClientGetter(client client.Client) ClientGetter {
	return &FakeClientGetter{
		client: client,
	}
}

func (g *FakeClientGetter) Client(ctx context.Context) (client.Client, error) {
	return g.client, nil
}
