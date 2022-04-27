module github.com/weaveworks/weave-gitops/gitops-server

go 1.17

require (
	github.com/coreos/go-oidc/v3 v3.1.0
	github.com/fluxcd/helm-controller/api v0.19.0
	github.com/fluxcd/kustomize-controller/api v0.23.0
	github.com/fluxcd/pkg/apis/meta v0.12.1
	github.com/fluxcd/pkg/ssa v0.6.0
	github.com/fluxcd/source-controller/api v0.22.5
	github.com/go-logr/logr v1.2.2
	github.com/golang-jwt/jwt/v4 v4.0.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.7.1
	github.com/oauth2-proxy/mockoidc v0.0.0-20210703044157-382d3faf2671
	github.com/onsi/ginkgo/v2 v2.1.3
	github.com/sethvargo/go-limiter v0.7.2
	github.com/spf13/cobra v1.3.0
	golang.org/x/crypto v0.0.0-20211215153901-e495a2d5b3d3
	google.golang.org/genproto v0.0.0-20220228195345-15d65a4533f7
	google.golang.org/grpc v1.45.0-dev.0.20220209221444-a354b1eec350
	k8s.io/api v0.23.4
	sigs.k8s.io/cli-utils v0.26.0
	sigs.k8s.io/controller-runtime v0.11.1
	sigs.k8s.io/kustomize/kstatus v0.0.2
)

require (
	github.com/go-logr/zapr v1.2.2
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/go-cmp v0.5.7
	github.com/onsi/gomega v1.18.1
	github.com/pkg/errors v0.9.1 // indirect
	github.com/stretchr/testify v1.7.1 // indirect
	go.uber.org/zap v1.21.0
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8
	google.golang.org/protobuf v1.27.1
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	k8s.io/apiextensions-apiserver v0.23.4 // indirect
	k8s.io/apimachinery v0.23.4
	k8s.io/client-go v0.23.4
	sigs.k8s.io/yaml v1.3.0 // indirect
)

require (
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/benbjohnson/clock v1.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/evanphx/json-patch v4.12.0+incompatible // indirect
	github.com/fluxcd/pkg/apis/acl v0.0.3 // indirect
	github.com/fluxcd/pkg/apis/kustomize v0.3.2 // indirect
	github.com/go-errors/errors v1.4.0 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.19.5 // indirect
	github.com/go-openapi/swag v0.19.14 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kr/pretty v0.2.1 // indirect
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.12.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/xlab/treeprint v0.0.0-20181112141820-a009c3971eca // indirect
	go.starlark.net v0.0.0-20200306205701-8dd3e2ee1dd5 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	golang.org/x/net v0.0.0-20220225172249-27dd8689420f // indirect
	golang.org/x/sys v0.0.0-20220319134239-a9b59b0215f8 // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	google.golang.org/appengine v1.6.7 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/square/go-jose.v2 v2.5.1 // indirect
	k8s.io/cli-runtime v0.23.4 // indirect
	k8s.io/klog/v2 v2.50.0 // indirect
	k8s.io/kube-openapi v0.0.0-20211115234752-e816edb12b65 // indirect
	k8s.io/kubectl v0.22.1 // indirect
	k8s.io/utils v0.0.0-20220210201930-3a6ce19ff2f9 // indirect
	sigs.k8s.io/json v0.0.0-20211208200746-9f7c6b3444d2 // indirect
	sigs.k8s.io/kustomize/api v0.10.1 // indirect
	sigs.k8s.io/kustomize/kyaml v0.13.0 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.1 // indirect
)

// fix CVE-2021-41103
// fix GHSA-mvff-h3cj-wj9c
replace github.com/containerd/containerd => github.com/containerd/containerd v1.5.9

// fix CVE-2021-30465
// fix GHSA-v95c-p5hm-xq8f
replace github.com/opencontainers/runc => github.com/opencontainers/runc v1.0.3

// https://github.com/gorilla/websocket/security/advisories/GHSA-jf24-p9p9-4rjh
replace github.com/gorilla/websocket v0.0.0 => github.com/gorilla/websocket v1.4.1

// https://security.snyk.io/vuln/SNYK-GOLANG-GITHUBCOMOPENCONTAINERSIMAGESPECSPECSGOV1-1922832
replace github.com/opencontainers/image-spec v1.0.1 => github.com/opencontainers/image-spec v1.0.2

// Fix for CVE-2020-29652: https://github.com/golang/crypto/commit/8b5274cf687fd9316b4108863654cc57385531e8
// Fix for CVE-2021-43565: https://github.com/golang/crypto/commit/5770296d904e90f15f38f77dfc2e43fdf5efc083
replace golang.org/x/crypto => golang.org/x/crypto v0.0.0-20211215153901-e495a2d5b3d3

// Fix for CVE-2020-26160: https://github.com/advisories/GHSA-w73w-5m7g-f7qc
replace github.com/dgrijalva/jwt-go v3.2.0+incompatible => github.com/golang-jwt/jwt/v4 v4.1.0

// CVE-2014-3499
replace github.com/docker/docker v0.7.3-0.20190327010347-be7ac8be2ae0 => github.com/docker/docker v1.6.1

// CVE-2020-8552
replace k8s.io/apimachinery v0.0.0-20190913080033-27d36303b655 => k8s.io/apimachinery v0.19.0

// CVE-2020-8552
replace k8s.io/apiserver v0.0.0-20190918160949-bfa5e2e684ad => k8s.io/apiserver v0.17.3

replace k8s.io/apiextensions-apiserver v0.0.0-20190918161926-8f644eb6e783 => k8s.io/apiextensions-apiserver v0.20.2

replace k8s.io/api v0.0.0-20190918155943-95b840bb6a1f => k8s.io/api v0.17.0
