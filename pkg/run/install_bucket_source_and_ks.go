package run

import (
	"context"
	"time"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func SetupBucketSourceAndKS(log logger.Logger, kubeClient *kube.KubeHTTP, namespace string, path string) error {
	const devBucketCredentials = "dev-bucket-credentials"

	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      devBucketCredentials,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			"accesskey": []byte("user"),
			"secretkey": []byte("doesn't matter"),
		},
		Type: "Opaque",
	}
	source := sourcev1.Bucket{
		ObjectMeta: metav1.ObjectMeta{
			Name:      devBucket,
			Namespace: namespace,
		},
		Spec: sourcev1.BucketSpec{
			Interval:   metav1.Duration{Duration: 30 * 24 * time.Hour}, // 30 days
			Provider:   "generic",
			BucketName: devBucket,
			Endpoint:   "dev-bucket.dev-bucket.svc.cluster.local:9000",
			Insecure:   true,
			SecretRef:  &meta.LocalObjectReference{Name: devBucketCredentials},
		},
	}
	ks := kustomizev1.Kustomization{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dev-ks",
			Namespace: namespace,
		},
		Spec: kustomizev1.KustomizationSpec{
			Interval: metav1.Duration{Duration: 30 * 24 * time.Hour}, // 30 days
			Prune:    true,                                           // GC the kustomization
			SourceRef: kustomizev1.CrossNamespaceSourceReference{
				Kind: "Bucket",
				Name: devBucket,
			},
			Path: path,
			Wait: true,
		},
	}

	// create secret
	log.Actionf("Checking secret %s ...", secret.Name)

	if err := kubeClient.Get(context.Background(), client.ObjectKeyFromObject(&secret), &secret); err != nil && apierrors.IsNotFound(err) {
		if err := kubeClient.Create(context.Background(), &secret); err != nil {
			log.Failuref("Error creating secret %s: %v", secret.Name, err.Error())
		} else {
			log.Successf("Created secret %s", secret.Name)
		}
	} else if err == nil {
		log.Successf("Secret %s already existed", secret.Name)
	}

	// create source
	log.Actionf("Checking bucket source %s ...", source.Name)

	if err := kubeClient.Get(context.Background(), client.ObjectKeyFromObject(&source), &source); err != nil && apierrors.IsNotFound(err) {
		if err := kubeClient.Create(context.Background(), &source); err != nil {
			log.Failuref("Error creating source %s: %v", source.Name, err.Error())
		} else {
			log.Successf("Created source %s", source.Name)
		}
	} else if err == nil {
		log.Successf("Source %s already existed", source.Name)
	}

	// create ks
	log.Actionf("Checking Kustomization %s ...", ks.Name)

	if err := kubeClient.Get(context.Background(), client.ObjectKeyFromObject(&ks), &ks); err != nil && apierrors.IsNotFound(err) {
		if err := kubeClient.Create(context.Background(), &ks); err != nil {
			log.Failuref("Error creating ks %s: %v", ks.Name, err.Error())
		} else {
			log.Successf("Created ks %s", ks.Name)
		}
	} else if err == nil {
		log.Successf("Kustomization %s already existed", source.Name)
	}

	log.Successf("Setup Bucket Source and Kustomization successfully")

	return nil
}

// CleanupBucketSourceAndKS removes the bucket source and ks
func CleanupBucketSourceAndKS(log logger.Logger, kubeClient *kube.KubeHTTP, namespace string) error {
	const (
		devBucketCredentials = "dev-bucket-credentials"
		devBucket            = "dev-bucket"
		devKS                = "dev-ks"
	)

	// delete secret
	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      devBucketCredentials,
			Namespace: namespace,
		},
	}

	log.Actionf("Deleting secret %s ...", secret.Name)

	if err := kubeClient.Delete(context.Background(), &secret); err != nil {
		log.Failuref("Error deleting secret %s: %v", secret.Name, err.Error())
	} else {
		log.Successf("Deleted secret %s", secret.Name)
	}

	// delete source
	source := sourcev1.Bucket{
		ObjectMeta: metav1.ObjectMeta{
			Name:      devBucket,
			Namespace: namespace,
		},
	}

	log.Actionf("Deleting source %s ...", source.Name)

	if err := kubeClient.Delete(context.Background(), &source); err != nil {
		log.Failuref("Error deleting source %s: %v", source.Name, err.Error())
	} else {
		log.Successf("Deleted source %s", source.Name)
	}

	// delete ks
	ks := kustomizev1.Kustomization{
		ObjectMeta: metav1.ObjectMeta{
			Name:      devKS,
			Namespace: namespace,
		},
	}

	log.Actionf("Deleting ks %s ...", ks.Name)

	if err := kubeClient.Delete(context.Background(), &ks); err != nil {
		log.Failuref("Error deleting ks %s: %v", ks.Name, err.Error())
	} else {
		log.Successf("Deleted ks %s", ks.Name)
	}

	log.Successf("Cleanup Bucket Source and Kustomization successfully")

	return nil
}
