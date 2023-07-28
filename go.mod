module github.com/weaveworks/weave-gitops

go 1.20

require (
	github.com/Masterminds/semver/v3 v3.2.0
	github.com/NYTimes/gziphandler v1.1.1
	github.com/charmbracelet/bubbles v0.14.0
	github.com/charmbracelet/bubbletea v0.22.1
	github.com/charmbracelet/lipgloss v0.6.0
	github.com/cheshir/ttlcache v1.0.1-0.20220504185148-8ceeff21b789
	github.com/coreos/go-oidc/v3 v3.1.0
	github.com/fluxcd/go-git-providers v0.16.0
	github.com/fluxcd/helm-controller/api v0.35.0
	github.com/fluxcd/image-automation-controller/api v0.33.1
	github.com/fluxcd/image-reflector-controller/api v0.27.2
	github.com/fluxcd/kustomize-controller/api v1.0.0
	github.com/fluxcd/notification-controller/api v1.0.0
	github.com/fluxcd/pkg/apis/meta v1.1.1
	github.com/fluxcd/pkg/runtime v0.35.0
	github.com/fluxcd/pkg/ssa v0.27.0
	github.com/fluxcd/source-controller/api v1.0.0
	github.com/go-git/go-git/v5 v5.6.1
	github.com/go-logr/logr v1.2.4
	github.com/go-logr/zapr v1.2.4
	github.com/go-resty/resty/v2 v2.7.0
	github.com/golang-jwt/jwt/v4 v4.4.3
	github.com/google/go-cmp v0.5.9
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.7.1
	github.com/grpc-ecosystem/protoc-gen-grpc-gateway-ts v1.1.1
	github.com/hashicorp/go-multierror v1.1.1
	github.com/johannesboyne/gofakes3 v0.0.0-20220627085814-c3ac35da23b2
	github.com/manifoldco/promptui v0.9.0
	github.com/mattn/go-tty v0.0.4
	github.com/maxbrunsfeld/counterfeiter/v6 v6.4.1
	github.com/minio/minio-go/v7 v7.0.31
	github.com/mitchellh/go-ps v1.0.0
	github.com/oauth2-proxy/mockoidc v0.0.0-20220308204021-b9169deeb282
	github.com/onsi/ginkgo/v2 v2.9.5
	github.com/onsi/gomega v1.27.7
	github.com/pkg/browser v0.0.0-20210911075715-681adbf594b8
	github.com/pkg/errors v0.9.1
	github.com/sabhiram/go-gitignore v0.0.0-20210923224102-525f6e181f06
	github.com/slok/go-http-metrics v0.10.0
	github.com/spf13/cobra v1.7.0
	github.com/spf13/viper v1.13.0
	github.com/tomwright/dasel v1.22.1
	github.com/weaveworks/policy-agent/api v1.0.5
	github.com/weaveworks/tf-controller/tfctl v0.0.0-20230728124123-d0a010c4e700
	github.com/yannh/kubeconform v0.5.0
	go.uber.org/zap v1.24.0
	golang.org/x/crypto v0.8.0
	golang.org/x/oauth2 v0.7.0
	google.golang.org/genproto/googleapis/api v0.0.0-20230525234035-dd9d682886f9
	google.golang.org/grpc v1.57.0
	google.golang.org/protobuf v1.30.0
	gopkg.in/square/go-jose.v2 v2.6.0
	gopkg.in/yaml.v3 v3.0.1
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.27.3
	k8s.io/apiextensions-apiserver v0.27.3
	k8s.io/apimachinery v0.27.3
	k8s.io/cli-runtime v0.26.3
	k8s.io/client-go v0.27.3
	sigs.k8s.io/cli-utils v0.34.0
	sigs.k8s.io/controller-runtime v0.15.0
	sigs.k8s.io/kustomize/api v0.12.1
	sigs.k8s.io/yaml v1.3.0
)

require (
	github.com/Masterminds/sprig v2.22.0+incompatible // indirect
	github.com/atotto/clipboard v0.1.4 // indirect
	github.com/aws/aws-sdk-go v1.44.137 // indirect
	github.com/benbjohnson/clock v1.3.0 // indirect
	github.com/chai2010/gettext-go v1.0.2 // indirect
	github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e // indirect
	github.com/cloudflare/circl v1.3.3 // indirect
	github.com/containerd/console v1.0.3 // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/emicklei/go-restful/v3 v3.10.0 // indirect
	github.com/evanphx/json-patch/v5 v5.6.0 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/golang/glog v1.1.0 // indirect
	github.com/google/gnostic v0.6.9 // indirect
	github.com/google/go-github/v52 v52.0.0 // indirect
	github.com/google/pprof v0.0.0-20210720184732-4bb14d4b1be1 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/hako/durafmt v0.0.0-20210608085754-5c1018a4e16b // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/iancoleman/strcase v0.1.2 // indirect
	github.com/klauspost/cpuid v1.3.1 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/matryer/is v1.4.0 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-localereader v0.0.1 // indirect
	github.com/minio/md5-simd v1.1.0 // indirect
	github.com/minio/sha256-simd v0.1.1 // indirect
	github.com/muesli/ansi v0.0.0-20211018074035-2e021307bc4b // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/muesli/reflow v0.3.0 // indirect
	github.com/muesli/termenv v0.11.1-0.20220212125758-44cd13922739 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/pelletier/go-toml/v2 v2.0.5 // indirect
	github.com/pjbgf/sha1cd v0.3.0 // indirect
	github.com/rs/xid v1.2.1 // indirect
	github.com/ryszard/goskiplist v0.0.0-20150312221310-2dfbae5fcf46 // indirect
	github.com/shabbyrobe/gocovmerge v0.0.0-20180507124511-f6ea450bfb63 // indirect
	github.com/sirupsen/logrus v1.9.0 // indirect
	github.com/skeema/knownhosts v1.1.0 // indirect
	github.com/theckman/yacspin v0.13.12 // indirect
	github.com/weaveworks/tf-controller/api v0.0.0-20230416092146-4a7dfa5b6cc4 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	google.golang.org/genproto v0.0.0-20230526161137-0005af68ea54 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230525234030-28d5490b6b19 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Microsoft/go-winio v0.6.0 // indirect
	github.com/ProtonMail/go-crypto v0.0.0-20230426101702-58e86b294756 // indirect
	github.com/acomagu/bufpipe v1.0.4 // indirect
	github.com/alecthomas/chroma v0.9.2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/clbanning/mxj/v2 v2.3.3-0.20201214204241-e937bdee5a3e // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/danwakefield/fnmatch v0.0.0-20160403171240-cbb64ac3d964 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dlclark/regexp2 v1.4.0 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/evanphx/json-patch v5.6.0+incompatible // indirect
	github.com/exponent-io/jsonpath v0.0.0-20210407135951-1de76d718b3f // indirect
	github.com/fluxcd/pkg/apis/acl v0.1.0 // indirect
	github.com/fluxcd/pkg/apis/kustomize v1.1.1 // indirect
	github.com/fsnotify/fsnotify v1.6.0
	github.com/go-errors/errors v1.4.2 // indirect
	github.com/go-git/gcfg v1.5.0 // indirect
	github.com/go-git/go-billy/v5 v5.4.1 // indirect
	github.com/go-openapi/jsonpointer v0.19.6 // indirect
	github.com/go-openapi/jsonreference v0.20.1 // indirect
	github.com/go-openapi/swag v0.22.3 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.3
	github.com/google/btree v1.1.2 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2
	github.com/hashicorp/go-retryablehttp v0.7.2 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/imdario/mergo v0.3.13 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/klauspost/compress v1.16.0 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/magiconair/properties v1.8.6 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-isatty v0.0.18 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/spdystream v0.2.0 // indirect
	github.com/moby/term v0.0.0-20221205130635-1aeaba878587 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/prometheus/client_golang v1.15.1
	github.com/prometheus/client_model v0.4.0 // indirect
	github.com/prometheus/common v0.42.0 // indirect
	github.com/prometheus/procfs v0.9.0 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sergi/go-diff v1.3.1 // indirect
	github.com/sethvargo/go-limiter v0.7.2
	github.com/spf13/afero v1.8.2 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5
	github.com/subosito/gotenv v1.4.1 // indirect
	github.com/xanzy/go-gitlab v0.83.0 // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	github.com/xlab/treeprint v1.1.0 // indirect
	go.starlark.net v0.0.0-20221028183056-acb66ad56dd2 // indirect
	go.uber.org/atomic v1.10.0
	go.uber.org/multierr v1.8.0 // indirect
	golang.org/x/mod v0.10.0 // indirect
	golang.org/x/net v0.10.0
	golang.org/x/sys v0.8.0 // indirect
	golang.org/x/term v0.8.0
	golang.org/x/text v0.9.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	golang.org/x/tools v0.9.1 // indirect
	gomodules.xyz/jsonpatch/v2 v2.3.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	k8s.io/component-base v0.27.3 // indirect
	k8s.io/klog/v2 v2.100.1 // indirect
	k8s.io/kube-openapi v0.0.0-20230501164219-8b0f38b5fd1f // indirect
	k8s.io/kubectl v0.26.3
	k8s.io/utils v0.0.0-20230505201702-9f6742963106 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/kustomize/kyaml v0.13.9
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
)

// Fix for CVE-2022-1996
// replace k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20220614142933-1062c7ade5f8

// Use patched version that fixed recursive gets, and force delete for buckets
replace github.com/johannesboyne/gofakes3 => github.com/chanwit/gofakes3 v0.0.0-20220715114300-3f51f1961f7b

// Replace k8s.io packages v0.26 to downgrade controller-runtime to v0.14.6
replace (
	k8s.io/api => k8s.io/api v0.26.3
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.26.3
	k8s.io/apimachinery => k8s.io/apimachinery v0.26.3
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.26.3
	k8s.io/client-go => k8s.io/client-go v0.26.3

	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.14.6
)
