---
title: Deploying CAPA with EKS

---

# Deploying CAPA with EKS ~ENTERPRISE~

Weave GitOps Enterprise can leverage [Cluster API](https://cluster-api.sigs.k8s.io/introduction.html) providers to enable leaf cluster creation. Cluster API provides declarative APIs, controllers, and tooling to manage the lifecycle of Kubernetes clusters across a large number of [infrastructure providers](https://cluster-api.sigs.k8s.io/reference/providers.html#infrastructure). Cluster API custom resource definitions (CRDs) are platform-independent as each provider implementation handles the creation of virtual machines, VPCs, networks, and other required infrastructure parts—enabling consistent and repeatable cluster deployments.

As an AWS advanced technology partner, Weaveworks has been working tirelessly to ensure that deploying EKS **anywhere** is smooth and removes the barriers to application modernization.

## Prerequisites

You'll need to install the following software before continuing with these instructions:

- `github cli` >= 2.3.0 [(source)](https://cli.github.com/)
- `kubectl` [(source)](https://kubernetes.io/docs/tasks/tools/#kubectl)
- `eksctl` [(source)](https://github.com/weaveworks/eksctl/releases)
- the AWS Command Line Interface/`aws cli` [(source)](https://aws.amazon.com/cli/)
- `clusterctl` >= v1.1.3 [(source)](https://github.com/kubernetes-sigs/cluster-api/releases); follow [these steps](https://cluster-api-aws.sigs.k8s.io/getting-started.html#install-clusterctl) to initialise the cluster and enable feature gates
- `clusterawsadm` >= v1.1.0, following [Cluster API's instructions](https://github.com/kubernetes-sigs/cluster-api-provider-aws/releases)
- Make sure you have a management cluster. If you followed the Weave GitOps Enterprise [installation guide](../enterprise/install-enterprise.md), you'll have done this already.
- Configure your `AWS_ACCESS_KEY_ID`and `AWS_SECRET_ACCESS_KEY` with either `aws configure` or by exporting it in the current shell.
- Set the `GITHUB_TOKEN` as an environment variable in the current shell. It should have permissions to create Pull Requests against the cluster config repo.

## Multitenancy

Some Cluster API providers allow you to choose the account or identity that the new cluster will be created with. This is often referred to as _Multi-tenancy_ in the CAPI world. Weave GitOps currently supports:

- [**AWS** multi-tenancy](https://cluster-api-aws.sigs.k8s.io/topics/multitenancy.html)
- [**Azure** multi-tenancy](https://capz.sigs.k8s.io/topics/multitenancy.html)
- [**vSphere** multi-tenancy](https://github.com/kubernetes-sigs/cluster-api-provider-vsphere/blob/master/docs/identity_management.md)

## 1. Add Common RBAC to Your Repository

When a cluster is provisioned, by default it will reconcile all the manifests in `./clusters/<cluster-namespace>/<cluster-name>` and `./clusters/bases`.

To display Applications and Sources in the UI we need to give the logged in user permissions to inspect the new cluster.

Adding common RBAC rules to `./clusters/bases/rbac` is an easy way to configure this!

import WegoAdmin from "!!raw-loader!./assets/rbac/wego-admin.yaml";

```curl
curl -o clusters/bases/rbac/wego-admin.yaml https://docs.gitops.weave.works/assets/files/wego-admin-c80945c1acf9908fe6e61139ef65c62e.yaml
```

??? example "Expand to see full template yaml"
    ``` title="clusters/bases/rbac/wego-admin.yaml"

    ```

    <CodeBlock
    title="clusters/bases/rbac/wego-admin.yaml"
    className="language-yaml"
    >
    {WegoAdmin}
    </CodeBlock>

## 2. Build a Kubernetes Platform with Built-in Components Preconfigured for Your Organization

To do this, go to Weaveworks' [Profiles Catalog](https://github.com/weaveworks/profiles-catalog).

See [CAPI Templates](../gitops-templates/index.md) page for more details on this topic. Once we load a template we can use it in the UI to create clusters!

import CapaTemplate from "!!raw-loader!./assets/templates/capa-template.yaml";

Download the template below to your config repository path, then commit and push to your Git origin.

```curl
curl -o clusters/management/capi/templates/capa-template.yaml https://docs.gitops.weave.works/assets/files/capa-template-49001fbae51e2a9f365b80caebd6f341.yaml
```

``` title="clusters/management/apps/capi/templates/capa-template.yaml"
    --8<-- "./assets/templates/capa-template.yaml"
    {% include '/assets/templates/capa-template.yaml' %}
```

<CodeBlock
  title="clusters/management/apps/capi/templates/capa-template.yaml"
  className="language-yaml"
>
  {CapaTemplate}
</CodeBlock>

## 3. Add a Cluster Bootstrap Config

This step ensures that Flux gets installed into your cluster. Create a cluster bootstrap config as follows:

```bash
 kubectl create secret generic my-pat --from-literal GITHUB_TOKEN=$GITHUB_TOKEN
```

import CapiGitopsCDC from "!!raw-loader!./assets/bootstrap/capi-gitops-cluster-bootstrap-config.yaml";

Download the config with:

```curl
curl -o clusters/management/capi/bootstrap/capi-gitops-cluster-bootstrap-config.yaml https://docs.gitops.weave.works/assets/files/capi-gitops-cluster-bootstrap-config-d9934a1e6872a5b7ee5559d2d97a3d83.yaml
```

Then update the `GITOPS_REPO` variable to point to your cluster

??? example "Expand to see full yaml"

    ``` title="clusters/management/capi/boostrap/capi-gitops-cluster-bootstrap-config.yaml"
        --8<-- "./assets/bootstrap/capi-gitops-cluster-bootstrap-config.yaml"
    ```

    <CodeBlock
    title="clusters/management/capi/boostrap/capi-gitops-cluster-bootstrap-config.yaml"
    className="language-yaml"
    >
    {CapiGitopsCDC}
    </CodeBlock>

## 4. Delete a Cluster with the Weave GitOps Enterprise UI

Here are the steps:

- Select the clusters you want to delete
- Press the `Create a PR to delete clusters` button
- Either update the deletion PR values or leave the default values, depending on your situation
- Press the `Remove clusters` button
- Merge the PR for clusters deletion

Note that you can't apply an _empty_ repository to a cluster. If you have Cluster API clusters and other manifests committed to this repository, and then _delete all of them_ so there are zero manifests left, then the apply will fail and the resources will not be removed from the cluster.
A workaround is to add a dummy _ConfigMap_ back to the Git repository after deleting everything else so that there is at least one manifest to apply.

## 5. Disable CAPI Support

If you do not need CAPI-based cluster management support, you can disable CAPI
via the Helm Chart values.

Update your Weave GitOps Enterprise `HelmRelease` object with the
`global.capiEnabled` value set to `false`:

```yaml title="clusters/management/weave-gitops-enterprise.yaml"
---
apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: HelmRepository
metadata:
  name: weave-gitops-enterprise-charts
  namespace: flux-system
spec:
  interval: 60m
  secretRef:
    name: weave-gitops-enterprise-credentials
  url: https://charts.dev.wkp.weave.works/releases/charts-v3
---
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  name: weave-gitops-enterprise
  namespace: flux-system
spec:
  chart:
    spec:
      interval: 65m
      chart: mccp
      sourceRef:
        kind: HelmRepository
        name: weave-gitops-enterprise-charts
        namespace: flux-system
      version: 0.12.0
  install:
    crds: CreateReplace
  upgrade:
    crds: CreateReplace
  interval: 50m
  values:
    global:
      capiEnabled: false
```

And that's it!
