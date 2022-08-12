package cmderrors

import "errors"

var (
	ErrNoWGEEndpoint          = errors.New("the Weave GitOps Enterprise HTTP API endpoint flag (--endpoint) has not been set")
	ErrNoURL                  = errors.New("the URL flag (--url) has not been set")
	ErrNoTLSCertOrKey         = errors.New("flags --tls-cert-file and --tls-private-key-file cannot be empty")
	ErrNoFilePath             = errors.New("the filepath has not been set")
	ErrMultipleFilePaths      = errors.New("only one filepath is allowed")
	ErrNoContextForKubeConfig = errors.New("no context provided for the kubeconfig")
	ErrNoCluster              = errors.New("no cluster in the kube config")
	ErrGetKubeClient          = errors.New("error getting Kube HTTP client")
	ErrNoDashboardName        = errors.New("the dashboard name has not been set")
	ErrMultipleDashboardNames = errors.New("only one dashboard name is allowed")
)
