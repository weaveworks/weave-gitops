package fluxexec

type ClusterDomainOption struct {
	clusterDomain string
}

// ClusterDomain represents the --cluster-domain flag.
func ClusterDomain(clusterDomain string) *ClusterDomainOption {
	return &ClusterDomainOption{clusterDomain}
}

type ComponentsOption struct {
	components []Component
}

// Components represents the --components flag.
func Components(components ...Component) *ComponentsOption {
	allowedComponents := map[Component]struct{}{
		ComponentSourceController:       {},
		ComponentKustomizeController:    {},
		ComponentHelmController:         {},
		ComponentNotificationController: {},
	}
	// validate components
	validatedComponents := []Component{}

	for _, c := range components {
		if _, ok := allowedComponents[c]; !ok {
			continue
		}

		validatedComponents = append(validatedComponents, c)
	}

	return &ComponentsOption{validatedComponents}
}

type ComponentsExtraOption struct {
	componentsExtra []ComponentExtra
}

// ComponentsExtra represents the --components-extra flag.
func ComponentsExtra(componentsExtra ...ComponentExtra) *ComponentsExtraOption {
	allowedComponents := map[ComponentExtra]struct{}{
		ComponentImageAutomationController: {},
		ComponentImageReflectorController:  {},
	}
	// validate components
	validatedComponents := []ComponentExtra{}

	for _, c := range componentsExtra {
		if _, ok := allowedComponents[c]; !ok {
			continue
		}

		validatedComponents = append(validatedComponents, c)
	}

	return &ComponentsExtraOption{validatedComponents}
}

type ExportOption struct {
	export bool
}

// Export represents the --export flag.
func Export(export bool) *ExportOption {
	return &ExportOption{export}
}

type ImagePullSecretOption struct {
	imagePullSecret string
}

// ImagePullSecret represents the --image-pull-secret flag.
func ImagePullSecret(imagePullSecret string) *ImagePullSecretOption {
	return &ImagePullSecretOption{imagePullSecret}
}

type LogLevelOption struct {
	logLevel string
}

// LogLevel represents the --log-level flag.
func LogLevel(logLevel string) *LogLevelOption {
	return &LogLevelOption{logLevel}
}

type NetworkPolicyOption struct {
	networkPolicy bool
}

// NetworkPolicy represents the --network-policy flag.
func NetworkPolicy(networkPolicy bool) *NetworkPolicyOption {
	return &NetworkPolicyOption{networkPolicy}
}

type RegistryOption struct {
	registry string
}

// Registry represents the --registry flag.
func Registry(registry string) *RegistryOption {
	return &RegistryOption{registry}
}

type TolerationKeysOption struct {
	tolerationKeys []string
}

// TolerationKeys represents the --toleration-keys flag.
func TolerationKeys(tolerationKeys ...string) *TolerationKeysOption {
	return &TolerationKeysOption{tolerationKeys}
}

type WatchAllNamespacesOption struct {
	watchAllNamespaces bool
}

// WatchAllNamespaces represents the --watch-all-namespaces flag.
func WatchAllNamespaces(watchAllNamespaces bool) *WatchAllNamespacesOption {
	return &WatchAllNamespacesOption{watchAllNamespaces}
}

type HostnameOption struct {
	hostname string
}

// Hostname represents the --hostname flag.
func Hostname(hostname string) *HostnameOption {
	return &HostnameOption{hostname}
}

type IntervalOption struct {
	interval string
}

// Interval represents the --interval flag.
func Interval(interval string) *IntervalOption {
	return &IntervalOption{interval}
}

type OwnerOption struct {
	owner string
}

// Owner represents the --owner flag.
func Owner(owner string) *OwnerOption {
	return &OwnerOption{owner}
}

type PathOption struct {
	path string
}

// Path represents the --path flag.
func Path(path string) *PathOption {
	return &PathOption{path}
}

type PersonalOption struct {
	personal bool
}

// Personal represents the --personal flag.
func Personal(personal bool) *PersonalOption {
	return &PersonalOption{personal}
}

type PrivateOption struct {
	private bool
}

// Private represents the --private flag.
func Private(private bool) *PrivateOption {
	return &PrivateOption{private}
}

type ReadWriteKeyOption struct {
	readWriteKey bool
}

// ReadWriteKey represents the --read-write-key flag.
func ReadWriteKey(readWriteKey bool) *ReadWriteKeyOption {
	return &ReadWriteKeyOption{readWriteKey}
}

type ReconcileOption struct {
	reconcile bool
}

// Reconcile represents the --reconcile flag.
func Reconcile(reconcile bool) *ReconcileOption {
	return &ReconcileOption{reconcile}
}

type RepositoryOption struct {
	repository string
}

// Repository represents the --repository flag.
func Repository(repository string) *RepositoryOption {
	return &RepositoryOption{repository}
}

type TeamOption struct {
	team []string
}

// Team represents the --team flag.
func Team(team ...string) *TeamOption {
	return &TeamOption{team}
}

type AuthorEmailOption struct {
	authorEmail string
}

// AuthorEmail represents the --author-email flag.
func AuthorEmail(authorEmail string) *AuthorEmailOption {
	return &AuthorEmailOption{authorEmail}
}

type AuthorNameOption struct {
	authorName string
}

// AuthorName represents the --author-name flag.
func AuthorName(authorName string) *AuthorNameOption {
	return &AuthorNameOption{authorName}
}

type BranchOption struct {
	branch string
}

// Branch represents the --branch flag.
func Branch(branch string) *BranchOption {
	return &BranchOption{branch}
}

type CaFileOption struct {
	caFile string
}

// CaFile represents the --ca-file flag.
func CaFile(caFile string) *CaFileOption {
	return &CaFileOption{caFile}
}

type CommitMessageAppendixOption struct {
	commitMessageAppendix string
}

// CommitMessageAppendix represents the --commit-message-appendix flag.
func CommitMessageAppendix(commitMessageAppendix string) *CommitMessageAppendixOption {
	return &CommitMessageAppendixOption{commitMessageAppendix}
}

type GroupOption struct {
	group []string
}

// Group represents the --group flag.
// Bitbucket Server groups to be given write access (also accepts comma-separated values)
func Group(group ...string) *GroupOption {
	return &GroupOption{group}
}

type GpgKeyIdOption struct {
	gpgKeyId string
}

// GpgKeyId represents the --gpg-key-id flag.
func GpgKeyId(gpgKeyId string) *GpgKeyIdOption {
	return &GpgKeyIdOption{gpgKeyId}
}

type GpgKeyRingOption struct {
	gpgKeyRing string
}

// GpgKeyRing represents the --gpg-key-ring flag.
func GpgKeyRing(gpgKeyRing string) *GpgKeyRingOption {
	return &GpgKeyRingOption{gpgKeyRing}
}

type GpgPassphraseOption struct {
	gpgPassphrase string
}

// GpgPassphrase represents the --gpg-passphrase flag.
func GpgPassphrase(gpgPassphrase string) *GpgPassphraseOption {
	return &GpgPassphraseOption{gpgPassphrase}
}

type PrivateKeyFileOption struct {
	privateKeyFile string
}

// PrivateKeyFile represents the --private-key-file flag.
func PrivateKeyFile(privateKeyFile string) *PrivateKeyFileOption {
	return &PrivateKeyFileOption{privateKeyFile}
}

type RecurseSubmodulesOption struct {
	recurseSubmodules bool
}

// RecurseSubmodules represents the --recurse-submodules flag.
func RecurseSubmodules(recurseSubmodules bool) *RecurseSubmodulesOption {
	return &RecurseSubmodulesOption{recurseSubmodules}
}

type SecretNameOption struct {
	secretName string
}

// SecretName represents the --secret-name flag.
func SecretName(secretName string) *SecretNameOption {
	return &SecretNameOption{secretName}
}

type SshEcdsaCurveOption struct {
	sshEcdsaCurve EcdsaCurve
}

// SshEcdsaCurve represents the --ssh-ecdsa-curve flag.
func SshEcdsaCurve(ecdsaCurve EcdsaCurve) *SshEcdsaCurveOption {
	return &SshEcdsaCurveOption{ecdsaCurve}
}

type SshHostnameOption struct {
	sshHostname string
}

// SshHostname represents the --ssh-hostname flag.
func SshHostname(sshHostname string) *SshHostnameOption {
	return &SshHostnameOption{sshHostname}
}

type SshKeyAlgorithmOption struct {
	sshKeyAlgorithm KeyAlgorithm
}

// SshKeyAlgorithm represents the --ssh-key-algorithm flag.
func SshKeyAlgorithm(sshKeyAlgorithm KeyAlgorithm) *SshKeyAlgorithmOption {
	return &SshKeyAlgorithmOption{sshKeyAlgorithm}
}

type SshRsaBitsOption struct {
	sshRsaBits int
}

// SshRsaBits represents the --ssh-rsa-bits flag.
func SshRsaBits(sshRsaBits int) *SshRsaBitsOption {
	return &SshRsaBitsOption{sshRsaBits}
}

type TokenAuthOption struct {
	tokenAuth bool
}

// TokenAuth represents the --token-auth flag.
func TokenAuth(tokenAuth bool) *TokenAuthOption {
	return &TokenAuthOption{tokenAuth}
}

type UsernameOption struct {
	username string
}

// Username represents the --username flag.
func Username(username string) *UsernameOption {
	return &UsernameOption{username}
}

type PasswordOption struct {
	password string
}

// Password represents the --password flag.
func Password(password string) *PasswordOption {
	return &PasswordOption{password}
}

type AllowInsecureHTTPOption struct {
	allowInsecureHTTP bool
}

// AllowInsecureHTTP represents the --allow-insecure-http flag.
func AllowInsecureHTTP(allowInsecureHTTP bool) *AllowInsecureHTTPOption {
	return &AllowInsecureHTTPOption{allowInsecureHTTP: allowInsecureHTTP}
}

type SilentOption struct {
	silent bool
}

// Silent represents the --silent flag.
func Silent(silent bool) *SilentOption {
	return &SilentOption{silent: silent}
}

type URLOption struct {
	url string
}

// URL represents the --url flag.
func URL(url string) *URLOption {
	return &URLOption{url: url}
}
