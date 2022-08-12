package run

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/logger/loggerfakes"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
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
	secret    = "test-secret"
	namespace = "test-namespace"

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

		err := InstallDashboard(fakeLogger, fakeContext, man, namespace, "admin", secret, "3.0.0")
		Expect(err).NotTo(HaveOccurred())
	})

	It("should return an apply all error if the resource manager returns an apply all error", func() {
		man := &mockResourceManagerForApply{state: stateApplyAllReturnErr}

		err := InstallDashboard(fakeLogger, fakeContext, man, namespace, "admin", secret, "3.0.0")
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
		helmChart := getDashboardHelmChart(fakeLogger, fakeContext, &mockClientForGetDashboardHelmChart{}, namespace)
		Expect(helmChart).ToNot(BeNil())
		Expect(helmChart.Namespace).To(Equal("test-namespace"))
		Expect(helmChart.Name).To(Equal("test-namespace-ww-gitops"))
	})

	It("returns nil if there is an error when getting the helmchart", func() {
		helmChart := getDashboardHelmChart(fakeLogger, fakeContext, &mockClientForGetDashboardHelmChart{state: stateGetDashboardHelmChartGetReturnErr}, namespace)
		Expect(helmChart).To(BeNil())
	})
})

var _ = Describe("generateManifestsForDashboard", func() {
	var fakeLogger *loggerfakes.FakeLogger

	BeforeEach(func() {
		fakeLogger = &loggerfakes.FakeLogger{}
	})

	It("generates manifests successfully", func() {
		valuesMap := map[string]interface{}{
			"adminUser": map[string]interface{}{
				"create":       true,
				"passwordHash": "test-secret",
				"username":     "admin",
			},
		}

		helmRepository := &sourcev1.HelmRepository{
			TypeMeta: metav1.TypeMeta{
				Kind:       sourcev1.HelmRepositoryKind,
				APIVersion: sourcev1.GroupVersion.Identifier(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ww-gitops",
				Namespace: "test-namespace",
			},
			Spec: sourcev1.HelmRepositorySpec{
				URL: "https://helm.gitops.weave.works",
				Interval: metav1.Duration{
					Duration: time.Minute,
				},
			},
		}

		values, err := json.Marshal(valuesMap)
		Expect(err).NotTo(HaveOccurred())

		helmRelease := &helmv2.HelmRelease{
			TypeMeta: metav1.TypeMeta{
				Kind:       helmv2.HelmReleaseKind,
				APIVersion: helmv2.GroupVersion.Identifier(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ww-gitops",
				Namespace: "test-namespace",
			},
			Spec: helmv2.HelmReleaseSpec{
				Interval: metav1.Duration{
					Duration: time.Minute,
				},
				Chart: helmv2.HelmChartTemplate{
					Spec: helmv2.HelmChartTemplateSpec{
						Chart:   "weave-gitops",
						Version: "3.0.0",
						SourceRef: helmv2.CrossNamespaceObjectReference{
							Kind: "HelmRepository",
							Name: "ww-gitops",
						},
						ReconcileStrategy: "ChartVersion",
					},
				},
				Values: &v1.JSON{Raw: values},
			},
		}

		expectedHelmRepository, err := yaml.Marshal(helmRepository)
		Expect(err).NotTo(HaveOccurred())

		expectedHelmRelease, err := yaml.Marshal(helmRelease)
		Expect(err).NotTo(HaveOccurred())

		divider := []byte("---\n")

		expected := append(divider, expectedHelmRepository...)
		expected = append(expected, divider...)
		expected = append(expected, expectedHelmRelease...)

		actual, err := generateManifestsForDashboard(fakeLogger, secret, helmRepository, helmRelease)
		Expect(err).NotTo(HaveOccurred())
		Expect(actual).To(Equal(expected))
	})
})

var _ = Describe("makeHelmRelease", func() {
	var fakeLogger *loggerfakes.FakeLogger

	BeforeEach(func() {
		fakeLogger = &loggerfakes.FakeLogger{}
	})

	It("creates helmrelease successfully", func() {
		valuesMap := map[string]interface{}{
			"adminUser": map[string]interface{}{
				"create":       true,
				"passwordHash": "test-secret",
				"username":     "admin",
			},
		}

		values, err := json.Marshal(valuesMap)
		Expect(err).NotTo(HaveOccurred())

		helmRelease := &helmv2.HelmRelease{
			TypeMeta: metav1.TypeMeta{
				Kind:       helmv2.HelmReleaseKind,
				APIVersion: helmv2.GroupVersion.Identifier(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ww-gitops",
				Namespace: "test-namespace",
			},
			Spec: helmv2.HelmReleaseSpec{
				Interval: metav1.Duration{
					Duration: time.Minute,
				},
				Chart: helmv2.HelmChartTemplate{
					Spec: helmv2.HelmChartTemplateSpec{
						Chart:   "weave-gitops",
						Version: "3.0.0",
						SourceRef: helmv2.CrossNamespaceObjectReference{
							Kind: "HelmRepository",
							Name: "ww-gitops",
						},
						ReconcileStrategy: "ChartVersion",
					},
				},
				Values: &v1.JSON{Raw: values},
			},
		}

		expected, err := json.Marshal(helmRelease)
		Expect(err).NotTo(HaveOccurred())

		actualHelmRelease, err := makeHelmRelease(fakeLogger, namespace, "admin", secret, "3.0.0")
		Expect(err).NotTo(HaveOccurred())

		actual, err := json.Marshal(actualHelmRelease)
		Expect(err).NotTo(HaveOccurred())
		Expect(actual).To(Equal(expected))
	})
})

var _ = Describe("makeHelmRepository", func() {
	It("creates helmrepository successfully", func() {
		helmRepository := &sourcev1.HelmRepository{
			TypeMeta: metav1.TypeMeta{
				Kind:       sourcev1.HelmRepositoryKind,
				APIVersion: sourcev1.GroupVersion.Identifier(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ww-gitops",
				Namespace: "test-namespace",
			},
			Spec: sourcev1.HelmRepositorySpec{
				URL: "https://helm.gitops.weave.works",
				Interval: metav1.Duration{
					Duration: time.Minute,
				},
			},
		}

		expected, err := json.Marshal(helmRepository)
		Expect(err).NotTo(HaveOccurred())

		actual, err := json.Marshal(makeHelmRepository(namespace))
		Expect(err).NotTo(HaveOccurred())
		Expect(actual).To(Equal(expected))
	})
})

var _ = Describe("makeValues", func() {
	It("creates values successfully", func() {
		valuesMap := map[string]interface{}{
			"adminUser": map[string]interface{}{
				"create":       true,
				"passwordHash": "test-secret",
				"username":     "testUser",
			},
		}

		expected, err := json.Marshal(valuesMap)
		Expect(err).NotTo(HaveOccurred())

		actual, err := makeValues("testUser", secret)
		Expect(err).NotTo(HaveOccurred())
		Expect(actual).To(Equal(expected))
	})
})
