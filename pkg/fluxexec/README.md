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
	if c.export {
		args = append(args, "--export")
	}
```
