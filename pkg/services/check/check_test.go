package check

import (
	"context"
	"errors"

	"github.com/weaveworks/weave-gitops/pkg/flux"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"

	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/flux/fluxfakes"
)

var _ = Describe("Check", func() {

	var fakeFluxClient *fluxfakes.FakeFlux
	var fakeKubeClient *kubefakes.FakeKube
	someError := errors.New("some error")
	var context context.Context
	const successfulFluxPreCheckOutput = `► checking prerequisites
✔ Kubernetes 1.21.1 >=1.19.0-0
✔ prerequisites checks passed`
	const expectedFluxVersion = "1.21.1"
	BeforeEach(func() {
		fakeFluxClient = &fluxfakes.FakeFlux{}
		fakeKubeClient = &kubefakes.FakeKube{}
	})

	It("should fail running flux.PreCheck", func() {
		fakeFluxClient.PreCheckReturns("", someError)

		out, err := Pre(context, fakeKubeClient, fakeFluxClient, expectedFluxVersion)
		Expect(err.Error()).To(ContainSubstring(someError.Error()))
		Expect(out).Should(BeEmpty())
	})

	It("should show flux is not installed", func() {

		fakeFluxClient.PreCheckReturns(successfulFluxPreCheckOutput, nil)

		fakeKubeClient.GetNamespacesReturns(&corev1.NamespaceList{
			Items: []corev1.Namespace{},
		}, nil)

		actualOutput, err := Pre(context, fakeKubeClient, fakeFluxClient, expectedFluxVersion)
		Expect(err).ShouldNot(HaveOccurred())

		expectedOutput := `✔ Kubernetes 1.21.1 >=1.19.0-0
✔ Flux is not installed`
		Expect(actualOutput).To(Equal(expectedOutput))
	})

	It("should fail while getting namespaces", func() {

		fakeFluxClient.PreCheckReturns(successfulFluxPreCheckOutput, nil)

		fakeKubeClient.GetNamespacesReturns(nil, someError)

		out, err := Pre(context, fakeKubeClient, fakeFluxClient, expectedFluxVersion)
		Expect(err.Error()).To(ContainSubstring(someError.Error()))
		Expect(out).Should(BeEmpty())
	})

	It("should fail parsing actual version", func() {

		output, err := validateFluxVersion("", "")
		Expect(err).Should(HaveOccurred())
		Expect(output).To(BeEmpty())

	})

	It("should fail parsing expected version", func() {

		output, err := validateFluxVersion("v0.0.0", "")
		Expect(err).Should(HaveOccurred())
		Expect(output).To(BeEmpty())

	})

	It("should fail when there is no Kubernetes when running flux pre check", func() {

		fakeFluxClient.PreCheckReturns("", nil)

		output, err := Pre(context, fakeKubeClient, fakeFluxClient, expectedFluxVersion)
		Expect(err).Should(Equal(ErrKubernetesNotFound))
		Expect(output).To(BeEmpty())

	})

	It("should fail when expected flux version is not valid", func() {

		fakeFluxClient.PreCheckReturns(successfulFluxPreCheckOutput, nil)

		fakeKubeClient.GetNamespacesReturns(&corev1.NamespaceList{
			Items: []corev1.Namespace{
				{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							flux.PartOfLabelKey:  flux.PartOfLabelValue,
							flux.VersionLabelKey: expectedFluxVersion,
						},
					},
				},
			},
		}, nil)

		output, err := Pre(context, fakeKubeClient, fakeFluxClient, "")
		Expect(err).Should(HaveOccurred())
		Expect(output).To(BeEmpty())

	})

	It("should show flux version is compatible", func() {

		fakeFluxClient.PreCheckReturns(successfulFluxPreCheckOutput, nil)

		fakeKubeClient.GetNamespacesReturns(&corev1.NamespaceList{
			Items: []corev1.Namespace{
				{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							flux.PartOfLabelKey:  flux.PartOfLabelValue,
							flux.VersionLabelKey: expectedFluxVersion,
						},
					},
				},
			},
		}, nil)

		actualOutput, err := Pre(context, fakeKubeClient, fakeFluxClient, expectedFluxVersion)
		Expect(err).ShouldNot(HaveOccurred())

		expectedOutput := `✔ Kubernetes 1.21.1 >=1.19.0-0
✔ Flux 1.21.1 ~=1.21.1
` + FluxCompatibleMessage
		Expect(actualOutput).To(Equal(expectedOutput))
	})

	It("should show that the current flux version is not compatible", func() {

		fakeFluxClient.PreCheckReturns(successfulFluxPreCheckOutput, nil)

		differentFluxVersion := "v0.0.0"

		fakeKubeClient.GetNamespacesReturns(&corev1.NamespaceList{
			Items: []corev1.Namespace{
				{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app.kubernetes.io/part-of": "flux",
							"app.kubernetes.io/version": differentFluxVersion,
						},
					},
				},
			},
		}, nil)

		actualOutput, err := Pre(context, fakeKubeClient, fakeFluxClient, expectedFluxVersion)
		Expect(err).ShouldNot(HaveOccurred())

		expectedOutput := `✔ Kubernetes 1.21.1 >=1.19.0-0
✗ Flux 0.0.0 !=~1.21.x
` + FluxNotCompatibleMessage
		Expect(actualOutput).To(Equal(expectedOutput))
	})

})

//var _ = Describe("validateFluxVersion", func() {
//	It("should XYZ", func() {
//
//		constrain, err := semver.NewConstraint("~0.21.x")
//		Expect(err).ShouldNot(HaveOccurred())
//
//		v0, err := semver.NewVersion("0.19.987")
//		Expect(err).ShouldNot(HaveOccurred())
//
//		b := constrain.Check(v0)
//		fmt.Println("B", b)
//
//		// ~0.21.x 0.21.3   true
//		// ~0.21.x 0.22.0   false
//		// ~0.21.x 0.19.987 false
//
//		//v1, err := semver.NewVersion("v0.21.3")
//		//Expect(err).ShouldNot(HaveOccurred())
//
//		// v0.21.0 v0.22.3 -1 => v0.21.0 is smaller than v0.22.3
//		// v0.21.3 v0.21      => v0.21.3 is larger than v0.21
//		// v0.21 v0.21.3      => v0.21   is smaller than v0.21.3
//
//		//r := v0.Compare(v1)
//		//Expect(r).To(Equal("2"))
//
//		//str, err := validateFluxVersion("v0.21.1", "v0.21.2")
//		//Expect(err).ShouldNot(HaveOccurred())
//		//fmt.Println("str", str)
//	})
//})
