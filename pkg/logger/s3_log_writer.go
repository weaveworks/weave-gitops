package logger

import (
	"context"
	"fmt"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/minio/minio-go/v7"
	"github.com/weaveworks/weave-gitops/pkg/s3"
)

type S3LogWriter struct {
	id    string
	s3cli *minio.Client
	log0  Logger
}

const logBucketName = "gitops-run-logs"

func (l *S3LogWriter) L() logr.Logger {
	return l.log0.L()
}

func NewInsecureS3LogWriter(id, endpoint string, accessKey, secretKey string, log0 Logger) (Logger, error) {
	minioClient, err := minio.New(
		endpoint,
		&minio.Options{
			Creds:        credentials.NewStaticV4(accessKey, secretKey, ""),
			Secure:       false,
			BucketLookup: minio.BucketLookupPath,
		},
	)

	if err != nil {
		return nil, err
	}

	if err := minioClient.MakeBucket(context.Background(), logBucketName, minio.MakeBucketOptions{}); err != nil {
		return nil, err
	}

	return &S3LogWriter{
		id:    id,
		s3cli: minioClient,
		log0:  log0,
	}, nil
}

func NewS3LogWriter(id, endpoint string, accessKey, secretKey, caCert []byte, log0 Logger) (Logger, error) {
	minioClient, err := s3.NewMinioClient(endpoint, accessKey, secretKey, caCert)
	if err != nil {
		return nil, err
	}

	if err := minioClient.MakeBucket(context.Background(), logBucketName, minio.MakeBucketOptions{}); err != nil {
		return nil, err
	}

	return &S3LogWriter{
		id:    id,
		s3cli: minioClient,
		log0:  log0,
	}, nil
}

func (l *S3LogWriter) putLog(msg string) {
	// append new line at the end of each log
	msg = msg + "\n"
	_, err := l.s3cli.PutObject(context.Background(),
		logBucketName,
		// This funny pattern 20060102-150405.00000 is the loyout needed by time.Format
		fmt.Sprintf("%s/%s.txt", l.id, time.Now().Format("20060102-150405.00000")),
		strings.NewReader(msg), int64(len(msg)), minio.PutObjectOptions{})

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
