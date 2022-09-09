package fluxexec

import (
	"context"
	"os/exec"
	"strings"
)

type bootstrapBitbucketServerConfig struct {
	globalOptions    []GlobalOption
	bootstrapOptions []BootstrapOption

	group        []string
	hostname     string
	interval     string
	owner        string
	path         string
	personal     bool
	private      bool
	readWriteKey bool
	reconcile    bool
	repository   string
	username     string
}

var defaultBootstrapBitbucketServerOptions = bootstrapBitbucketServerConfig{
	interval: "1m0s",
	private:  true,
	username: "git",
}

// BootstrapBitbucketServerOption represents options used in the BootstrapBitbucketServer method.
type BootstrapBitbucketServerOption interface {
	configureBootstrapBitbucketServer(*bootstrapBitbucketServerConfig)
}

func (opt *GroupOption) configureBootstrapBitbucketServer(conf *bootstrapBitbucketServerConfig) {
	conf.group = opt.group
}

func (opt *HostnameOption) configureBootstrapBitbucketServer(conf *bootstrapBitbucketServerConfig) {
	conf.hostname = opt.hostname
}

func (opt *IntervalOption) configureBootstrapBitbucketServer(conf *bootstrapBitbucketServerConfig) {
	conf.interval = opt.interval
}

func (opt *OwnerOption) configureBootstrapBitbucketServer(conf *bootstrapBitbucketServerConfig) {
	conf.owner = opt.owner
}

func (opt *PathOption) configureBootstrapBitbucketServer(conf *bootstrapBitbucketServerConfig) {
	conf.path = opt.path
}

func (opt *PersonalOption) configureBootstrapBitbucketServer(conf *bootstrapBitbucketServerConfig) {
	conf.personal = opt.personal
}

func (opt *PrivateOption) configureBootstrapBitbucketServer(conf *bootstrapBitbucketServerConfig) {
	conf.private = opt.private
}

func (opt *ReadWriteKeyOption) configureBootstrapBitbucketServer(conf *bootstrapBitbucketServerConfig) {
	conf.readWriteKey = opt.readWriteKey
}

func (opt *ReconcileOption) configureBootstrapBitbucketServer(conf *bootstrapBitbucketServerConfig) {
	conf.reconcile = opt.reconcile
}

func (opt *RepositoryOption) configureBootstrapBitbucketServer(conf *bootstrapBitbucketServerConfig) {
	conf.repository = opt.repository
}

func (opt *UsernameOption) configureBootstrapBitbucketServer(conf *bootstrapBitbucketServerConfig) {
	conf.username = opt.username
}

func (flux *Flux) BootstrapBitbucketServer(ctx context.Context, opts ...BootstrapBitbucketServerOption) error {
	bootstrapBitbucketServerCmd, err := flux.bootstrapBitbucketServerCmd(ctx, opts...)
	if err != nil {
		return err
	}

	if err := flux.runFluxCmd(ctx, bootstrapBitbucketServerCmd); err != nil {
		return err
	}

	return nil
}

func (flux *Flux) bootstrapBitbucketServerCmd(ctx context.Context, opts ...BootstrapBitbucketServerOption) (*exec.Cmd, error) {
	c := defaultBootstrapBitbucketServerOptions
	for _, opt := range opts {
		opt.configureBootstrapBitbucketServer(&c)
	}

	args := []string{"bootstrap", "bitbucket-server"}

	// Add the global args first.
	globalArgs := flux.globalArgs(c.globalOptions...)
	args = append(args, globalArgs...)

	// The add the bootstrap args.
	bootstrapArgs := flux.bootstrapArgs(c.bootstrapOptions...)
	args = append(args, bootstrapArgs...)

	if len(c.group) > 0 {
		args = append(args, "--group", strings.Join(c.group, ","))
	}

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

	if c.username != "" {
		args = append(args, "--username", c.username)
	}

	return flux.buildFluxCmd(ctx, nil, args...), nil
}
