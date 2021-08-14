package app_test

import (
	"fmt"
	"testing"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

type testOsys struct {
	osys.Osys
	userHome string
}

func (o *testOsys) UserHomeDir() (string, error) {
	return o.userHome, nil
}

type actualFluxRunner struct {
	runner.Runner
}

func (r *actualFluxRunner) Run(command string, args ...string) ([]byte, error) {
	cmd := "../../pkg/flux/bin/flux"

	return r.Runner.Run(cmd, args...)
}

var testEnv *envtest.Environment
var k client.Client

var _ = BeforeSuite(func() {
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			"../../manifests/crds",
			"../../tools/testcrds",
		},
	}

	var err error
	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())

	scheme := kube.CreateScheme()

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		ClientDisableCacheFor: []client.Object{
			&wego.Application{},
			&corev1.Namespace{},
			&corev1.Secret{},
			&kustomizev1.Kustomization{},
			&sourcev1.GitRepository{},
		},
		Scheme: scheme,
	})
	Expect(err).NotTo(HaveOccurred())
	go func() {
		if err := k8sManager.Start(ctrl.SetupSignalHandler()); err != nil {
			fmt.Printf("error starting k8s manager: %s", err.Error())
		}
	}()

	k = k8sManager.GetClient()

})

var _ = AfterSuite(func() {
	// TODO check to see if more teardown is required.
	testEnv.Stop()
})

func TestAppIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "App Integration Tests")
}
