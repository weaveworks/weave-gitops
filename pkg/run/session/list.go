package session

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

func Get(kubeClient client.Client, name string, namespace string) (*InternalSession, error) {
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
		if labels["app"] != "vcluster" || labels["app.kubernetes.io/part-of"] != "gitops-run" {
			return nil, fmt.Errorf("%s/%s is an invalid GitOps Run session", namespace, name)
		}
	}

	annotations := statefulSet.GetAnnotations()
	result = &InternalSession{
		Name:        statefulSet.Name,
		Namespace:   statefulSet.Namespace,
		Command:     annotations["run.weave.works/command"],
		CliVersion:  annotations["run.weave.works/cli-version"],
		PortForward: strings.Split(annotations["run.weave.works/port-forward"], ","),
	}

	return result, nil
}

func List(kubeClient client.Client, namespace string) ([]*InternalSession, error) {
	var result []*InternalSession

	statefulSets := appsv1.StatefulSetList{}
	if err := kubeClient.List(context.Background(), &statefulSets,
		client.InNamespace(namespace),
		client.MatchingLabels(map[string]string{
			"app":                       "vcluster",
			"app.kubernetes.io/part-of": "gitops-run",
		}),
	); err != nil {
		return nil, err
	}

	for _, s := range statefulSets.Items {
		annotations := s.GetAnnotations()

		result = append(result, &InternalSession{
			Name:        s.Name,
			Namespace:   s.Namespace,
			Command:     annotations["run.weave.works/command"],
			CliVersion:  annotations["run.weave.works/cli-version"],
			PortForward: strings.Split(annotations["run.weave.works/port-forward"], ","),
		})
	}

	return result, nil
}
