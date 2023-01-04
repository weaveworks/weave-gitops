package fluxexec

import (
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"time"
)

type globalConfig struct {
	as                    string
	asGroup               []string
	asUID                 string
	cacheDir              string
	certificateAuthority  string
	clientCertificate     string
	clientKey             string
	cluster               string
	context               string
	insecureSkipTLSVerify bool
	kubeAPIBurst          int
	kubeAPIQPS            float32
	kubeconfig            string
	namespace             string
	server                string
	timeout               time.Duration
	tlsServerName         string
	token                 string
	user                  string
	verbose               bool
	version               string
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
	kubeAPIBurst: 100,
	kubeAPIQPS:   50,
	namespace:    "flux-system",
	timeout:      5 * time.Minute,
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

func (opt *AsUIDOption) configureGlobal(conf *globalConfig) {
	conf.asUID = opt.asUID
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

func (opt *InsecureSkipTLSVerifyOption) configureGlobal(conf *globalConfig) {
	conf.insecureSkipTLSVerify = opt.insecureSkipTLSVerify
}

func (opt *KubeAPIBurstOption) configureGlobal(conf *globalConfig) {
	conf.kubeAPIBurst = opt.kubeAPIBurst
}

func (opt *KubeAPIQPSOption) configureGlobal(conf *globalConfig) {
	conf.kubeAPIQPS = opt.kubeAPIQPS
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

func (opt *VersionOption) configureGlobal(conf *globalConfig) {
	conf.version = opt.version
}

// GlobalOptions is a special set of options for the (parent) global command.
type GlobalOptions struct {
	globalOptions []GlobalOption
}

func (opt *GlobalOptions) configureBootstrapGitHub(conf *bootstrapGitHubConfig) {
	conf.globalOptions = opt.globalOptions
}

func (opt *GlobalOptions) configureBootstrapGitLab(conf *bootstrapGitLabConfig) {
	conf.globalOptions = opt.globalOptions
}

func (opt *GlobalOptions) configureBootstrapBitbucketServer(conf *bootstrapBitbucketServerConfig) {
	conf.globalOptions = opt.globalOptions
}

func (opt *GlobalOptions) configureBootstrapGit(conf *bootstrapGitConfig) {
	conf.globalOptions = opt.globalOptions
}

func (opt *GlobalOptions) configureInstall(conf *installConfig) {
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

	if c.as != "" && !reflect.DeepEqual(c.as, defaultGlobalOptions.as) {
		args = append(args, "--as", c.as)
	}

	if len(c.asGroup) > 0 && !reflect.DeepEqual(c.asGroup, defaultGlobalOptions.asGroup) {
		// this flag can be repeated to add multiple groups
		for _, g := range c.asGroup {
			args = append(args, "--as-group", g)
		}
	}

	if c.asUID != "" && !reflect.DeepEqual(c.asUID, defaultGlobalOptions.asUID) {
		args = append(args, "--as-uid", c.asUID)
	}

	if c.cacheDir != "" && !reflect.DeepEqual(c.cacheDir, defaultGlobalOptions.cacheDir) {
		args = append(args, "--cache-dir", c.cacheDir)
	}

	if c.certificateAuthority != "" && !reflect.DeepEqual(c.certificateAuthority, defaultGlobalOptions.certificateAuthority) {
		args = append(args, "--certificate-authority", c.certificateAuthority)
	}

	if c.clientCertificate != "" && !reflect.DeepEqual(c.clientCertificate, defaultGlobalOptions.clientCertificate) {
		args = append(args, "--client-certificate", c.clientCertificate)
	}

	if c.clientKey != "" && !reflect.DeepEqual(c.clientKey, defaultGlobalOptions.clientKey) {
		args = append(args, "--client-key", c.clientKey)
	}

	if c.cluster != "" && !reflect.DeepEqual(c.cluster, defaultGlobalOptions.cluster) {
		args = append(args, "--cluster", c.cluster)
	}

	if c.context != "" && !reflect.DeepEqual(c.context, defaultGlobalOptions.context) {
		args = append(args, "--context", c.context)
	}

	if c.insecureSkipTLSVerify && !reflect.DeepEqual(c.insecureSkipTLSVerify, defaultGlobalOptions.insecureSkipTLSVerify) {
		args = append(args, "--insecure-skip-tls-verify")
	}

	if c.kubeAPIBurst != 0 && !reflect.DeepEqual(c.kubeAPIBurst, defaultGlobalOptions.kubeAPIBurst) {
		args = append(args, "--kube-api-burst", strconv.Itoa(c.kubeAPIBurst))
	}

	if c.kubeAPIQPS != 0 && !reflect.DeepEqual(c.kubeAPIQPS, defaultGlobalOptions.kubeAPIQPS) {
		args = append(args, "--kube-api-qps", strconv.FormatFloat(float64(c.kubeAPIQPS), 'f', -1, 32))
	}

	if c.kubeconfig != "" && !reflect.DeepEqual(c.kubeconfig, defaultGlobalOptions.kubeconfig) {
		args = append(args, "--kubeconfig", c.kubeconfig)
	}

	if c.namespace != "" && !reflect.DeepEqual(c.namespace, defaultGlobalOptions.namespace) {
		args = append(args, "--namespace", c.namespace)
	}

	if c.server != "" && !reflect.DeepEqual(c.server, defaultGlobalOptions.server) {
		args = append(args, "--server", c.server)
	}

	if c.timeout.Nanoseconds() != 0 && !reflect.DeepEqual(c.timeout, defaultGlobalOptions.timeout) {
		args = append(args, "--timeout", c.timeout.String())
	}

	if c.tlsServerName != "" && !reflect.DeepEqual(c.tlsServerName, defaultGlobalOptions.tlsServerName) {
		args = append(args, "--tls-server-name", c.tlsServerName)
	}

	if c.token != "" && !reflect.DeepEqual(c.token, defaultGlobalOptions.token) {
		args = append(args, "--token", c.token)
	}

	if c.user != "" && !reflect.DeepEqual(c.user, defaultGlobalOptions.user) {
		args = append(args, "--user", c.user)
	}

	if c.verbose && !reflect.DeepEqual(c.verbose, defaultGlobalOptions.verbose) {
		args = append(args, "--verbose")
	}

	if c.version != "" && !reflect.DeepEqual(c.version, defaultGlobalOptions.version) {
		args = append(args, "--version", c.version)
	}

	return args
}
