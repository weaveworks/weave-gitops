module github.com/weaveworks/weave-gitops/controllers/wego-controller

go 1.15

require (
	github.com/fluxcd/source-controller/api v0.12.1
	github.com/go-logr/logr v0.3.0
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	k8s.io/apimachinery v0.20.4
	k8s.io/client-go v0.20.2
	sigs.k8s.io/controller-runtime v0.8.3
)
