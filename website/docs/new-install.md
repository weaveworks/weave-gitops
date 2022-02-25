---
title: Install Weave GitOps via Helm Chart sidebar_position: 1 hide_title: true
---

import TierLabel from "./_components/TierLabel";

<h1>
  {frontMatter.title} <TierLabel tiers="All tiers" />
</h1>

### Weave GitOps Application

The Weave GitOps Application can be installed via a Flux `GitRepository` and `HelmRelease` objects. In this document we
are going to install the application via GitOps.

## Pre-requisites

### Kubernetes Cluster

Weave GitOps is compatible with conformant Kubernetes distributions which match the minimum required version level
of [Flux](https://fluxcd.io/docs/installation/#prerequisites).

## Install Flux

To install the `Flux` CLI, please follow the following steps:

{{% tabs %}} {{% tab "Homebrew" %}}

With Homebrew for macOS and Linux:

brew install fluxcd/tap/flux {{% /tab %}} {{% tab "bash" %}}

With Bash for macOS and Linux:

curl -s https://fluxcd.io/install.sh | sudo bash

### Flux Bootstrap

Using the flux bootstrap command you can install Flux on a Kubernetes cluster and configure it to manage itself from a
Git repository.

If the Flux components are present on the cluster, the bootstrap command will perform an upgrade if needed. The
bootstrap is idempotent, it's safe to run the command as many times as you want.

```
flux bootstrap git \
  --url=ssh://git@<host>/<org>/<repository> \
  --branch=<my-branch> \
  --path=clusters/my-cluster
```

### Install Weave GitOps

You can take the following yaml and add it to the git repository you defined. What is the yaml below? First, we have
a `GitRepository` object which is pointing to our public `weave-gitops` repository. This is also ignoring everything in
the repository except for the `charts` directory. Next, we have a `HelmRelease` object which is pointing to the path
where the chart resides. Flux will use these two objects to generate the application on your cluster once you commit the
yaml file to your gitops configuration repository that you defined during the `flux bootstrap` process.

```
---
apiVersion: source.toolkit.fluxcd.io/v1beta1
kind: GitRepository
metadata:
  name: ww-gitops
  namespace: flux-system
spec:
  interval: 1m0s
  ref:
    branch: v2
  secretRef:
    name: flux-system
  url: ssh://git@github.com/weaveworks/weave-gitops
  ignore: |
    # exclude all
    /*
    # include charts directory
    !/charts/
---
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  name: ww-gitops
  namespace: flux-system
spec:
  chart:
    spec:
      chart: ./charts/weave-gitops
      sourceRef:
        kind: GitRepository
        name: ww-gitops
  interval: 1m0s
```

Now, that you have commited, pushed, and merged these changes you should now see `wego-app` as a `Pod` running in your
cluster.