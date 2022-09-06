package fluxexec

import (
	"os"
	"path/filepath"
	"strconv"
)

type globalConfig struct {
	as                    string
	asGroup               []string
	asUid                 string
	cacheDir              string
	certificateAuthority  string
	clientCertificate     string
	clientKey             string
	cluster               string
	context               string
	insecureSkipTlsVerify bool
	kubeApiBurst          int
	kubeApiQps            float32
	kubeconfig            string
	namespace             string
	server                string
	timeout               string
	tlsServerName         string
	token                 string
	user                  string
	verbose               bool
}

func defaultCacheDir() string {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return filepath.Join(homedir, ".kube", "cache")
}

var defaultGlobalOptions = globalConfig{
	cacheDir:     defaultCacheDir(),
	kubeApiBurst: 100,
	kubeApiQps:   50,
	namespace:    "flux-system",
	timeout:      "5m0s",
}

// GlobalOption represents options used in the Global* methods.
type GlobalOption interface {
	configureGlobal(*globalConfig)
}

func (opt *AsOption) configureGlobal(conf *globalConfig) {
	conf.as = opt.as
}

func (opt *AsGroupOption) configureGlobal(conf *globalConfig) {
	conf.asGroup = opt.asGroup
}

func (opt *AsUidOption) configureGlobal(conf *globalConfig) {
	conf.asUid = opt.asUid
}

func (opt *CacheDirOption) configureGlobal(conf *globalConfig) {
	conf.cacheDir = opt.cacheDir
}

func (opt *CertificateAuthorityOption) configureGlobal(conf *globalConfig) {
	conf.certificateAuthority = opt.certificateAuthority
}

func (opt *ClientCertificateOption) configureGlobal(conf *globalConfig) {
	conf.clientCertificate = opt.clientCertificate
}

func (opt *ClientKeyOption) configureGlobal(conf *globalConfig) {
	conf.clientKey = opt.clientKey
}

func (opt *ClusterOption) configureGlobal(conf *globalConfig) {
	conf.cluster = opt.cluster
}

func (opt *ContextOption) configureGlobal(conf *globalConfig) {
	conf.context = opt.context
}

func (opt *InsecureSkipTlsVerifyOption) configureGlobal(conf *globalConfig) {
	conf.insecureSkipTlsVerify = opt.insecureSkipTlsVerify
}

func (opt *KubeApiBurstOption) configureGlobal(conf *globalConfig) {
	conf.kubeApiBurst = opt.kubeApiBurst
}

func (opt *KubeApiQpsOption) configureGlobal(conf *globalConfig) {
	conf.kubeApiQps = opt.kubeApiQps
}

func (opt *KubeconfigOption) configureGlobal(conf *globalConfig) {
	conf.kubeconfig = opt.kubeconfig
}

func (opt *NamespaceOption) configureGlobal(conf *globalConfig) {
	conf.namespace = opt.namespace
}

func (opt *ServerOption) configureGlobal(conf *globalConfig) {
	conf.server = opt.server
}

func (opt *TimeoutOption) configureGlobal(conf *globalConfig) {
	conf.timeout = opt.timeout
}

func (opt *TLSServerNameOption) configureGlobal(conf *globalConfig) {
	conf.tlsServerName = opt.tlsServerName
}

func (opt *TokenOption) configureGlobal(conf *globalConfig) {
	conf.token = opt.token
}

func (opt *UserOption) configureGlobal(conf *globalConfig) {
	conf.user = opt.user
}

func (opt *VerboseOption) configureGlobal(conf *globalConfig) {
	conf.verbose = opt.verbose
}

// GlobalOptions is a special set of options for the (parent) global command.
type GlobalOptions struct {
	globalOptions []GlobalOption
}

func (opt *GlobalOptions) configureBootstrapGitHub(conf *bootstrapGitHubConfig) {
	conf.globalOptions = opt.globalOptions
}

func WithGlobalOptions(opts ...GlobalOption) *GlobalOptions {
	return &GlobalOptions{opts}
}

func (flux *Flux) globalArgs(opts ...GlobalOption) []string {
	c := defaultGlobalOptions
	for _, o := range opts {
		o.configureGlobal(&c)
	}

	args := []string{}

	if c.as != "" {
		args = append(args, "--as", c.as)
	}

	if len(c.asGroup) > 0 {
		// this flag can be repeated to add multiple groups
		for _, g := range c.asGroup {
			args = append(args, "--as-group", g)
		}
	}

	if c.asUid != "" {
		args = append(args, "--as-uid", c.asUid)
	}

	if c.cacheDir != "" {
		args = append(args, "--cache-dir", c.cacheDir)
	}

	if c.certificateAuthority != "" {
		args = append(args, "--certificate-authority", c.certificateAuthority)
	}

	if c.clientCertificate != "" {
		args = append(args, "--client-certificate", c.clientCertificate)
	}

	if c.clientKey != "" {
		args = append(args, "--client-key", c.clientKey)
	}

	if c.cluster != "" {
		args = append(args, "--cluster", c.cluster)
	}

	if c.context != "" {
		args = append(args, "--context", c.context)
	}

	if c.insecureSkipTlsVerify {
		args = append(args, "--insecure-skip-tls-verify")
	}

	if c.kubeApiBurst != 0 {
		args = append(args, "--kube-api-burst", strconv.Itoa(c.kubeApiBurst))
	}

	if c.kubeApiQps != 0 {
		args = append(args, "--kube-api-qps", strconv.FormatFloat(float64(c.kubeApiQps), 'f', -1, 32))
	}

	if c.kubeconfig != "" {
		args = append(args, "--kubeconfig", c.kubeconfig)
	}

	if c.namespace != "" {
		args = append(args, "--namespace", c.namespace)
	}

	if c.server != "" {
		args = append(args, "--server", c.server)
	}

	if c.timeout != "" {
		args = append(args, "--timeout", c.timeout)
	}

	if c.tlsServerName != "" {
		args = append(args, "--tls-server-name", c.tlsServerName)
	}

	if c.token != "" {
		args = append(args, "--token", c.token)
	}

	if c.user != "" {
		args = append(args, "--user", c.user)
	}

	if c.verbose {
		args = append(args, "--verbose")
	}

	return args
}
