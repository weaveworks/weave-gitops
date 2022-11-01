package install

import (
	"context"
	runclient "github.com/fluxcd/pkg/runtime/client"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/core/server"
	coretypes "github.com/weaveworks/weave-gitops/core/server/types"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/run"
	v1 "k8s.io/api/core/v1"
)

const (
	testVersion = "test-version"
)

var _ = Describe("GetFluxVersion", func() {
	var fakeLogger logger.Logger

	BeforeEach(func() {
		fakeLogger = logger.From(logr.Discard())
	})

	It("gets flux version", func() {
		kubeclientOpts := &runclient.Options{
			QPS:   1,
			Burst: 1,
		}

		contextName := "test-context"

		kubeClient, err := run.GetKubeClient(fakeLogger, contextName, k8sEnv.Rest, kubeclientOpts)
		Expect(err).NotTo(HaveOccurred())

		ctx := context.Background()

		fluxNs := &v1.Namespace{}
		fluxNs.Name = "flux-ns-test"
		fluxNs.Labels = map[string]string{
			coretypes.PartOfLabel: server.FluxNamespacePartOf,
			flux.VersionLabelKey:  testVersion,
		}

		Expect(kubeClient.Create(ctx, fluxNs)).To(Succeed())

		fluxVersion, err := GetFluxVersion(fakeLogger, ctx, kubeClient)

		Expect(err).NotTo(HaveOccurred())
		Expect(fluxVersion).To(Equal(testVersion))
	})
})
