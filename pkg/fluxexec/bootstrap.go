package fluxexec

import (
	"strconv"
	"strings"
)

type bootstrapConfig struct {
	authorEmail           string
	authorName            string
	branch                string
	caFile                string
	clusterDomain         string
	commitMessageAppendix string
	components            []Component
	componentsExtra       []ComponentExtra
	gpgKeyID              string
	gpgKeyRing            string
	gpgPassphrase         string
	imagePullSecret       string
	logLevel              string
	networkPolicy         bool
	privateKeyFile        string
	recurseSubmodules     bool
	registry              string
	secretName            string
	sshECDSACurve         ECDSACurve
	sshHostname           string
	sshKeyAlgorithm       KeyAlgorithm
	sshRSABits            int
	tokenAuth             bool
	tolerationKeys        []string
	watchAllNamespaces    bool
}

var defaultBootstrapOptions = bootstrapConfig{
	authorName:         "Flux",
	branch:             "main",
	clusterDomain:      "cluster.local",
	components:         []Component{ComponentSourceController, ComponentKustomizeController, ComponentHelmController, ComponentNotificationController},
	componentsExtra:    []ComponentExtra{},
	logLevel:           "info",
	networkPolicy:      true,
	recurseSubmodules:  false,
	registry:           "ghcr.io/fluxcd",
	secretName:         "flux-system",
	sshECDSACurve:      ECDSACurveP384,
	sshKeyAlgorithm:    KeyAlgorithmECDSA,
	sshRSABits:         2048,
	tokenAuth:          false,
	tolerationKeys:     []string{},
	watchAllNamespaces: true,
}

// BootstrapOption represents options used in the Bootstrap* methods.
type BootstrapOption interface {
	configureBootstrap(*bootstrapConfig)
}

func (opt *AuthorEmailOption) configureBootstrap(conf *bootstrapConfig) {
	conf.authorEmail = opt.authorEmail
}

func (opt *AuthorNameOption) configureBootstrap(conf *bootstrapConfig) {
	conf.authorName = opt.authorName
}

func (opt *BranchOption) configureBootstrap(conf *bootstrapConfig) {
	conf.branch = opt.branch
}

func (opt *CaFileOption) configureBootstrap(conf *bootstrapConfig) {
	conf.caFile = opt.caFile
}

func (opt *ClusterDomainOption) configureBootstrap(conf *bootstrapConfig) {
	conf.clusterDomain = opt.clusterDomain
}

func (opt *CommitMessageAppendixOption) configureBootstrap(conf *bootstrapConfig) {
	conf.commitMessageAppendix = opt.commitMessageAppendix
}

func (opt *ComponentsOption) configureBootstrap(conf *bootstrapConfig) {
	// allows only source-controller, kustomize-controller, helm-controller, notification-controller
	conf.components = opt.components
}

func (opt *ComponentsExtraOption) configureBootstrap(conf *bootstrapConfig) {
	conf.componentsExtra = opt.componentsExtra
}

func (opt *GPGKeyIDOption) configureBootstrap(conf *bootstrapConfig) {
	conf.gpgKeyID = opt.gpgKeyID
}

func (opt *GPGKeyRingOption) configureBootstrap(conf *bootstrapConfig) {
	conf.gpgKeyRing = opt.gpgKeyRing
}

func (opt *GPGPassphraseOption) configureBootstrap(conf *bootstrapConfig) {
	conf.gpgPassphrase = opt.gpgPassphrase
}

func (opt *ImagePullSecretOption) configureBootstrap(conf *bootstrapConfig) {
	conf.imagePullSecret = opt.imagePullSecret
}

func (opt *LogLevelOption) configureBootstrap(conf *bootstrapConfig) {
	conf.logLevel = opt.logLevel
}

func (opt *NetworkPolicyOption) configureBootstrap(conf *bootstrapConfig) {
	conf.networkPolicy = opt.networkPolicy
}

func (opt *PrivateKeyFileOption) configureBootstrap(conf *bootstrapConfig) {
	conf.privateKeyFile = opt.privateKeyFile
}

func (opt *RecurseSubmodulesOption) configureBootstrap(conf *bootstrapConfig) {
	conf.recurseSubmodules = opt.recurseSubmodules
}

func (opt *RegistryOption) configureBootstrap(conf *bootstrapConfig) {
	conf.registry = opt.registry
}

func (opt *SecretNameOption) configureBootstrap(conf *bootstrapConfig) {
	conf.secretName = opt.secretName
}

func (opt *SSHECDSACurveOption) configureBootstrap(conf *bootstrapConfig) {
	conf.sshECDSACurve = opt.sshECDSACurve
}

func (opt *SSHHostnameOption) configureBootstrap(conf *bootstrapConfig) {
	conf.sshHostname = opt.sshHostname
}

func (opt *SSHKeyAlgorithmOption) configureBootstrap(conf *bootstrapConfig) {
	conf.sshKeyAlgorithm = opt.sshKeyAlgorithm
}

func (opt *SSHRSABitsOption) configureBootstrap(conf *bootstrapConfig) {
	conf.sshRSABits = opt.sshRSABits
}

func (opt *TokenAuthOption) configureBootstrap(conf *bootstrapConfig) {
	conf.tokenAuth = opt.tokenAuth
}

func (opt *TolerationKeysOption) configureBootstrap(conf *bootstrapConfig) {
	conf.tolerationKeys = opt.tolerationKeys
}

func (opt *WatchAllNamespacesOption) configureBootstrap(conf *bootstrapConfig) {
	conf.watchAllNamespaces = opt.watchAllNamespaces
}

// BootstrapOptions is a special set of options for the (parent) bootstrap command.
type BootstrapOptions struct {
	bootstrapOptions []BootstrapOption
}

// TODO: add more git provider options for the bootstrap command.
func (opt *BootstrapOptions) configureBootstrapGitHub(conf *bootstrapGitHubConfig) {
	conf.bootstrapOptions = opt.bootstrapOptions
}

func WithBootstrapOptions(opts ...BootstrapOption) *BootstrapOptions {
	return &BootstrapOptions{opts}
}

func (flux *Flux) bootstrapArgs(opts ...BootstrapOption) []string {
	c := defaultBootstrapOptions
	for _, o := range opts {
		o.configureBootstrap(&c)
	}

	args := []string{}

	if c.authorEmail != "" {
		args = append(args, "--author-email", c.authorEmail)
	}

	if c.authorName != "" {
		args = append(args, "--author-name", c.authorName)
	}

	if c.branch != "" {
		args = append(args, "--branch", c.branch)
	}

	if c.caFile != "" {
		args = append(args, "--ca-file", c.caFile)
	}

	if c.clusterDomain != "" {
		args = append(args, "--cluster-domain", c.clusterDomain)
	}

	if c.commitMessageAppendix != "" {
		args = append(args, "--commit-message-appendix", c.commitMessageAppendix)
	}

	if c.components != nil {
		var comps []string
		for _, c := range c.components {
			comps = append(comps, string(c))
		}

		args = append(args, "--components", strings.Join(comps, ","))
	}

	if len(c.componentsExtra) > 0 {
		var extras []string
		for _, c := range c.componentsExtra {
			extras = append(extras, string(c))
		}

		args = append(args, "--components-extra", strings.Join(extras, ","))
	}

	if c.gpgKeyID != "" {
		args = append(args, "--gpg-key-id", c.gpgKeyID)
	}

	if c.gpgKeyRing != "" {
		args = append(args, "--gpg-key-ring", c.gpgKeyRing)
	}

	if c.gpgPassphrase != "" {
		args = append(args, "--gpg-passphrase", c.gpgPassphrase)
	}

	if c.imagePullSecret != "" {
		args = append(args, "--image-pull-secret", c.imagePullSecret)
	}

	if c.logLevel != "" {
		args = append(args, "--log-level", c.logLevel)
	}

	if c.networkPolicy {
		args = append(args, "--network-policy")
	}

	if c.privateKeyFile != "" {
		args = append(args, "--private-key-file", c.privateKeyFile)
	}

	if c.recurseSubmodules {
		args = append(args, "--recurse-submodules")
	}

	if c.registry != "" {
		args = append(args, "--registry", c.registry)
	}

	if c.secretName != "" {
		args = append(args, "--secret-name", c.secretName)
	}

	if c.sshECDSACurve != "" {
		args = append(args, "--ssh-ecdsa-curve", string(c.sshECDSACurve))
	}

	if c.sshHostname != "" {
		args = append(args, "--ssh-hostname", c.sshHostname)
	}

	if c.sshKeyAlgorithm != "" {
		args = append(args, "--ssh-key-algorithm", string(c.sshKeyAlgorithm))
	}

	if c.sshRSABits != 0 {
		args = append(args, "--ssh-rsa-bits", strconv.Itoa(c.sshRSABits))
	}

	if c.tokenAuth {
		args = append(args, "--token-auth")
	}

	if len(c.tolerationKeys) > 0 {
		args = append(args, "--toleration-keys", strings.Join(c.tolerationKeys, ","))
	}

	if c.watchAllNamespaces {
		args = append(args, "--watch-all-namespaces")
	}

	return args
}
