---
title: Join a Cluster with Azure Flux

---



# Joining a Cluster with Azure Flux ~ENTERPRISE~

## Prerequisites

See also our [guide to installing Weave GitOps Enterprise on Azure](install-enterprise-azure.md):
- An Azure cluster deployed with either the Azure Portal or Azure CLI tools.
- Azure Flux add-on deployed by adding a GitOps configuration, either via the Azure Portal or the CLI tool.

Note that this documentation applies to both Azure AKS and Azure ARC clusters.

## Initial Status

The Azure cluster already has the Azure Flux add-on installed. This differs from [CNCF Flux](https://fluxcd.io/) in that there are two additional controllers:
- fluxconfig-agent
- fluxconfig-controller

These controllers have CRDs that define the version of Flux and any Flux Kustomizations that are managed via the [Azure CLI](https://learn.microsoft.com/en-us/cli/azure/install-azure-cli).

The CRDs are all apiVersion: clusterconfig.azure.com/v1beta1. 

The Kinds are:
- FluxConfig
- FluxConfigSyncStatus

The FluxConfig Kind configures Flux itself and creates any Kustomizations that refer to a single-source GitRepository. This guide assumes that this process is already completed and that a top-level Kustomization has been configured for the fleet repo cluster directory already set up at
`clusters/default/CLUSTER_NAME/manifests`.

The CRDs that this FluxConfig generates are Flux CRDs, as follows:
- GitRepositories
- Kustomizations

These generated resources are viewable through Weave GitOps Enterprise.

Weave GitOps itself is deployed by Flux using a HelmRelease that pulls the Helm Chart. It doesn’t need to install Flux, as it is assumed that Flux is already deployed. Therefore it can use the Azure Flux add-on, which poses no conflicts with WGE itself.

Incompatibilities exist between the Azure Flux add-on and CNCF Flux. They should not be run at the same time, on the same cluster, due to conflicts in the CRD management. If the Flux bootstrapping process IS run on a cluster with Azure Flux add-on, it will override the Azure Flux add-on with the Flux version used in the bootstrap. Also, it would add Flux manifests to the source Git repository. This would be undesirable.

Azure Flux add-on-enabled clusters keep the Azure Flux add-on in place.

## Joining a Cluster to WGE

### Setting up a Service Account

To join a cluster, you'll set up a service account with permissions and create a kubeconfig for the service account. This service account does not need cluster admin permissions unless you are bootstrapping Flux into the cluster. The bootstrapping process will either be A) carried out before joining the cluster to WGE; or B) configured specifically for Flux to be bootstrapped into the cluster from WGE.

If you already have Flux running, you can create the service account in your fleet repo:

1. Create a service account file:

??? example "Expand to see role manifests"

    ```yaml
    apiVersion: v1
    kind: ServiceAccount
    metadata:
    name: wgesa
    namespace: default
    ---
    apiVersion: v1
    kind: Secret
    type: kubernetes.io/service-account-token
    metadata:
    name: wgesa-secret
    namespace: default
    annotations:
        kubernetes.io/service-account.name: "wgesa"
    ```


2. Create a roles file:

??? example "Expand to see role manifests"

    ```yaml
    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRoleBinding
    metadata:
        name: impersonate-user-groups
    subjects:
        - kind: ServiceAccount
        name: wgesa
        namespace: default
    roleRef:
        kind: ClusterRole
        name: user-groups-impersonator
        apiGroup: rbac.authorization.k8s.io
    ---
    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRole
    metadata:
        name: user-groups-impersonator
    rules:
        - apiGroups: [""]
        resources: ["users", "groups"]
        verbs: ["impersonate"]
        - apiGroups: [""]
        resources: ["namespaces"]
        verbs: ["get", "list"]
    ```


3. Commit to your fleet repo to sync.

4. Create a secret to store the kubeconfig, and a GitopsCluster object in the WGE management cluster that points to the kubeconfig secret. This allows you to connect to the target cluster and read various Kubernetes objects—including the Flux objects, such as:
- GitRepositories
- HelmReleases
- Kustomizations
- Providers
- Alerts
- Receivers

Kubernetes 1.24+ [will not create secrets for Service Accounts for you](https://stackoverflow.com/questions/75692230/secret-for-a-kubernetes-service-accounts-is-not-getting-created), so you have to add it yourself.

5. Add a new secret for the service account by adding to the service account yaml file in step 1.

6. Create a kubeconfig secret. We'll use a helper script to generate the kubeconfig, and then save it into `static-kubeconfig.sh`:

	??? example "Expand to see script"

        ```bash title="static-kubeconfig.sh"
        #!/bin/bash

        if [[ -z "$CLUSTER_NAME" ]]; then
                echo "Ensure CLUSTER_NAME has been set"
                exit 1
        fi

        if [[ -z "$CA_CERTIFICATE" ]]; then
                echo "Ensure CA_CERTIFICATE has been set to the path of the CA certificate"
                exit 1
        fi

        if [[ -z "$ENDPOINT" ]]; then
                echo "Ensure ENDPOINT has been set"
                exit 1
        fi

        if [[ -z "$TOKEN" ]]; then
                echo "Ensure TOKEN has been set"
                exit 1
        fi

        export CLUSTER_CA_CERTIFICATE=$(cat "$CA_CERTIFICATE" | base64)

        envsubst <<EOF
        apiVersion: v1
        kind: Config
        clusters:
        - name: $CLUSTER_NAME
            cluster:
                server: https://$ENDPOINT
                certificate-authority-data: $CLUSTER_CA_CERTIFICATE
        users:
        - name: $CLUSTER_NAME
            user:
                token: $TOKEN
        contexts:
        - name: $CLUSTER_NAME
            context:
                cluster: $CLUSTER_NAME
                user: $CLUSTER_NAME
        current-context: $CLUSTER_NAME

        EOF
        ```

	</details>

7. Create a secret for the generated kubeconfig in the WGE management cluster:

	```bash
	kubectl create secret generic demo-01-kubeconfig \
	--from-file=value=./demo-01-kubeconfig
	```

You can also take care of this step in WGE's [Secrets UI](https://docs.gitops.weave.works/docs/next/secrets/intro/), setting up a a secret in [SOPS](https://docs.gitops.weave.works/docs/next/secrets/setup-sops/) or [ESO](https://docs.gitops.weave.works/docs/next/secrets/setup-eso/).

Flux CRDs are compatible with the Azure Flux Configuration CRDs. This means that there are no compatibility issues between WGE and Azure Flux.

8. Create a GitopsCluster object. It must NOT be bootstrapped. Remove the annotation for bootstrap so it will not deploy Flux.

9. Commit to your fleet repo and sync.

10. Log in to your WGE management cluster to see if the cluster has appeared.

## Using WGE to Deploy Clusters

### With Cluster API

MSFT maintains CAPZ, the Azure CAPI provider. Currently there is no support for Azure Flux. A CAPI-based cluster will continue to run the Flux bootstrap process on cluster creation when managed by WGE, because there is no Azure Flux option.

### With Terraform Provider

WGE uses [TF-controller](https://github.com/weaveworks/tf-controller) to deploy Terraform resources. For WGE to use the cluster as a target requires A) a resource created in the management cluster and B) a kubeconfig that maps to a service account in the target cluster. The Terraform cluster build typically creates this service account and then outputs to a secret store or local secret so that WGE can use it as a cluster. The Flux bootstrap process can be initiated directly with the Flux Terraform module, which deploys CNCF Flux to the target cluster.

Alternatively, you can apply an Azure Policy to provide the Azure Flux add-on. This is an example of how you can use the policy controls. This means you could come across clusters that are deployed with Terraform with the Azure Flux add-on already installed and would not run the Flux bootstrap process.

Either way, it is typical that Terraform-deployed clusters do not run the Flux bootstrap process at all, because it is usually already installed.

### With Crossplane

The Azure Flux add-on is supported under [Crossplane](https://www.crossplane.io/)-deployed Azure clusters. Any clusters deployed with Crossplane that have the Azure Flux add-on enabled would also be added to WGE without running the bootstrap process.
