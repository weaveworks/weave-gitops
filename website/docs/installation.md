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
curl --silent --location "https://github.com/weaveworks/weave-gitops/releases/download/v0.6.2/gitops-$(uname)-$(uname -m).tar.gz" | tar xz -C /tmp
sudo mv /tmp/gitops /usr/local/bin
gitops version
```

You should see:

```console
Current Version: v0.6.2
GitCommit: cf7b11b8
BuildTime: 2022-01-20_20:15:15
Branch: HEAD
Flux Version: v0.24.1
```
