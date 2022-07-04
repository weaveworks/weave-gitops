package config

type Options struct {
	Endpoint              string
	OverrideInCluster     bool
	GitHostTypes          map[string]string
	InsecureSkipTLSVerify bool
	Username              string
	Password              string
	KubeconfigLocation    string
}
