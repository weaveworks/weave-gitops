package fluxexec

type AsOption struct {
	as string
}

// As represents the --as flag.
func As(as string) *AsOption {
	return &AsOption{as}
}

type AsGroupOption struct {
	asGroup []string
}

// AsGroup represents the --as-group flag.
func AsGroup(asGroup ...string) *AsGroupOption {
	return &AsGroupOption{asGroup}
}

type AsUIDOption struct {
	asUID string
}

// AsUID represents the --as-uid flag.
func AsUID(asUID string) *AsUIDOption {
	return &AsUIDOption{asUID}
}

type CacheDirOption struct {
	cacheDir string
}

// CacheDir represents the --cache-dir flag.
func CacheDir(cacheDir string) *CacheDirOption {
	return &CacheDirOption{cacheDir}
}

type CertificateAuthorityOption struct {
	certificateAuthority string
}

// CertificateAuthority represents the --certificate-authority flag.
func CertificateAuthority(certificateAuthority string) *CertificateAuthorityOption {
	return &CertificateAuthorityOption{certificateAuthority}
}

type ClientCertificateOption struct {
	clientCertificate string
}

// ClientCertificate represents the --client-certificate flag.
func ClientCertificate(clientCertificate string) *ClientCertificateOption {
	return &ClientCertificateOption{clientCertificate}
}

type ClientKeyOption struct {
	clientKey string
}

// ClientKey represents the --client-key flag.
func ClientKey(clientKey string) *ClientKeyOption {
	return &ClientKeyOption{clientKey}
}

type ClusterOption struct {
	cluster string
}

// Cluster represents the --cluster flag.
func Cluster(cluster string) *ClusterOption {
	return &ClusterOption{cluster}
}

type ContextOption struct {
	context string
}

// KubeContext represents the --context flag.
func KubeContext(context string) *ContextOption {
	return &ContextOption{context}
}

type InsecureSkipTLSVerifyOption struct {
	insecureSkipTLSVerify bool
}

// InsecureSkipTLSVerify represents the --insecure-skip-tls-verify flag.
func InsecureSkipTLSVerify(insecureSkipTLSVerify bool) *InsecureSkipTLSVerifyOption {
	return &InsecureSkipTLSVerifyOption{insecureSkipTLSVerify}
}

type KubeAPIBurstOption struct {
	kubeAPIBurst int
}

// KubeAPIBurst represents the --kube-api-burst flag.
func KubeAPIBurst(kubeAPIBurst int) *KubeAPIBurstOption {
	return &KubeAPIBurstOption{kubeAPIBurst}
}

type KubeAPIQPSOption struct {
	kubeAPIQPS float32
}

// KubeAPIQPS represents the --kube-api-qps flag.
func KubeAPIQPS(kubeAPIQPS float32) *KubeAPIQPSOption {
	return &KubeAPIQPSOption{kubeAPIQPS}
}

type KubeconfigOption struct {
	kubeconfig string
}

// Kubeconfig represents the --kubeconfig flag.
func Kubeconfig(kubeconfig string) *KubeconfigOption {
	return &KubeconfigOption{kubeconfig}
}

type NamespaceOption struct {
	namespace string
}

// Namespace represents the --namespace flag.
func Namespace(namespace string) *NamespaceOption {
	return &NamespaceOption{namespace}
}

type ServerOption struct {
	server string
}

// Server represents the --server flag.
func Server(server string) *ServerOption {
	return &ServerOption{server}
}

type TimeoutOption struct {
	timeout string
}

// Timeout represents the --timeout flag.
func Timeout(timeout string) *TimeoutOption {
	return &TimeoutOption{timeout}
}

type TLSServerNameOption struct {
	tlsServerName string
}

// TLSServerName represents the --tls-server-name flag.
func TLSServerName(tlsServerName string) *TLSServerNameOption {
	return &TLSServerNameOption{tlsServerName}
}

type TokenOption struct {
	token string
}

// Token represents the --token flag.
func Token(token string) *TokenOption {
	return &TokenOption{token}
}

type UserOption struct {
	user string
}

// User represents the --user flag.
func User(user string) *UserOption {
	return &UserOption{user}
}

type VerboseOption struct {
	verbose bool
}

// Verbose represents the --verbose flag.
func Verbose(verbose bool) *VerboseOption {
	return &VerboseOption{verbose}
}
