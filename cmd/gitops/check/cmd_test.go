package check

import (
	"context"
	"errors"
	"io"
	"strings"

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
	var output io.Writer
	const successfulFluxPreCheckOutput = `► checking prerequisites
✔ Kubernetes 1.21.1 >=1.19.0-0
✔ prerequisites checks passed`
	BeforeEach(func() {
		fakeFluxClient = &fluxfakes.FakeFlux{}
		fakeKubeClient = &kubefakes.FakeKube{}
		output = new(strings.Builder)
	})

	It("should fail running flux.PreCheck", func() {
		fakeFluxClient.PreCheckReturns("", someError)

		_, err := getCurrentFluxVersion(output, context, fakeFluxClient, fakeKubeClient)
		Expect(err.Error()).To(ContainSubstring(someError.Error()))
	})

	It("should fail while getting namespaces", func() {

		fakeFluxClient.PreCheckReturns(successfulFluxPreCheckOutput, nil)

		fakeKubeClient.GetNamespacesReturns(nil, someError)

		_, err := getCurrentFluxVersion(output, context, fakeFluxClient, fakeKubeClient)
		Expect(err.Error()).To(ContainSubstring(someError.Error()))
	})

	It("should return flux is not installed", func() {

		fakeFluxClient.PreCheckReturns(successfulFluxPreCheckOutput, nil)

		fakeKubeClient.GetNamespacesReturns(&corev1.NamespaceList{}, nil)

		_, err := getCurrentFluxVersion(output, context, fakeFluxClient, fakeKubeClient)
		Expect(err).To(Equal(ErrFluxNotFound))
	})

	It("should succeed and show current flux version is valid", func() {

		fakeFluxClient.PreCheckReturns(successfulFluxPreCheckOutput, nil)

		expectedFluxVersion := "v0.21.0"

		fakeKubeClient.GetNamespacesReturns(&corev1.NamespaceList{
			Items: []corev1.Namespace{
				{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app.kubernetes.io/part-of": "flux",
							"app.kubernetes.io/version": expectedFluxVersion,
						},
					},
				},
			},
		}, nil)

		actualFluxVersion, err := getCurrentFluxVersion(output, context, fakeFluxClient, fakeKubeClient)
		Expect(err).ToNot(HaveOccurred())
		Expect(actualFluxVersion).To(Equal(expectedFluxVersion))
	})

})
