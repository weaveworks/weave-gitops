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
Weave GitOps currently supports SaaS versions of GitHub and GitLab (CLI only).

## Installing the Weave GitOps CLI

To install the `Gitops` CLI, please follow the following steps:

```console
curl -L "https://github.com/weaveworks/weave-gitops/releases/download/v0.5.0/gitops-$(uname)-$(uname -m).tar.gz" | tar -xz -C /tmp
chmod +x /tmp/gitops
sudo mv /tmp/gitops /usr/local/bin/gitops
gitops version
```

Alternatively, macOS users can use Homebrew:

```console
brew tap weaveworks/tap
brew install weaveworks/tap/gitops
gitops version
```

You should see:

```console
Current Version: v0.5.0-rc0.4.1-13-gb563109
GitCommit: b563109
BuildTime: 2021-12-02_19:21:32
Branch: HEAD
Flux Version: v0.21.0
```
