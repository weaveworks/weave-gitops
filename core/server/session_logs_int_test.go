package server

import (
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"
	. "github.com/onsi/gomega"
	logger2 "github.com/weaveworks/weave-gitops/pkg/logger"
	"net/http/httptest"
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
	s3logger, err := logger2.NewInsecureS3LogWriter("session-id", strings.TrimPrefix(s.URL, "http://"), "gitops", "gitops123", log0)
	g.Expect(err).ShouldNot(HaveOccurred())

	s3logger.Actionf("test action")
	s3logger.Failuref("test failure")
	s3logger.Successf("test success")
	s3logger.Waitingf("test waiting")

	minioClient, err := minio.New(
		strings.TrimPrefix(s.URL, "http://"),
		&minio.Options{
			Creds:        credentials.NewStaticV4("test", "test", ""),
			Secure:       false,
			BucketLookup: minio.BucketLookupPath,
		},
	)
	g.Expect(err).ShouldNot(HaveOccurred())

	lines, next, err := getLogs(context.Background(),
		"session-id",
		"",
		asS3Reader(minioClient),
		"gitops-run-logs",
	)

	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(lines).Should(HaveLen(4))
	g.Expect(lines[0]).Should(ContainSubstring("► test action\n"))
	g.Expect(lines[1]).Should(ContainSubstring("✗ test failure\n"))
	g.Expect(lines[2]).Should(ContainSubstring("✔ test success\n"))
	g.Expect(lines[3]).Should(ContainSubstring("◎ test waiting\n"))

	s3logger.Actionf("round 2 - test action")
	s3logger.Failuref("round 2 - test failure")

	round2, _, err := getLogs(context.Background(),
		"session-id",
		next,
		asS3Reader(minioClient),
		"gitops-run-logs",
	)

	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(round2).Should(HaveLen(2))

	g.Expect(round2[0]).Should(ContainSubstring("► round 2 - test action\n"))
	g.Expect(round2[1]).Should(ContainSubstring("✗ round 2 - test failure\n"))
}
