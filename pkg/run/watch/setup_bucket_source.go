package watch

import (
	"context"
	"fmt"
	sourcev1b2 "github.com/fluxcd/source-controller/api/v1beta2"
	"time"

	"github.com/fluxcd/pkg/apis/meta"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/run/constants"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func createBucketAndSecretObjects(params SetupRunObjectParams) (corev1.Secret, sourcev1b2.Bucket) {
	// create a secret
	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.RunDevBucketCredentials,
			Namespace: params.Namespace,
		},
		Data: map[string][]byte{
			"accesskey": params.AccessKey,
			"secretkey": params.SecretKey,
		},
		Type: "Opaque",
	}
	source := sourcev1b2.Bucket{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.RunDevBucketName,
			Namespace: params.Namespace,
			Annotations: map[string]string{
				"metadata.weave.works/description": "This is a temporary Bucket created by GitOps Run. This will be cleaned up when this instance of GitOps Run is ended.",
				"metadata.weave.works/run-id":      params.SessionName,
				"metadata.weave.works/username":    params.Username,
			},
		},
		Spec: sourcev1b2.BucketSpec{
			Interval:   metav1.Duration{Duration: 30 * 24 * time.Hour}, // 30 days
			Provider:   "generic",
			BucketName: constants.RunDevBucketName,
			Endpoint:   fmt.Sprintf("%s.%s.svc.cluster.local:%d", constants.RunDevBucketName, constants.GitOpsRunNamespace, params.DevBucketPort),
			Insecure:   true,
			Timeout:    &metav1.Duration{Duration: params.Timeout},
			SecretRef:  &meta.LocalObjectReference{Name: constants.RunDevBucketCredentials},
		},
	}

	return secret, source
}

func reconcileBucketAndSecretObjects(ctx context.Context, log logger.Logger, kubeClient client.Client, secret corev1.Secret, source sourcev1b2.Bucket) error {
	// create secret
	log.Actionf("Checking secret %s ...", secret.Name)

	if err := kubeClient.Get(ctx, client.ObjectKeyFromObject(&secret), &secret); err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed fetching secret %s/%s: %w", secret.Namespace, secret.Name, err)
		}

		if err := kubeClient.Create(ctx, &secret); err != nil {
			return fmt.Errorf("couldn't create secret %s: %v", secret.Name, err.Error())
		}

		log.Successf("Created secret %s", secret.Name)
	}

	log.Successf("Secret %s already existed", secret.Name)

	// create source
	log.Actionf("Checking bucket source %s ...", source.Name)

	if err := kubeClient.Get(ctx, client.ObjectKeyFromObject(&source), &source); err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed fetching bucket source %s/%s: %w", source.Namespace, source.Name, err)
		}

		if err := kubeClient.Create(ctx, &source); err != nil {
			return fmt.Errorf("couldn't create source %s: %v", source.Name, err.Error())
		}

		log.Successf("Created source %s", source.Name)
	}

	log.Successf("Source %s already existed", source.Name)

	return nil
}

func cleanupBucketAndSecretObjects(ctx context.Context, log logger.Logger, kubeClient client.Client, namespace string) {
	// delete secret
	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.RunDevBucketCredentials,
			Namespace: namespace,
		},
	}

	log.Actionf("Deleting secret %s ...", secret.Name)

	if err := kubeClient.Delete(ctx, &secret); err != nil {
		if !apierrors.IsNotFound(err) {
			log.Failuref("Error deleting secret %s: %v", secret.Name, err.Error())
		}
	} else {
		log.Successf("Deleted secret %s", secret.Name)
	}

	// delete source
	source := sourcev1b2.Bucket{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.RunDevBucketName,
			Namespace: namespace,
		},
	}

	log.Actionf("Deleting source %s ...", source.Name)

	if err := kubeClient.Delete(ctx, &source); err != nil {
		if !apierrors.IsNotFound(err) {
			log.Failuref("Error deleting source %s: %v", source.Name, err.Error())
		}
	} else {
		log.Successf("Deleted source %s", source.Name)
	}
}
