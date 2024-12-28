---
title: Install Weave GitOps Enterprise via CLI
---

# Install Weave GitOps Enterprise via CLI

!!! warning
    **This feature is in alpha and certain aspects will change**
    We're very excited for people to use this feature.
    However, please note that changes in the API, behaviour and security will evolve.
    The feature is suitable to use in controlled testing environments.

You could install Weave GitOps Enterprise via `gitops-ee bootstrap` CLI command which is suitable for two main scenarios:

1. **Day 0**: you want to get started quickly for discovery with the less knowledge possible.
2. **Day 1**: you have done discovery and want to set it up in your organisation.

Each scenario is supported by an operation modes:

1. **Interactive:** guides you step-by-step through the process until Weave GitOps Enterprise is up and running.
2. **Non-interactive:** for your automated workflows where you are already familiar with install process and have the configuration.

For those seeking other scenarios or fine-grain customisation [Weave GitOps Enterprise manual install](../install-enterprise) would be the recommended.

## Getting Started

### Prerequisites

Before you start make sure the following requirements are met:

- [ ] **Management Cluster**: a Kubernetes cluster with a Kubeconfig that has Admin permissions to be able to create resources.
- [ ] **Git Repository with SSH access**: this is the configuration repo that WeaveGitOps will use to sync configuration manifests from.
- [ ] **Flux CLI**: is [installed](https://fluxcd.io/flux/installation/#install-the-flux-cli) locally. It will be used for reconciling Flux resources.
- [ ] **Flux Bootstrapped** in your Management cluster via ssh. See [Flux Bootstrap](https://fluxcd.io/flux/installation/bootstrap/generic-git-server/) for more info.
- [ ] **Weave GitOps Enterprise Entitlements** are installed in the management cluster. Contact [Sales](/help-and-support/) for help on getting them.

#### Install `gitops-ee` CLI (> v0.35)

Weave GitOps Enterprise Bootstrap functionality is available on Weave GitOps Enterprise CLI starting from version v0.35. If you haven't already, please install the latest `gitops-ee` CLI using this command.

```bash
brew install weaveworks/tap/gitops-ee
```

#### Bootstrap Weave GitOps Enterprise

Please use the following command to start the installation wizard of Weave GitOps Enterprise.

=== "Interactive"

    ```bash
    gitops bootstrap
    ```
    The bootstrap wizard will take you step-by-step into configuring Weave GitOps Enterprise. To understand more about the CLI configurations experience, check the below sections [here](#cli-configurations).

=== "Non-Interactive"

   You could run the bootstrap command in non-interactive mode by providing the required configurations as flags. The following gives you an example to get started that you could adapt to your own context

   ```bash
    gitops bootstrap \
       --kubeconfig=$HOME/.kube/config \
       --private-key=$HOME/.ssh/id_rsa --private-key-password="" \
       --version="0.35.0" \
       --domain-type="localhost" \
       --password="admin123"
   ```

   For more information about the CLI configurations, check the below sections [here](#cli-configurations)

## Appendix

### Understanding `gitops-ee bootstrap`

`gitops-ee bootstrap` is a workflow that will take you through the following stages:

1. [Verify Flux](#verify-flux): verify Flux installation on the Management cluster.
2. [Verify Entitlement](#verify-entitlement): verify the Entitlements secret content (username, password, entitlement).
3. [Configure Git Access](#configure-git-access): configure the access to your configuration repo.
4. [Select WGE version](#select-wge-version): from the latest 3 available releases.
5. [Create Cluster User](#create-cluster-user): create a Secret with the username and password for the emergency cluster user.
6. [Configure Dashboard Access](#configure-dashboard-access): choose between 2 methods to access the dashboard either local or external.
7. [Access the dashboard](#access-the-dashboard): via the link from the installation success message.
8. (Optional) [Configure OIDC](#optional-configure-oidc): to enable login to dashboard via OIDC providers.

#### Verify Flux

Weave GitOps Enterprise runs on top of flux, the bootstrap CLI will check if flux is installed on the management cluster, and it will verify that it has the right version with valid git repository setup, and it is able to reconcile flux components properly.
If flux is installed, but doesn't have a valid installation, the bootstrap CLI will terminate pending the fix or uninstall of current flux installation.

#### Verify Entitlement

Weave GitOps Enterprise Entitlement is your obtained license to use our product. The Entitlements file is a Kubernetes secret that contains your licence.
`Bootstrapping` checks that the secret exists on the management cluster, and that it is valid will check if it has valid content and the entitlement is not expired.
To get the entitlement secret please contact *<sales@weave.works>*, then apply it on your management cluster with the name `weave-gitops-enterprise-credentials` under `flux-system` namespace.

#### Configure Git Access

In order for `gitops-ee bootstrap` to push WGE resources to the management cluster's git repository, you will be prompted to provide the private key used to access your repo via ssh. If the private key is encrypted, you will also be asked to provide the private key password.

!!! info
    Disclaimer: The bootstrap CLI will ONLY use the private key to push WGE resources to your repo, and won't use it in any other way that can comprimise your repo or clusters security.

#### Select WGE version

The bootstrap CLI will prompt you to choose from the latest 3 versions of Weave GitOps Enterprise.

#### Create Cluster User

You will be prompt to provide admin username and password, which will be used to access the dashboard. This will create admin secret with the credentials. If you already have previous admin credentials on your cluster, the installation will prompt you if you want to continue with the old credentials or exit and revoke them and re-run the installation.

#### Configure Dashboard Access

To access Weave GitOps Enterprise dashboard, you have the two following options available:

1. **Service**: this option is called `localhost` in the cli and the dashboard will be available through a [ClusterIP Service](https://kubernetes.io/docs/concepts/services-networking/service/#type-clusterip).
2. **Ingress**: this option is called `externaldns` the dashboard will be available through an [Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) with the following considerations:
    - An Ingress controller needs to exist.
    - A host-based ingress will be created of the ingress class `public-nginx`.
    - An [ExternalDNS](https://github.com/kubernetes-sigs/external-dns) annotation will be added with the value of the provided domain.

#### Access the dashboard

After installation is successful. The CLI will print out the URL where you can access the dashboard.

#### (Optional) Configure OIDC

OIDC configuration will enable you to login with OIDC provider beside, or instead of the admin credentials. Afte the installation is complete, you will be prompt if you want to configure OIDC access. If you don't want to set it up right away, you can do it later by running `gitops-ee bootstrap auth --type=oidc` command.

To configure OIDC access, you will be asked to provide the following values:
`DiscoveryUrl` this will verify that OIDC is accessible and get the issuerUrl from the OIDC settings.
`clientID` & `clientSecret` that you have configured on your OIDC static-clients.

!!! note
    Please don't forget to add a new static-client on your OIDC provider settings with the redirectURI `your-domain/oauth2/callback` for example `http://localhost:3000/oauth2/callback`

### CLI configurations

- `--kube-config`:                  allows to choose the Kubeconfig for your cluster, default would be ~/.kube/config
- `-d`, `--domain externaldns`:     indicate the domain to use in case of using externaldns
- `-t`, `--domain-type`:            dashboard domain type: could be 'localhost' or 'externaldns'
- `-h`, `--help`:                   help for bootstrap
- `-p`, `--password`:               Dashboard admin password
- `-k`, `--private-key`:            Private key path. This key will be used to push the Weave GitOps Enterprise's resources to the default cluster repository
- `-c`, `--private-key-password`:   Private key password. If the private key is encrypted using password
- `-u`, `--username`:               Dashboard admin username
- `-v`, `--version`:                Weave GitOps Enterprise version to install
