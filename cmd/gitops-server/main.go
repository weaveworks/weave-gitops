package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/helm/watcher"
	"github.com/weaveworks/weave-gitops/pkg/helm/watcher/cache"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/server"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	command := NewAPIServerCommand()

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}

const addr = "0.0.0.0:8000"

func NewAPIServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "gitops-server",
		Long: `The gitops-server handles HTTP requests for Weave GitOps Applications`,

		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			flux.New(osys.New(), &runner.CLIRunner{}).SetupBin()

			appConfig, err := server.DefaultApplicationsConfig()
			if err != nil {
				return err
			}

			rest, clusterName, err := kube.RestConfig()
			if err != nil {
				return fmt.Errorf("could not create client config: %w", err)
			}

			_, rawClient, err := kube.NewKubeHTTPClientWithConfig(rest, clusterName)
			if err != nil {
				return fmt.Errorf("could not create kube http client: %w", err)
			}

			helmCache := cache.NewCache()
			helmWatcher, err := watcher.NewWatcher(rawClient, helmCache)
			if err != nil {
				return err
			}
			// TODO: check this error
			go helmWatcher.StartWatcher()

			// Create the cache here as well and pass it in through the profiles Config thing.
			runtimeNamespace := os.Getenv("RUNTIME_NAMESPACE")
			if runtimeNamespace == "" {
				runtimeNamespace = "default"
			}
			profilesConfig := server.NewProfilesConfig(rawClient, helmCache, runtimeNamespace, "weaveworks-charts")

			s, err := server.NewHandlers(context.Background(), &server.Config{AppConfig: appConfig, ProfilesConfig: profilesConfig})
			if err != nil {
				return err
			}

			appConfig.Logger.Info("server starting", "address", addr)
			return http.ListenAndServe(addr, s)
		},
	}

	return cmd
}
