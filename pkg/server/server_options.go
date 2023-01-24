package server

import (
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/services/crd"
)

// ApplicationsOptions includes all the options that can be set for an
// ApplicationsServer.
type ApplicationsOptions struct {
	ClientGetter kube.ClientGetter
	CRDFetcher   crd.Fetcher
}

// ApplicationsOption defines the signature of a function that can be used
// to set an option for an ApplicationsServer.
type ApplicationsOption func(*ApplicationsOptions)

// WithClientGetter allows for setting a ClientGetter.
func WithClientGetter(clientGetter kube.ClientGetter) ApplicationsOption {
	return func(args *ApplicationsOptions) {
		args.ClientGetter = clientGetter
	}
}

// WithCRDFetcher allows for setting a CRDFetcher.
func WithCRDFetcher(fetcher crd.Fetcher) ApplicationsOption {
	return func(args *ApplicationsOptions) {
		args.CRDFetcher = fetcher
	}
}
