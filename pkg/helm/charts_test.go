package helm_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	helmv2beta1 "github.com/fluxcd/helm-controller/api/v2beta1"
	fluxmeta "github.com/fluxcd/pkg/apis/meta"
	sourcev1beta1 "github.com/fluxcd/source-controller/api/v1beta1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	pb "github.com/weaveworks/weave-gitops/pkg/api/profiles"
	"github.com/weaveworks/weave-gitops/pkg/helm"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/repo"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/yaml"
)

const (
	testNamespace  = "testing"
	testSecretName = "https-credentials"
)

var _ = Describe("ScanCharts", func() {
	Context("GetCharts", func() {

		var scanner *helm.RepoManager

		BeforeEach(func() {
			scanner = &helm.RepoManager{}
		})

		It("returns all profiles in the repository", func() {
			testServer := httptest.NewServer(http.FileServer(http.Dir("testdata/with_profiles")))
			profiles, err := scanner.GetCharts(context.TODO(), makeTestHelmRepository(testServer.URL), helm.Profiles)
			Expect(err).NotTo(HaveOccurred())

			Expect(profiles).To(ConsistOf(&pb.Profile{
				Name:        "demo-profile",
				Home:        "https://helm.sh/helm",
				Sources:     []string{"https://github.com/helm/charts"},
				Description: "Simple demo profile",
				Keywords:    []string{"gitops", "demo"},
				Maintainers: []*pb.Maintainer{
					{
						Name:  "WeaveWorks",
						Email: "maintainers@weave.works",
						Url:   "",
					},
					{
						Name:  "CNCF",
						Email: "",
						Url:   "cncf.io",
					},
				},
				Icon: "https://helm.sh/icon.png",
				AvailableVersions: []string{
					"1.1.0",
					"1.1.2",
				},
			}))
		})

		When("no charts exist with the profiles tag", func() {
			It("returns an empty list", func() {
				testServer := httptest.NewServer(http.FileServer(http.Dir("testdata/no_profiles")))
				profiles, err := scanner.GetCharts(context.TODO(), makeTestHelmRepository(testServer.URL), helm.Profiles)
				Expect(err).NotTo(HaveOccurred())
				Expect(profiles).To(BeEmpty())
			})
		})

		When("server isn't a valid helm repository", func() {
			It("errors", func() {
				testServer := httptest.NewServer(http.FileServer(http.Dir("testdata")))
				_, err := scanner.GetCharts(context.TODO(), makeTestHelmRepository(testServer.URL), helm.Profiles)
				Expect(err).To(MatchError(ContainSubstring("fetching profiles from HelmRepository testing/test-ns")))
				Expect(err).To(MatchError(ContainSubstring("404")))
			})
		})

		When("the URL is invalid", func() {
			It("errors", func() {
				_, err := scanner.GetCharts(context.TODO(), makeTestHelmRepository("http://[::1]:namedport/index.yaml"), helm.Profiles)
				Expect(err).To(MatchError(ContainSubstring("invalid port")))
			})
		})

		When("the scheme is unsupported", func() {
			It("errors", func() {
				_, err := scanner.GetCharts(context.TODO(), makeTestHelmRepository("sftp://localhost:4222/index.yaml"), helm.Profiles)
				Expect(err).To(MatchError(ContainSubstring("no provider for scheme: sftp")))
			})
		})

		When("the index file doesn't contain an API version", func() {
			It("errors", func() {
				testServer := httptest.NewServer(http.FileServer(http.Dir("testdata")))
				_, err := scanner.GetCharts(context.TODO(), makeTestHelmRepository(testServer.URL+"/invalid"), helm.Profiles)
				Expect(err).To(MatchError(ContainSubstring("no API version specified")))
			})
		})

		When("the index file isn't valid yaml", func() {
			It("errors", func() {
				testServer := httptest.NewServer(http.FileServer(http.Dir("testdata")))
				_, err := scanner.GetCharts(context.TODO(), makeTestHelmRepository(testServer.URL+"/brokenyaml"), helm.Profiles)
				Expect(err).To(MatchError(ContainSubstring("cannot decode")))
			})
		})
	})

	Context("GetValuesFile", func() {
		var tempDir string

		BeforeEach(func() {
			var err error
			tempDir, err = ioutil.TempDir("", "values-test")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			Expect(os.RemoveAll(tempDir)).To(Succeed())
		})

		It("returns the values file for a chart", func() {
			testServer := httptest.NewServer(makeServeMux())
			helmRepo := makeTestHelmRepository(testServer.URL)
			chartReference := &helm.ChartReference{Chart: "demo-profile", Version: "0.0.1", SourceRef: referenceForRepository(helmRepo)}
			chartClient := helm.NewRepoManager(makeTestClient(), testNamespace, tempDir)

			values, err := chartClient.GetValuesFile(context.TODO(), helmRepo, chartReference, "values.yaml")
			Expect(err).NotTo(HaveOccurred())
			Expect(string(values)).To(Equal("favoriteDrink: coffee\n"))
		})

		When("the chart version doesn't exist", func() {
			It("errors", func() {
				testServer := httptest.NewServer(makeServeMux())
				helmRepo := makeTestHelmRepository(testServer.URL)
				chartReference := &helm.ChartReference{Chart: "demo-profile", Version: "0.0.2", SourceRef: referenceForRepository(helmRepo)}
				chartClient := helm.NewRepoManager(makeTestClient(), testNamespace, tempDir)

				_, err := chartClient.GetValuesFile(context.TODO(), helmRepo, chartReference, "values.yaml")
				Expect(err).To(MatchError(ContainSubstring(`chart "demo-profile" version "0.0.2" not found`)))
			})
		})

		When("the chart doesn't exist", func() {
			It("errors", func() {
				testServer := httptest.NewServer(makeServeMux(func(ri *repo.IndexFile) {
					ri.Entries["demo-profile"][0].Metadata.Version = "0.0.2"
					ri.Entries["demo-profile"][0].URLs = nil
				}))

				helmRepo := makeTestHelmRepository(testServer.URL)
				chartReference := &helm.ChartReference{Chart: "demo-profile", Version: "0.0.2", SourceRef: referenceForRepository(helmRepo)}
				chartClient := helm.NewRepoManager(makeTestClient(), testNamespace, tempDir)

				_, err := chartClient.GetValuesFile(context.TODO(), helmRepo, chartReference, "values.yaml")
				Expect(err).To(MatchError(ContainSubstring(`chart "demo-profile" version "0.0.2" has no downloadable URLs`)))
			})
		})

		When("the chart URL is invalid", func() {
			It("errors", func() {
				helmRepo := makeTestHelmRepository("http://[::1]:namedport/index.yaml")
				chartReference := &helm.ChartReference{Chart: "demo-profile", Version: "0.0.1", SourceRef: referenceForRepository(helmRepo)}
				chartClient := helm.NewRepoManager(makeTestClient(), testNamespace, tempDir)

				_, err := chartClient.GetValuesFile(context.TODO(), helmRepo, chartReference, "values.yaml")
				Expect(err).To(MatchError(ContainSubstring("invalid chart URL format")))
			})
		})

		When("the credentials to access the repository are missing", func() {
			It("errors", func() {
				testServer := httptest.NewServer(basicAuthHandler(makeServeMux(), "test", "password"))
				helmRepo := makeTestHelmRepository(testServer.URL, func(hr *sourcev1beta1.HelmRepository) {
					hr.Spec.SecretRef = &fluxmeta.LocalObjectReference{
						Name: testSecretName,
					}
				})
				chartReference := &helm.ChartReference{Chart: "demo-profile", Version: "0.0.1", SourceRef: referenceForRepository(helmRepo)}
				chartClient := helm.NewRepoManager(makeTestClient(), testNamespace, tempDir)

				_, err := chartClient.GetValuesFile(context.TODO(), helmRepo, chartReference, "values.yaml")
				Expect(err).To(MatchError(ContainSubstring(`repository authentication: secrets "https-credentials" not found`)))
			})
		})
	})
})

func makeTestHelmRepository(base string, opts ...func(*sourcev1beta1.HelmRepository)) *sourcev1beta1.HelmRepository {
	hr := &sourcev1beta1.HelmRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       sourcev1beta1.HelmRepositoryKind,
			APIVersion: sourcev1beta1.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testing",
			Namespace: "test-ns",
		},
		Spec: sourcev1beta1.HelmRepositorySpec{
			URL:      base + "/charts",
			Interval: metav1.Duration{Duration: time.Minute * 10},
		},
		Status: sourcev1beta1.HelmRepositoryStatus{
			URL: base + "/index.yaml",
		},
	}

	for _, o := range opts {
		o(hr)
	}

	return hr
}

func makeServeMux(opts ...func(*repo.IndexFile)) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/charts/index.yaml", func(w http.ResponseWriter, req *http.Request) {
		b, err := yaml.Marshal(makeTestChartIndex(opts...))
		Expect(err).NotTo(HaveOccurred())
		w.Write(b)
	})
	mux.Handle("/", http.FileServer(http.Dir("testdata")))

	return mux
}

func referenceForRepository(s *sourcev1beta1.HelmRepository) helmv2beta1.CrossNamespaceObjectReference {
	return helmv2beta1.CrossNamespaceObjectReference{
		APIVersion: s.TypeMeta.APIVersion,
		Kind:       s.TypeMeta.Kind,
		Name:       s.ObjectMeta.Name,
		Namespace:  s.ObjectMeta.Name,
	}
}

func makeTestChartIndex(opts ...func(*repo.IndexFile)) *repo.IndexFile {
	ri := &repo.IndexFile{
		APIVersion: "v1",
		Entries: map[string]repo.ChartVersions{
			"demo-profile": repo.ChartVersions{
				{
					Metadata: &chart.Metadata{
						Annotations: map[string]string{
							helm.ProfileAnnotation: "demo-profile",
						},
						Description: "Simple demo profile",
						Home:        "https://example.com/testing",
						Name:        "demo-profile",
						Sources: []string{
							"https://example.com/testing",
						},
						Version: "0.0.1",
					},
					Created: time.Now(),
					Digest:  "aaff4545f79d8b2913a10cb400ebb6fa9c77fe813287afbacf1a0b897cdffffff",
					URLs: []string{
						"/charts/demo-profile-0.1.0.tgz",
					},
				},
			},
		},
	}

	for _, o := range opts {
		o(ri)
	}

	return ri
}

func basicAuthHandler(next http.Handler, user, pass string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if ok && (u == user && p == pass) {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm="test"`))
		w.WriteHeader(401)
		w.Write([]byte("401 Unauthorized\n"))
	})
}

func makeTestClient(objs ...runtime.Object) client.Client {
	s := runtime.NewScheme()
	Expect(corev1.AddToScheme(s)).To(Succeed())

	return fake.NewClientBuilder().WithScheme(s).WithRuntimeObjects(objs...).Build()
}

// Based on https://fluxcd.io/docs/components/source/helmrepositories/
func makeTestSecret(user, pass string) *corev1.Secret {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		Type: corev1.SecretTypeOpaque,
		ObjectMeta: metav1.ObjectMeta{
			Name:      testSecretName,
			Namespace: testNamespace,
		},
		Data: map[string][]byte{
			"username": []byte(user),
			"password": []byte(pass),
		},
	}
}
