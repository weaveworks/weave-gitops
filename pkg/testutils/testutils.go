package testutils

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	fakelogr "github.com/weaveworks/weave-gitops/pkg/vendorfakes/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta1"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
)

type K8sTestEnv struct {
	Client     client.Client
	DynClient  dynamic.Interface
	RestMapper *restmapper.DeferredDiscoveryRESTMapper
	Stop       func()
}

// Starts a local k8s test environment for testing Kubernetes operations such as Create, Get, Delete, etc
func StartK8sTestEnvironment() (*K8sTestEnv, error) {
	testEnv := &envtest.Environment{
		CRDDirectoryPaths: []string{
			"../../manifests/crds",
			"../../tools/testcrds",
		},
	}

	var err error
	cfg, err := testEnv.Start()

	if err != nil {
		return nil, fmt.Errorf("could not start testEnv: %w", err)
	}

	scheme := kube.CreateScheme()

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		ClientDisableCacheFor: []client.Object{
			&wego.Application{},
			&corev1.Namespace{},
			&corev1.Secret{},
			&appsv1.Deployment{},
			&kustomizev1.Kustomization{},
		},
		Scheme: scheme,
	})
	if err != nil {
		return nil, fmt.Errorf("could not create controller manager: %w", err)
	}

	go func() {
		err := k8sManager.Start(ctrl.SetupSignalHandler())
		Expect(err).ToNot(HaveOccurred())
	}()

	dc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize discovery client: %s", err)
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize dynamic client: %s", err)
	}

	return &K8sTestEnv{
		Client:     k8sManager.GetClient(),
		DynClient:  dyn,
		RestMapper: mapper,
		Stop: func() {
			err := testEnv.Stop()
			Expect(err).NotTo(HaveOccurred())
		},
	}, nil
}

// MakeFakeLogr returns an API compliant logr object that can be used for unit testing.
// Without these stubs filled in, a nil pointer exception will be thrown on log.V().
func MakeFakeLogr() *fakelogr.FakeLogger {
	log := &fakelogr.FakeLogger{}
	log.WithValuesStub = func(i ...interface{}) logr.Logger {
		return log
	}
	log.VStub = func(i int) logr.Logger {
		return log
	}

	return log
}
