module github.com/weaveworks/weave-gitops

go 1.16

require (
	github.com/deepmap/oapi-codegen v1.8.1
	github.com/dnaeon/go-vcr v1.2.0
	github.com/fluxcd/go-git-providers v0.2.1-0.20210908163615-833963032787
	github.com/fluxcd/helm-controller/api v0.11.1
	github.com/fluxcd/kustomize-controller/api v0.13.2
	github.com/fluxcd/pkg/apis/meta v0.10.0
	github.com/fluxcd/source-controller/api v0.15.3
	github.com/go-git/go-billy/v5 v5.3.1
	github.com/go-git/go-git/v5 v5.4.2
	github.com/go-logr/logr v1.1.0
	github.com/go-logr/zapr v1.1.0
	github.com/golang-jwt/jwt/v4 v4.0.0
	github.com/google/go-cmp v0.5.6
	github.com/google/uuid v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.5.0
	github.com/grpc-ecosystem/protoc-gen-grpc-gateway-ts v1.1.1
	github.com/jandelgado/gcov2lcov v1.0.5
	github.com/lithammer/dedent v1.1.0
	github.com/mattn/go-isatty v0.0.13
	github.com/maxbrunsfeld/counterfeiter/v6 v6.4.1
	github.com/olekukonko/tablewriter v0.0.5
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	github.com/ory/go-acc v0.2.6
	github.com/pkg/browser v0.0.0-20210706143420-7d21f8c997e2
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.7.0
	github.com/weaveworks/go-checkpoint v0.0.0-20170503165305-ebbb8b0518ab
	go.uber.org/zap v1.19.0
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97 // indirect
	golang.org/x/net v0.0.0-20210510120150-4163338589ed // indirect
	golang.org/x/oauth2 v0.0.0-20210615190721-d04028783cf1
	golang.org/x/term v0.0.0-20210615171337-6886f2dfbf5b // indirect
	google.golang.org/genproto v0.0.0-20210617175327-b9e0b3197ced
	google.golang.org/grpc v1.38.0
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.1.0
	google.golang.org/protobuf v1.27.1
	k8s.io/api v0.21.2
	k8s.io/apiextensions-apiserver v0.21.2
	k8s.io/apimachinery v0.21.2
	k8s.io/client-go v0.21.2
	k8s.io/klog/v2 v2.9.0 // indirect
	k8s.io/kube-openapi v0.0.0-20210421082810-95288971da7e // indirect
	sigs.k8s.io/controller-runtime v0.9.1
	sigs.k8s.io/controller-tools v0.4.1
	sigs.k8s.io/kustomize/kstatus v0.0.2
	sigs.k8s.io/yaml v1.2.0

)

// https://github.com/gorilla/websocket/security/advisories/GHSA-jf24-p9p9-4rjh
replace github.com/gorilla/websocket v0.0.0 => github.com/gorilla/websocket v1.4.1

replace github.com/go-logr/logr v1.1.0 => github.com/go-logr/logr v0.4.0

replace github.com/go-logr/zapr v1.1.0 => github.com/go-logr/zapr v0.4.0
