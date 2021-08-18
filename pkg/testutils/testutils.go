package testutils

import (
	"fmt"

	"github.com/weaveworks/weave-gitops/pkg/kube"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

// Starts a local k8s test environment for testing Kubernetes operations such as Create, Get, Delete, etc
func StartK8sTestEnvironment() (client.Client, func(), error) {
	testEnv := &envtest.Environment{
		CRDDirectoryPaths: []string{
			"../../manifests/crds",
			"../../tools/testcrds",
		},
	}

	var err error
	cfg, err := testEnv.Start()
	if err != nil {
		return nil, nil, fmt.Errorf("could not start testEnv: %w", err)
	}

	scheme := kube.CreateScheme()

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		ClientDisableCacheFor: []client.Object{
			&wego.Application{},
			&corev1.Namespace{},
			&corev1.Secret{},
		},
		Scheme: scheme,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("could not create controller manager: %w", err)
	}

	go func() {
		err := k8sManager.Start(ctrl.SetupSignalHandler())
		Expect(err).ToNot(HaveOccurred())
	}()

	return k8sManager.GetClient(), func() {
		err := testEnv.Stop()
		Expect(err).NotTo(HaveOccurred())
	}, nil
}
