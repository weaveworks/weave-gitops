package run

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/logger/loggerfakes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

// mock controller-runtime client
type mockClientForGetDashboardHelmChart struct {
	client.Client
	state stateGetDashboardHelmChart
}

type stateGetDashboardHelmChart string

const (
	testDashboardName = "ww-gitops"
	testNamespace     = "test-namespace"
	testAdminUser     = "testUser"
	testSecret        = "test-secret"

	stateGetDashboardHelmChartGetReturnErr stateGetDashboardHelmChart = "get-return-err"

	getDashboardErrorMsg = "get dashboard error"
)

var _ = Describe("InstallDashboard", func() {
	var fakeLogger *loggerfakes.FakeLogger
	var fakeContext context.Context

	BeforeEach(func() {
		fakeLogger = &loggerfakes.FakeLogger{}
		fakeContext = context.Background()
	})

	It("should install dashboard successfully", func() {
		man := &mockResourceManagerForApply{}

		manifests, err := CreateDashboardObjects(fakeLogger, testDashboardName, testNamespace, testAdminUser, testSecret, "3.0.0")
		Expect(err).NotTo(HaveOccurred())

		err = InstallDashboard(fakeLogger, fakeContext, man, manifests)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should return an apply all error if the resource manager returns an apply all error", func() {
		man := &mockResourceManagerForApply{state: stateApplyAllReturnErr}

		manifests, err := CreateDashboardObjects(fakeLogger, testDashboardName, testNamespace, testAdminUser, testSecret, "3.0.0")
		Expect(err).NotTo(HaveOccurred())

		err = InstallDashboard(fakeLogger, fakeContext, man, manifests)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal(applyAllErrorMsg))
	})
})

func (man *mockClientForGetDashboardHelmChart) Get(_ context.Context, key client.ObjectKey, obj client.Object) error {
	switch man.state {
	case stateGetDashboardHelmChartGetReturnErr:
		return errors.New(getDashboardErrorMsg)

	default:
		switch obj := obj.(type) {
		case *sourcev1.HelmChart:
			helmChart := sourcev1.HelmChart{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test-namespace",
					Name:      "test-namespace-ww-gitops",
				},
			}
			helmChart.DeepCopyInto(obj)
		}
	}

	return nil
}

var _ = Describe("getDashboardHelmChart", func() {
	var fakeLogger *loggerfakes.FakeLogger
	var fakeContext context.Context

	BeforeEach(func() {
		fakeLogger = &loggerfakes.FakeLogger{}
		fakeContext = context.Background()
	})

	It("returns the dashboard helmchart if there is no error when getting the helmchart", func() {
		helmChart := getDashboardHelmChart(fakeLogger, fakeContext, &mockClientForGetDashboardHelmChart{}, testDashboardName, testNamespace)
		Expect(helmChart).ToNot(BeNil())
		Expect(helmChart.Namespace).To(Equal("test-namespace"))
		Expect(helmChart.Name).To(Equal("test-namespace-ww-gitops"))
	})

	It("returns nil if there is an error when getting the helmchart", func() {
		helmChart := getDashboardHelmChart(fakeLogger, fakeContext, &mockClientForGetDashboardHelmChart{state: stateGetDashboardHelmChartGetReturnErr}, testDashboardName, testNamespace)
		Expect(helmChart).To(BeNil())
	})
})

var _ = Describe("generateManifestsForDashboard", func() {
	var fakeLogger *loggerfakes.FakeLogger

	BeforeEach(func() {
		fakeLogger = &loggerfakes.FakeLogger{}
	})

	It("generates manifests successfully", func() {
		helmRepository := makeHelmRepository(testDashboardName, testNamespace)

		helmChartVersion := "3.0.0"
		helmRelease, err := makeHelmRelease(fakeLogger, testDashboardName, testNamespace, testAdminUser, testSecret, helmChartVersion)
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
		Expect(chart.Spec.Chart).To(Equal(helmChartName))
		Expect(chart.Spec.SourceRef.Name).To(Equal(testDashboardName))
		Expect(chart.Spec.Version).To(Equal(helmChartVersion))
	})
})

var _ = Describe("makeHelmRelease", func() {
	var fakeLogger *loggerfakes.FakeLogger

	BeforeEach(func() {
		fakeLogger = &loggerfakes.FakeLogger{}
	})

	It("creates helmrelease with chart version and values successfully", func() {
		helmChartVersion := "3.0.0"
		actual, err := makeHelmRelease(fakeLogger, testDashboardName, testNamespace, testAdminUser, testSecret, helmChartVersion)
		Expect(err).NotTo(HaveOccurred())
		Expect(actual.Kind).To(Equal(helmv2.HelmReleaseKind))
		Expect(actual.APIVersion).To(Equal(helmv2.GroupVersion.Identifier()))
		Expect(actual.Name).To(Equal(testDashboardName))
		Expect(actual.Namespace).To(Equal(testNamespace))

		Expect(actual.Spec.Interval.Duration).To(Equal(60 * time.Minute))

		chart := actual.Spec.Chart
		Expect(chart.Spec.Chart).To(Equal(helmChartName))
		Expect(chart.Spec.SourceRef.Name).To(Equal(testDashboardName))
		Expect(chart.Spec.SourceRef.Kind).To(Equal("HelmRepository"))
		Expect(chart.Spec.Version).To(Equal(helmChartVersion))

		values := map[string]interface{}{}
		err = json.Unmarshal(actual.Spec.Values.Raw, &values)
		Expect(err).NotTo(HaveOccurred())

		adminUser := values["adminUser"].(map[string]interface{})
		Expect(adminUser["create"]).To(BeTrue())
		Expect(adminUser["username"]).To(Equal(testAdminUser))
		Expect(adminUser["passwordHash"]).To(Equal(testSecret))
	})

	It("creates helmrelease without chart version successfully", func() {
		actual, err := makeHelmRelease(fakeLogger, testDashboardName, testNamespace, testAdminUser, testSecret, "")
		Expect(err).NotTo(HaveOccurred())

		chart := actual.Spec.Chart
		Expect(chart.Spec.Version).To(BeEmpty())
	})

	It("does not add values to helmrelease if both username and secret are empty successfully", func() {
		actual, err := makeHelmRelease(fakeLogger, testDashboardName, testNamespace, "", "", "3.0.0")
		Expect(err).NotTo(HaveOccurred())
		Expect(actual.Spec.Values).To(BeNil())
	})
})

var _ = Describe("makeHelmRepository", func() {
	It("creates helmrepository successfully", func() {
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
	It("creates values successfully", func() {
		values, err := makeValues(testAdminUser, testSecret)
		Expect(err).NotTo(HaveOccurred())

		actual := map[string]interface{}{}
		err = json.Unmarshal(values, &actual)
		Expect(err).NotTo(HaveOccurred())

		adminUser := actual["adminUser"].(map[string]interface{})
		Expect(adminUser["create"]).To(BeTrue())
		Expect(adminUser["username"]).To(Equal(testAdminUser))
		Expect(adminUser["passwordHash"]).To(Equal(testSecret))
	})
})

var _ = Describe("sanitizeResourceData", func() {
	var fakeLogger *loggerfakes.FakeLogger

	BeforeEach(func() {
		fakeLogger = &loggerfakes.FakeLogger{}
	})

	It("sanitizes helmrepository data", func() {
		helmRepository := makeHelmRepository(testDashboardName, testNamespace)

		resData, err := yaml.Marshal(helmRepository)
		Expect(err).NotTo(HaveOccurred())

		resStr := string(resData)
		Expect(strings.Contains(resStr, "status")).To(BeTrue())
		Expect(strings.Contains(resStr, "creationTimestamp")).To(BeTrue())

		sanitizedResData, err := sanitizeResourceData(fakeLogger, resData)
		Expect(err).NotTo(HaveOccurred())

		sanitizedResStr := string(sanitizedResData)
		Expect(strings.Contains(sanitizedResStr, "status")).To(BeFalse())
		Expect(strings.Contains(sanitizedResStr, "creationTimestamp")).To(BeFalse())
	})

	It("sanitizes helmrelease data", func() {
		helmChartVersion := "3.0.0"
		helmRelease, err := makeHelmRelease(fakeLogger, testDashboardName, testNamespace, testAdminUser, testSecret, helmChartVersion)
		Expect(err).NotTo(HaveOccurred())

		resData, err := yaml.Marshal(helmRelease)
		Expect(err).NotTo(HaveOccurred())

		resStr := string(resData)
		Expect(strings.Contains(resStr, "status")).To(BeTrue())
		Expect(strings.Contains(resStr, "creationTimestamp")).To(BeTrue())

		sanitizedResData, err := sanitizeResourceData(fakeLogger, resData)
		Expect(err).NotTo(HaveOccurred())

		sanitizedResStr := string(sanitizedResData)
		Expect(strings.Contains(sanitizedResStr, "status")).To(BeFalse())
		Expect(strings.Contains(sanitizedResStr, "creationTimestamp")).To(BeFalse())
	})
})
