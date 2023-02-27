package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/weaveworks/weave-gitops/pkg/compositehash"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	coretypes "github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/run/constants"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// identical string can be used in the UI to test for the secret not found condition.
	secretNotFound = "secret not found"
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

	sessionNamespace := msg.GetSessionNamespace()
	sessionName := msg.GetSessionId()

	var clusterName string
	if sessionName == "no-session" {
		clusterName = cluster.DefaultCluster
	} else {
		// this is the virtual cluster name
		clusterName = sessionNamespace + "/" + sessionName
	}

	cli, err := clustersClient.Scoped(clusterName)
	if err != nil {
		retErr := fmt.Errorf("session %s not found: %w", clusterName, err)
		return &pb.GetSessionLogsResponse{Error: retErr.Error()}, retErr
	}

	fluxNamespace, err := cs.getFluxNamespace(ctx, cli)
	if err != nil {
		// assume flux-system if we can't find the flux namespace
		fluxNamespace = "flux-system"
	}

	logSourceFilter := msg.GetLogSourceFilter()
	isLoadingGitOpsRunLogs := logSourceFilter == "" || logSourceFilter == logger.SessionLogSource

	if isLoadingGitOpsRunLogs {
		// check if we can get session logs already
		// if the secret is not created yet, we should not display an error in the browser console
		if err = isSecretCreated(ctx, cli, constants.GitOpsRunNamespace, constants.RunDevBucketCredentials); err != nil {
			return &pb.GetSessionLogsResponse{Error: err.Error()}, nil
		}

		if err = isSecretCreated(ctx, cli, fluxNamespace, constants.RunDevBucketCredentials); err != nil {
			return &pb.GetSessionLogsResponse{Error: err.Error()}, nil
		}
	}

	info, err := getBucketConnectionInfo(ctx, clusterName, fluxNamespace, cli)
	if err != nil {
		return &pb.GetSessionLogsResponse{Error: err.Error()}, err
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
		return &pb.GetSessionLogsResponse{Error: err.Error()}, err
	}

	var (
		logEntries []*pb.LogEntry
		firstToken string
	)

	if isLoadingGitOpsRunLogs {
		// get gitops-run logs
		gitopsRunLogs, token, err := getGitOpsRunLogs(
			ctx,
			sessionName,
			msg.GetToken(),
			asS3Reader(minioClient),
			logger.SessionLogBucketName,
			msg.GetLogLevelFilter())
		if err != nil {
			return &pb.GetSessionLogsResponse{Error: err.Error()}, err
		}
		logEntries = append(logEntries, gitopsRunLogs...)

		firstToken = token
	}

	// get pod-logs
	podLogs, secondToken, logSources, err := getPodLogs(
		ctx,
		msg.GetToken(),
		asS3Reader(minioClient),
		logger.PodLogBucketName,
		logSourceFilter,
		msg.GetLogLevelFilter())
	if err != nil {
		return &pb.GetSessionLogsResponse{Error: err.Error()}, err
	}
	logEntries = append(logEntries, podLogs...)

	// we sort the logs by timestamp
	sort.Slice(logEntries, func(i, j int) bool {
		tsi, _ := strconv.ParseInt(logEntries[i].GetSortingKey(), 10, 64)
		tsj, _ := strconv.ParseInt(logEntries[j].GetSortingKey(), 10, 64)
		return tsi < tsj
	})

	return &pb.GetSessionLogsResponse{
		Logs:       logEntries,
		NextToken:  firstToken + "," + secondToken,
		LogSources: append([]string{logger.SessionLogSource}, logSources...),
	}, nil
}

func isSecretCreated(ctx context.Context, cli client.Client, namespace string, name string) error {
	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	if err := cli.Get(ctx, client.ObjectKeyFromObject(&secret), &secret); err != nil {
		return fmt.Errorf("%s in the namespace %s: %w", secretNotFound, namespace, err)
	}

	return nil
}

func detectLogLevel(message string) string {
	message = strings.ToLower(message)
	errorRegex := regexp.MustCompile(`(err(or)?|fatal|ftl)`)
	warnRegex := regexp.MustCompile(`(warn(ing)?|wrn)`)

	if errorRegex.MatchString(message) {
		return "error"
	} else if warnRegex.MatchString(message) {
		return "warning"
	} else {
		return "info"
	}
}

type PodLog struct {
	Date       time.Time `json:"date"`
	Time       time.Time `json:"time"`
	Log        string    `json:"log"`
	Msg        string    `json:"msg"`
	Level      string    `json:"level"`
	Kubernetes struct {
		PodName        string            `json:"pod_name"`
		NamespaceName  string            `json:"namespace_name"`
		PodID          string            `json:"pod_id"`
		Labels         map[string]string `json:"labels"`
		Annotations    map[string]string `json:"annotations"`
		Host           string            `json:"host"`
		ContainerName  string            `json:"container_name"`
		DockerID       string            `json:"docker_id"`
		ContainerHash  string            `json:"container_hash"`
		ContainerImage string            `json:"container_image"`
	} `json:"kubernetes"`
}

func getPodLogs(ctx context.Context, nextToken string, minioClient s3Reader, bucketName string, logSourceFilter string, logLevelFilter string) ([]*pb.LogEntry, string, []string, error) {
	// we use the second part of the token as the startAfter value
	if strings.Contains(nextToken, ",") {
		parts := strings.SplitN(nextToken, ",", 2)
		nextToken = parts[1]
	}

	var (
		logs      []*pb.LogEntry
		lastToken string
	)

	tmpLogSources := map[string]struct{}{}
	for obj := range minioClient.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Prefix:     "fluent-bit-logs",
		StartAfter: nextToken,
		Recursive:  true,
	}) {
		if obj.Err != nil {
			return nil, "", nil, obj.Err
		}

		content, err := minioClient.GetObject(ctx, bucketName, obj.Key, minio.GetObjectOptions{})
		if err != nil {
			return nil, "", nil, err
		}

		b, err := io.ReadAll(content)
		if err != nil {
			return nil, "", nil, err
		}

		messages := strings.Split(string(b), "\n")
		for _, message := range messages {
			if strings.TrimSpace(message) == "" {
				continue
			}

			podLog := &PodLog{}
			if err := json.Unmarshal([]byte(message), podLog); err != nil {
				return nil, "", nil, err
			}

			innerMessage := podLog.Log
			if innerMessage == "" {
				innerMessage = podLog.Msg
			}

			if innerMessage == "&{   }" || innerMessage == "" {
				// found an empty log entry,
				continue
			}

			var loggingSource string
			if podLog.Kubernetes.Labels != nil {
				loggingSource = podLog.Kubernetes.NamespaceName + "/" + podLog.Kubernetes.Labels["app"]
			} else {
				loggingSource = podLog.Kubernetes.NamespaceName + "/" + podLog.Kubernetes.PodName
			}

			tmpLogSources[loggingSource] = struct{}{}

			var logLevel string
			if podLog.Level == "" {
				logLevel = detectLogLevel(innerMessage)
			} else {
				logLevel = detectLogLevel(podLog.Level)
			}

			if logSourceFilter != "" && !strings.Contains(loggingSource, logSourceFilter) {
				continue
			}

			if logLevelFilter != "" && !strings.Contains(logLevel, logLevelFilter) {
				continue
			}

			hash, err := compositehash.New(innerMessage, podLog.Time)
			if err != nil {
				return nil, "", nil, err
			}

			logs = append(logs, &pb.LogEntry{
				SortingKey: fmt.Sprintf("%d", hash),
				Timestamp:  podLog.Time.Format(time.RFC3339),
				Level:      logLevel,
				Message:    innerMessage,
				Source:     loggingSource,
			})
		}
		lastToken = obj.Key
	}

	var logSources []string
	for k := range tmpLogSources {
		logSources = append(logSources, k)
	}
	sort.Strings(logSources)

	return logs, lastToken, logSources, nil
}

func getGitOpsRunLogs(ctx context.Context, sessionID string, nextToken string, minioClient s3Reader, bucketName string, logLevelFilter string) ([]*pb.LogEntry, string, error) {
	// we use the first part of the token as the startAfter value
	if strings.Contains(nextToken, ",") {
		parts := strings.SplitN(nextToken, ",", 2)
		nextToken = parts[0]
	}

	var (
		logs      []*pb.LogEntry
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

		logEntry := &pb.LogEntry{}
		err = json.Unmarshal(b, logEntry)
		if err != nil {
			return nil, "", err
		}

		// filter by log level
		// blank log level filter means we want all logs
		if logLevelFilter == "" || logEntry.Level == logLevelFilter {
			logs = append(logs, logEntry)
		}
		lastToken = obj.Key
	}

	return logs, lastToken, nil
}

func getBucketConnectionInfo(ctx context.Context, clusterName string, fluxNamespace string, cli client.Client) (*bucketConnectionInfo, error) {
	// get secret
	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.RunDevBucketCredentials,
			Namespace: fluxNamespace,
		},
	}

	if err := cli.Get(ctx, client.ObjectKeyFromObject(&secret), &secret); err != nil {
		return nil, err
	}

	// get bucket source
	bucket := sourcev1.Bucket{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.RunDevBucketName,
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
