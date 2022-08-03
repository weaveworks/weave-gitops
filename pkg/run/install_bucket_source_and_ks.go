package run

import (
	"context"
	"fmt"
	"strings"
	"time"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/cli-utils/pkg/object"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func SetupBucketSourceAndKS(log logger.Logger, kubeClient client.Client, namespace string, path string, timeout time.Duration) error {
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
			Timeout:    &metav1.Duration{Duration: timeout},
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
			Timeout: &metav1.Duration{Duration: timeout},
			Path:    path,
			Wait:    true,
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
func CleanupBucketSourceAndKS(log logger.Logger, kubeClient client.Client, namespace string) error {
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

// FindConditionMessages finds the messages in the condition of objects in the inventory.
func FindConditionMessages(kubeClient client.Client, ks *kustomizev1.Kustomization) ([]string, error) {
	if ks.Status.Inventory == nil {
		return nil, fmt.Errorf("inventory is nil")
	}

	gvks := map[string]schema.GroupVersionKind{}
	// collect gvk of the objects
	for _, entry := range ks.Status.Inventory.Entries {
		objMeta, err := object.ParseObjMetadata(entry.ID)
		if err != nil {
			return nil, fmt.Errorf("invalid inventory item '%s', error: %w", entry.ID, err)
		}

		gvkId := strings.Join([]string{objMeta.GroupKind.Group, entry.Version, objMeta.GroupKind.Kind}, "_")

		if _, exist := gvks[gvkId]; !exist {
			gvks[gvkId] = schema.GroupVersionKind{
				Group:   objMeta.GroupKind.Group,
				Version: entry.Version,
				Kind:    objMeta.GroupKind.Kind,
			}
		}
	}

	var messages []string

	for _, gvk := range gvks {
		unstructuredList := &unstructured.UnstructuredList{}
		unstructuredList.SetGroupVersionKind(gvk)

		if err := kubeClient.List(context.Background(), unstructuredList,
			client.MatchingLabelsSelector{
				Selector: labels.Set(
					map[string]string{
						"kustomize.toolkit.fluxcd.io/name":      ks.Name,
						"kustomize.toolkit.fluxcd.io/namespace": ks.Namespace,
					},
				).AsSelector(),
			},
		); err != nil {
			return nil, err
		}

		for _, u := range unstructuredList.Items {
			if conditions, found, err := unstructured.NestedSlice(u.UnstructuredContent(), "status", "conditions"); err == nil && found {
				for _, condition := range conditions {
					c := condition.(map[string]interface{})
					if status, found, err := unstructured.NestedString(c, "status"); err == nil && found {
						if status != "True" {
							if message, found, err := unstructured.NestedString(c, "message"); err == nil && found {
								messages = append(messages, fmt.Sprintf("%s %s/%s: %s", u.GetKind(), u.GetNamespace(), u.GetName(), message))
							}
						}
					}
				}
			}
		}
	}

	return messages, nil
}
