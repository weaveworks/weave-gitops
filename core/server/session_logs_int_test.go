package server

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"net/http/httptest"

	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"
	. "github.com/onsi/gomega"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	logger2 "github.com/weaveworks/weave-gitops/pkg/logger"
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

	lines, next, err := getLogs(context.Background(),
		"session-id",
		"",
		asS3Reader(minioClient),
		logger2.SessionLogBucketName,
	)

	g.Expect(err).ShouldNot(HaveOccurred())

	logEntries := []*pb.LogEntry{}

	for _, line := range lines {
		logEntry := &pb.LogEntry{}
		err = json.Unmarshal([]byte(line), &logEntry)
		g.Expect(err).ShouldNot(HaveOccurred())

		logEntries = append(logEntries, logEntry)
	}

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

	lines2, _, err := getLogs(context.Background(),
		"session-id",
		next,
		asS3Reader(minioClient),
		logger2.SessionLogBucketName,
	)

	g.Expect(err).ShouldNot(HaveOccurred())

	logEntries = nil

	for _, line := range lines2 {
		logEntry := &pb.LogEntry{}
		err = json.Unmarshal([]byte(line), &logEntry)
		g.Expect(err).ShouldNot(HaveOccurred())

		logEntries = append(logEntries, logEntry)
	}

	g.Expect(lines2).Should(HaveLen(2))

	g.Expect(logEntries[0].Message).Should(Equal("► round 2 - test action"))
	g.Expect(logEntries[0].Level).Should(Equal("info"))
	g.Expect(logEntries[0].Source).Should(Equal(logger2.SessionLogSource))

	g.Expect(logEntries[1].Message).Should(Equal("✗ round 2 - test failure"))
	g.Expect(logEntries[1].Level).Should(Equal("error"))
	g.Expect(logEntries[1].Source).Should(Equal(logger2.SessionLogSource))
}
