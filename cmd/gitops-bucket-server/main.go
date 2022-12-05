package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"
	"github.com/weaveworks/weave-gitops/pkg/http"
	"github.com/weaveworks/weave-gitops/pkg/s3"
)

func main() {
	logger := log.New(os.Stdout, "", 0)

	awsAccessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	if awsAccessKeyID == "" {
		minioRootUser := os.Getenv("MINIO_ROOT_USER")
		if minioRootUser == "" {
			logger.Fatal("AWS_ACCESS_KEY_ID or MINIO_ROOT_USER must be set")
			return
		}

		awsAccessKeyID = minioRootUser
	}

	awsSecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if awsSecretAccessKey == "" {
		minioRootPassword := os.Getenv("MINIO_ROOT_PASSWORD")
		if minioRootPassword == "" {
			logger.Fatal("AWS_SECRET_ACCESS_KEY or MINIO_ROOT_PASSWORD must be set")
			return
		}

		awsSecretAccessKey = minioRootPassword
	}

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM)
	defer cancel()

	s3Server := gofakes3.New(s3mem.New(),
		gofakes3.WithAutoBucket(true),
		gofakes3.WithLogger(
			gofakes3.StdLog(
				logger,
				gofakes3.LogErr,
				gofakes3.LogWarn,
				gofakes3.LogInfo,
			))).Server()

	var (
		httpPort, httpsPort int
		certFile, keyFile   string
	)

	flag.IntVar(&httpPort, "http-port", 9000, "TCP port to listen on for HTTP connections")
	flag.IntVar(&httpsPort, "https-port", 9443, "TCP port to listen on for HTTPS connections")
	flag.StringVar(&certFile, "cert-file", "", "Path to the HTTPS server certificate file")
	flag.StringVar(&keyFile, "key-file", "", "Path to the HTTPS server certificate key file")
	flag.Parse()

	if certFile == "" {
		logger.Fatalf("please specify the path to the HTTPS server certificate file")
	}

	if keyFile == "" {
		logger.Fatalf("please specify the path to the HTTPS server certificate key file")
	}

	srv := http.MultiServer{
		HTTPPort:  httpPort,
		HTTPSPort: httpsPort,
		CertFile:  certFile,
		KeyFile:   keyFile,
		Logger:    logger,
	}

	if err := srv.Start(ctx, s3.AuthMiddleware(awsAccessKeyID, awsSecretAccessKey, s3Server)); err != nil {
		logger.Fatalf("server exited unexpectedly: %s", err)
	}
}
