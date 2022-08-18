---
title: "Security Documentation"
linkTitle: "Security"
description: "Flux Security documentation."
weight: 140
---

<!-- For doc writers: Step-by-step security instructions should live on the appropriate documentation pages.
To fulfil our promise to end users, we should briefly outline the context here,
and link to the more detailed instruction pages from each relevant part of this outline. -->

## Introduction

Flux has a multi-component design, and integrates with many other systems.

This document outlines an overview of security considerations for Flux components,
project processes, artifacts, as well as Flux configurable options and
what they enable for both Kubernetes cluster and external system security.

See our [security processes document](/security) for vulnerability reporting, handling,
and disclosure of information for the Flux project and community.

Please also have a look at [our security-related blog posts](/tags/security/). We are writing there to inform you what we are doing to keep Flux and you safe!

## Signed container images

The Flux CLI and the controllers' images are signed using [Sigstore](https://www.sigstore.dev/) Cosign and GitHub OIDC.
The container images along with their signatures are published on GitHub Container Registry and Docker Hub.

To verify the authenticity of Flux's container images,
install [cosign](https://docs.sigstore.dev/cosign/installation/) and run:

```console
$ COSIGN_EXPERIMENTAL=1 cosign verify ghcr.io/fluxcd/source-controller:v0.21.1

Verification for ghcr.io/fluxcd/source-controller:v0.21.1 --
The following checks were performed on each of these signatures:
  - The cosign claims were validated
  - Existence of the claims in the transparency log was verified offline
  - Any certificates were verified against the Fulcio roots.
```

We also wrote [a blog post](/blog/2022/02/security-image-provenance/) which discusses the this in some more detail.

## Software Bill of Materials

For the Flux project we publish a Software Bill of Materials (SBOM) with each release.
The SBOM is generated with [Syft](https://github.com/anchore/syft) in the [SPDX](https://spdx.dev/) format.

The `spdx.json` file is available for download on the GitHub release page e.g.:

```shell
curl -sL https://github.com/fluxcd/flux2/releases/download/v0.25.3/flux_0.25.3_sbom.spdx.json | jq
```

Please also refer to [the blog post](/blog/2022/02/security-the-value-of-sboms/) which discusses the idea and value of SBOMs.

## Pod security standard

The controller deployments are configured in conformance with the
Kubernetes [restricted pod security standard](https://kubernetes.io/docs/concepts/security/pod-security-standards/#restricted):

- all Linux capabilities are dropped
- the root filesystem is set to read-only
- the seccomp profile is set to the runtime default
- run as non-root is enabled
- the filesystem group is set to 1337
- the user and group ID is set to 65534

## Controller permissions

While Flux integrates with other systems it is built on Kubernetes core controller-runtime and properly adheres to Kubernetes security model including RBAC [^1].

Flux installs a set of [RBAC manifests](https://github.com/fluxcd/flux2/tree/main/manifests/rbac).
These include:

1. A `crd-controller` `ClusterRole`, which:
    - Has full access to all the Custom Resource Definitions defined by Flux controllers
    - Can get, list, and watch namespaces and secrets
    - Can get, list, watch, create, patch, and delete configmaps and their status
    - Can get, list, watch, create, patch, and delete coordination.k8s.io leases
2. A `crd-controller` `ClusterRoleBinding`:
    - References `crd-controller` `ClusterRole` above
    - Bound to a service accounts for every Flux controller
3. A `cluster-reconciler` `ClusterRoleBinding`:
    - References `cluster-admin` `ClusterRole`
    - Bound to service accounts for only `kustomize-controller` and `helm-controller`

Flux uses these two `ClusterRoleBinding` strategies in order to allow for clear access separation using tools
purpose-built for policy enforcement (OPA, Kyverno, admission controllers).

For example, the design allows all controllers to access Flux CRDs (binds to `crd-controller` `ClusterRole`),
but only binds the Flux reconciler controllers for Kustomize and Helm to `cluster-admin` `ClusterRole`,
as these are the only two controllers that manage resources in the cluster.

However in a [soft multi-tenancy setup]({{< relref "../get-started#multi-cluster-setup" >}}),
Flux does not reconcile a tenant's repo under the `cluster-admin` role.
Instead, you specify a different service account in your manifest, and the Flux controllers will use
the Kubernetes Impersonation API under `cluster-admin` to impersonate that service account [^2].
In this way, policy restrictions for this service account are applied to the manifests being reconciled.
If the binding is not defined for the correct service account and namespace, it will fail.
The roles and permissions for this multi-tenancy approach
are described in detail here: <https://github.com/fluxcd/flux2-multi-tenancy>.

## Further securing Flux Deployments

Beyond the security features that Flux has backed into it, there are further best
practices that can be implemented to ensure your Flux deployment is as secure
as it can be. For more information, checkout the [Flux Security Best Practices]({{< relref "./best-practices" >}}).

[^1]: However, by design cross-namespace references are an exception to RBAC.
Platform admins have to option to turnoff cross-namespace references as described in the
[installation documentation](../installation/_index.md#multi-tenancy-lockdown).
[^2]: Platform admins have to option to enforce impersonation as described in the
[installation documentation](../installation/_index.md#multi-tenancy-lockdown).
