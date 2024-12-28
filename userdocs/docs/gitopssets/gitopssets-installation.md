---
title: Installation

---



# Installation ~ENTERPRISE~

The gitopssets-controller can be installed in two ways:

- As part of the Weave GitOps Enterprise installation. (installed by default)
- As a standalone installation using a Helm chart.

The standalone installation can be useful for leaf clusters that don't have Weave GitOps Enterprise installed.

## Prerequisites

Before installing the gitopssets-controller, ensure that you've installed [Flux](https://github.com/fluxcd/flux2).

## Installing the gitopssets-controller

To install the gitopssets-controller using a Helm chart, use the following HelmRelease:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: gitopssets-system
---
apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: HelmRepository
metadata:
  name: weaveworks-oci-charts
  namespace: gitopssets-system
spec:
  interval: 1m
  type: oci
  url: oci://ghcr.io/weaveworks/charts
---
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  name: gitopssets-controller
  namespace: gitopssets-system
spec:
  interval: 10m
  chart:
    spec:
      chart: gitopssets-controller
      sourceRef:
        kind: HelmRepository
        name: weaveworks-oci-charts
        namespace: gitopssets-system
      version: 0.15.3
  install:
    crds: CreateReplace
  upgrade:
    crds: CreateReplace
```

After adding the Namespace, HelmRepository and HelmRelease to a Git repository synced by Flux, commit the changes to complete the installation process.

## Customising the Generators

Not all generators are enabled by default, this is because not all CRDs are required by the generators.

You might want to enable or disable individual generators via the Helm Chart:

```yaml
gitopssets-controller:
  enabled: true
  controllerManager:
    manager:
      args:
        - --health-probe-bind-address=:8081
        - --metrics-bind-address=127.0.0.1:8080
        - --leader-elect
        # enable the cluster generator which is not enabled by default
        - --enabled-generators=GitRepository,Cluster,PullRequests,List,APIClient,Matrix,Config
```
