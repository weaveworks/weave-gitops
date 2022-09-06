package server_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	grpcruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc/metadata"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	pb "github.com/weaveworks/weave-gitops/pkg/api/profiles"
	"github.com/weaveworks/weave-gitops/pkg/helm/watcher/cache/cachefakes"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"
	"github.com/weaveworks/weave-gitops/pkg/server"
	"github.com/weaveworks/weave-gitops/pkg/testutils"
)

var _ = Describe("ProfilesServer", func() {
	var (
		fakeCache  *cachefakes.FakeCache
		s          *server.ProfilesServer
		helmRepo   *sourcev1.HelmRepository
		kubeClient client.Client
	)
	var profileName = "observability"
	var profileVersion = "latest"

	BeforeEach(func() {
		scheme := runtime.NewScheme()
		schemeBuilder := runtime.SchemeBuilder{
			sourcev1.AddToScheme,
		}
		Expect(schemeBuilder.AddToScheme(scheme)).To(Succeed())

		kubeClient = fake.NewClientBuilder().WithScheme(scheme).Build()

		fakeCache = &cachefakes.FakeCache{}
		fakeClientGetter := kubefakes.NewFakeClientGetter(kubeClient)
		log, _ := testutils.MakeFakeLogr()
		s = &server.ProfilesServer{
			Log:               log,
			HelmRepoName:      "helmrepo",
			HelmRepoNamespace: "default",
			HelmCache:         fakeCache,
			ClientGetter:      fakeClientGetter,
		}

		helmRepo = makeHelmRepo("default", "helmrepo")
	})

	Describe("GetProfiles", func() {
		When("the HelmRepository exists", func() {
			BeforeEach(func() {
				Expect(kubeClient.Create(context.TODO(), helmRepo)).To(Succeed())
				Expect(kubeClient.Create(context.TODO(), makeHelmRepo("test-namespace", "test-name"))).To(Succeed())
			})

			It("Returns a list of profiles in the HelmRepository", func() {
				profiles := []*pb.Profile{
					{
						Name: profileName,
					},
				}
				fakeCache.ListProfilesReturns(profiles, nil)

				profilesResp, err := s.GetProfiles(context.TODO(), &pb.GetProfilesRequest{})
				Expect(err).NotTo(HaveOccurred())
				Expect(profilesResp).NotTo(BeNil())
				Expect(profilesResp.Profiles).To(Equal(profiles))
				Expect(fakeCache.ListProfilesCallCount()).To(Equal(1))
				_, _, namespacedName := fakeCache.ListProfilesArgsForCall(0)
				Expect(namespacedName.Namespace).To(Equal(helmRepo.Namespace))
				Expect(namespacedName.Name).To(Equal(helmRepo.Name))
			})

			It("Passes the correct values to the Helm Cache Querier", func() {
				profiles := []*pb.Profile{
					{
						Name: profileName,
					},
				}
				fakeCache.ListProfilesReturns(profiles, nil)

				profilesResp, err := s.GetProfiles(context.TODO(), &pb.GetProfilesRequest{HelmRepoName: "test-name", HelmRepoNamespace: "test-namespace"})
				Expect(err).NotTo(HaveOccurred())
				Expect(profilesResp).NotTo(BeNil())
				Expect(profilesResp.Profiles).To(Equal(profiles))
				Expect(fakeCache.ListProfilesCallCount()).To(Equal(1))
				_, _, namespacedName := fakeCache.ListProfilesArgsForCall(0)
				Expect(namespacedName.Namespace).To(Equal("test-namespace"))
				Expect(namespacedName.Name).To(Equal("test-name"))
			})

			When("scanning for helmcharts errors", func() {
				It("errors", func() {
					fakeCache.ListProfilesReturns(nil, fmt.Errorf("foo"))
					_, err := s.GetProfiles(context.TODO(), &pb.GetProfilesRequest{})
					Expect(err).To(MatchError("failed to scan HelmRepository \"default\"/\"helmrepo\" for charts: foo"))
				})
			})
		})

		When("the HelmRepository doesn't exist", func() {
			It("errors", func() {
				_, err := s.GetProfiles(context.TODO(), &pb.GetProfilesRequest{})
				Expect(err).To(MatchError("HelmRepository \"default\"/\"helmrepo\" does not exist"))
				Expect(err).To(BeAssignableToTypeOf(&grpcruntime.HTTPStatusError{}))
				//TODO why do we return 200 when the HelmRepository doesn't exist
				Expect(err.(*grpcruntime.HTTPStatusError).HTTPStatus).To(Equal(http.StatusOK))
			})
		})
	})

	Describe("GetProfileValues", func() {
		When("the HelmRepository exists", func() {
			BeforeEach(func() {
				Expect(kubeClient.Create(context.TODO(), helmRepo)).To(Succeed())
				Expect(kubeClient.Create(context.TODO(), makeHelmRepo("test-namespace", "test-name"))).To(Succeed())
			})

			When("it retrieves the values file from Helm chart", func() {

				It("Passes the correct values to the Helm Cache Querier", func() {
					fakeCache.GetProfileValuesReturns([]byte("values"), nil)

					resp, err := s.GetProfileValues(context.TODO(), &pb.GetProfileValuesRequest{ProfileName: profileName,
						ProfileVersion: profileVersion, HelmRepoName: "test-name", HelmRepoNamespace: "test-namespace"})
					Expect(err).NotTo(HaveOccurred())
					Expect(resp.ContentType).To(Equal(server.JsonType))
					valuesResp := &pb.GetProfileValuesResponse{}
					err = json.Unmarshal(resp.Data, valuesResp)
					Expect(err).NotTo(HaveOccurred())
					profileValues, err := base64.StdEncoding.DecodeString(valuesResp.Values)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(profileValues)).To(Equal("values"))
					_, _, namespacedName, pname, pversion := fakeCache.GetProfileValuesArgsForCall(0)
					Expect(namespacedName.Namespace).To(Equal("test-namespace"))
					Expect(namespacedName.Name).To(Equal("test-name"))
					Expect(pname).To(Equal(profileName))
					Expect(pversion).To(Equal(profileVersion))
				})

				When("header does not contain 'application/octet-stream'", func() {
					It("returns a values file in base64-encoded json", func() {
						fakeCache.GetProfileValuesReturns([]byte("values"), nil)

						resp, err := s.GetProfileValues(context.TODO(), &pb.GetProfileValuesRequest{
							ProfileName:    profileName,
							ProfileVersion: profileVersion,
						})
						Expect(err).NotTo(HaveOccurred())
						Expect(resp.ContentType).To(Equal(server.JsonType))
						valuesResp := &pb.GetProfileValuesResponse{}
						err = json.Unmarshal(resp.Data, valuesResp)
						Expect(err).NotTo(HaveOccurred())
						profileValues, err := base64.StdEncoding.DecodeString(valuesResp.Values)
						Expect(err).NotTo(HaveOccurred())
						Expect(string(profileValues)).To(Equal("values"))
					})
				})

				When("header contains 'application/octet-stream'", func() {
					It("returns a values file in binary", func() {
						fakeCache.GetProfileValuesReturns([]byte("values"), nil)
						ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("accept", server.OctetStreamType))

						resp, err := s.GetProfileValues(ctx, &pb.GetProfileValuesRequest{
							ProfileName:    profileName,
							ProfileVersion: profileVersion,
						})
						Expect(err).NotTo(HaveOccurred())
						Expect(resp.ContentType).To(Equal(server.OctetStreamType))
						Expect(string(resp.Data)).To(Equal("values"))
					})
				})

				When("it cannot retrieve the values file from Helm chart", func() {
					It("errors", func() {
						fakeCache.GetProfileValuesReturns(nil, fmt.Errorf("err"))
						_, err := s.GetProfileValues(context.TODO(), &pb.GetProfileValuesRequest{
							ProfileName:    profileName,
							ProfileVersion: profileVersion,
						})
						Expect(err).To(MatchError("failed to retrieve values file from Helm chart 'observability' (latest): err"))
					})
				})
			})
		})

		When("the HelmRepository doesn't exist", func() {
			It("errors", func() {
				_, err := s.GetProfileValues(context.TODO(), &pb.GetProfileValuesRequest{})
				Expect(err).To(MatchError("HelmRepository \"default\"/\"helmrepo\" does not exist"))
				Expect(err).To(BeAssignableToTypeOf(&grpcruntime.HTTPStatusError{}))
				Expect(err.(*grpcruntime.HTTPStatusError).HTTPStatus).To(Equal(http.StatusOK))
			})
		})
	})
})

func makeHelmRepo(namespace, name string) *sourcev1.HelmRepository {
	return &sourcev1.HelmRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       sourcev1.HelmRepositoryKind,
			APIVersion: sourcev1.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: sourcev1.HelmRepositorySpec{
			URL:      "example.com/charts",
			Interval: metav1.Duration{Duration: time.Minute * 10},
		},
		Status: sourcev1.HelmRepositoryStatus{
			URL: "example.com/index.yaml",
		},
	}
}
