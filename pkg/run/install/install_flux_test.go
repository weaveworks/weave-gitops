package install

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/core/server"
	coretypes "github.com/weaveworks/weave-gitops/core/server/types"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/run"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	testVersion = "test-version"
)

var _ = Describe("GetFluxVersion", func() {
	var fakeLogger logger.Logger

	BeforeEach(func() {
		fakeLogger = logger.From(logr.Discard())
	})

	It("guess flux version", func() {
		kubeClientOpts := run.GetKubeClientOptions()

		contextName := "test-context"

		kubeClient, err := run.GetKubeClient(fakeLogger, contextName, k8sEnv.Rest, kubeClientOpts)
		Expect(err).NotTo(HaveOccurred())

		ctx := context.Background()

		fluxNs := &v1.Namespace{}
		fluxNs.Name = "flux-system"
		fluxNs.Labels = map[string]string{
			coretypes.PartOfLabel: server.FluxNamespacePartOf,
		}

		Expect(kubeClient.Create(ctx, fluxNs)).To(Succeed())

		source := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "source-controller",
				Namespace: "flux-system",
				Labels: map[string]string{
					"app": "source-controller",
				},
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "source-controller",
					},
				},
				Template: v1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": "source-controller",
						},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name:  "source-controller",
								Image: "ghcr.io/fluxcd/source-controller:v0.33.0",
							},
						},
					},
				},
			},
		}
		Expect(kubeClient.Create(ctx, source)).To(Succeed())

		fluxVersionInfo, guessed, err := GetFluxVersion(ctx, fakeLogger, kubeClient)

		Expect(err).NotTo(HaveOccurred())
		Expect(fluxVersionInfo.FluxVersion).To(Equal("v0.38.0"))
		Expect(fluxVersionInfo.FluxNamespace).To(Equal("flux-system"))
		Expect(guessed).To(BeTrue())

		Eventually(kubeClient.Delete(ctx, fluxNs)).ProbeEvery(1 * time.Second).Should(Succeed())
	})

	It("gets flux version", func() {
		kubeClientOpts := run.GetKubeClientOptions()

		contextName := "test-context"

		kubeClient, err := run.GetKubeClient(fakeLogger, contextName, k8sEnv.Rest, kubeClientOpts)
		Expect(err).NotTo(HaveOccurred())

		ctx := context.Background()

		fluxNs := &v1.Namespace{}
		fluxNs.Name = "flux-ns-test"
		fluxNs.Labels = map[string]string{
			coretypes.PartOfLabel: server.FluxNamespacePartOf,
			flux.VersionLabelKey:  testVersion,
		}

		Expect(kubeClient.Create(ctx, fluxNs)).To(Succeed())

		fluxVersionInfo, guessed, err := GetFluxVersion(ctx, fakeLogger, kubeClient)

		Expect(err).NotTo(HaveOccurred())
		Expect(fluxVersionInfo.FluxVersion).To(Equal(testVersion))
		Expect(fluxVersionInfo.FluxNamespace).To(Equal(fluxNs.Name))
		Expect(guessed).To(BeFalse())

		Eventually(kubeClient.Delete(ctx, fluxNs)).ProbeEvery(1 * time.Second).Should(Succeed())
	})

})
