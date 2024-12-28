---
title: Setup ESO

---

# Setup ESO ~ENTERPRISE~

Weave GitOps Enterprise now supports managing secrets using [External Secrets Operator](https://external-secrets.io/v0.8.1/) from the [UI](./manage-secrets-ui.md#external-secrets). External Secrets Operator is a Kubernetes operator that allows users to use secrets from external secrets management systems by reading their information using external APIs and injecting their values into Kubernetes secrets. To be able to use this functionality, users need to configure their External Secrets Operator and SecretStores using one of the guides below.

## Prerequisites

### SecretStores

You should have your [SecretStore CRs](https://external-secrets.io/v0.8.1/) defined in a git repository. Those CRs will be installed to your cluster in the following steps and used by the creation UI.

### ESO Profile

The [ESO profile](https://github.com/weaveworks/weave-gitops-profile-examples/tree/main/charts/external-secrets) is packaged with the [weaveworks-charts](https://github.com/weaveworks/weave-gitops-profile-examples). If you have the usual profiles setup, you will not need to do anything extra.
This profile installs the ESO controller, all the required CRDs, and the SecretStore CRs defined in the [previous](./#secretstores) step.

### Secrets

There are several Kubernetes Secrets that need to exist on your management cluster for the whole flow to work.

If your SecretStores repository is private then you'll need a Secret, that contains the repo credentials, to access the repository. This is usually the Secret you created while bootstrapping Flux on the management cluster and is copied to your leaf cluster during creation.

For each SecretStore CR, you'll need to add a Secret, that follows the format expected by this CR, to allow the operator to access the defined External Secret Store.

Follow this [guide](/secrets/bootstrapping-secrets.md) for bootstrapping those secrets on leaf clusters.

## Installation Steps

### Install ESO on management cluster or existing leaf cluster

To install the ESO profile on an exisitng cluster, use `Add an application` from the `Applications` page and select `external-secrets` from `weaveworks-charts`. Check the [Profile values](./#profile-values) section for more info about configuring the `values.yaml`.

### Install ESO on leaf cluster

To bootstrap the ESO profile on a leaf cluster, select `external-secrets` from the profiles dropdown in the `Create Cluster` page. Check the [Profile values](./#profile-values) section for more info about configuring the `values.yaml`.

### Profile values

You should then configure the `values.yaml` to install the `SecretStores` on the cluster from a `GitRepository`.
This is done by configuring the `secretStores` section.

??? example "Expand to see an example that creates a new git source from a specific tag"

    ```yaml
    secretStores:
    enabled: true
    url: ssh://git@github.com/github-owner/repo-name   # url for the git repository that contains the SecretStores
    tag: v1.0.0
    path: ./    # could be a path to the secrets dir or a kustomization.yaml file for the SecretStore in the GitRepository
    secretRef: my-pat   # the name of the Secret containing the repo credentials for private repositories
    ```

??? example "Expand to see an example that uses an existing git source"

    ```yaml
    secretStores:
    enabled: true
    sourceRef: # Specify the name for an existing GitSource reference
        kind: GitRepository
        name: flux-system
        namespace: flux-system
    ```
