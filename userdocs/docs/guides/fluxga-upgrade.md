---
title: Upgrade to Flux GA

---

# Upgrade to Flux GA

We are very excited for the release of the [Flux v2.0 GA!](https://github.com/fluxcd/flux2/releases)

This guide aims to answer some [common questions](#faq) before starting the upgrade, and provides step-by-step
instructions.

## Before Starting the Upgrade

Useful terms used in this guide:

- `Flux Beta or Flux v0.x` as the [latest Flux Beta Release](https://github.com/fluxcd/flux2/releases/tag/v0.41.2).
- `Flux GA` as the [latest Flux GA Release Candidate](https://github.com/fluxcd/flux2/releases/tag/v2.0.0-rc.3)
- `Weave GitOps` as the [latest Weave GitOps Enterprise release](https://github.com/weaveworks/weave-gitops-enterprise/releases/latest)

## FAQ

Here you can find the most common questions around upgrading.

### Why Upgrade to Flux GA

Although Flux Beta APIs have been stable and used in production for quite some time, Flux GA is the main supported API version for new features and development. Features like [horizontal scaling](https://fluxcd.io/flux/cheatsheets/sharding/)
are only available in Flux GA. Also, beta APIs will be removed after six months.

### Can I Use Weave GitOps with Flux GA?

Yes. This has been possible since Weave Gitops v0.22.0. Use the [latest available release](https://github.com/weaveworks/weave-gitops/releases) for the best experience.

### Can I Use Weave GitOps Enterprise with Flux GA?

Yes. This has been possible since Weave GitOps Enterprise v0.22.0. Use the [latest available release](https://docs.gitops.weave.works/docs/enterprise/releases-enterprise/) for the best experience.

The following limitations are knowns by version:

#### v0.23.0 onwards

No limitations

#### v0.22.0

If you are using GitOpsSets, upgrade that component to v0.10.0 for Flux GA compatibility.
Update the Weave GitOps Enterprise HelmRelease values to use the new version.

```yaml
gitopssets-controller:
  controllerManager:
    manager:
      image:
        tag: v0.10.0
```

### Can I Use Weave GitOps with Flux v2 0.x (pre-GA versions)?

As of Weave GitOps v0.29, only Flux v2.0 GA is supported. Please follow the [Upgrade](#upgrade) section to help you with the process.

Earlier versions of Weave GitOps work with both Flux v2 GA and Flux v2 0.x (the pre-GA ones), but it is encouraged that you upgrade to the latest version for the best experience.

## Upgrade

!!! info "Hosted flux?"
    If you are using a hosted Flux version, please check with your provider if they support Flux GA before upgrading following this guide.
    Known hosted Flux providers:

    - EKS Anywhere
    - [Azure AKS Flux-GitOps extension](https://learn.microsoft.com/en-us/azure/azure-arc/kubernetes/extensions-release#flux-gitops)

    As of writing they do not yet support the new version, so please wait before upgrading to Flux GA.

Below, we'll take you through the multiple steps required to migrate to your system to Flux GA. After each step the cluster will be
in a working state, so you can take your time to complete the migration.

1. Upgrade to Flux GA on your existing leaf clusters and management clusters
2. Upgrade to Flux GA in `ClusterBootstrapConfig`s.
3. Upgrade to [latest Weave GitOps](https://docs.gitops.weave.works/docs/enterprise/releases-enterprise/).
4. Upgrade GitopsTemplates, GitopsSets and ClusterBootstrapConfigs.

### 1. Upgrade to Flux GA on your existing leaf clusters and management clusters

Follow the upgrade instructions from the [Flux v2.0.0 release notes](https://github.com/fluxcd/flux2/releases/tag/v2.0.0).

At minimum, you'll need to rerun the `flux bootstrap` command on your leaf clusters and management clusters.

You'll also need to bump API versions in your manifests to `v1` as described in the Flux upgrade instructions:

> Bumping the APIs version in manifests can be done gradually. It is advised to not delay this procedure as the beta
> versions will be removed after 6 months.

At this stage all clusters are running Flux GA.

### 2. Upgrade to Flux GA in ClusterBootstrapConfigs

First, we ensure any new clusters are bootstrapped with Flux GA. Then we'll upgrade the existing clusters.

`ClusterBootstrapConfig` will most often contain an invocation of `flux bootstrap`. Make sure the image is using `v2`.

??? example "Expand to see example"

    ```patch
    diff --git a/tools/dev-resources/user-guide/cluster-bootstrap-config.yaml b/tools/dev-resources/user-guide/cluster-bootstrap-config.yaml
    index bd41ec036..1b21df860 100644
    --- a/tools/dev-resources/user-guide/cluster-bootstrap-config.yaml
    +++ b/tools/dev-resources/user-guide/cluster-bootstrap-config.yaml
    @@ -1,34 +1,34 @@
    apiVersion: capi.weave.works/v1alpha1
    kind: ClusterBootstrapConfig
    metadata:
    name: capi-gitops
    namespace: default
    spec:
    clusterSelector:
        matchLabels:
        weave.works/capi: bootstrap
    jobTemplate:
        generateName: "run-gitops-{{ .ObjectMeta.Name }}"
        spec:
        containers:
    -        - image: ghcr.io/fluxcd/flux-cli:v0.34.0
    +        - image: ghcr.io/fluxcd/flux-cli:v2.0.0
            name: flux-bootstrap
            ...
    ```

At this stage, your new bootstrapped clusters will run Flux GA.

### 3. Upgrade to latest WGE

Use your regular WGE upgrade procedure to bring it to the [latest version](https://docs.gitops.weave.works/docs/enterprise/releases-enterprise/)

At this stage you have Weave GitOps running Flux GA.

### 4. Upgrade GitOpsTemplates, GitOpsSets, and ClusterBootstrapConfigs

Bumping the APIs version in manifests can be done gradually. We advise against delaying this procedure as the Beta versions will be removed after six months.

#### `GitOpsTemplate` and `CAPITemplate`

Update `GitRepository` and `Kustomization` CRs in the `spec.resourcetemplates` to `v1` as described in the flux upgrade instructions.

#### `GitOpsSets`

Update `GitRepository` and `Kustomization` CRs in the `spec.template` of your `GitOpsSet` resources to `v1` as described in the Flux upgrade instructions.

### 5. Future steps

If you haven't done it yet, plan to update your  `Kustomization` , `GitRepository` and `Receiver` resources to `v1`, you can also upgrade to the future release of Flux that will drop support for `< v1` APIs.

## Contact us

If you find any issues, please let us know via [support](https://docs.gitops.weave.works/help-and-support/).