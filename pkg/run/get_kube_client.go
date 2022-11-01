package run

import (
	runclient "github.com/fluxcd/pkg/runtime/client"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"k8s.io/client-go/rest"
)

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
