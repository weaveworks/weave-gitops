package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	pb "github.com/weaveworks/weave-gitops/pkg/api/profiles"
	fakes "github.com/weaveworks/weave-gitops/pkg/charts/chartsfakes"
	"github.com/weaveworks/weave-gitops/pkg/testutils"

	sourcev1beta1 "github.com/fluxcd/source-controller/api/v1beta1"
	grpcruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc/metadata"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("ProfilesServer", func() {
	var (
		fakeChartScanner *fakes.FakeChartScanner
		fakeChartClient  *fakes.FakeChartClient
		s                *ProfilesServer
		helmRepo         *sourcev1beta1.HelmRepository
		kubeClient       client.Client
	)
	var profileName = "observability"
	var profileVersion = "latest"

	BeforeEach(func() {
		scheme := runtime.NewScheme()
		schemeBuilder := runtime.SchemeBuilder{
			sourcev1beta1.AddToScheme,
		}
		schemeBuilder.AddToScheme(scheme)

		kubeClient = fake.NewClientBuilder().WithScheme(scheme).Build()

		fakeChartScanner = &fakes.FakeChartScanner{}
		fakeChartClient = &fakes.FakeChartClient{}
		s = &ProfilesServer{
			kubeClient:        kubeClient,
			log:               testutils.MakeFakeLogr(),
			helmRepoName:      "helmrepo",
			helmRepoNamespace: "default",
			chartScanner:      fakeChartScanner,
			chartClient:       fakeChartClient,
		}

		helmRepo = &sourcev1beta1.HelmRepository{
			TypeMeta: metav1.TypeMeta{
				Kind:       sourcev1beta1.HelmRepositoryKind,
				APIVersion: sourcev1beta1.GroupVersion.Identifier(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "helmrepo",
				Namespace: "default",
			},
			Spec: sourcev1beta1.HelmRepositorySpec{
				URL:      "example.com/charts",
				Interval: metav1.Duration{Duration: time.Minute * 10},
			},
			Status: sourcev1beta1.HelmRepositoryStatus{
				URL: "example.com/index.yaml",
			},
		}
	})

	Describe("GetProfiles", func() {
		When("the HelmRepository exists", func() {
			BeforeEach(func() {
				Expect(kubeClient.Create(context.TODO(), helmRepo)).To(Succeed())
			})

			It("Returns a list of profiles in the HelmRepository", func() {
				profiles := []*pb.Profile{
					{
						Name: profileName,
					},
				}
				fakeChartScanner.ScanChartsReturns(profiles, nil)

				profilesResp, err := s.GetProfiles(context.TODO(), &pb.GetProfilesRequest{})
				Expect(err).NotTo(HaveOccurred())
				Expect(profilesResp).NotTo(BeNil())
				Expect(profilesResp.Profiles).To(Equal(profiles))
				Expect(fakeChartScanner.ScanChartsCallCount()).To(Equal(1))
				_, helmRepoArg, _ := fakeChartScanner.ScanChartsArgsForCall(0)
				Expect(helmRepoArg.Name).To(Equal(helmRepo.Name))
				Expect(helmRepoArg.Namespace).To(Equal(helmRepo.Namespace))
			})

			When("scanning for helmcharts errors", func() {
				It("errors", func() {
					fakeChartScanner.ScanChartsReturns(nil, fmt.Errorf("foo"))
					_, err := s.GetProfiles(context.TODO(), &pb.GetProfilesRequest{})
					Expect(err).To(MatchError("failed to scan HelmRepository \"default\"/\"helmrepo\" for charts: foo"))
				})
			})
		})

		When("the HelmRepository doesn't exist", func() {
			It("errors", func() {
				_, err := s.GetProfiles(context.TODO(), &pb.GetProfilesRequest{})
				Expect(err).To(MatchError("cannot find HelmRepository \"default\"/\"helmrepo\""))
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
			})

			When("the Helm cache updates", func() {
				When("it retrieves the values file from Helm chart", func() {
					When("header does not contain 'application/octet-stream'", func() {
						It("returns a values file in base64-encoded json", func() {
							fakeChartClient.UpdateCacheReturns(nil)
							fakeChartClient.FileFromChartReturns([]byte("values"), nil)

							resp, err := s.GetProfileValues(context.TODO(), &pb.GetProfileValuesRequest{
								ProfileName:    profileName,
								ProfileVersion: profileVersion,
							})
							Expect(err).NotTo(HaveOccurred())
							Expect(resp.ContentType).To(Equal(jsonType))
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
							fakeChartClient.UpdateCacheReturns(nil)
							fakeChartClient.FileFromChartReturns([]byte("values"), nil)
							ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("accept", octetStreamType))

							resp, err := s.GetProfileValues(ctx, &pb.GetProfileValuesRequest{
								ProfileName:    profileName,
								ProfileVersion: profileVersion,
							})
							Expect(err).NotTo(HaveOccurred())
							Expect(resp.ContentType).To(Equal(octetStreamType))
							Expect(string(resp.Data)).To(Equal("values"))
						})
					})
				})

				When("it cannot retrieve the values file from Helm chart", func() {
					It("errors", func() {
						fakeChartClient.UpdateCacheReturns(nil)
						fakeChartClient.FileFromChartReturns([]byte{}, fmt.Errorf("err"))
						_, err := s.GetProfileValues(context.TODO(), &pb.GetProfileValuesRequest{
							ProfileName:    profileName,
							ProfileVersion: profileVersion,
						})
						Expect(err).To(MatchError("failed to retrieve values file from Helm chart 'observability' (latest): err"))
					})
				})
			})

			When("the Helm cache fails to update", func() {
				It("errors", func() {
					fakeChartClient.UpdateCacheReturns(fmt.Errorf("err"))
					_, err := s.GetProfileValues(context.TODO(), &pb.GetProfileValuesRequest{})
					Expect(err).To(MatchError("failed to update Helm cache: err"))
				})
			})
		})

		When("the HelmRepository doesn't exist", func() {
			It("errors", func() {
				_, err := s.GetProfileValues(context.TODO(), &pb.GetProfileValuesRequest{})
				Expect(err).To(MatchError("cannot find HelmRepository \"default\"/\"helmrepo\""))
				Expect(err).To(BeAssignableToTypeOf(&grpcruntime.HTTPStatusError{}))
				//TODO why do we return 200 when the HelmRepository doesn't exist
				Expect(err.(*grpcruntime.HTTPStatusError).HTTPStatus).To(Equal(http.StatusOK))
			})
		})
	})
})
