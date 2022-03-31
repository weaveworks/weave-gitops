# Weave GitOps - simplifying application operations on Kubernetes with GitOps using Flux 

![Test status](https://github.com/weaveworks/weave-gitops/actions/workflows/pr.yaml/badge.svg)
[![LICENSE](https://img.shields.io/github/license/weaveworks/weave-gitops)](https://github.com/weaveworks/weave-gitops/blob/master/LICENSE)
[![Contributors](https://img.shields.io/github/contributors/weaveworks/weave-gitops)](https://github.com/weaveworks/weave-gitops/graphs/contributors)
[![Release](https://img.shields.io/github/v/release/weaveworks/weave-gitops?include_prereleases)](https://github.com/weaveworks/weave-gitops/releases/latest)
[![FOSSA Status](https://app.fossa.com/api/projects/custom%2B19155%2Fgithub.com%2Fweaveworks%2Fweave-gitops.svg?type=shield)](https://app.fossa.com/reports/005da7c4-1f10-4889-9432-8b97c2084e41)


Weave GitOps is a powerful extension to [Flux](https://fluxcd.io), a leading GitOps engine and CNCF project, which provides insights into your application deployments, and makes continuous delivery with GitOps easier to adopt and scale across your teams.

The web UI surfaces key information to **help application operators easily discover and resolve issues**. The intuitive interface provides a guided experience to **build understanding** and **simplify getting started for new users**; they can easily discover the relationship between Flux objects and navigate to deeper levels of information as required.

Weave GitOps is an open source project sponsored by [Weaveworks](https://weave.works) - the GitOps company, and original creators of [Flux](https://fluxcd.io).

### Why adopt GitOps?
> "GitOps is the best thing since configuration as code. Git changed how we collaborate, but declarative configuration is the key to dealing with infrastructure at scale, and sets the stage for the next generation of management tools"  
--<cite>Kelsey Hightower, Staff Developer Advocate, Google. [Source](https://twitter.com/kelseyhightower/status/1164192321891528704?s=20&t=FkRyvLThKm8Ns7yhHh7UQg).</cite>

Adopting GitOps can bring a number of key benefits:
- Faster and more frequent deployments
- Easy recovery from failures
- Improved security and auditability

To learn more about GitOps, check out these resources:
- [GitOps for absolute beginners](https://go.weave.works/WebContent-EB-GitOps-for-Beginners.html) - eBook from Weaveworks
- [Guide to GitOps](https://www.weave.works/technologies/gitops/) - from Weaveworks
- [OpenGitOps](https://opengitops.dev/) - CNCF Sandbox project aiming to define a vendor-neutral, principle-led meaning of GitOps.
- [gitops.tech](https://www.gitops.tech/) - supported by Innoq

### See Weave GitOps in action
*Video coming soon!*

## WIP - Quick start
For a full walkthrough, please check out our [Getting Started guide](https://docs.gitops.weave.works/docs/getting-started).

### Installing Weave GitOps on Kubernetes

1. First, you will need to [install Flux](https://fluxcd.io/docs/installation/).  
Both Weave GitOps and Flux work on any conformant Kubernetes distribution, for minimum supported versions see the [Flux pre-requisites](https://fluxcd.io/docs/installation/#prerequisites). We highly recommend using `flux bootstrap` to commit the Flux manifests to a Git Repository and have Flux itself be managed through GitOps, however for testing purposes you can simply issue `flux install` to install Flux without storing its manifests in a repository.

2. Configure authentication to the web UI, by either integrating an [OIDC provider](https://docs.gitops.weave.works/docs/next/gitops-dashboard#login-via-an-oidc-provider) or using the [cluster user account](https://docs.gitops.weave.works/docs/next/gitops-dashboard#login-via-a-cluster-user-account).

3. Weave GitOps is available as a Helm Chart and can be installed in the same manner as any other resource with Flux, namely a [Source](https://fluxcd.io/docs/concepts/#sources) and a [reconciliation](https://fluxcd.io/docs/concepts/#reconciliation) object, in our case a [HelmRelease](https://fluxcd.io/docs/components/helm/helmreleases/).  
Again, we recommend committing the following to a repository being reconciled by Flux, however you can also directly apply the resources to your cluster:

```
kubectl apply -n flux-system -f - << EOF
apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: HelmRepository
metadata:
  name: ww-gitops
  namespace: flux-system
spec:
  interval: 1m0s
  url: https://helm.gitops.weave.works
---
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  name: ww-gitops
  namespace: flux-system
spec:
  chart:
    spec:
      chart: weave-gitops
      sourceRef:
        kind: HelmRepository
        name: ww-gitops
  interval: 1m0s
  values:
    additionalArgs:
    - --insecure
```

## Installing the GitOps CLI

The `gitops` CLI provides a set of commands to make it easier to interact with Weave GitOps, including both the free open source project, and commercial [Weave GitOps Enterprise](https://www.weave.works/product/gitops-enterprise/) product.

Mac / Linux

```console
curl --silent --location "https://github.com/weaveworks/weave-gitops/releases/download/v0.6.2/gitops-$(uname)-$(uname -m).tar.gz" | tar xz -C /tmp
sudo mv /tmp/gitops /usr/local/bin
gitops version
```

Homebrew:

```console
brew tap weaveworks/tap
brew install weaveworks/tap/gitops
```

## CLI Reference

```console
Weave GitOps
Command line utility for managing Kubernetes applications via GitOps.

Usage:
  gitops [command]

Available Commands:
  add         Add a new Weave GitOps resource
  check       Validates flux compatibility
  completion  Generate the autocompletion script for the specified shell
  delete      Delete one or many Weave GitOps resources
  get         Display one or many Weave GitOps resources
  help        Help about any command
  update      Update a Weave GitOps resource
  upgrade     Upgrade to Weave GitOps Enterprise
  version     Display gitops version

Flags:
  -h, --help               Help for gitops
      --namespace string   The namespace scope for this operation (default "flux-system").
  -v, --verbose            Enable verbose output

Use "gitops [command] --help" for more information about a command.
```

For more information, please see the [docs](https://docs.gitops.weave.works/docs/cli-reference).

## Contribution

Need help or want to contribute? Please see the links below.

- Need help?
  - Talk to us in the [#weave-gitops channel](https://app.slack.com/client/T2NDH1D9D/C0248LVC719/thread/C2ND76PAA-1621532937.019800) on Weaveworks Community Slack. [Invite yourself](https://slack.weave.works/) if you haven't joined yet.
- Have feature proposals or want to contribute?
  - Please create an [issue](https://github.com/weaveworks/weave-gitops/issues).
  - Check out our [Contributing guide](CONTRIBUTING.md).

## Commercial support

Weaveworks provides [Weave GitOps Enterprise](https://www.weave.works/product/gitops-enterprise/), a continuous operations product that makes it easy to deploy and manage Kubernetes clusters and applications at scale in any environment. The single management console automates trusted application delivery and secure infrastructure operations on premise, in the cloud and at the edge.

To discuss your support needs, please contact us at [sales@weave.works](mailto:sales@weave.works).

## License scan details

[![FOSSA Status](https://app.fossa.com/api/projects/custom%2B19155%2Fgithub.com%2Fweaveworks%2Fweave-gitops.svg?type=large)](https://app.fossa.com/reports/005da7c4-1f10-4889-9432-8b97c2084e41)

