package watch

import (
	"context"
	"fmt"
	sourcev1b2 "github.com/fluxcd/source-controller/api/v1beta2"
	"path/filepath"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/fluxcd/pkg/apis/meta"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/run"
	"github.com/weaveworks/weave-gitops/pkg/run/constants"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func SetupBucketSourceAndHelm(ctx context.Context, log logger.Logger, kubeClient client.Client, params SetupRunObjectParams) error {
	secret, source := createBucketAndSecretObjects(params)

	helm := helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.RunDevHelmName,
			Namespace: params.Namespace,
			Annotations: map[string]string{
				"metadata.weave.works/description": "This is a temporary HelmRelease created by GitOps Run. This will be cleaned up when this instance of GitOps Run is ended.",
				"metadata.weave.works/run-id":      params.SessionName,
				"metadata.weave.works/username":    params.Username,
			},
		},
		Spec: helmv2.HelmReleaseSpec{
			Interval: metav1.Duration{Duration: 30 * 24 * time.Hour}, // 30 days
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart:             params.Path,
					ReconcileStrategy: "Revision",
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Kind: sourcev1b2.BucketKind,
						Name: constants.RunDevBucketName,
					},
					// relative to the root of SourceRef
					ValuesFiles: []string{
						filepath.Join(params.Path, "values.yaml"),
					},
				},
			},
			Timeout: &metav1.Duration{Duration: params.Timeout},
		},
	}

	err := reconcileBucketAndSecretObjects(ctx, log, kubeClient, secret, source)
	if err != nil {
		return err
	}

	// create ks
	log.Actionf("Checking HelmRelease %s ...", helm.Name)

	if err := kubeClient.Get(ctx, client.ObjectKeyFromObject(&helm), &helm); err != nil && apierrors.IsNotFound(err) {
		if err := kubeClient.Create(ctx, &helm); err != nil {
			return fmt.Errorf("couldn't create HelmRelease %s: %v", helm.Name, err.Error())
		} else {
			log.Successf("Created HelmRelease %s", helm.Name)
		}
	} else if err == nil {
		log.Successf("HelmRelease %s already existed", source.Name)
	}

	log.Successf("Setup Bucket Source and HelmRelease successfully")

	return nil
}

// CleanupBucketSourceAndHelm removes the bucket source and ks
func CleanupBucketSourceAndHelm(ctx context.Context, log logger.Logger, kubeClient client.Client, namespace string) error {
	// delete ks
	helm := helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.RunDevHelmName,
			Namespace: namespace,
		},
	}

	log.Actionf("Deleting HelmRelease %s ...", helm.Name)

	if err := kubeClient.Delete(ctx, &helm); err != nil {
		if !apierrors.IsNotFound(err) {
			log.Failuref("Error deleting HelmRelease %s: %v", helm.Name, err.Error())
		}
	} else {
		log.Successf("Deleted HelmRelease %s", helm.Name)
	}

	cleanupBucketAndSecretObjects(ctx, log, kubeClient, namespace)

	log.Successf("Cleanup Bucket Source and HelmRelease successfully")

	return nil
}

// ReconcileDevBucketSourceAndHelm reconciles the dev-bucket and dev-helm asynchronously.
func ReconcileDevBucketSourceAndHelm(ctx context.Context, log logger.Logger, kubeClient client.Client, namespace string, timeout time.Duration) error {
	const interval = 4 * time.Second

	log.Actionf("Start reconciling %s and %s ...", constants.RunDevBucketName, constants.RunDevHelmName)

	// reconcile dev-bucket
	sourceRequestedAt, err := run.RequestReconciliation(ctx, kubeClient,
		types.NamespacedName{
			Name:      constants.RunDevBucketName,
			Namespace: namespace,
		}, schema.GroupVersionKind{
			Group:   sourcev1b2.GroupVersion.Group,
			Version: sourcev1b2.GroupVersion.Version,
			Kind:    sourcev1b2.BucketKind,
		})
	if err != nil {
		return err
	}

	log.Actionf("Reconciling %s ...", constants.RunDevBucketName)

	// wait for the reconciliation of dev-bucket to be done
	if err := wait.ExponentialBackoff(wait.Backoff{
		Duration: interval,
		Factor:   2,
		Jitter:   1,
		Steps:    10,
		Cap:      timeout,
	}, func() (done bool, err error) {
		devBucket := &sourcev1b2.Bucket{}
		if err := kubeClient.Get(ctx, types.NamespacedName{
			Name:      constants.RunDevBucketName,
			Namespace: namespace,
		}, devBucket); err != nil {
			return false, err
		}

		return devBucket.Status.GetLastHandledReconcileRequest() == sourceRequestedAt, nil
	}); err != nil {
		return err
	}

	log.Successf("Reconciled %s", constants.RunDevBucketName)

	// wait for devBucket to be ready
	if err := wait.ExponentialBackoff(wait.Backoff{
		Duration: interval,
		Factor:   2,
		Jitter:   1,
		Steps:    10,
		Cap:      timeout,
	}, func() (done bool, err error) {
		devBucket := &sourcev1b2.Bucket{}
		if err := kubeClient.Get(ctx, types.NamespacedName{
			Name:      constants.RunDevBucketName,
			Namespace: namespace,
		}, devBucket); err != nil {
			return false, err
		}

		return apimeta.IsStatusConditionPresentAndEqual(devBucket.Status.Conditions, meta.ReadyCondition, metav1.ConditionTrue), nil
	}); err != nil {
		return err
	}

	log.Successf("Bucket %s is ready", constants.RunDevBucketName)

	// reconcile dev-ks
	helmRequestedAt, err := run.RequestReconciliation(ctx, kubeClient,
		types.NamespacedName{
			Name:      constants.RunDevHelmName,
			Namespace: namespace,
		}, schema.GroupVersionKind{
			Group:   helmv2.GroupVersion.Group,
			Version: helmv2.GroupVersion.Version,
			Kind:    helmv2.HelmReleaseKind,
		})
	if err != nil {
		return err
	}

	log.Actionf("Reconciling %s ...", constants.RunDevHelmName)

	if err := wait.ExponentialBackoff(wait.Backoff{
		Duration: interval,
		Factor:   2,
		Jitter:   1,
		Steps:    10,
		Cap:      timeout,
	}, func() (done bool, err error) {
		devHelm := &helmv2.HelmRelease{}
		if err := kubeClient.Get(ctx, types.NamespacedName{
			Name:      constants.RunDevHelmName,
			Namespace: namespace,
		}, devHelm); err != nil {
			return false, err
		}

		return devHelm.Status.GetLastHandledReconcileRequest() == helmRequestedAt, nil
	}); err != nil {
		return err
	}

	log.Successf("Reconciled %s", constants.RunDevHelmName)

	devHelm := &helmv2.HelmRelease{}

	devHelmErr := wait.ExponentialBackoff(wait.Backoff{
		Duration: interval,
		Factor:   2,
		Jitter:   1,
		Steps:    10,
		Cap:      timeout,
	}, func() (done bool, err error) {
		if err := kubeClient.Get(ctx, types.NamespacedName{
			Name:      constants.RunDevHelmName,
			Namespace: namespace,
		}, devHelm); err != nil {
			return false, err
		}

		log.Actionf("Waiting for %s to be ready ...", constants.RunDevHelmName)

		cond := apimeta.FindStatusCondition(devHelm.Status.Conditions, meta.ReadyCondition)
		if cond == nil {
			return false, nil
		}

		if cond.Status != "False" && cond.Status != "True" {
			log.Waitingf("Waiting for HelmRelease %s to be ready: %s", devHelm.Name, cond.Message)
			return false, nil
		} else if cond.Status == "False" {
			log.Failuref("HelmRelease %s is not ready: %s", devHelm.Name, cond.Message)
			return false, fmt.Errorf("HelmRelease %s is not ready: %s", devHelm.Name, cond.Message)
		}

		return true, nil
	})

	return devHelmErr
}
