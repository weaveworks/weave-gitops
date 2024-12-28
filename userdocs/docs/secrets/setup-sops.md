---
title: Setup SOPS
---

import CodeBlock from "@theme/CodeBlock";

import SopsBootstrapJob from "!!raw-loader!./assets/sops-bootstrap-job.yaml";
import TemplateParams from "!!raw-loader!./assets/template-params.yaml";
import TemplateAnnotations from "!!raw-loader!./assets/template-annotations.yaml";

# Setup SOPS ~ENTERPRISE~

Weave GitOps Enterprise now supports managing secrets using SOPS, a tool that encrypts and decrypts secrets using various key management services, from the [UI](manage-secrets-ui.md#sops). To be able to use this functionality, users need to configure their private and public key-pairs using one of the guides below.

## Setup SOPS on management cluster or existing leaf cluster

In this section, we will cover the prerequisites for using [SOPS](https://github.com/mozilla/sops) with Weave GitOps Enterprise, and how to configure SOPS for your existing Kubernetes cluster to work with GPG and age keys.

For a more advanced setup for SOPS with flux, please refer to this [guide](https://fluxcd.io/flux/guides/mozilla-sops/).

### Encrypting secrets using GPG/OpenPGP

OpenPGP is a way of using SOPS to encrypt and decrypt secrets with Weave GitOps Enterprise.

Here are the steps to generate an OpenPGP key and configure your cluster to work with Weave GitOps Enterprise secrets management.

1- Generate a gpg key pairs

??? example "Expand for instructions"

    ```bash
    export KEY_NAME="gpg-key"
    export KEY_COMMENT="gpg key"

    gpg --batch --full-generate-key <<EOF
    %no-protection
    Key-Type: 1
    Key-Length: 4096
    Subkey-Type: 1
    Subkey-Length: 4096
    Expire-Date: 0
    Name-Comment: ${KEY_COMMENT}
    Name-Real: ${KEY_NAME}
    EOF
    ```

2- Export the key pairs fingerprint in the shell

```bash
gpg --list-secret-keys "${KEY_NAME}"

sec   rsa4096 2020-09-06 [SC]
      710DC0DB6C1662F707095FC30233CB21E656A3CB

export KEY_FP="710DC0DB6C1662F707095FC30233CB21E656A3CB"
```

3- Export the generated private key to a kubernetes secret `sops-gpg-private-key` which will be used by flux's kustomize-controller to decrypt the secrets using sops.

```bash
gpg --export-secret-keys --armor "${KEY_FP}" |
kubectl create secret generic sops-gpg-private-key \
--namespace=flux-system \
--from-file=sops.asc=/dev/stdin
```

4- Export the generated public key to a kubernetes secret `sops-gpg-public-key` which will be used by Weave GitOps Enterprise to encrypt the secrets created from the UI.

```bash
gpg --export --armor "${KEY_FP}" |
kubectl create secret generic sops-gpg-public-key \
--namespace=flux-system \
--from-file=sops.asc=/dev/stdin
```

!!! tip
    It's recommended to remove the secret from your machine

    ```bash
    gpg --delete-secret-keys "${KEY_FP}"
    ```

5- Create a kustomization for reconciling the secrets on the cluster and set the `--decryption-secret` flag to the name of the private key created in step 3.

```bash
flux create kustomization gpg-secrets \
--source=secrets \ # the git source to reconcile the secrets from
--path=./secrets/gpg \
--prune=true \
--interval=10m \
--decryption-provider=sops \
--decryption-secret=sops-gpg-private-key
```

6- Annotate the kustomization object created in the previous step with the name and namespace of the public key created in step 4.

```bash
kubectl annotate kustomization gpg-secrets \
sops-public-key/name=sops-gpg-public-key \
sops-public-key/namespace=flux-system \
-n flux-system
```

??? example "Expand to see the expected kustomization object"

    ```yaml
    apiVersion: kustomize.toolkit.fluxcd.io/v1beta2
    kind: Kustomization
    metadata:
    name: gpg-secrets
    namespace: flux-system
    annotations:
        sops-public-key/name: sops-gpg-public-key
        sops-public-key/namespace: flux-system
    spec:
    interval: 10m
    sourceRef:
        kind: GitRepository
        name: secrets
    path: ./secrets/gpg
    decryption:
        provider: sops
        secretRef:
        name: sops-gpg-private-key
    prune: true
    validation: server
    ```

!!! note
    This is an essential step in order to allow other operators and developers to utilize WeaveGitOps UI to encrypt SOPS secrets using the public key secret in the cluster.

### Encrypting secrets using age

[age](https://github.com/FiloSottile/age) is a simple, modern and secure file encryption tool, that can be used to encrypt secrets using Weave GitOps Enterprise.

Here are the steps to generate an age key and configure your cluster to work with Weave GitOps Enterprise secrets management.

1- Generate an age key with age-keygen

```bash
age-keygen -o age.agekey

Public key: <public key>
```

2- Export the generated private key to a kubernetes secret `sops-age-private-key` which will be used by flux's kustomize-controller to decrypt the secrets using sops.

```bash
cat age.agekey |
kubectl create secret generic sops-age-private-key \
--namespace=flux-system \
--from-file=age.agekey=/dev/stdin
```

4- Export the generated public key to a kubernetes secret `sops-age-public-key` which will be used by Weave GitOps Enterprise to encrypt the secrets created from the UI.

```bash
echo "<public key>" |
kubectl create secret generic sops-age-public-key \
--namespace=flux-system \
--from-file=age.agekey=/dev/stdin
```

!!! tip
    It's recommended to remove the secret from your machine

    ```bash
    rm -f age.ageKey
    ```

5- Create a kustomization for reconciling the secrets on the cluster and set the `--decryption-secret` flag to the name of the private key created in step 2.

```bash
flux create kustomization age-secrets \
--source=secrets \ # the git source to reconcile the secrets from
--path=./secrets/age \
--prune=true \
--interval=10m \
--decryption-provider=sops \
--decryption-secret=sops-age-private-key
```

6- Annotate the kustomization object created in the previous step with the name and namespace of the public key created in step 4.

```bash
kubectl annotate kustomization age-secrets \
sops-public-key/name=sops-age-public-key \
sops-public-key/namespace=flux-system \
-n flux-system
```

??? example "Expand to see the expected kustomization object"

    ```yaml
    apiVersion: kustomize.toolkit.fluxcd.io/v1beta2
    kind: Kustomization
    metadata:
    name: age-secrets
    namespace: flux-system
    annotations:
        sops-public-key/name: sops-age-public-key
        sops-public-key/namespace: flux-system
    spec:
    interval: 10m
    sourceRef:
        kind: GitRepository
        name: secrets
    path: ./secrets/age
    decryption:
        provider: sops
        secretRef:
        name: sops-age-private-key
    prune: true
    validation: server
    ```


!!! note
    This is an essential step in order to allow other operators and developers to utilize WeaveGitOps UI to encrypt SOPS secrets using the public key secret in the cluster.

!!! tip
    In case of using OpenPGP and age in the same cluster, you need to make the kustomizations point to different directories. This is because flux's kustomize-controller expects that all the secrets in the kustomization's path are encrypted with the same key.

## Bootstrapping SOPS to leaf clusters

Bootstrapping SOPS to leaf clusters in WGE can be done by utilizing `ClusterBootstrapConfig` job to bootstrap Flux and SOPS.
The job is a container which generates SOPS secrets key pair, creates a kubernetes secret with the private key, creates a kubernetes secret with the public key (to be used in self-serve flow) and the proper rbac for it.
As well as an option to push the public key to the git repository via a PR (to be distributed).

### Prerequisites

#### ClusterBootstrapConfig job

The following example is using GPG encryption to install SOPS and generate keys when bootstrapping leaf clusters. Create the following `ClusterBootstrapConfig` CR and push it to your fleet repo.

??? example "Expand to view "

    <CodeBlock
    title="clusters/management/capi/boostrap/sops-bootstrap-job.yaml"
    className="language-yaml"
    >
    {SopsBootstrapJob}
    </CodeBlock>


#### Cluster template updates

In order to bootstrap SOPS to leaf clusters, we need some modifications to the cluster template to allow creating a [Kustomization](https://fluxcd.io/flux/guides/mozilla-sops/#configure-in-cluster-secrets-decryption)
for reconciling the secrets on the cluster using SOPS and to run the `ClusterBootstrapConfig` job during cluster creation.

The template metadata should have annotation, it will be used by WGE to create the Kustomization with the cluster files.

```yaml
templates.weave.works/sops-enabled: "true"
```

The template should have the following parameters that are needed for the Kustomization

??? example "Expand to view "

    <CodeBlock
    title="clusters/management/capi/templates/template.yaml"
    className="language-yaml"
    >
    {TemplateParams}
    </CodeBlock>


The template should have the following annotations under `GitOpsCluster` to be used in the bootstrap job

??? example "Expand to view "

    <CodeBlock
    title="clusters/management/capi/templates/template.yaml"
    className="language-yaml"
    >
    {TemplateAnnotations}
    </CodeBlock>


### Installation Steps

To bootstrap SOPS on a leaf cluster, create a new cluster using the SOPS template from the `Create Cluster` page and fill in the following SOPS-related values in the form:

- `SOPS_KUSTOMIZATION_NAME`: This Kustomization will be used to decrypt SOPS secrets from this path `clusters/default/leaf-cluster/sops/` after reconciling on the cluster. example (`my-secrets`)
- `SOPS_SECRET_REF`: The private key secret name that will be generated by SOPS in the bootstrap job. example (`sops-gpg`)
- `SOPS_SECRET_REF_NAMESPACE`: The private key secret namespace this secret will be generated by SOPS in the bootstrap job. example (`flux-system`)
- `SOPS_KEY_NAME`: SOPS key name. This will be used to generate SOPS keys. example (`test.yourdomain.com`)
- `SOPS_KEY_COMMENT`: SOPS key comment. This will be used to generate SOPS keys. example (`sops secret comment`)
- `SOPS_PUSH_TO_GIT`: Option to push the public key to the git repository. expected values (`true`, `false`)

![Bootstrap SOPS](../img/sops.png)

### What to expect

- A leaf cluster created with Flux & SOPS bootstrapped
- A secret created on leaf cluster `sops-gpg` to decrypt secrets
- A secret created on leaf cluster `sops-gpg-pub` to encrypt secrets
- A Kustomization with `decryption` defined in it to `SOPS` location in the cluster repo location
- Added Role for the public key to be accessed through management cluster
- A PR is created to the cluster repo with the public key and SOPS creation rules (optional)
- Visit the Secrets Page and start managing your secrets via the [UI](manage-secrets-ui.md)

## Security Recommendations

Access to sops decryption secrets should be restricted and allowed only to be read by flux's kustomize controller. This can be done using Kubernetes RBAC.

Here's an example of how you can use RBAC to restrict access to sops decryption secrets:

1. Create a new Kubernetes role that grants read access to sops decryption secrets

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: sops-secrets-role
rules:
- apiGroups: [""]
  resources: ["secrets"]
  resourceNames: ["sops-gpg-private-key", "sops-age-private-key"]
  verbs: ["get", "watch", "list"]
```

2. Bind the role to the service account of the flux's kustomize-controller

??? example "Expand to view the RoleBinding "

    ```yaml
    apiVersion: rbac.authorization.k8s.io/v1
    kind: RoleBinding
    metadata:
        name: sops-secrets-rolebinding
    roleRef:
        apiGroup: rbac.authorization.k8s.io
        kind: Role
        name: sops-secrets-role
    subjects:
    - kind: ServiceAccount
        name: kustomize-controller
    ```


!!! warning
    You would need to ensure that no other rolebindings or clusterrolebndings would allow reading the the decryption secret at any time. This could be achieved by leveraging policy capabilities to detect existing and prevent future creation of roles that would grant read secrets permissions.
