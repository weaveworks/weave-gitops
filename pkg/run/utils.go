package run

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fluxcd/pkg/apis/meta"
	"github.com/fsnotify/fsnotify"
	"github.com/minio/minio-go/v7"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// FindGitRepoDir finds git repo root directory
func FindGitRepoDir() (string, error) {
	gitDir := "."

	for {
		if _, err := os.Stat(filepath.Join(gitDir, ".git")); err == nil {
			break
		}

		gitDir = filepath.Join(gitDir, "..")

		if gitDir == "/" {
			return "", errors.New("not in a git repo")
		}
	}

	return filepath.Abs(gitDir)
}

// GetRelativePathToRootDir gets relative path to a directory from the git root. It returns an error if there's no git repo.
func GetRelativePathToRootDir(rootDir string, path string) (string, error) {
	absGitDir, err := filepath.Abs(rootDir)

	if err != nil { // not in a git repo
		return "", err
	}

	return filepath.Rel(absGitDir, path)
}

// WatchAndSync watches files recursively, and re-sync the whole directory to the bucket using minio library
func WatchAndSync(log logger.Logger, dir string, bucket string, client *minio.Client) (chan<- bool, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	cancel := make(chan bool)

	go func() {
		for {
			select {
			case <-watcher.Events:
				if err := SyncDir(log, dir, bucket, client); err != nil {
					log.Failuref("Error syncing directory: %v", err)
				}
			case err := <-watcher.Errors:
				log.Failuref("Error: %v", err)
			case <-cancel:
				if err := watcher.Close(); err != nil {
					log.Failuref("Error closing watcher: %v", err)
				}
			}
		}
	}()

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// if it's a hidden directory, ignore it
			if strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}

			if err := watcher.Add(path); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		close(cancel)

		if err := watcher.Close(); err != nil {
			log.Failuref("Error closing watcher: %v", err)
		}

		return nil, err
	}

	return cancel, nil
}

// SyncDir recursively uploads all files in a directory to an S3 bucket with minio library
func SyncDir(log logger.Logger, dir string, bucket string, client *minio.Client) error {
	log.Actionf("Refreshing bucket %s ...", bucket)

	if err := client.RemoveBucketWithOptions(context.Background(), bucket, minio.RemoveBucketOptions{
		ForceDelete: true,
	}); err != nil {
		// if error is not bucket not found, return error
		if !strings.Contains(err.Error(), "NoSuchBucket") {
			return err
		}
	}

	if err := client.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{}); err != nil {
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
		// upload the file
		_, err = client.FPutObject(context.Background(), bucket, objectName, path, minio.PutObjectOptions{})
		if err != nil {
			return err
		}
		uploadCount = uploadCount + 1
		if uploadCount%10 == 0 {
			fmt.Print(".")
		}
		return nil
	})

	fmt.Println()
	log.Actionf("Uploaded %d files", uploadCount)

	if err != nil {
		log.Failuref("Error syncing directory: %v", err)
		return err
	}

	return nil
}

func RequestReconciliation(ctx context.Context, kubeClient client.Client, namespacedName types.NamespacedName, gvk schema.GroupVersionKind) (string, error) {
	requestAt := time.Now().Format(time.RFC3339Nano)

	return requestAt, retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		object := &metav1.PartialObjectMetadata{}
		object.SetGroupVersionKind(gvk)
		object.SetName(namespacedName.Name)
		object.SetNamespace(namespacedName.Namespace)
		if err := kubeClient.Get(ctx, namespacedName, object); err != nil {
			return err
		}
		patch := client.MergeFrom(object.DeepCopy())
		if ann := object.GetAnnotations(); ann == nil {
			object.SetAnnotations(map[string]string{
				meta.ReconcileRequestAnnotation: requestAt,
			})
		} else {
			ann[meta.ReconcileRequestAnnotation] = requestAt
			object.SetAnnotations(ann)
		}
		err = kubeClient.Patch(ctx, object, patch)
		return err
	})
}

// IsLocalCluster checks if it's a local cluster. See https://skaffold.dev/docs/environment/local-cluster/
func IsLocalCluster(kubeClient *kube.KubeHTTP) bool {
	const (
		// cluster context prefixes
		kindPrefix = "kind-"
		k3dPrefix  = "k3d-"

		// cluster context names
		minikube         = "minikube"
		dockerForDesktop = "docker-for-desktop"
		dockerDesktop    = "docker-desktop"
	)

	if strings.HasPrefix(kubeClient.ClusterName, kindPrefix) ||
		strings.HasPrefix(kubeClient.ClusterName, k3dPrefix) ||
		kubeClient.ClusterName == minikube ||
		kubeClient.ClusterName == dockerForDesktop ||
		kubeClient.ClusterName == dockerDesktop {
		return true
	} else {
		return false
	}
}

// IsPodStatusConditionPresentAndEqual returns true when conditionType is present and equal to status.
func IsPodStatusConditionPresentAndEqual(conditions []corev1.PodCondition, conditionType corev1.PodConditionType, status corev1.ConditionStatus) bool {
	for _, condition := range conditions {
		if condition.Type == conditionType {
			return condition.Status == status
		}
	}

	return false
}
