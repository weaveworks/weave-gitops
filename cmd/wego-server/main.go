package main

import (
	"math/rand"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/weave-gitops/cmd/wego-server/app"
)

func init() {
	if os.Getenv("LOG_LEVEL") == "DEBUG" {
		// Only log the warning severity or above.
		log.SetLevel(log.DebugLevel)
	} else if os.Getenv("LOG_LEVEL") == "WARN" {
		// Only log the warning severity or above.
		log.SetLevel(log.WarnLevel)
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	command := app.NewAPIServerCommand()

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
