package fluxexec

import (
	"context"
	"os/exec"
	"strings"
)

type bootstrapGitLabConfig struct {
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

var defaultBootstrapGitLabOptions = bootstrapGitLabConfig{
	hostname: "gitlab.com",
	interval: "1m0s",
	private:  true,
}

// BootstrapGitLabOption represents options used in the BootstrapGitLab method.
type BootstrapGitLabOption interface {
	configureBootstrapGitLab(*bootstrapGitLabConfig)
}

func (opt *HostnameOption) configureBootstrapGitLab(conf *bootstrapGitLabConfig) {
	conf.hostname = opt.hostname
}

func (opt *IntervalOption) configureBootstrapGitLab(conf *bootstrapGitLabConfig) {
	conf.interval = opt.interval
}

func (opt *OwnerOption) configureBootstrapGitLab(conf *bootstrapGitLabConfig) {
	conf.owner = opt.owner
}

func (opt *PathOption) configureBootstrapGitLab(conf *bootstrapGitLabConfig) {
	conf.path = opt.path
}

func (opt *PersonalOption) configureBootstrapGitLab(conf *bootstrapGitLabConfig) {
	conf.personal = opt.personal
}

func (opt *PrivateOption) configureBootstrapGitLab(conf *bootstrapGitLabConfig) {
	conf.private = opt.private
}

func (opt *ReadWriteKeyOption) configureBootstrapGitLab(conf *bootstrapGitLabConfig) {
	conf.readWriteKey = opt.readWriteKey
}

func (opt *ReconcileOption) configureBootstrapGitLab(conf *bootstrapGitLabConfig) {
	conf.reconcile = opt.reconcile
}

func (opt *RepositoryOption) configureBootstrapGitLab(conf *bootstrapGitLabConfig) {
	conf.repository = opt.repository
}

func (opt *TeamOption) configureBootstrapGitLab(conf *bootstrapGitLabConfig) {
	conf.team = opt.team
}

func (flux *Flux) bootstrapGitLabCmd(ctx context.Context, opts ...BootstrapGitLabOption) (*exec.Cmd, error) {
	c := defaultBootstrapGitLabOptions
	for _, opt := range opts {
		opt.configureBootstrapGitLab(&c)
	}

	args := []string{"bootstrap", "gitlab"}

	// Add the global args first.
	globalArgs := flux.globalArgs(c.globalOptions...)
	args = append(args, globalArgs...)

	// The add the bootstrap args.
	bootstrapArgs := flux.bootstrapArgs(c.bootstrapOptions...)
	args = append(args, bootstrapArgs...)

	if c.hostname != "" {
		args = append(args, "--hostname", c.hostname)
	}

	if c.interval != "" {
		args = append(args, "--interval", c.interval)
	}

	if c.owner != "" {
		args = append(args, "--owner", c.owner)
	}

	if c.path != "" {
		args = append(args, "--path", c.path)
	}

	if c.personal {
		args = append(args, "--personal")
	}

	if c.private {
		args = append(args, "--private")
	}

	if c.readWriteKey {
		args = append(args, "--read-write-key")
	}

	if c.reconcile {
		args = append(args, "--reconcile")
	}

	if c.repository != "" {
		args = append(args, "--repository", c.repository)
	}

	if len(c.team) > 0 {
		args = append(args, "--team", strings.Join(c.team, ","))
	}

	return flux.buildFluxCmd(ctx, nil, args...), nil
}

func (flux *Flux) BootstrapGitlab(ctx context.Context, opts ...BootstrapGitLabOption) error {
	bootstrapGitLabCmd, err := flux.bootstrapGitLabCmd(ctx, opts...)
	if err != nil {
		return err
	}

	if err := flux.runFluxCmd(ctx, bootstrapGitLabCmd); err != nil {
		return err
	}

	return nil
}
