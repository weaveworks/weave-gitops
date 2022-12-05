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
)

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM)
	defer cancel()

	logger := log.New(os.Stdout, "", 0)
	backend := s3mem.New()
	s3 := gofakes3.New(backend,
		gofakes3.WithAutoBucket(true),
		gofakes3.WithLogger(gofakes3.StdLog(logger, gofakes3.LogErr, gofakes3.LogWarn, gofakes3.LogInfo)))
	s3Server := s3.Server()

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

	if err := srv.Start(ctx, s3Server); err != nil {
		logger.Fatalf("server exited unexpectedly: %s", err)
	}
}
