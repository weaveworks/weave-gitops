package install

import (
	"context"
	"fmt"
	sourcev1b2 "github.com/fluxcd/source-controller/api/v1beta2"
	"os"
	"path/filepath"
	"strings"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/weaveworks/weave-gitops/cmd/gitops/version"
	coretypes "github.com/weaveworks/weave-gitops/core/server/types"
	"github.com/weaveworks/weave-gitops/pkg/run/constants"
	"github.com/weaveworks/weave-gitops/pkg/run/session"
	appsv1 "k8s.io/api/apps/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func makeVClusterHelmRepository(namespace string) *sourcev1b2.HelmRepository {
	helmRepository := &sourcev1b2.HelmRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "loft-sh",
			Namespace: namespace,
		},
		Spec: sourcev1b2.HelmRepositorySpec{
			URL: "https://charts.loft.sh",
		},
	}

	return helmRepository
}

func makeVClusterHelmRelease(name, namespace, fluxNamespace, command string, portForwards []string, automationKind string) *helmv2.HelmRelease {
	annotations := []string{
		`"run.weave.works/cli-version": "` + version.Version + `"`,
		`"run.weave.works/command": "` + command + `"`,
		`"run.weave.works/automation-kind": "` + automationKind + `"`,
		`"run.weave.works/namespace": "` + namespace + `"`,
		`"run.weave.works/flux-namespace": "` + fluxNamespace + `"`,
	}

	if len(portForwards) > 0 {
		annotations = append(annotations, `"run.weave.works/port-forward":"`+strings.Join(portForwards, ",")+`"`)
	}

	helmRelease := &helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				coretypes.AppLabel:    "vcluster",
				coretypes.PartOfLabel: "gitops-run",
			},
		},
		Spec: helmv2.HelmReleaseSpec{
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart: "vcluster",
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Kind: "HelmRepository",
						Name: "loft-sh",
					},
				},
			},
			Interval:        metav1.Duration{Duration: 1 * time.Hour},
			ReleaseName:     name,
			TargetNamespace: namespace,
			Install: &helmv2.Install{
				CRDs: helmv2.Create,
			},
			Upgrade: &helmv2.Upgrade{
				CRDs: helmv2.CreateReplace,
			},
			Values: &apiextensions.JSON{Raw: []byte(fmt.Sprintf(`
{
  "labels": {
    "app.kubernetes.io/part-of": "gitops-run"
  },
  "annotations": {
    %s
  },
  "hostpathMapper": {
    "enabled": true
  },
  "mapServices": {
    "fromVirtual": [
      {
        "from": "%s/%s",
        "to": "%s-bucket"
      }
    ]
  }
}`,
				strings.Join(annotations, ", \n    "),
				constants.GitOpsRunNamespace,
				constants.RunDevBucketName,
				name,
			))},
		},
	}

	return helmRelease
}

func installVCluster(kubeClient client.Client, name, namespace, fluxNamespace string, portForwards []string, automationKind string) error {
	helmRepo := makeVClusterHelmRepository(namespace)

	if err := kubeClient.Create(context.Background(), helmRepo); err != nil {
		if apierrors.IsAlreadyExists(err) {
			// do nothing
		} else {
			return err
		}
	}

	args := append([]string{filepath.Base(os.Args[0])}, os.Args[1:]...)
	command := strings.Join(args, " ")

	helmRelease := makeVClusterHelmRelease(name, namespace, fluxNamespace, command, portForwards, automationKind)

	if err := kubeClient.Create(context.Background(), helmRelease); err != nil {
		if apierrors.IsAlreadyExists(err) {
			// do nothing
		} else {
			return err
		}
	}

	//nolint:staticcheck // deprecated, tracking issue: https://github.com/weaveworks/weave-gitops/issues/3812
	if err := wait.Poll(2*time.Second, 5*time.Minute, func() (bool, error) {
		instance := appsv1.StatefulSet{}
		if err := kubeClient.Get(
			context.Background(),
			types.NamespacedName{
				Name:      name,
				Namespace: namespace,
			}, &instance); err != nil {
			if apierrors.IsNotFound(err) {
				return false, nil
			} else {
				return false, err
			}
		}

		if instance.Status.ReadyReplicas >= 1 {
			return true, nil
		}

		return false, nil
	}); err != nil {
		return err
	}

	return nil
}

func uninstallVcluster(kubeClient client.Client, name, namespace string) error {
	// clean up the session resources using functions from the session package
	internalSession, err := session.Get(kubeClient, name, namespace)
	if err != nil {
		return err
	}

	if err := session.Remove(kubeClient, internalSession); err != nil {
		return err
	}

	// clean up repo
	helmRepo := makeVClusterHelmRepository(namespace)

	if err := kubeClient.Delete(context.Background(), helmRepo); err != nil {
		if apierrors.IsNotFound(err) {
			// Do nothing because the helmRepo can be deleted by other GitOps Run instances
		} else {
			return err
		}
	}

	return nil
}
