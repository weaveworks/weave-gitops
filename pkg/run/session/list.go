package session

import (
	"context"
	"strings"

	"github.com/weaveworks/weave-gitops/core/server/types"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func List(kubeClient client.Client, targetNamespace string) ([]*InternalSession, error) {
	var result []*InternalSession

	statefulSets := appsv1.StatefulSetList{}
	if err := kubeClient.List(context.Background(), &statefulSets,
		client.InNamespace(targetNamespace),
		client.MatchingLabels(map[string]string{
			types.AppLabel:    "vcluster",
			types.PartOfLabel: "gitops-run",
		}),
	); err != nil {
		return nil, err
	}

	for _, s := range statefulSets.Items {
		annotations := s.GetAnnotations()

		result = append(result, &InternalSession{
			SessionName:      s.Name,
			SessionNamespace: s.Namespace,
			Command:          annotations["run.weave.works/command"],
			CliVersion:       annotations["run.weave.works/cli-version"],
			PortForward:      strings.Split(annotations["run.weave.works/port-forward"], ","),
			Namespace:        annotations["run.weave.works/namespace"],
		})
	}

	return result, nil
}
