package kube

import (
	"context"

	"k8s.io/client-go/rest"
)

// ConfigGetter implementations should extract the details from a context and
// create a *rest.Config for use in clients.
type ConfigGetter interface {
	Config(ctx context.Context) *rest.Config
}
