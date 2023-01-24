package server

import (
	"context"
	"fmt"
	"io"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type s3Reader interface {
	ListObjects(ctx context.Context, bucketName string, opts minio.ListObjectsOptions) <-chan minio.ObjectInfo
	GetObject(ctx context.Context, bucketName string, objectName string, opts minio.GetObjectOptions) (io.ReadCloser, error)
}

type s3ReaderWrapper struct {
	r *minio.Client
}

func asS3Reader(r *minio.Client) s3Reader {
	return &s3ReaderWrapper{r: r}
}

func (s *s3ReaderWrapper) ListObjects(ctx context.Context, bucketName string, opts minio.ListObjectsOptions) <-chan minio.ObjectInfo {
	return s.r.ListObjects(ctx, bucketName, opts)
}

func (s *s3ReaderWrapper) GetObject(ctx context.Context, bucketName string, objectName string, opts minio.GetObjectOptions) (io.ReadCloser, error) {
	return s.r.GetObject(ctx, bucketName, objectName, opts)
}

type bucketConnectionInfo struct {
	accessKey      string
	secretKey      string
	bucketEndpoint string
	bucketInsecure bool
}

// GetSessionLogs returns the logs for a session.
func (cs *coreServer) GetSessionLogs(ctx context.Context, msg *pb.GetSessionLogsRequest) (*pb.GetSessionLogsResponse, error) {
	clustersClient, err := cs.clustersManager.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		return nil, fmt.Errorf("error getting impersonating client: %w", err)
	}

	cli, err := clustersClient.Scoped(msg.GetClusterName())
	if err != nil {
		return nil, fmt.Errorf("getting cluster client: %w", err)
	}

	info, err := getBucketConnectionInfo(ctx, msg.GetNamespace(), cli)
	if err != nil {
		return nil, err
	}

	minioClient, err := minio.New(
		info.bucketEndpoint,
		&minio.Options{
			Creds:        credentials.NewStaticV4(info.accessKey, info.secretKey, ""),
			Secure:       !info.bucketInsecure,
			BucketLookup: minio.BucketLookupPath,
		},
	)
	if err != nil {
		return nil, err
	}

	logs, lastToken, err := getLogs(ctx, msg.GetSessionId(), msg.GetToken(), asS3Reader(minioClient), logger.SessionLogBucketName)
	if err != nil {
		return nil, err
	}

	return &pb.GetSessionLogsResponse{
		Logs:      logs,
		NextToken: lastToken,
	}, nil
}

func getLogs(ctx context.Context, sessionID string, nextToken string, minioClient s3Reader, bucketName string) ([]string, string, error) {
	var (
		logs      []string
		lastToken string
	)

	for obj := range minioClient.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Prefix:     sessionID,
		StartAfter: nextToken,
		Recursive:  true,
	}) {
		if obj.Err != nil {
			return nil, "", obj.Err
		}

		content, err := minioClient.GetObject(ctx, bucketName, obj.Key, minio.GetObjectOptions{})
		if err != nil {
			return nil, "", err
		}

		b, err := io.ReadAll(content)
		if err != nil {
			return nil, "", err
		}

		if err := content.Close(); err != nil {
			return nil, "", err
		}

		logs = append(logs, string(b))
		lastToken = obj.Key
	}

	return logs, lastToken, nil
}

func getBucketConnectionInfo(ctx context.Context, namespace string, cli client.Client) (*bucketConnectionInfo, error) {
	const sourceName = "run-dev-bucket"

	var secretName = sourceName + "-credentials"

	// get secret
	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
	}

	// get secret
	if err := cli.Get(ctx, client.ObjectKeyFromObject(&secret), &secret); err != nil {
		return nil, err
	}

	// get bucket source
	bucket := sourcev1.Bucket{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sourceName,
			Namespace: namespace,
		},
	}

	if err := cli.Get(ctx, client.ObjectKeyFromObject(&bucket), &bucket); err != nil {
		return nil, err
	}

	return &bucketConnectionInfo{
		accessKey:      string(secret.Data["accesskey"]),
		secretKey:      string(secret.Data["secretkey"]),
		bucketEndpoint: bucket.Spec.Endpoint,
		bucketInsecure: bucket.Spec.Insecure,
	}, nil
}
