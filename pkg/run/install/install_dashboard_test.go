package install

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	coretypes "github.com/weaveworks/weave-gitops/core/server/types"
	"github.com/weaveworks/weave-gitops/pkg/config"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/yaml"
)

const (
	testDashboardName = "ww-gitops"
	testNamespace     = "test-namespace"
	testAdminUser     = "testUser"
	testPasswordHash  = "test-password-hash"
	testUserID        = "abcdefgh90"
	helmChartVersion  = "3.0.0"

	objectCreationErrorMsg = " \"\" is invalid: metadata.name: Required value: name is required"
)

var helmReleaseFixtures = []runtime.Object{
	&helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespace,
			Name:      "dashboard-1",
		},
	},
	&helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespace,
			Name:      "dashboard-2",
		},
		Spec: helmv2.HelmReleaseSpec{
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart: ossDashboardHelmChartName,
				},
			},
		},
	},
	&helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespace,
			Name:      "dashboard-3",
		},
		Spec: helmv2.HelmReleaseSpec{
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart: enterpriseDashboardHelmChartName,
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Name: enterpriseDashboardHelmRepositoryName,
					},
				},
			},
		},
	},
	&helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespace,
			Name:      "dashboard-4",
		},
	},
}

var deploymentFixtures = []runtime.Object{
	&appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespace,
			Name:      "dashboard-1",
		},
	}, &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespace,
			Name:      "dashboard-2",
			Labels: map[string]string{
				coretypes.PartOfLabel:       dashboardPartOfName,
				coretypes.DashboardAppLabel: ossDashboardAppName,
			},
		},
	}, &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespace,
			Name:      "dashboard-3",
			Labels: map[string]string{
				coretypes.PartOfLabel:       dashboardPartOfName,
				coretypes.DashboardAppLabel: enterpriseDashboardAppName,
			},
		},
	}, &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespace,
			Name:      "dashboard-4",
		},
	},
}

type ErroringFakeClient struct {
	client.Client
}

func (p *ErroringFakeClient) List(_ context.Context, _ client.ObjectList, _ ...client.ListOption) error {
	return errors.New("error listing objects")
}

var _ = Describe("InstallDashboard", func() {
	var fakeContext context.Context
	var fakeLogger logger.Logger
	var fakeClient client.WithWatch

	BeforeEach(func() {
		fakeContext = context.Background()
		fakeLogger = logger.From(logr.Discard())
		scheme, err := kube.CreateScheme()
		Expect(err).NotTo(HaveOccurred())
		fakeClient = fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(helmReleaseFixtures...).Build()
	})

	It("should install dashboard successfully", func() {
		manifests, err := CreateDashboardObjects(fakeLogger, testDashboardName, testNamespace, testAdminUser, testPasswordHash, helmChartVersion, "")
		Expect(err).NotTo(HaveOccurred())

		err = InstallDashboard(fakeContext, fakeLogger, fakeClient, manifests)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should return an apply all error if the resource manager returns an apply all error", func() {
		manifests := &DashboardObjects{
			Manifests:      []byte{},
			HelmRepository: &sourcev1.HelmRepository{},
			HelmRelease:    &helmv2.HelmRelease{},
		}
		err := InstallDashboard(fakeContext, fakeLogger, fakeClient, manifests)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal(objectCreationErrorMsg))
	})
})

var _ = Describe("GetInstalledDashboard", func() {
	var (
		fakeContext                context.Context
		fakeClientWithHelmReleases client.WithWatch
		fakeClientWithDeployments  client.WithWatch
		blankClient                client.WithWatch
		errorClient                ErroringFakeClient
	)

	BeforeEach(func() {
		fakeContext = context.Background()
		scheme, err := kube.CreateScheme()
		Expect(err).NotTo(HaveOccurred())

		fakeClientWithHelmReleases = fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(helmReleaseFixtures...).Build()
		fakeClientWithDeployments = fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(deploymentFixtures...).Build()
		blankClient = fake.NewClientBuilder().WithScheme(scheme).Build()
		errorClient = ErroringFakeClient{}
	})

	It("returns the oss dashboard type if the dashboard is installed with a helmrelease", func() {
		dashboardType, dashboardName, err := GetInstalledDashboard(fakeContext, fakeClientWithHelmReleases, testNamespace, map[DashboardType]bool{
			DashboardTypeOSS: true,
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(dashboardType).To(Equal(DashboardTypeOSS))
		Expect(dashboardName).To(Equal("dashboard-2"))
	})

	It("returns the enterprise dashboard type if the dashboard is installed with a helmrelease", func() {
		dashboardType, dashboardName, err := GetInstalledDashboard(fakeContext, fakeClientWithHelmReleases, testNamespace, map[DashboardType]bool{
			DashboardTypeEnterprise: true,
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(dashboardType).To(Equal(DashboardTypeEnterprise))
		Expect(dashboardName).To(Equal("dashboard-3"))
	})

	It("returns the enterprise dashboard type if both dashboards are installed with a helmrelease", func() {
		dashboardType, dashboardName, err := GetInstalledDashboard(fakeContext, fakeClientWithHelmReleases, testNamespace, map[DashboardType]bool{
			DashboardTypeOSS:        true,
			DashboardTypeEnterprise: true,
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(dashboardType).To(Equal(DashboardTypeEnterprise))
		Expect(dashboardName).To(Equal("dashboard-3"))
	})

	It("returns the oss dashboard type if the dashboard is installed with a deployment only", func() {
		dashboardType, dashboardName, err := GetInstalledDashboard(fakeContext, fakeClientWithDeployments, testNamespace, map[DashboardType]bool{
			DashboardTypeOSS: true,
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(dashboardType).To(Equal(DashboardTypeOSS))
		Expect(dashboardName).To(BeEmpty())
	})

	It("returns the enterprise dashboard type if the dashboard is installed with a deployment only", func() {
		dashboardType, dashboardName, err := GetInstalledDashboard(fakeContext, fakeClientWithDeployments, testNamespace, map[DashboardType]bool{
			DashboardTypeEnterprise: true,
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(dashboardType).To(Equal(DashboardTypeEnterprise))
		Expect(dashboardName).To(BeEmpty())
	})

	It("returns the enterprise dashboard type if both dashboards are installed with a deployment only", func() {
		dashboardType, dashboardName, err := GetInstalledDashboard(fakeContext, fakeClientWithDeployments, testNamespace, map[DashboardType]bool{
			DashboardTypeOSS:        true,
			DashboardTypeEnterprise: true,
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(dashboardType).To(Equal(DashboardTypeEnterprise))
		Expect(dashboardName).To(BeEmpty())
	})

	It("returns none dashboard type with no error if no objects are returned", func() {
		dashboardType, dashboardName, err := GetInstalledDashboard(fakeContext, blankClient, testNamespace, map[DashboardType]bool{
			DashboardTypeOSS:        true,
			DashboardTypeEnterprise: true,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(dashboardType).To(Equal(DashboardTypeNone))
		Expect(dashboardName).To(BeEmpty())
	})

	It("returns none dashboard type if there is an error when listing objects", func() {
		dashboardType, dashboardName, err := GetInstalledDashboard(fakeContext, &errorClient, testNamespace, map[DashboardType]bool{
			DashboardTypeOSS:        true,
			DashboardTypeEnterprise: true,
		})
		Expect(err).To(HaveOccurred())
		Expect(dashboardType).To(Equal(DashboardTypeNone))
		Expect(dashboardName).To(BeEmpty())
	})
})

var _ = Describe("generateManifestsForDashboard", func() {
	var fakeLogger logger.Logger

	BeforeEach(func() {
		fakeLogger = logger.From(logr.Discard())
	})

	It("generates manifests", func() {
		helmRepository := makeHelmRepository(testDashboardName, testNamespace)

		helmRelease, err := makeHelmRelease(fakeLogger, testDashboardName, testNamespace, testAdminUser, testPasswordHash, helmChartVersion, "")
		Expect(err).NotTo(HaveOccurred())

		manifestsData, err := generateManifestsForDashboard(fakeLogger, helmRepository, helmRelease)
		Expect(err).NotTo(HaveOccurred())
		Expect(manifestsData).NotTo(BeNil())

		manifests := strings.Split(string(manifestsData), "---\n")
		Expect(len(manifests)).To(Equal(2))

		var actualHelmRepository sourcev1.HelmRepository
		err = yaml.Unmarshal([]byte(manifests[0]), &actualHelmRepository)
		Expect(err).NotTo(HaveOccurred())
		Expect(actualHelmRepository.Name).To(Equal(testDashboardName))
		Expect(actualHelmRepository.Namespace).To(Equal(testNamespace))

		var actualHelmRelease helmv2.HelmRelease
		err = yaml.Unmarshal([]byte(manifests[1]), &actualHelmRelease)
		Expect(err).NotTo(HaveOccurred())
		Expect(actualHelmRelease.Name).To(Equal(testDashboardName))
		Expect(actualHelmRelease.Namespace).To(Equal(testNamespace))

		Expect(actualHelmRelease.Spec.Interval.Duration).To(Equal(60 * time.Minute))

		chart := actualHelmRelease.Spec.Chart
		Expect(chart.Spec.Chart).To(Equal(ossDashboardHelmChartName))
		Expect(chart.Spec.SourceRef.Name).To(Equal(testDashboardName))
		Expect(chart.Spec.Version).To(Equal(helmChartVersion))
	})
})

var _ = Describe("makeHelmRelease", func() {
	var fakeLogger logger.Logger

	BeforeEach(func() {
		fakeLogger = logger.From(logr.Discard())
	})

	It("creates helmrelease with chart version and values", func() {
		config.SetConfig(&config.GitopsCLIConfig{
			UserID:    testUserID,
			Analytics: true,
		})

		actual, err := makeHelmRelease(fakeLogger, testDashboardName, testNamespace, testAdminUser, testPasswordHash, helmChartVersion, "")
		Expect(err).NotTo(HaveOccurred())
		Expect(actual.Kind).To(Equal(helmv2.HelmReleaseKind))
		Expect(actual.APIVersion).To(Equal(helmv2.GroupVersion.Identifier()))
		Expect(actual.Name).To(Equal(testDashboardName))
		Expect(actual.Namespace).To(Equal(testNamespace))

		Expect(actual.Spec.Interval.Duration).To(Equal(60 * time.Minute))

		chart := actual.Spec.Chart
		Expect(chart.Spec.Chart).To(Equal(ossDashboardHelmChartName))
		Expect(chart.Spec.SourceRef.Name).To(Equal(testDashboardName))
		Expect(chart.Spec.SourceRef.Kind).To(Equal("HelmRepository"))
		Expect(chart.Spec.Version).To(Equal(helmChartVersion))

		values := map[string]interface{}{}
		err = json.Unmarshal(actual.Spec.Values.Raw, &values)
		Expect(err).NotTo(HaveOccurred())

		adminUser := values["adminUser"].(map[string]interface{})
		Expect(adminUser["create"]).To(BeTrue())
		Expect(adminUser["username"]).To(Equal(testAdminUser))
		Expect(adminUser["passwordHash"]).To(Equal(testPasswordHash))

		Expect(values["WEAVE_GITOPS_FEATURE_TELEMETRY"]).To(Equal("true"))
	})

	It("creates helmrelease without chart version", func() {
		actual, err := makeHelmRelease(fakeLogger, testDashboardName, testNamespace, testAdminUser, testPasswordHash, "", "")
		Expect(err).NotTo(HaveOccurred())

		chart := actual.Spec.Chart
		Expect(chart.Spec.Version).To(BeEmpty())
	})

	It("does not add values to helmrelease if username and secret are empty and analytics is off", func() {
		config.SetConfig(&config.GitopsCLIConfig{
			UserID:    testUserID,
			Analytics: false,
		})

		actual, err := makeHelmRelease(fakeLogger, testDashboardName, testNamespace, "", "", helmChartVersion, "")
		Expect(err).NotTo(HaveOccurred())
		Expect(actual.Spec.Values).To(BeNil())
	})

	It("adds only telemetry value to helmrelease if username and secret are empty but analytics is on", func() {
		config.SetConfig(&config.GitopsCLIConfig{
			UserID:    testUserID,
			Analytics: true,
		})

		actual, err := makeHelmRelease(fakeLogger, testDashboardName, testNamespace, "", "", helmChartVersion, "")
		Expect(err).NotTo(HaveOccurred())

		values := map[string]interface{}{}
		err = json.Unmarshal(actual.Spec.Values.Raw, &values)
		Expect(err).NotTo(HaveOccurred())

		Expect(values).NotTo(BeNil())
		Expect(values["adminUser"]).To(BeNil())

		Expect(values["WEAVE_GITOPS_FEATURE_TELEMETRY"]).To(Equal("true"))
	})

	It("does not add telemetry value to helmrelease if analytics is off", func() {
		config.SetConfig(&config.GitopsCLIConfig{
			UserID:    testUserID,
			Analytics: false,
		})

		actual, err := makeHelmRelease(fakeLogger, testDashboardName, testNamespace, testAdminUser, testPasswordHash, helmChartVersion, "")
		Expect(err).NotTo(HaveOccurred())

		values := map[string]interface{}{}
		err = json.Unmarshal(actual.Spec.Values.Raw, &values)
		Expect(err).NotTo(HaveOccurred())

		Expect(values).NotTo(BeNil())

		adminUser := values["adminUser"].(map[string]interface{})
		Expect(adminUser["create"]).To(BeTrue())
		Expect(adminUser["username"]).To(Equal(testAdminUser))
		Expect(adminUser["passwordHash"]).To(Equal(testPasswordHash))

		Expect(values["WEAVE_GITOPS_FEATURE_TELEMETRY"]).To(BeNil())
	})
})

var _ = Describe("makeHelmRepository", func() {
	It("creates helmrepository", func() {
		actual := makeHelmRepository(testDashboardName, testNamespace)
		Expect(actual.Kind).To(Equal(sourcev1.HelmRepositoryKind))
		Expect(actual.APIVersion).To(Equal(sourcev1.GroupVersion.Identifier()))
		Expect(actual.Name).To(Equal(testDashboardName))
		Expect(actual.Namespace).To(Equal(testNamespace))

		labels := actual.Labels
		Expect(labels["app.kubernetes.io/name"]).NotTo(BeEmpty())
		Expect(labels["app.kubernetes.io/component"]).NotTo(BeEmpty())
		Expect(labels["app.kubernetes.io/part-of"]).NotTo(BeEmpty())
		Expect(labels["app.kubernetes.io/created-by"]).NotTo(BeEmpty())

		annotations := actual.Annotations
		Expect(annotations["metadata.weave.works/description"]).NotTo(BeEmpty())

		Expect(actual.Spec.URL).To(Equal(helmRepositoryURL))
		Expect(actual.Spec.Interval.Duration).To(Equal(60 * time.Minute))
	})
})

var _ = Describe("makeValues", func() {
	It("creates all values", func() {
		config.SetConfig(&config.GitopsCLIConfig{
			UserID:    testUserID,
			Analytics: true,
		})

		values, err := makeValues(testAdminUser, testPasswordHash, "")
		Expect(err).NotTo(HaveOccurred())

		actual := map[string]interface{}{}
		err = json.Unmarshal(values, &actual)
		Expect(err).NotTo(HaveOccurred())

		adminUser := actual["adminUser"].(map[string]interface{})
		Expect(adminUser["create"]).To(BeTrue())
		Expect(adminUser["username"]).To(Equal(testAdminUser))
		Expect(adminUser["passwordHash"]).To(Equal(testPasswordHash))

		Expect(actual["WEAVE_GITOPS_FEATURE_TELEMETRY"]).To(Equal("true"))
	})
})

var _ = Describe("SanitizeResourceData", func() {
	var fakeLogger logger.Logger

	BeforeEach(func() {
		fakeLogger = logger.From(logr.Discard())
	})

	It("sanitizes helmrepository data", func() {
		helmRepository := makeHelmRepository(testDashboardName, testNamespace)

		resData, err := yaml.Marshal(helmRepository)
		Expect(err).NotTo(HaveOccurred())

		resStr := string(resData)
		Expect(strings.Contains(resStr, "status")).To(BeTrue())
		Expect(strings.Contains(resStr, "creationTimestamp")).To(BeTrue())

		sanitizedResData, err := SanitizeResourceData(fakeLogger, resData)
		Expect(err).NotTo(HaveOccurred())

		sanitizedResStr := string(sanitizedResData)
		Expect(strings.Contains(sanitizedResStr, "status")).To(BeFalse())
		Expect(strings.Contains(sanitizedResStr, "creationTimestamp")).To(BeFalse())
	})

	It("sanitizes helmrelease data", func() {
		helmRelease, err := makeHelmRelease(fakeLogger, testDashboardName, testNamespace, testAdminUser, testPasswordHash, helmChartVersion, "")
		Expect(err).NotTo(HaveOccurred())

		resData, err := yaml.Marshal(helmRelease)
		Expect(err).NotTo(HaveOccurred())

		resStr := string(resData)
		Expect(strings.Contains(resStr, "status")).To(BeTrue())
		Expect(strings.Contains(resStr, "creationTimestamp")).To(BeTrue())

		sanitizedResData, err := SanitizeResourceData(fakeLogger, resData)
		Expect(err).NotTo(HaveOccurred())

		sanitizedResStr := string(sanitizedResData)
		Expect(strings.Contains(sanitizedResStr, "status")).To(BeFalse())
		Expect(strings.Contains(sanitizedResStr, "creationTimestamp")).To(BeFalse())
	})
})
