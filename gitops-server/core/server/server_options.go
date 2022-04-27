package server

import "github.com/weaveworks/weave-gitops/gitops-server/pkg/kube"

// CoreOptions includes all the options that can be set for an
// CoreServer.
type CoreOptions struct {
	ClientGetter kube.ClientGetter
}

// CoreOption defines the signature of a function that can be used
// to set an option for an CoreServer.
type CoreOption func(*CoreOptions)

// WithClientGetter allows for setting a ClientGetter.
func WithClientGetter(clientGetter kube.ClientGetter) CoreOption {
	return func(args *CoreOptions) {
		args.ClientGetter = clientGetter
	}
}
