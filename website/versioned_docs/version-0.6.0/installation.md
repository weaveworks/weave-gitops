---
sidebar_position: 1
---
# Installation

## Pre-requisites

### Kubernetes Cluster
Weave GitOps is compatible with conformant Kubernetes distributions which match the minimum required version level of [Flux](https://fluxcd.io/docs/installation/#prerequisites).

### CLI
The `gitops` command-line interface is currently supported on Mac (x86 and Arm), and Linux including WSL.

Windows support is a [planned enhancement](https://github.com/weaveworks/weave-gitops/issues/663).

### Git Providers
Weave GitOps currently supports SaaS versions of GitHub and GitLab.

## Installing the Weave GitOps CLI

To install the `Gitops` CLI, please follow the following steps:

```console
curl --silent --location "https://github.com/weaveworks/weave-gitops/releases/download/v0.6.0/gitops-$(uname)-$(uname -m).tar.gz" | tar xz -C /tmp
sudo mv /tmp/gitops /usr/local/bin
gitops version
```

Alternatively, macOS users can use Homebrew to install the latest version:

```console
brew tap weaveworks/tap
brew install weaveworks/tap/gitops
gitops version
```

You should see:

```console
Current Version: 0.6.0
GitCommit: 15349f1b392f08804d30419372b695d2b9a8318b
BuildTime: 2021-12-16T19:13:11Z
Branch: HEAD
Flux Version: v0.21.0
```
