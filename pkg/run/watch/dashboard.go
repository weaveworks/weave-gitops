package watch

import (
	"context"

	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/run"
	"github.com/weaveworks/weave-gitops/pkg/server"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// EnablePortForwardingForDashboard enables port forwarding for the GitOps Dashboard.
func EnablePortForwardingForDashboard(ctx context.Context, log logger.Logger, kubeClient client.Client, config *rest.Config, namespace string, podName string, dashboardPort string) (func(), error) {
	specMap := &PortForwardSpec{
		Namespace:     namespace,
		Name:          podName,
		Kind:          "deployment",
		HostPort:      dashboardPort,
		ContainerPort: server.DefaultPort,
	}
	// get pod from specMap
	namespacedName := types.NamespacedName{Namespace: specMap.Namespace, Name: specMap.Name}

	pod, err := run.GetPodFromResourceDescription(ctx, namespacedName, specMap.Kind, kubeClient)
	if err != nil {
		log.Failuref("Error getting pod from specMap: %v", err)
	}

	if pod != nil {
		waitFwd := make(chan struct{}, 1)
		readyChannel := make(chan struct{})
		cancelPortFwd := func() {
			close(waitFwd)
		}

		log.Actionf("Port forwarding to pod %s/%s ...", pod.Namespace, pod.Name)

		go func() {
			if err := ForwardPort(pod, config, specMap, waitFwd, readyChannel); err != nil {
				log.Failuref("Error forwarding port: %v", err)
			}
		}()
		<-readyChannel

		log.Successf("Port forwarding for dashboard is ready.")

		return cancelPortFwd, nil
	}

	return nil, run.ErrDashboardPodNotFound
}
