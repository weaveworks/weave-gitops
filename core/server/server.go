package server

import (
	"context"
	"fmt"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/weaveworks/weave-gitops/core/clientset"
	"github.com/weaveworks/weave-gitops/core/gitops/app"
	"github.com/weaveworks/weave-gitops/core/gitops/kustomize"
	"github.com/weaveworks/weave-gitops/core/gitops/source"
	pb "github.com/weaveworks/weave-gitops/pkg/api/app"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

func Hydrate(ctx context.Context, mux *runtime.ServeMux, config *rest.Config) error {
	appKubeCreator := app.NewKubeCreator()
	appFetcher := app.NewKubeAppFetcher()

	kustCreator := kustomize.NewK8sCreator()
	kustFetcher := kustomize.NewKustomizationFetcher()

	sourceCreator := source.NewKubeCreator()
	sourceFetcher := source.NewSourceFetcher()

	clientSet := clientset.NewClientSets(config)

	appsServer := NewAppServer(clientSet, appKubeCreator, kustCreator, sourceCreator, appFetcher, kustFetcher, sourceFetcher)
	if err := pb.RegisterAppsHandlerServer(ctx, mux, appsServer); err != nil {
		return fmt.Errorf("could not register new app server: %w", err)
	}

	fluxServer := NewFluxServer(clientSet, kustCreator, sourceCreator, kustFetcher, sourceFetcher)
	if err := pb.RegisterFluxHandlerServer(ctx, mux, fluxServer); err != nil {
		return fmt.Errorf("could not register new kustomization server: %w", err)
	}

	return nil
}

func intervalDuration(input *pb.Interval) metav1.Duration {
	return metav1.Duration{Duration: time.Duration(input.Hours)*time.Hour + time.Duration(input.Minutes)*time.Minute + time.Duration(input.Seconds)*time.Second}
}
