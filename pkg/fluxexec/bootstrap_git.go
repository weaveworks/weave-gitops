package fluxexec

import (
	"context"
	"os/exec"
)

type bootstrapGitConfig struct {
	globalOptions    []GlobalOption
	bootstrapOptions []BootstrapOption

	allowInsecureHTTP bool
	interval          string
	password          string
	path              string
	silent            bool
	url               string
	username          string
}

var defaultBootstrapGitOptions = bootstrapGitConfig{
	interval: "1m0s",
	username: "git",
}

// BootstrapGitOption represents options used in the BootstrapGit method.
type BootstrapGitOption interface {
	configureBootstrapGit(*bootstrapGitConfig)
}

func (opt *AllowInsecureHTTPOption) configureBootstrapGit(conf *bootstrapGitConfig) {
	conf.allowInsecureHTTP = opt.allowInsecureHTTP
}

func (opt *IntervalOption) configureBootstrapGit(conf *bootstrapGitConfig) {
	conf.interval = opt.interval
}

func (opt *PasswordOption) configureBootstrapGit(conf *bootstrapGitConfig) {
	conf.password = opt.password
}

func (opt *PathOption) configureBootstrapGit(conf *bootstrapGitConfig) {
	conf.path = opt.path
}

func (opt *SilentOption) configureBootstrapGit(conf *bootstrapGitConfig) {
	conf.silent = opt.silent
}

func (opt *URLOption) configureBootstrapGit(conf *bootstrapGitConfig) {
	conf.url = opt.url
}

func (opt *UsernameOption) configureBootstrapGit(conf *bootstrapGitConfig) {
	conf.username = opt.username
}

func (flux *Flux) BootstrapGit(ctx context.Context, opts ...BootstrapGitOption) error {
	bootstrapGitCmd, err := flux.bootstrapGitCmd(ctx, opts...)
	if err != nil {
		return err
	}

	if err := flux.runFluxCmd(ctx, bootstrapGitCmd); err != nil {
		return err
	}

	return nil
}

func (flux *Flux) bootstrapGitCmd(ctx context.Context, opts ...BootstrapGitOption) (*exec.Cmd, error) {
	c := defaultBootstrapGitOptions
	for _, o := range opts {
		o.configureBootstrapGit(&c)
	}

	args := []string{"bootstrap", "git"}

	// Add the global args first.
	globalArgs := flux.globalArgs(c.globalOptions...)
	args = append(args, globalArgs...)

	// The add the bootstrap args.
	bootstrapArgs := flux.bootstrapArgs(c.bootstrapOptions...)
	args = append(args, bootstrapArgs...)

	if c.allowInsecureHTTP {
		args = append(args, "--allow-insecure-http")
	}

	if c.interval != "" {
		args = append(args, "--interval", c.interval)
	}

	if c.password != "" {
		args = append(args, "--password", c.password)
	}

	if c.path != "" {
		args = append(args, "--path", c.path)
	}

	if c.silent {
		args = append(args, "--silent")
	}

	if c.url != "" {
		args = append(args, "--url", c.url)
	}

	if c.username != "" {
		args = append(args, "--username", c.username)
	}

	return flux.buildFluxCmd(ctx, nil, args...), nil
}
