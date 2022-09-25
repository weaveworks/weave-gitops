# Flux Exec (fluxexec)

## Introduction

This package provides a wrapper of the command line interface to execute Flux commands.

## How to add new flags

Add the new type `Option` struct to the `options.go` of `fluxexec` package. 
Then create a function to represent it.

The below example shows how to add the `--export` flag.
We have `ExportOption` struct to whole the value of the `--export` flag, 
and we have `Export` function to represent this flag during the command creation.

```go
type ExportOption struct {
	export bool
}

// Export represents the --export flag.
func Export(export bool) *ExportOption {
	return &ExportOption{export}
}
```

After that we go to each command wrapper file, for example we have `install.go` file as the wrapper of `flux install` command,
and we have `bootstrap_github.go` file as the wrapper of the `flux bootstrap github` command.

Following the above example, we add the `--export` flag to the wrapper of the `install` command.

First, find the `installConfig` struct in the `install.go` file.
Then, add the `export` field to the struct.

```go
type installConfig struct {
	clusterDomain      string
	components         []string
	componentsExtra    []string
	export             bool     // Here's our new export field
	imagePullSecret    string
	logLevel           string
	networkPolicy      bool
	registry           string
	tolerationKeys     []string
	watchAllNamespaces bool
}
```

Then, we add the `configureInstall` function for the `ExportOption` struct.
This function will be called to configure the `installConfig` struct, by setting the value of the `export` field.

```go
func (opt *ExportOption) configureInstall(conf *installConfig) {
	conf.export = opt.export
}
```

Next, locate the `installCmd` function.

```go
func (flux *Flux) installCmd(ctx context.Context, opts ...InstallOption) (*exec.Cmd, error) {
```

This function will be called to create the command, with all the flags as args.
We add the following if statement to help this command generate the `--export` flag.

```go
	if c.export && !reflect.DeepEqual(c.export, defaultInstallOptions.export) {
		args = append(args, "--export")
	}
```

### Revisiting the package for a new Flux version

When a new version of Flux is released, the `fluxexec` package needs to be revisited to ensure that all the new flags are added.
To do so, we need to check the `flux install` and `flux bootstrap` commands to see if there are any new flags added.

For example, running the following command to check if there's a new default value: `flux install --help | grep "(default"`

```bash
$ flux install --help | grep "(default"
      --cluster-domain string      internal cluster domain (default "cluster.local")
      --components strings         list of components, accepts comma-separated values (default [source-controller,kustomize-controller,helm-controller,notification-controller])
      --log-level logLevel         log level, available options are: (debug, info, error) (default info)
      --network-policy             deny ingress access to the toolkit controllers from other namespaces using network policies (default true)
      --registry string            container registry where the toolkit images are published (default "ghcr.io/fluxcd")
      --watch-all-namespaces       watch for custom resources in all namespaces, if set to false it will only watch the namespace where the toolkit is installed (default true)
```

Here's the new flags from the global scope in the `flux install` command.
```bash
      --cache-dir string               Default cache directory (default "/home/user/.kube/cache")
      --kube-api-burst int             The maximum burst queries-per-second of requests sent to the Kubernetes API. (default 100)
      --kube-api-qps float32           The maximum queries-per-second of requests sent to the Kubernetes API. (default 50)
  -n, --namespace string               If present, the namespace scope for this CLI request (default "flux-system")
      --timeout duration               timeout for this operation (default 5m0s)
```

### Future Improvements

The current implementation of the `fluxexec` package is not ideal.
It is not easy to add new flags, and it is not easy to maintain.
We need to find a better way to implement this package.
For example, we would write a script to generate the code of this package for us, instead of writing the code manually.
It would be done by parsing the `flux install` and `flux bootstrap` commands, and generate the code based on the flags.
