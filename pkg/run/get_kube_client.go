package run

import (
	"context"

	runclient "github.com/fluxcd/pkg/runtime/client"
	"github.com/fluxcd/pkg/ssa"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/cli-utils/pkg/kstatus/polling"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func GetKubeConfigArgs() *genericclioptions.ConfigFlags {
	kubeConfigArgs := genericclioptions.NewConfigFlags(false)

	// Prevent AddFlags from configuring unnecessary flags.
	kubeConfigArgs.Insecure = nil
	kubeConfigArgs.Timeout = nil
	kubeConfigArgs.KubeConfig = nil
	kubeConfigArgs.CacheDir = nil
	kubeConfigArgs.ClusterName = nil
	kubeConfigArgs.AuthInfoName = nil
	kubeConfigArgs.Namespace = nil
	kubeConfigArgs.APIServer = nil
	kubeConfigArgs.TLSServerName = nil
	kubeConfigArgs.CertFile = nil
	kubeConfigArgs.KeyFile = nil
	kubeConfigArgs.CAFile = nil
	kubeConfigArgs.BearerToken = nil
	kubeConfigArgs.Impersonate = nil
	kubeConfigArgs.ImpersonateUID = nil
	kubeConfigArgs.ImpersonateGroup = nil

	return kubeConfigArgs
}

func GetKubeClientOptions() *runclient.Options {
	kubeClientOpts := new(runclient.Options)

	return kubeClientOpts
}

func GetKubeClient(log logger.Logger, contextName string, cfg *rest.Config, kubeClientOpts *runclient.Options) (*kube.KubeHTTP, error) {
	// avoid throttling request when some Flux CRDs are not registered
	cfg.QPS = kubeClientOpts.QPS
	cfg.Burst = kubeClientOpts.Burst

	kubeClient, err := kube.NewKubeHTTPClientWithConfig(cfg, contextName)
	if err != nil {
		log.Failuref("Kubernetes client initialization failed: %v", err.Error())
		return nil, err
	}

	return kubeClient, nil
}

func newManager(log logger.Logger, ctx context.Context, kubeClient ctrlclient.Client, kubeConfigArgs genericclioptions.RESTClientGetter) (*ssa.ResourceManager, error) {
	restMapper, err := kubeConfigArgs.ToRESTMapper()
	if err != nil {
		log.Failuref("Error getting a restmapper")
		return nil, err
	}

	kubePoller := polling.NewStatusPoller(kubeClient, restMapper, polling.Options{})

	return ssa.NewResourceManager(kubeClient, kubePoller, ssa.Owner{
		Field: "flux",
		Group: "fluxcd.io",
	}), nil
}
