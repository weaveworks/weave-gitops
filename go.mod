module github.com/weaveworks/weave-gitops

go 1.16

require (
	github.com/dnaeon/go-vcr v1.2.0
	github.com/fluxcd/go-git-providers v0.1.1
	github.com/go-git/go-billy/v5 v5.3.1
	github.com/go-git/go-git/v5 v5.4.1
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/gorilla/websocket v1.4.1 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.5.0
	github.com/hashicorp/go-retryablehttp v0.6.8 // indirect
	github.com/json-iterator/go v1.1.11 // indirect
	github.com/lithammer/dedent v1.1.0
	github.com/maxbrunsfeld/counterfeiter/v6 v6.4.1
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.0.0
	github.com/stretchr/testify v1.7.0
	github.com/weaveworks/go-checkpoint v0.0.0-20170503165305-ebbb8b0518ab
	golang.org/x/net v0.0.0-20210525063256-abc453219eb5 // indirect
	golang.org/x/term v0.0.0-20210615171337-6886f2dfbf5b
	golang.org/x/time v0.0.0-20210611083556-38a9dc6acbc6 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20210617175327-b9e0b3197ced
	google.golang.org/grpc v1.39.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	k8s.io/api v0.21.0-alpha.2 // indirect
	k8s.io/apimachinery v0.21.0-alpha.2
	k8s.io/klog/v2 v2.8.0 // indirect
	sigs.k8s.io/controller-runtime v0.7.2
	sigs.k8s.io/structured-merge-diff/v4 v4.1.0 // indirect
	sigs.k8s.io/yaml v1.2.0

)

// https://github.com/gorilla/websocket/security/advisories/GHSA-jf24-p9p9-4rjh
replace github.com/gorilla/websocket v0.0.0 => github.com/gorilla/websocket v1.4.1
