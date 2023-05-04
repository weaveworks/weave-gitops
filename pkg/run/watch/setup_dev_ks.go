package watch

import (
	"context"
	"errors"
	"fmt"
	sourcev1b2 "github.com/fluxcd/source-controller/api/v1beta2"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	"github.com/fluxcd/pkg/apis/meta"
	"github.com/fsnotify/fsnotify"
	"github.com/minio/minio-go/v7"
	ignore "github.com/sabhiram/go-gitignore"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/run"
	"github.com/weaveworks/weave-gitops/pkg/run/constants"
	"github.com/weaveworks/weave-gitops/pkg/sourceignore"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/cli-utils/pkg/object"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SetupRunObjectParams struct {
	Namespace         string
	Path              string
	Timeout           time.Duration
	DevBucketPort     int32
	SessionName       string
	Username          string
	AccessKey         []byte
	SecretKey         []byte
	DecryptionKeyFile string
}

func SetupBucketSourceAndKS(ctx context.Context, log logger.Logger, kubeClient client.Client, params SetupRunObjectParams) error {
	secret, source := createBucketAndSecretObjects(params)

	ks := kustomizev1.Kustomization{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.RunDevKsName,
			Namespace: params.Namespace,
			Annotations: map[string]string{
				"metadata.weave.works/description": "This is a temporary Kustomization created by GitOps Run. This will be cleaned up when this instance of GitOps Run is ended.",
				"metadata.weave.works/run-id":      params.SessionName,
				"metadata.weave.works/username":    params.Username,
			},
		},
		Spec: kustomizev1.KustomizationSpec{
			Interval: metav1.Duration{Duration: 30 * 24 * time.Hour}, // 30 days
			Prune:    true,                                           // GC the kustomization
			SourceRef: kustomizev1.CrossNamespaceSourceReference{
				Kind: sourcev1b2.BucketKind,
				Name: constants.RunDevBucketName,
			},
			Timeout: &metav1.Duration{Duration: params.Timeout},
			Path:    params.Path,
			Wait:    true,
		},
	}

	if err := setupDecryption(ctx, params, kubeClient, &ks); err != nil {
		return fmt.Errorf("failed setting up decryption: %w", err)
	}

	err := reconcileBucketAndSecretObjects(ctx, log, kubeClient, secret, source)
	if err != nil {
		return err
	}

	// create ks
	log.Actionf("Checking Kustomization %s ...", ks.Name)

	if err := kubeClient.Get(ctx, client.ObjectKeyFromObject(&ks), &ks); err != nil && apierrors.IsNotFound(err) {
		if err := kubeClient.Create(ctx, &ks); err != nil {
			return fmt.Errorf("couldn't create kustomization %s: %v", ks.Name, err.Error())
		} else {
			log.Successf("Created Kustomization %s", ks.Name)
		}
	} else if err == nil {
		log.Successf("Kustomization %s already existed", source.Name)
	}

	log.Successf("Setup Bucket Source and Kustomization successfully")

	return nil
}

func setupDecryption(ctx context.Context, params SetupRunObjectParams, kubeClient client.Client, ks *kustomizev1.Kustomization) error {
	if params.DecryptionKeyFile == "" {
		return nil
	}

	ageKeyData, err := os.ReadFile(params.DecryptionKeyFile)
	if err != nil {
		return fmt.Errorf("failed reading age key file: %w", err)
	}

	var secretKey string

	switch filepath.Ext(params.DecryptionKeyFile) {
	case ".agekey":
		secretKey = "age.agekey"
	case ".asc":
		secretKey = "identity.asc"
	default:
		return fmt.Errorf("failed determining decryption key type from filename %s", filepath.Base(params.DecryptionKeyFile))
	}

	decSecret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.RunDevKsDecryption,
			Namespace: params.Namespace,
		},
		Data: map[string][]byte{
			secretKey: ageKeyData,
		},
	}
	if err := kubeClient.Create(ctx, &decSecret); err != nil {
		return fmt.Errorf("failed creating age key Secret: %w", err)
	}

	ks.Spec.Decryption = &kustomizev1.Decryption{
		Provider: "sops",
		SecretRef: &meta.LocalObjectReference{
			Name: decSecret.Name,
		},
	}

	return nil
}

// SyncDir recursively uploads all files in a directory to an S3 bucket with minio library
func SyncDir(ctx context.Context, log logger.Logger, dir, bucket string, client *minio.Client, ignorer *ignore.GitIgnore) error {
	log.Actionf("Refreshing bucket %s ...", bucket)

	if err := client.RemoveBucketWithOptions(ctx, bucket, minio.RemoveBucketOptions{
		ForceDelete: true,
	}); err != nil {
		// if error is not bucket not found, return error
		if !strings.Contains(err.Error(), "NoSuchBucket") {
			return err
		}
	}

	if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
		return err
	}

	uploadCount := 0
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Failuref("Error walking directory: %v", err)
			return err
		}

		if info.IsDir() {
			// if it's a hidden directory, ignore it
			if strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}

			return nil
		}

		objectName, err := filepath.Rel(dir, path)
		if err != nil {
			log.Failuref("Error getting relative path: %v", err)
			return err
		}
		if ignorer.MatchesPath(path) {
			return nil
		}
		// upload the file
		_, err = client.FPutObject(ctx, bucket, objectName, path, minio.PutObjectOptions{})

		if err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
			errResp, ok := err.(minio.ErrorResponse)
			if ok && errResp.Code == "MissingContentLength" {
				// This happens when the file was empty - this is OK
				return nil
			}
			// Report the error, but continue anyway - this could be e.g.
			// a file with odd permissions, which isn't necessarily a problem
			log.Failuref("Couldn't upload %v: %v", path, err)
			return nil
		}
		uploadCount++
		if uploadCount%10 == 0 {
			fmt.Print(".")
		}
		return nil
	})

	fmt.Println()
	log.Actionf("Uploaded %d files", uploadCount)

	if err != nil && !errors.Is(err, context.Canceled) {
		log.Failuref("Error syncing directory: %v", err)
		return err
	}

	return nil
}

// CleanupBucketSourceAndKS removes the bucket source and ks
func CleanupBucketSourceAndKS(ctx context.Context, log logger.Logger, kubeClient client.Client, namespace string) error {
	// delete ks
	ks := kustomizev1.Kustomization{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.RunDevKsName,
			Namespace: namespace,
		},
	}

	log.Actionf("Deleting Kustomization %s ...", ks.Name)

	if err := kubeClient.Delete(ctx, &ks); err != nil {
		if !apierrors.IsNotFound(err) {
			log.Failuref("Error deleting Kustomization %s: %v", ks.Name, err.Error())
		}
	} else {
		log.Successf("Deleted Kustomization %s", ks.Name)
	}

	cleanupBucketAndSecretObjects(ctx, log, kubeClient, namespace)

	log.Successf("Cleanup Bucket Source and Kustomization successfully")

	return nil
}

// findConditionMessages finds the messages in the condition of objects in the inventory.
func findConditionMessages(ctx context.Context, kubeClient client.Client, ks *kustomizev1.Kustomization) ([]string, error) {
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

		gvkID := strings.Join([]string{objMeta.GroupKind.Group, entry.Version, objMeta.GroupKind.Kind}, "_")

		if _, exist := gvks[gvkID]; !exist {
			gvks[gvkID] = schema.GroupVersionKind{
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

		if err := kubeClient.List(ctx, unstructuredList,
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

func WatchDirsForFileWalker(watcher *fsnotify.Watcher, ignorer *ignore.GitIgnore) func(path string, info os.FileInfo, err error) error {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error walking path: %v", err)
		}

		if info.IsDir() {
			// if it's a hidden directory, ignore it
			if strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}

			if ignorer.MatchesPath(path) {
				return filepath.SkipDir
			}

			if err := watcher.Add(path); err != nil {
				return err
			}
		}

		return nil
	}
}

// InitializeRootDir initializes the root directory (creates the .sourceignore file in it).
func InitializeRootDir(log logger.Logger, rootPath string) error {
	stat, err := os.Stat(rootPath)

	if err != nil {
		return err
	} else if !stat.IsDir() {
		return fmt.Errorf("root must be a directory")
	} else {
		f, err := os.Open(rootPath)
		if err != nil {
			return err
		}

		_, err = f.Readdirnames(1)
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}
	}

	err = sourceignore.CreateIgnoreFile(rootPath, sourceignore.IgnoreFilename, []string{})
	if err == nil {
		log.Successf("%s file created. Please add ignore patterns to ignore specific YAML files or directories during validation to it", sourceignore.IgnoreFilename)
	}

	return err
}

// InitializeTargetDir initializes the target directory (creates the entrypoint Kustomization in it).
func InitializeTargetDir(targetPath string) error {
	shouldCreate := false
	stat, err := os.Stat(targetPath)

	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	} else if err != nil {
		err := os.MkdirAll(targetPath, 0o755)
		if err != nil {
			return err
		}
		shouldCreate = true
	} else if !stat.IsDir() {
		return fmt.Errorf("target must be a directory")
	} else {
		f, err := os.Open(targetPath)
		if err != nil {
			return err
		}

		_, err = f.Readdirnames(1)
		if err != nil && errors.Is(err, io.EOF) {
			shouldCreate = true
		} else if err != nil {
			return err
		}
	}

	if shouldCreate {
		f, err := os.Create(filepath.Join(targetPath, "kustomization.yaml"))
		if err != nil {
			return fmt.Errorf("error creating entrypoint kustomization.yaml: %w", err)
		}

		_, err = f.Write([]byte(`---
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources: [] # 👋 Start adding the resources you want to sync here
`))
		if err != nil {
			return err
		}
	}

	return nil
}

// ReconcileDevBucketSourceAndKS reconciles the dev-bucket and dev-ks asynchronously.
func ReconcileDevBucketSourceAndKS(ctx context.Context, log logger.Logger, kubeClient client.Client, namespace string, timeout time.Duration) error {
	const interval = 3 * time.Second / 2

	// reconcile dev-bucket
	sourceRequestedAt, err := run.RequestReconciliation(ctx, kubeClient,
		types.NamespacedName{
			Name:      constants.RunDevBucketName,
			Namespace: namespace,
		}, schema.GroupVersionKind{
			Group:   "source.toolkit.fluxcd.io",
			Version: "v1beta2",
			Kind:    sourcev1b2.BucketKind,
		})
	if err != nil {
		return err
	}

	// wait for the reconciliation of dev-bucket to be done
	if err := wait.Poll(interval, timeout, func() (bool, error) {
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

	// wait for devBucket to be ready
	if err := wait.Poll(interval, timeout, func() (bool, error) {
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

	// reconcile dev-ks
	ksRequestedAt, err := run.RequestReconciliation(ctx, kubeClient,
		types.NamespacedName{
			Name:      constants.RunDevKsName,
			Namespace: namespace,
		}, schema.GroupVersionKind{
			Group:   "kustomize.toolkit.fluxcd.io",
			Version: "v1beta2",
			Kind:    kustomizev1.KustomizationKind,
		})
	if err != nil {
		return err
	}

	if err := wait.Poll(interval, timeout, func() (bool, error) {
		devKs := &kustomizev1.Kustomization{}
		if err := kubeClient.Get(ctx, types.NamespacedName{
			Name:      constants.RunDevKsName,
			Namespace: namespace,
		}, devKs); err != nil {
			return false, err
		}

		return devKs.Status.GetLastHandledReconcileRequest() == ksRequestedAt, nil
	}); err != nil {
		return err
	}

	devKs := &kustomizev1.Kustomization{}
	devKsErr := wait.Poll(interval, timeout, func() (bool, error) {
		if err := kubeClient.Get(ctx, types.NamespacedName{
			Name:      constants.RunDevKsName,
			Namespace: namespace,
		}, devKs); err != nil {
			return false, err
		}

		healthy := apimeta.IsStatusConditionPresentAndEqual(
			devKs.Status.Conditions,
			kustomizev1.HealthyCondition,
			metav1.ConditionTrue,
		)
		return healthy, nil
	})

	if devKsErr != nil {
		messages, err := findConditionMessages(ctx, kubeClient, devKs)
		if err != nil {
			return err
		}

		for _, msg := range messages {
			log.Failuref(msg)
		}
	}

	return devKsErr
}

func CreateIgnorer(gitRootDir string) *ignore.GitIgnore {
	ignoreFile := filepath.Join(gitRootDir, ".gitignore")

	var ignorer *ignore.GitIgnore = nil
	if _, err := os.Stat(ignoreFile); err == nil {
		ignorer, err = ignore.CompileIgnoreFile(ignoreFile)
		if err != nil {
			// If we couldn't parse gitignore, just ignore nothing
			ignorer = nil
		}
	}

	if ignorer == nil {
		// Whether there was no gitignore file or the one that was there was broken,
		// fall back to just no ignore lines - this is just a pass-through
		ignorer = ignore.CompileIgnoreLines()
	}

	return ignorer
}
