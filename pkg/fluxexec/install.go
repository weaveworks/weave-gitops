package fluxexec

import (
	"context"
	"fmt"
	"os/exec"
	"reflect"
	"strings"
)

type installConfig struct {
	globalOptions []GlobalOption

	clusterDomain      string
	components         []Component
	componentsExtra    []ComponentExtra
	export             bool
	imagePullSecret    string
	logLevel           string
	networkPolicy      bool
	registry           string
	tolerationKeys     []string
	watchAllNamespaces bool
}

var defaultInstallOptions = installConfig{
	clusterDomain:      "cluster.local",
	components:         []Component{ComponentSourceController, ComponentKustomizeController, ComponentHelmController, ComponentNotificationController},
	componentsExtra:    []ComponentExtra{},
	export:             false,
	logLevel:           "info",
	networkPolicy:      true,
	registry:           "ghcr.io/fluxcd",
	tolerationKeys:     []string{},
	watchAllNamespaces: true,
}

// InstallOption represents options used in the Install method.
type InstallOption interface {
	configureInstall(*installConfig)
}

func (opt *ClusterDomainOption) configureInstall(conf *installConfig) {
	conf.clusterDomain = opt.clusterDomain
}

func (opt *ComponentsOption) configureInstall(conf *installConfig) {
	conf.components = opt.components
}

func (opt *ComponentsExtraOption) configureInstall(conf *installConfig) {
	conf.componentsExtra = opt.componentsExtra
}

func (opt *ExportOption) configureInstall(conf *installConfig) {
	conf.export = opt.export
}

func (opt *ImagePullSecretOption) configureInstall(conf *installConfig) {
	conf.imagePullSecret = opt.imagePullSecret
}

func (opt *LogLevelOption) configureInstall(conf *installConfig) {
	conf.logLevel = opt.logLevel
}

func (opt *NetworkPolicyOption) configureInstall(conf *installConfig) {
	conf.networkPolicy = opt.networkPolicy
}

func (opt *RegistryOption) configureInstall(conf *installConfig) {
	conf.registry = opt.registry
}

func (opt *TolerationKeysOption) configureInstall(conf *installConfig) {
	conf.tolerationKeys = opt.tolerationKeys
}

func (opt *WatchAllNamespacesOption) configureInstall(conf *installConfig) {
	conf.watchAllNamespaces = opt.watchAllNamespaces
}

func (flux *Flux) Install(ctx context.Context, opts ...InstallOption) error {
	installCmd := flux.installCmd(ctx, opts...)

	if err := flux.runFluxCmd(ctx, installCmd); err != nil {
		return err
	}

	return nil
}

func (flux *Flux) installCmd(ctx context.Context, opts ...InstallOption) *exec.Cmd {
	c := defaultInstallOptions
	for _, o := range opts {
		o.configureInstall(&c)
	}

	args := []string{"install"}

	// Add the global args first.
	globalArgs := flux.globalArgs(c.globalOptions...)
	args = append(args, globalArgs...)

	if !c.watchAllNamespaces && !reflect.DeepEqual(c.watchAllNamespaces, defaultInstallOptions.watchAllNamespaces) {
		args = append(args, "--watch-all-namespaces", fmt.Sprintf("%v", c.watchAllNamespaces))
	}

	if c.clusterDomain != "" && !reflect.DeepEqual(c.clusterDomain, defaultInstallOptions.clusterDomain) {
		args = append(args, "--cluster-domain", c.clusterDomain)
	}

	if c.export && !reflect.DeepEqual(c.export, defaultInstallOptions.export) {
		args = append(args, "--export")
	}

	if c.imagePullSecret != "" && !reflect.DeepEqual(c.imagePullSecret, defaultInstallOptions.imagePullSecret) {
		args = append(args, "--image-pull-secret", c.imagePullSecret)
	}

	if c.logLevel != "" && !reflect.DeepEqual(c.logLevel, defaultInstallOptions.logLevel) {
		args = append(args, "--log-level", c.logLevel)
	}

	if !reflect.DeepEqual(c.networkPolicy, defaultInstallOptions.networkPolicy) {
		args = append(args, "--network-policy", fmt.Sprintf("%v", c.networkPolicy))
	}

	if c.registry != "" && !reflect.DeepEqual(c.registry, defaultInstallOptions.registry) {
		args = append(args, "--registry", c.registry)
	}

	if len(c.tolerationKeys) > 0 && !reflect.DeepEqual(c.tolerationKeys, defaultInstallOptions.tolerationKeys) {
		args = append(args, "--toleration-keys", strings.Join(c.tolerationKeys, ","))
	}

	if len(c.components) > 0 && !reflect.DeepEqual(c.components, defaultInstallOptions.components) {
		var comps []string
		for _, c := range c.components {
			comps = append(comps, string(c))
		}

		args = append(args, "--components", strings.Join(comps, ","))
	}

	if len(c.componentsExtra) > 0 && !reflect.DeepEqual(c.componentsExtra, defaultInstallOptions.componentsExtra) {
		var extras []string
		for _, e := range c.componentsExtra {
			extras = append(extras, string(e))
		}

		args = append(args, "--components-extra", strings.Join(extras, ","))
	}

	return flux.buildFluxCmd(ctx, nil, args...)
}
