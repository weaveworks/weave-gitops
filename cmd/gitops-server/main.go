package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/weaveworks/weave-gitops/pkg/helm/watcher"
	"github.com/weaveworks/weave-gitops/pkg/helm/watcher/cache"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/server"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	command := NewAPIServerCommand()

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}

const (
	addr                    = "0.0.0.0:8000"
	metricsBindAddress      = ":9980"
	healthzBindAddress      = ":9981"
	notificationBindAddress = "http://notification-controller./"
	watcherPort             = 9443
)

func NewAPIServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "gitops-server",
		Long: `The gitops-server handles HTTP requests for Weave GitOps Applications`,

		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
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

			tmpDir, err := os.MkdirTemp("", "profile_cache_location")
			if err != nil {
				return fmt.Errorf("failed to create helm cache: %w", err)
			}

			profileCache, err := cache.NewCache(tmpDir)
			if err != nil {
				return fmt.Errorf("failed to create profile cache: %w", err)
			}

			profileWatcher, err := watcher.NewWatcher(watcher.Options{
				KubeClient:                    rawClient,
				Cache:                         profileCache,
				MetricsBindAddress:            metricsBindAddress,
				HealthzBindAddress:            healthzBindAddress,
				NotificationControllerAddress: notificationBindAddress,
				WatcherPort:                   watcherPort,
			})
			if err != nil {
				return fmt.Errorf("failed to create watcher: %w", err)
			}

			go func() {
				if err := profileWatcher.StartWatcher(); err != nil {
					appConfig.Logger.Error(err, "failed to start profile watcher")
					os.Exit(1)
				}
			}()

			profilesConfig := server.NewProfilesConfig(kube.ClusterConfig{
				DefaultConfig: rest,
				ClusterName:   clusterName,
			}, profileCache, "default", "weaveworks-charts")

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
