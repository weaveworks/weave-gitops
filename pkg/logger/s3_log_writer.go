package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/minio/minio-go/v7"

	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/compositehash"
)

type S3LogWriter struct {
	id    string
	s3cli *minio.Client
	log0  Logger
}

const (
	SessionLogBucketName = "gitops-run-logs"
	PodLogBucketName     = "pod-logs"
	SessionLogSource     = "gitops-run-client"
)

func (l *S3LogWriter) L() logr.Logger {
	return l.log0.L()
}

func NewS3LogWriter(minioClient *minio.Client, id string, log0 Logger) (Logger, error) {
	return &S3LogWriter{
		id:    id,
		s3cli: minioClient,
		log0:  log0,
	}, nil
}

func CreateBucket(minioClient *minio.Client, bucketName string) error {
	return minioClient.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{})
}

func (l *S3LogWriter) putLog(msg string) {
	now := time.Now()

	level := "info"

	if msg != "" {
		if strings.HasPrefix(msg, "✗") {
			level = "error"
		} else if strings.HasPrefix(msg, "⚠️") {
			level = "warning"
		}
	}

	key, err := compositehash.New(msg, now)
	if err != nil {
		l.log0.Failuref("failed to create composite hash: %v", err)
	}
	result := &pb.LogEntry{
		SortingKey: fmt.Sprintf("%d", key),
		Timestamp:  now.Format(time.RFC3339),
		Source:     SessionLogSource,
		Level:      level,
		Message:    msg,
	}

	logData, _ := json.Marshal(result)

	// append new line at the end of each log
	logMsg := string(logData) + "\n"

	_, err = l.s3cli.PutObject(context.Background(),
		SessionLogBucketName,
		// This funny pattern 20060102-150405.00000 is the layout needed by time.Format
		fmt.Sprintf("%s/%s.txt", l.id, now.Format("20060102-150405.00000")),
		strings.NewReader(logMsg), int64(len(logMsg)), minio.PutObjectOptions{})
	if err != nil {
		l.log0.Failuref("failed to put log to s3: %v", err)
	}
}

func (l *S3LogWriter) Println(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	l.putLog(msg)
	l.log0.Println(msg)
}

func (l *S3LogWriter) Actionf(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	l.putLog("► " + msg)
	l.log0.Actionf(msg)
}

func (l *S3LogWriter) Failuref(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	l.putLog("✗ " + msg)
	l.log0.Failuref(msg)
}

func (l *S3LogWriter) Generatef(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	l.putLog("✚ " + msg)
	l.log0.Generatef(msg)
}

func (l *S3LogWriter) Successf(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	l.putLog("✔ " + msg)
	l.log0.Successf(msg)
}

func (l *S3LogWriter) Waitingf(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	l.putLog("◎ " + msg)
	l.log0.Waitingf(msg)
}

func (l *S3LogWriter) Warningf(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	l.putLog("⚠️ " + msg)
	l.log0.Warningf(msg)
}
