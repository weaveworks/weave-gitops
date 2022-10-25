package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"
	"net/http/httptest"
)

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGKILL)
	defer cancel()

	logger := log.New(os.Stdout, "", 0)
	backend := s3mem.New()
	s3 := gofakes3.New(backend,
		gofakes3.WithAutoBucket(true),
		gofakes3.WithLogger(gofakes3.StdLog(logger, gofakes3.LogErr, gofakes3.LogWarn, gofakes3.LogInfo)))

	port := "9000"
	// check args
	if len(os.Args) > 1 {
		port = os.Args[1]
		// part string to integer
		_, err := strconv.Atoi(port)
		if err != nil {
			log.Fatalf("Invalid port number: %s", port)
		}
	}

	// create a listener with the desired port.
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}

	ts := httptest.NewUnstartedServer(s3.Server())
	if err := ts.Listener.Close(); err != nil {
		log.Fatal(err)
	}

	ts.Listener = listener
	// Start the server.
	ts.Start()
	defer ts.Close()

	logger.Println(ts.URL)

	<-ctx.Done()
}
