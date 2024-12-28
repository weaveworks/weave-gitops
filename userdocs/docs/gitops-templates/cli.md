---
title: CLI

---



# Template CLI ~ENTERPRISE~

The Enterprise `gitops` CLI tool provides a set of commands to help you manage your templates.

Here we're going to talk about the `gitops create template` command that allows
you to render templates locally and airgapped, without a full WGE installation
in a Kubernetes cluster.

## Use cases

- In CI/CD systems where you want to render a template and then use the raw output in a pipeline
- For quickly debugging templates

## Restrictions

The `gitops create template` command only works with `GitOpsTemplate` objects.
It does not work with `CAPITemplate` objects. You should be able to migrate any
`CAPITemplate` objects to `GitOpsTemplate` with some small tweaks.

!!! info
    GitOpsTemplate or CAPITemplate?

    The only difference between `CAPITemplate` and `GitOpsTemplate` is the default
    value of these two annotations:

    | Annotation | default value for `CAPITemplate` | default value for `GitOpsTemplate` |
    | ----------- | ---------------- | ------------------ |
    | `templates.weave.works/add-common-bases`  | `"true"` | `"false"` |
    | `templates.weave.works/inject-prune-annotations` | `"true"` | `"false"` |

## Installation

See the Weave Gitops Enterprise [installation instructions](../enterprise/install-enterprise.md#7-install-the-cli) for details on how to install the EE `gitops` CLI tool.

## Getting started

Using a local `GitOpsTemplate` manifest with required parameters exported in the
environment, the command can render the template to one of the following:
1. The current kubecontext directly (default)
1. stdout with `--export`
1. The local file system with `--output-dir`, this will use the
	`spec.resourcestemplates[].path` fields in the template to determine where to
	write the rendered files.
	This is the recommended approach for GitOps as you can then commit the
	rendered files to your repository.

```bash
gitops create template \
  --template-file capd-template.yaml \
  --output-dir ./clusters/ \
  --values CLUSTER_NAME=foo
```

## Profiles

As in the UI you can add profiles to your template. However instead of reading
the latest version of a profile and its layers from a `HelmRepository` object
in the cluster, we instead read from your local helm cache.

```bash
helm repo add weaveworks-charts https://raw.githubusercontent.com/weaveworks/weave-gitops-profile-examples/gh-pages
helm repo update
```

This particular helm repo provides a version of the `cert-manager` repo and others.

### Supplying values to a profile

You can supply a `values.yaml` file to a profile using the `values` parameter.
For example we can supply `cert-manager`'s `values.yaml` with:

```bash
gitops create template \
  --template-file capd-template.yaml \
  --output-dir ./out \
  --values CLUSTER_NAME=foo \
  --profiles "name=cert-manager,namespace=foo,version=>0.1,values=cert-manager-values.yaml"
```

## Using a config file

Instead of specifying the parameters on the command line you can supply a
config file. For example the above invocation can be replaced like so:

```yaml title=config.yaml
template-file: capd-capi-template.yaml
output-dir: ./out
values:
  - CLUSTER_NAME=foo
profiles:
  - name=cert-manager,namespace=foo,version=>0.1,values=cert-manager-values.yaml
```

and executed with:

```bash
gitops create template --config config.yaml
```
