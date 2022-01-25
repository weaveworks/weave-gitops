package server

import (
	"context"
	"fmt"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	pb "github.com/weaveworks/weave-gitops/pkg/api/app"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

func Hydrate(ctx context.Context, mux *runtime.ServeMux, config *rest.Config) error {
	appsServer := NewAppServer(config)
	if err := pb.RegisterAppsHandlerServer(ctx, mux, appsServer); err != nil {
		return fmt.Errorf("could not register new app server: %w", err)
	}

	return nil
}

func intervalDuration(input *pb.Interval) metav1.Duration {
	return metav1.Duration{Duration: time.Duration(input.Hours)*time.Hour + time.Duration(input.Minutes)*time.Minute + time.Duration(input.Seconds)*time.Second}
}
