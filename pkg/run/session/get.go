package session

import (
	"context"
	"fmt"
	"strings"

	"github.com/weaveworks/weave-gitops/core/server/types"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Get(kubeClient client.Client, name, namespace string) (*InternalSession, error) {
	var result *InternalSession

	statefulSet := appsv1.StatefulSet{}
	if err := kubeClient.Get(context.Background(), client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}, &statefulSet); err != nil {
		if errors.IsNotFound(err) {
			return nil, fmt.Errorf("session %s/%s not found", namespace, name)
		}

		return nil, err
	}

	labels := statefulSet.GetLabels()
	if labels != nil {
		if labels[types.AppLabel] != "vcluster" || labels[types.PartOfLabel] != "gitops-run" {
			return nil, fmt.Errorf("%s/%s is an invalid GitOps Run session", namespace, name)
		}
	}

	annotations := statefulSet.GetAnnotations()
	result = &InternalSession{
		SessionName:      statefulSet.Name,
		SessionNamespace: statefulSet.Namespace,
		Command:          annotations["run.weave.works/command"],
		CliVersion:       annotations["run.weave.works/cli-version"],
		PortForward:      strings.Split(annotations["run.weave.works/port-forward"], ","),
		Namespace:        annotations["run.weave.works/namespace"],
	}

	return result, nil
}
