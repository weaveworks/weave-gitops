package server

import (
	"context"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"net/http/httptest"

	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	logger2 "github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/run/constants"
)

func TestGetSessionLogsIntegration(t *testing.T) {
	g := NewGomegaWithT(t)

	logger := log.New(os.Stdout, "", 0)

	s3Server := gofakes3.New(s3mem.New(),
		gofakes3.WithAutoBucket(true),
		gofakes3.WithLogger(
			gofakes3.StdLog(
				logger,
				gofakes3.LogErr,
				gofakes3.LogWarn,
				gofakes3.LogInfo,
			))).Server()

	s := httptest.NewServer(s3Server)
	defer s.Close()

	log0 := logger2.NewCLILogger(os.Stdout)
	insecureMinioClient, err := minio.New(
		strings.TrimPrefix(s.URL, "http://"),
		&minio.Options{
			Creds:        credentials.NewStaticV4("gitops", "gitops123", ""),
			Secure:       false,
			BucketLookup: minio.BucketLookupPath,
		},
	)
	g.Expect(err).ShouldNot(HaveOccurred())
	s3logger, err := logger2.NewS3LogWriter(insecureMinioClient, "session-id", log0)
	g.Expect(err).ShouldNot(HaveOccurred())

	s3logger.Actionf("test action")
	s3logger.Failuref("test failure")
	s3logger.Successf("test success")
	s3logger.Waitingf("test waiting")
	s3logger.Warningf("test warning")

	minioClient, err := minio.New(
		strings.TrimPrefix(s.URL, "http://"),
		&minio.Options{
			Creds:        credentials.NewStaticV4("test", "test", ""),
			Secure:       false,
			BucketLookup: minio.BucketLookupPath,
		},
	)
	g.Expect(err).ShouldNot(HaveOccurred())

	logEntries, next, err := getGitOpsRunLogs(context.Background(),
		"session-id",
		"",
		asS3Reader(minioClient),
		logger2.SessionLogBucketName,
		"",
	)

	g.Expect(err).ShouldNot(HaveOccurred())

	g.Expect(logEntries).Should(HaveLen(5))
	g.Expect(logEntries[0].Message).Should(Equal("► test action"))
	g.Expect(logEntries[0].Level).Should(Equal("info"))
	g.Expect(logEntries[0].Source).Should(Equal(logger2.SessionLogSource))

	g.Expect(logEntries[1].Message).Should(Equal("✗ test failure"))
	g.Expect(logEntries[1].Level).Should(Equal("error"))
	g.Expect(logEntries[1].Source).Should(Equal(logger2.SessionLogSource))

	g.Expect(logEntries[2].Message).Should(Equal("✔ test success"))
	g.Expect(logEntries[2].Level).Should(Equal("info"))
	g.Expect(logEntries[2].Source).Should(Equal(logger2.SessionLogSource))

	g.Expect(logEntries[3].Message).Should(Equal("◎ test waiting"))
	g.Expect(logEntries[3].Level).Should(Equal("info"))
	g.Expect(logEntries[3].Source).Should(Equal(logger2.SessionLogSource))

	g.Expect(logEntries[4].Message).Should(Equal("⚠️ test warning"))
	g.Expect(logEntries[4].Level).Should(Equal("warning"))
	g.Expect(logEntries[4].Source).Should(Equal(logger2.SessionLogSource))

	s3logger.Actionf("round 2 - test action")
	s3logger.Failuref("round 2 - test failure")

	logEntries, _, err = getGitOpsRunLogs(context.Background(),
		"session-id",
		next,
		asS3Reader(minioClient),
		logger2.SessionLogBucketName,
		"",
	)

	g.Expect(err).ShouldNot(HaveOccurred())

	g.Expect(logEntries[0].Message).Should(Equal("► round 2 - test action"))
	g.Expect(logEntries[0].Level).Should(Equal("info"))
	g.Expect(logEntries[0].Source).Should(Equal(logger2.SessionLogSource))

	g.Expect(logEntries[1].Message).Should(Equal("✗ round 2 - test failure"))
	g.Expect(logEntries[1].Level).Should(Equal("error"))
	g.Expect(logEntries[1].Source).Should(Equal(logger2.SessionLogSource))
}

func TestIsSecretCreatedSecretFound(t *testing.T) {
	g := NewGomegaWithT(t)

	scheme, err := kube.CreateScheme()
	g.Expect(err).ShouldNot(HaveOccurred())

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.RunDevBucketCredentials,
			Namespace: constants.GitOpsRunNamespace,
		},
		Data: map[string][]byte{
			"key": []byte("value"),
		},
	}

	cli := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(secret).Build()

	err = isSecretCreated(context.Background(), cli, constants.GitOpsRunNamespace, constants.RunDevBucketCredentials)

	g.Expect(err).ShouldNot(HaveOccurred())
}

func TestIsSecretCreatedSecretNotFound(t *testing.T) {
	g := NewGomegaWithT(t)

	scheme, err := kube.CreateScheme()
	g.Expect(err).ShouldNot(HaveOccurred())

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-name",
			Namespace: "test-namespace",
		},
		Data: map[string][]byte{
			"key": []byte("value"),
		},
	}

	cli := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(secret).Build()

	err = isSecretCreated(context.Background(), cli, constants.GitOpsRunNamespace, constants.RunDevBucketCredentials)

	g.Expect(err).Should(HaveOccurred())
}
