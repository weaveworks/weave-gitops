package app

import (
	"context"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev2 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/fluxcd/pkg/apis/meta"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var statusParams StatusParams
var _ = Describe("Status", func() {
	var _ = BeforeEach(func() {
		statusParams = StatusParams{
			Name:      "my-app",
			Namespace: "my-namespace",
		}

		fluxClient.GetAllResourcesStatusStub = func(s1, s2 string) ([]byte, error) {
			return []byte("status"), nil
		}

		kubeClient.GetApplicationStub = func(ctx context.Context, name types.NamespacedName) (*wego.Application, error) {
			return &wego.Application{
				Spec: wego.ApplicationSpec{DeploymentType: wego.DeploymentTypeKustomize},
			}, nil
		}
	})

	It("gets all flux resources status", func() {
		fluxOutput, _, err := appSrv.Status(statusParams)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(fluxOutput).To(Equal("status"))
	})

	Context("last successful reconciliation", func() {
		var (
			t time.Time = time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)

			conditions = []metav1.Condition{
				{
					Type:               meta.ReadyCondition,
					Status:             metav1.ConditionTrue,
					LastTransitionTime: metav1.NewTime(t),
				},
			}
		)

		It("returns when using kustomize", func() {
			kubeClient.GetResourceStub = func(c context.Context, nn types.NamespacedName, r kube.Resource) error {
				kust, ok := r.(*kustomizev2.Kustomization)
				Expect(ok).To(BeTrue())
				kust.Status.Conditions = conditions
				return nil
			}

			_, lastRecon, err := appSrv.Status(statusParams)
			Expect(err).ShouldNot(HaveOccurred())

			_, name, deploymentType := kubeClient.GetResourceArgsForCall(0)
			Expect(name).To(Equal(types.NamespacedName{Name: statusParams.Name, Namespace: statusParams.Namespace}))
			Expect(deploymentType).To(BeAssignableToTypeOf(&kustomizev2.Kustomization{}))

			Expect(lastRecon).To(Equal("2009-11-10 23:00:00 +0000 UTC"))
		})

		It("returns when using helm", func() {
			kubeClient.GetApplicationStub = func(ctx context.Context, name types.NamespacedName) (*wego.Application, error) {
				return &wego.Application{
					Spec: wego.ApplicationSpec{DeploymentType: wego.DeploymentTypeHelm},
				}, nil
			}

			kubeClient.GetResourceStub = func(c context.Context, nn types.NamespacedName, r kube.Resource) error {
				helm, ok := r.(*helmv2.HelmRelease)
				Expect(ok).To(BeTrue())
				helm.Status.Conditions = conditions
				return nil
			}

			_, lastRecon, err := appSrv.Status(statusParams)
			Expect(err).ShouldNot(HaveOccurred())

			_, name, deploymentType := kubeClient.GetResourceArgsForCall(0)
			Expect(name).To(Equal(types.NamespacedName{Name: statusParams.Name, Namespace: statusParams.Namespace}))
			Expect(deploymentType).To(BeAssignableToTypeOf(&helmv2.HelmRelease{}))

			Expect(lastRecon).To(Equal("2009-11-10 23:00:00 +0000 UTC"))
		})

		It("returns safe message when no succesfull reconciliation", func() {
			kubeClient.GetResourceStub = func(c context.Context, nn types.NamespacedName, r kube.Resource) error {
				kust, ok := r.(*kustomizev2.Kustomization)
				Expect(ok).To(BeTrue())
				kust.Status.Conditions = []metav1.Condition{}
				return nil
			}

			_, lastRecon, err := appSrv.Status(statusParams)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(lastRecon).To(Equal("No succesfull reconciliation"))
		})
	})
})
