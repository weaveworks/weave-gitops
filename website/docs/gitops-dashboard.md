---
title: GitOps Dashboard
sidebar_position: 4
hide_title: true
---

import TierLabel from "./\_components/TierLabel";

<h1>
  {frontMatter.title} <TierLabel tiers="All tiers" />
</h1>

Weave GitOps provides a web UI to help you quickly understand your Application deployments and perform common operations, such as adding a new Application to be deployed to your cluster. The `gitops` binary contains an HTTP server that can be used to start this browser interface as per the instructions below:

To run the dashboard:

```shell
$ gitops ui run
INFO[0000] Opening browser at https://0.0.0.0:9001/
INFO[0000] Serving on 0.0.0.0:9001
Opening in existing browser session.
```

Your browser should open to the Weave GitOps UI:

![Weave GitOps UI](/img/wego_ui.png)

## What information can I discover about my Applications?

Applications being managed by Weave GitOps are displayed in a list. Clicking the name of an Application allows you to view more details including:

- It's name, deployment type (Kustomize or Helm), URL for the source repository being synchronized to the cluster and the specific Path within the repository where we are looking for deployment manifests.
- A reconciliation graph detailing the on-cluster components which have been created as a result of the deployment.
- Information from Flux regarding the state of the reconciliation
- A list of the 10 most recent commits to the source git repository helping you to verify which change has been applied to the cluster. This includes a hyperlink back to your git provider for each commit.

## TLS configuration

By default the dashboard will listen on 0.0.0.0:9001 with TLS enabled. A self-signed certificate and key pair are generated when the server starts.
As the certificate is _self-signed_, Chrome and other browser will show a warning you will have to click through to view the dashboard.

| Parameter           | Type   | Description                                                     | Default |
| ------------------- | ------ | --------------------------------------------------------------- | ------- |
| `--no-tls`          | bool   | Disable TLS, access the dashboard on default port via http      | false   |
| `--tls-private-key` | string | Filename for the TLS certficate, in-memory generated if omitted |         |
| `--tls-cert-file`   | string | Filename for the TLS key, in-memory generated if omitted        |         |
| `--host`            | string | Host to listen on                                               | 0.0.0.0 |

## Dashboard Login

There are 2 supported methods for logging in to the dashboard:
- Login via an OIDC provider
- Login via Static User account

The recommended method is to integrate with an OIDC provider, as this will let you control permissions for existing users and groups that have already been configured to use OIDC. However, it is also possible to use a Static User account to login, if an OIDC provider is not available to use. Both methods work with standard Kubernetes RBAC.

:::note FEATURE TOGGLE
The following instructions describe a feature that is behind a feature toggle. To enable this feature, set the following OS environment variable:
```sh
export WEAVE_GITOPS_AUTH_ENABLED=true
```
:::

### Login via an OIDC provider

#### Securing the dashboard using OIDC and Kubernetes RBAC

You may decide to give your engineering teams access to the dashboard, in order to view and manage their workloads. In this case, you will want to secure access to the dashboard and restrict who can interact with it. Weave GitOps integrates with your OIDC provider and uses standard Kubernetes RBAC to give you fine-grained control of the permissions for the dashboard users.

#### Background

OIDC extends the OAuth2 authorization protocol by including an additional field (ID Token) that contains information (claims) about a user's identity. After a user successfully authenticates with the OIDC provider, this information is used by Weave GitOps to impersonate the user in any calls to the Kubernetes API. This allows cluster administrators to use RBAC rules to control access to the cluster and also the dashboard.

#### Configuration

In order to login via your OIDC provider, you need to create a Kubernetes secret to store the OIDC configuration. This configuration consists of the following parameters:

| Parameter         |  Description                                                                                                                      | Default   |
| ------------------|  -------------------------------------------------------------------------------------------------------------------------------- | --------- |
| `IssuerURL`       |  The URL of the issuer, typically the discovery URL without a path                                                                |           |
| `ClientID`        |  The client ID that has been setup for Weave GitOps in the issuer                                                                 |           |
| `ClientSecret`    |  The client secret that has been setup for Weave GitOps in the issuer                                                             |           |
| `RedirectURL`     |  The redirect URL that has been setup for Weave GitOps in the issuer, typically the dashboard URL followed by `/oauth2/callback ` |           |
| `TokenDuration`   |  The time duration that the ID Token will remain valid, after successful authentication                               | "1h0m0s"  |           |

Ensure that your OIDC provider has been setup with a client ID/secret and the redirect URL of the dashboard.

Create a secret named `oidc-auth` in the `flux-system` namespace with these parameters set:

```sh
kubectl create secret generic oidc-auth \
  --namespace flux-system \
  --from-literal=issuerURL=<oidc-issuer-url> \
  --from-literal=clientID=<client-id> \
  --from-literal=clientSecret=<client-secret> \
  --from-literal=redirectURL=<redirect-url> \
  --from-literal=tokenDuration=<token-duration>
```

Running `gitops ui run` should now load the dashboard and show the option to login via your OIDC provider.

Once the HTTP server starts, it will redirect unauthenticated users to the provider's login page to authenticate them. Upon successful authentication, the users' identity will be impersonated in any calls made to the Kubernetes API, as part of any action they take in the dashboard. At this point, the dashboard will fail to render correctly unless RBAC has been configured accordingly. Follow the instructions in the [RBAC authorization](#rbac-authorization) section in order to configure RBAC correctly.

### Login via a Static User account

Before you login via the Static User account, you need to generate a bcrypt hash for your chosen password and store it as a secret in Kubernetes with the `gitops.weave.works/entity: user` label. There are several different ways to generate a bcrypt hash, this guide uses an Docker image to generate one:

Run an Docker image interactively and supply the password of your choice as an environment variable:

```sh
$ docker run -e CLEAR_PASSWORD="super secret password" -it golang:1.17 go install github.com/bitnami/bcrypt-cli@latest && echo -n $CLEAR_PASSWORD | bcrypt-cli

$2a$10$cNkw6LZJOgv9LvgvLNbuz.gFxNjjFLMNJhZE95NHjzOTBD06a3LnG
```

Now create a Kubernetes secret to store your chosen username and the password hash and label it:

```sh
kubectl create secret generic admin-user \
  --namespace flux-system \
  --from-literal=username=admin \
  --from-literal=password='$2a$10$cNkw6LZJOgv9LvgvLNbuz.gFxNjjFLMNJhZE95NHjzOTBD06a3LnG'

kubectl label secret admin-user gitops.weave.works/entity=user
```

You should now be able to login via the cluster user account using your chosen username and password. Follow the instructions in the next section in order to configure RBAC correctly.

### RBAC authorization

Both login methods work with standard Kubernetes RBAC. The following roles represent the minimal set of permissions needed to view applications, commits and profiles from the dashboard:

```yaml title="apps-profiles-reader.yaml"
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: apps-reader
  namespace: flux-system
rules:
  - apiGroups: ["wego.weave.works"]
    resources: ["apps"]
    verbs: ["get", "list"]
  - apiGroups: ["source.toolkit.fluxcd.io"]
    resources: ["gitrepositories"]
    verbs: ["get"]
  - apiGroups: ["source.toolkit.fluxcd.io"]
    resources: ["helmrepositories"]
    verbs: ["get"]
  - apiGroups: ["kustomize.toolkit.fluxcd.io"]
    resources: ["kustomizations"]
    verbs: ["get"]
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get"]
    resourceNames: ["wego-github-dev-cluster"] # name of secret created by Weave GitOps that contains the deploy key for the git repository
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: profiles-reader
  namespace: flux-system
rules:
  - apiGroups: ["source.toolkit.fluxcd.io"]
    resources: ["helmrepositories"]
    verbs: ["get"]
    resourceNames: ["weaveworks-charts"]
```

The following role represents the minimal set of permissions needed to add applications from the dashboard:

```yaml title="apps-writer.yaml"
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: apps-writer
rules:
  - apiGroups: ["apiextensions.k8s.io"]
    resources: ["customresourcedefinitions"]
    verbs: ["get"]
    resourceNames: ["apps.wego.weave.works"]
```

The table below contains all the permissions that the dashboard uses:

| Resource                    | API Group                     | Action   | Description                                                                                  |
| --------------------------- | ----------------------------- | -------- | -------------------------------------------------------------------------------------------- |
| `apps`                      | `wego.weave.works`            | `list`   | Required to list all applications                                                            |
| `apps`                      | `wego.weave.works`            | `get`    | Required to retrieve a single application                                                    |
| `gitrepositories`           | `source.toolkit.fluxcd.io`    | `get`    | Required to retrieve a single application                                                    |
| `kustomizations`            | `kustomize.toolkit.fluxcd.io` | `get`    | Required to retrieve a single application                                                    |
| `gitrepositories`           | `source.toolkit.fluxcd.io`    | `update` | Required to sync an application                                                              |
| `helmrepositories`          | `source.toolkit.fluxcd.io`    | `update` | Required to sync an application                                                              |
| `kustomizations`            | `kustomize.toolkit.fluxcd.io` | `update` | Required to sync an application                                                              |
| `secrets`                   |                               | `get`    | Required to read deploy key secret in order to retrieve the list of commits                  |
| `customresourcedefinitions` | `apiextensions.k8s.io`        | `get`    | Required to read custom resources of type `apps.wego.weave.works` when adding an application |

In order to assign permissions to a user, create a `RoleBinding`/`ClusterRoleBinding`. For example, the following role bindings assign all dashboard permissions to the `admin` user:

```yaml title="admin-role-bindings.yaml"
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: read-apps
  namespace: flux-system
subjects:
- kind: User
  name: admin
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: Role
  name: apps-reader
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: read-profiles
  namespace: flux-system
subjects:
- kind: User
  name: admin
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: Role
  name: profiles-reader
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: write-apps
subjects:
- kind: User
  name: admin
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: Role
  name: apps-writer
  apiGroup: rbac.authorization.k8s.io
```

To test whether permissions have been setup correctly for a specific user/group use the `kubectl auth can-i` subcommand:
```sh
kubectl auth can-i list apps --as "admin" --namespace flux-system
```

For more information about RBAC authorization visit the [Kubernetes reference documentation](https://kubernetes.io/docs/reference/access-authn-authz/rbac/).

## Future development

The GitOps Dashboard is under active development, watch this space for exciting new features.
