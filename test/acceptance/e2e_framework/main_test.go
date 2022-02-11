package e2e_framework

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/weaveworks/weave-gitops/test/vcluster"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
)

var testenv env.Environment

func TestMain(m *testing.M) {
	testenv = env.New()
	clusterName := envconf.RandomName("vcluster", 16)

	testenv.Setup(
		waitK3sCluster(),
		crateVcluster(clusterName),
		envfuncs.CreateNamespace(clusterName),
	)

	testenv.Finish(
	// envfuncs.DeleteNamespace(namespace),
	// envfuncs.DestroyKindCluster(kindClusterName),
	)
	os.Exit(testenv.Run(m))
}

type vclusterContextKey string

func waitK3sCluster() env.Func {
	return func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
		if err := vcluster.UpdateHostKubeconfig(); err != nil {
			return ctx, fmt.Errorf("failed updating host kubeconfig: %w", err)
		}

		if err := vcluster.WaitClusterConnectivity(); err != nil {
			return ctx, fmt.Errorf("failed waiting cluster to be ready: %w", err)
		}

		if err := vcluster.InstallNginxIngressController(); err != nil {
			return ctx, fmt.Errorf("failed installing ingress controller: %w", err)
		}

		return ctx, nil
	}
}

func crateVcluster(name string) env.Func {
	return func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
		clusterFactory, err := vcluster.NewFactory()
		if err != nil {
			return ctx, fmt.Errorf("failed creating factory: %w", err)
		}

		_, kubeconfigFile, err := clusterFactory.Create(ctx, name)
		if err != nil {
			return ctx, fmt.Errorf("failed creating cluster: %w", err)
		}

		cfg.WithKubeconfigFile(kubeconfigFile)

		if err := vcluster.InstallFlux(kubeconfigFile); err != nil {
			return ctx, fmt.Errorf("failed installing flux: %w", err)
		}

		return ctx, nil
	}
}
