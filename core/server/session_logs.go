package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	coretypes "github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/flux"
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

func (cs *coreServer) getFluxNamespace(ctx context.Context, k8sClient client.Client) (string, error) {
	namespaceList := corev1.NamespaceList{}
	opts := client.MatchingLabels{
		coretypes.PartOfLabel: FluxNamespacePartOf,
	}

	var ns *corev1.Namespace

	err := k8sClient.List(ctx, &namespaceList, opts)
	if err != nil {
		return "", fmt.Errorf("error getting list of objects")
	} else {
		for _, item := range namespaceList.Items {
			if item.GetLabels()[flux.VersionLabelKey] != "" {
				ns = &item
				break
			}
		}
	}

	if ns == nil {
		return "", fmt.Errorf("no flux namespace found")
	}

	labels := ns.GetLabels()
	if labels == nil {
		return "", fmt.Errorf("error getting labels")
	}

	return ns.GetName(), nil
}

// GetSessionLogs returns the logs for a session.
func (cs *coreServer) GetSessionLogs(ctx context.Context, msg *pb.GetSessionLogsRequest) (*pb.GetSessionLogsResponse, error) {
	clustersClient, err := cs.clustersManager.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		return nil, fmt.Errorf("error getting impersonating client: %w", err)
	}

	// cli will be scoped to the vcluster
	virtualClusterName := msg.GetSessionNamespace() + "/" + msg.GetSessionId()
	cli, err := clustersClient.Scoped(virtualClusterName)
	if err != nil {
		return nil, fmt.Errorf("getting cluster client: %w", err)
	}

	fluxNamespace, err := cs.getFluxNamespace(ctx, cli)
	if err != nil {
		// assume flux-system if we can't find the flux namespace
		fluxNamespace = "flux-system"
	}

	info, err := getBucketConnectionInfo(ctx, virtualClusterName, fluxNamespace, cli)
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

	logEntries := []*pb.LogEntry{}

	for _, log := range logs {
		logEntry := &pb.LogEntry{}
		err = json.Unmarshal([]byte(log), &logEntry)
		if err != nil {
			return nil, err
		}

		logEntries = append(logEntries, logEntry)
	}

	return &pb.GetSessionLogsResponse{
		Logs:      logEntries,
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

func getBucketConnectionInfo(ctx context.Context, clusterName string, fluxNamespace string, cli client.Client) (*bucketConnectionInfo, error) {

	const sourceName = "run-dev-bucket"
	var secretName = sourceName + "-credentials"

	// get secret
	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: fluxNamespace,
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
			Namespace: fluxNamespace,
		},
	}

	if err := cli.Get(ctx, client.ObjectKeyFromObject(&bucket), &bucket); err != nil {
		return nil, err
	}

	// Endpoint format
	// fmt.Sprintf("%s.%s.svc.cluster.local:%d", RunDevBucketName, GitOpsRunNamespace, params.DevBucketPort),
	bucketEndpoint := bucket.Spec.Endpoint
	if clusterName != "Default" {
		parts := strings.Split(clusterName, "/")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid cluster name for vcluster session: %s", clusterName)
		}
		sessionNamespace := parts[0]
		sessionName := parts[1]

		parts = strings.Split(bucket.Spec.Endpoint, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid bucket endpoint for vcluster session: %s", bucketEndpoint)
		}
		port := parts[1]
		bucketEndpoint = fmt.Sprintf("%s-bucket.%s.svc:%s", sessionName, sessionNamespace, port)
	}

	return &bucketConnectionInfo{
		accessKey:      string(secret.Data["accesskey"]),
		secretKey:      string(secret.Data["secretkey"]),
		bucketEndpoint: bucketEndpoint,
		bucketInsecure: bucket.Spec.Insecure,
	}, nil
}
