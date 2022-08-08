package run

import (
	"context"
	"fmt"
	"time"

	"github.com/fluxcd/flux2/pkg/manifestgen/install"
	coretypes "github.com/weaveworks/weave-gitops/core/server/types"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func InstallFlux(log logger.Logger, ctx context.Context, installOptions install.Options, manager ResourceManagerForApply) error {
	log.Actionf("Installing Flux ...")

	manifests, err := install.Generate(installOptions, "")
	if err != nil {
		log.Failuref("Couldn't generate manifests")
		return err
	}

	content := []byte(manifests.Content)

	applyOutput, err := Apply(log, ctx, manager, content)
	if err != nil {
		log.Failuref("Flux install failed")
		return err
	}

	log.Println(applyOutput)

	return nil
}

func GetFluxVersion(log logger.Logger, ctx context.Context, kubeClient client.Client) (string, error) {
	log.Actionf("Getting Flux version ...")

	listResult := unstructured.UnstructuredList{}

	listResult.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "Namespace",
	})

	listOptions := client.MatchingLabels{
		coretypes.PartOfLabel: "flux",
	}

	u := unstructured.Unstructured{}

	if err := kubeClient.List(ctx, &listResult, listOptions); err != nil {
		log.Failuref("error getting the list of Flux objects")
		return "", err
	} else {
		for _, item := range listResult.Items {
			if item.GetLabels()[flux.VersionLabelKey] != "" {
				u = item
				break
			}
		}
	}

	labels := u.GetLabels()
	if labels == nil {
		return "", fmt.Errorf("error getting Flux labels")
	}

	fluxVersion := labels[flux.VersionLabelKey]
	if fluxVersion == "" {
		return "", fmt.Errorf("no flux version found")
	}

	return fluxVersion, nil
}

func WaitForDeploymentToBeReady(log logger.Logger, kubeClient client.Client, deploymentName string, namespace string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	deployment := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: namespace,
		},
	}

	if err := wait.ExponentialBackoff(wait.Backoff{
		Duration: 1 * time.Second,
		Factor:   2,
		Jitter:   1,
		Steps:    10,
	}, func() (done bool, err error) {
		d := deployment.DeepCopy()
		if err := kubeClient.Get(ctx, client.ObjectKeyFromObject(d), d); err != nil {
			return false, err
		}
		// Confirm the state we are observing is for the current generation
		if d.Generation != d.Status.ObservedGeneration {
			return false, nil
		}

		if d.Status.ReadyReplicas == d.Status.Replicas {
			return true, nil
		}

		return false, nil
	}); err != nil {
		return err
	}

	return nil
}
