package config

type Options struct {
	Endpoint          string
	OverrideInCluster bool
	GitHostTypes      map[string]string
	Username          string
	Password          string
}
