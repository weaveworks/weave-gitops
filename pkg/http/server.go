package http

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

// MultiServer lets you create and run an HTTP server that serves over both, HTTP and HTTPS. It is a convenience wrapper around net/http and crypto/tls.
type MultiServer struct {
	HTTPPort  int
	HTTPSPort int
	CertFile  string
	KeyFile   string
	Logger    *log.Logger
}

// Start creates listeners for HTTP and HTTPS and starts serving requests using the provided handler. The function blocks until both servers
// are properly shut down. A shutdown can be initiated by cancelling the given context.
func (srv MultiServer) Start(ctx context.Context, handler http.Handler) error {
	var wg sync.WaitGroup

	tlsListener, err := createTLSListener(srv.HTTPSPort, srv.CertFile, srv.KeyFile)
	if err != nil {
		return fmt.Errorf("failed to create TLS listener: %w", err)
	}

	wg.Add(1)

	go func() {
		defer wg.Done()
		startServer(ctx, handler, tlsListener, srv.Logger)
	}()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", srv.HTTPPort))
	if err != nil {
		return fmt.Errorf("failed to create TCP listener: %w", err)
	}

	wg.Add(1)

	go func() {
		defer wg.Done()
		startServer(ctx, handler, listener, srv.Logger)
	}()

	wg.Wait()

	return nil
}

func createTLSListener(port int, certFile, keyFile string) (net.Listener, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("unable to load TLS key pair: %w", err)
	}

	listener, err := tls.Listen("tcp", fmt.Sprintf(":%d", port), &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12})
	if err != nil {
		return nil, fmt.Errorf("unable to start TLS listener: %w", err)
	}

	return listener, nil
}

func startServer(ctx context.Context, hndlr http.Handler, listener net.Listener, logger *log.Logger) {
	srv := http.Server{
		Addr:              listener.Addr().String(),
		Handler:           hndlr,
		ReadHeaderTimeout: 5 * time.Second,
	}
	logger.Printf("https://%s", srv.Addr)

	go func() {
		if err := srv.Serve(listener); !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("server quit unexpectedly: %s", err)
		}
	}()
	<-ctx.Done()
	logger.Printf("shutting down %s", listener.Addr())

	if err := srv.Shutdown(ctx); err != nil && !errors.Is(err, context.Canceled) {
		logger.Printf("error shutting down %s: %s", listener.Addr(), err)
	}
}
