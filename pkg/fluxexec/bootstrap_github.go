package fluxexec

import (
	"context"
	"fmt"
	"os/exec"
	"reflect"
	"strings"
)

type bootstrapGitHubConfig struct {
	globalOptions    []GlobalOption
	bootstrapOptions []BootstrapOption

	hostname     string
	interval     string
	owner        string
	path         string
	personal     bool
	private      bool
	readWriteKey bool
	reconcile    bool
	repository   string
	team         []string
}

var defaultBootstrapGitHubOptions = bootstrapGitHubConfig{
	hostname: "github.com",
	interval: "1m0s",
	private:  true,
}

// BootstrapGitHubOption represents options used in the BootstrapGitHub method.
type BootstrapGitHubOption interface {
	configureBootstrapGitHub(*bootstrapGitHubConfig)
}

func (opt *HostnameOption) configureBootstrapGitHub(conf *bootstrapGitHubConfig) {
	conf.hostname = opt.hostname
}

func (opt *IntervalOption) configureBootstrapGitHub(conf *bootstrapGitHubConfig) {
	conf.interval = opt.interval
}

func (opt *OwnerOption) configureBootstrapGitHub(conf *bootstrapGitHubConfig) {
	conf.owner = opt.owner
}

func (opt *PathOption) configureBootstrapGitHub(conf *bootstrapGitHubConfig) {
	conf.path = opt.path
}

func (opt *PersonalOption) configureBootstrapGitHub(conf *bootstrapGitHubConfig) {
	conf.personal = opt.personal
}

func (opt *PrivateOption) configureBootstrapGitHub(conf *bootstrapGitHubConfig) {
	conf.private = opt.private
}

func (opt *ReadWriteKeyOption) configureBootstrapGitHub(conf *bootstrapGitHubConfig) {
	conf.readWriteKey = opt.readWriteKey
}

func (opt *ReconcileOption) configureBootstrapGitHub(conf *bootstrapGitHubConfig) {
	conf.reconcile = opt.reconcile
}

func (opt *RepositoryOption) configureBootstrapGitHub(conf *bootstrapGitHubConfig) {
	conf.repository = opt.repository
}

func (opt *TeamOption) configureBootstrapGitHub(conf *bootstrapGitHubConfig) {
	conf.team = append(conf.team, opt.team...)
}

func (flux *Flux) BootstrapGitHub(ctx context.Context, opts ...BootstrapGitHubOption) error {
	bootstrapGitHubCmd, err := flux.bootstrapGitHubCmd(ctx, opts...)
	if err != nil {
		return err
	}

	if err := flux.runFluxCmd(ctx, bootstrapGitHubCmd); err != nil {
		return err
	}

	return nil
}

func (flux *Flux) bootstrapGitHubCmd(ctx context.Context, opts ...BootstrapGitHubOption) (*exec.Cmd, error) {
	c := defaultBootstrapGitHubOptions
	for _, o := range opts {
		o.configureBootstrapGitHub(&c)
	}

	args := []string{"bootstrap", "github"}

	// Add the global args first.
	globalArgs := flux.globalArgs(c.globalOptions...)
	args = append(args, globalArgs...)

	// The add the bootstrap args.
	bootstrapArgs := flux.bootstrapArgs(c.bootstrapOptions...)
	args = append(args, bootstrapArgs...)

	// Then follow with the bootstrap github args.
	if c.hostname != "" && !reflect.DeepEqual(c.hostname, defaultBootstrapGitHubOptions.hostname) {
		args = append(args, "--hostname", c.hostname)
	}

	if c.interval != "" && !reflect.DeepEqual(c.interval, defaultBootstrapGitHubOptions.interval) {
		args = append(args, "--interval", c.interval)
	}

	if c.owner != "" && !reflect.DeepEqual(c.owner, defaultBootstrapGitHubOptions.owner) {
		args = append(args, "--owner", c.owner)
	}

	if c.path != "" && !reflect.DeepEqual(c.path, defaultBootstrapGitHubOptions.path) {
		args = append(args, "--path", c.path)
	}

	if c.personal && !reflect.DeepEqual(c.personal, defaultBootstrapGitHubOptions.personal) {
		args = append(args, "--personal")
	}

	if !reflect.DeepEqual(c.private, defaultBootstrapGitHubOptions.private) {
		args = append(args, "--private", fmt.Sprintf("%v", c.private))
	}

	if c.readWriteKey && !reflect.DeepEqual(c.readWriteKey, defaultBootstrapGitHubOptions.readWriteKey) {
		args = append(args, "--read-write-key")
	}

	if c.reconcile && !reflect.DeepEqual(c.reconcile, defaultBootstrapGitHubOptions.reconcile) {
		args = append(args, "--reconcile")
	}

	if c.repository != "" && !reflect.DeepEqual(c.repository, defaultBootstrapGitHubOptions.repository) {
		args = append(args, "--repository", c.repository)
	}

	if len(c.team) > 0 && !reflect.DeepEqual(c.team, defaultBootstrapGitHubOptions.team) {
		args = append(args, "--team", strings.Join(c.team, ","))
	}

	return flux.buildFluxCmd(ctx, flux.env, args...), nil
}
