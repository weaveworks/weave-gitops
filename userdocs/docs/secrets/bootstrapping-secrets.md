---
title: Bootstrapping Secrets

---

# Bootstrapping Secrets ~ENTERPRISE~

When accessing protected resources there is a need for a client to authenticate before
the access is granted and the resource is consumed. For authentication, a client presents
credentials that are either created manually or available through infrastructure. A common scenario
is to have a secrets store.

Weave Gitops allows you to provision the secret management infrastructure as part of its capabilities.
However, in order to provision, as any other application that has secrets, we need to create the secret needed for installing it.
This is known as a chicken-egg scenario that get addressed by providing an initial secret. This secret we call it
bootstrapping secret.

Bootstrapping secrets comes in handy, not only while provisioning your secrets management solution,
but also in any platform provisioning task where the existence of the secret is a prerequisite.
Another common example could be provisioning platform capabilities via [profiles](../cluster-management/profiles.md)
that are stored in [private repositories](https://fluxcd.io/flux/guides/helmreleases/#helm-repository-authentication-with-credentials).

Weave Gitops provides [SecretSync](#secretsync) as a solution to managing your bootstrapping secrets.

## SecretSync

!!! warning
    **This feature is in alpha and certain aspects will change**
    We're very excited for people to use this feature.
    However, please note that changes in the API, behaviour and security will evolve.
    The feature is suitable to use in controlled testing environments.

`SecretSync` is a [Kubernetes Custom Resource](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)
that provides semantics to sync [Kuberentes Secrets](https://kubernetes.io/docs/concepts/configuration/secret/) from management cluster to leaf clusters.

An example could be seen below:

```yaml
apiVersion: capi.weave.works/v1alpha1
kind: SecretSync
metadata:
  name: my-dev-secret-syncer
  namespace: default
spec:
  clusterSelector:
    matchLabels:
      environment: dev
  secretRef:
    name: my-dev-secret
  targetNamespace: my-namespace
```
Where you could find the following configuration sections:

1) Select the secret to sync:

```yaml
  secretRef:
    name: my-dev-secret
```

2) Specify the [GitopsClusters](../cluster-management/managing-clusters-without-capi.md)
that the secret will be synced to via labels:

```yaml
  clusterSelector:
    matchLabels:
      environment: dev
```

`Secretsync` reconciles secrets on clusters: any cluster at a future time matching the label selector will have
the secret reconciled too.

More info about the CRD spec [here](./spec/v1alpha1/secretSync.md)

### Working with SecretSync

#### Pre-requisites

1. You are using [Weave Gitops Enterprise version > v0.19.0](../enterprise/releases-enterprise.md)
2. You have a set of GitopsClusters representing the clusters that you want to sync the secret to. They have a set of labels to allow matching.

??? example "Expand to see example"

    ``` yaml
    apiVersion: gitops.weave.works/v1alpha1
    kind: GitopsCluster
    metadata:
    namespace: flux-system
    labels:
        environment: dev
    ```

3. You have created a secret that you want to sync from the management cluster.

??? example "Expand to see example"

    ```yaml
    apiVersion: v1
    kind: Secret
    metadata:
    name: my-dev-secret
    namespace: flux-system
    type: Opaque
    ```

!!! info
    Some restriction apply to the current version:
    - Resources should be in the same namespace
    - Leaf cluster nodes should be annotated with `node-role.kubernetes.io/control-plane`

#### Syncing secrets via SecretSync

1. Create SecretSync manifests that points to your secret. For example:

```yaml
apiVersion: capi.weave.works/v1alpha1
kind: SecretSync
metadata:
  name: my-dev-secret-syncer
  namespace: default
spec:
  clusterSelector:
    matchLabels:
      environment: dev
  secretRef:
    name: my-dev-secret
  targetNamespace: my-namespace
```

2. Check-in to your configuration repo within your management cluster

3. Create a PR, review and merge

4. See the progress

Once reconciled, the status section would show created secret resource version

```
status:
  versions:
    leaf-cluster-1: "211496"
```

5. See the secret created in your leaf cluster

Your secret has been created in the target leaf cluster

```bash
➜  kubectl get secret -n default
NAME               TYPE                                  DATA
my-dev-secret      Opaque                                1
```
