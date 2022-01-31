---
sidebar_position: 4
---

# GitOps Dashboard

Weave GitOps provides a web UI to help you quickly understand your Application deployments and perform common operations, such as adding a new Application to be deployed to your cluster. The `gitops` binary contains an HTTP server that can be used to start this browser interface as per the instructions below:

To run the dashboard:

```shell
$ gitops ui run
INFO[0000] Opening browser at http://0.0.0.0:9001/
INFO[0000] Serving on port 9001
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

## Securing the dashboard using OIDC and Kubernetes RBAC

:::note WORK IN PROGRESS 
The following instructions describe a feature that is being actively developed and is therefore behind a feature toggle. To enable this feature, set the following OS environment variable:
```sh
export WEAVE_GITOPS_AUTH_ENABLED=true
```
:::

You may decide to host the dashboard centrally to allow for your engineering teams to access it in order to manage their workloads. In this case, you will want to secure access to the dashboard and restrict who can interact with it. Weave GitOps integrates with your OIDC provider and uses standard Kubernetes RBAC to give you fine-grained control of the permissions for the dashboard users.

### Background

OIDC extends the OAuth2 authorization protocol by including an additional field (ID Token) that contains information (claims) about a user's identity. After a user successfully authenticates with the OIDC provider, this information is used by Weave GitOps to impersonate the user in any calls to the Kubernetes API. This allows cluster administrators to use RBAC rules to control access to the cluster and also the dashboard.

### Configuration

After enabling the feature, `gitops ui run` will require the following additional parameters:

| Parameter                     | Type | Description                                                           | Default   |
| ----------------------------- | - |--------------------------------------------------------------------- | --------- |
| `--oidc-issuer-url`           | string | The URL of the issuer, typically the discovery URL without a path     |           |
| `--oidc-client-id`     | string | The client ID that has been setup for Weave GitOps in the issuer      |           |
| `--oidc-client-secret` | string | The client secret that has been setup for Weave GitOps in the issuer  |           |
| `--oidc-redirect-url`         | string | The redirect URL that has been setup for Weave GitOps in the issuer, typically the dashboard URL followed by `/oauth2/callback ` |         |
| `--oidc-cookie-duration`      | duration | The time duration that the ID Token HTTP cookie will remain valid, after successful authentication | "1h0m0s" |

Ensure that your OIDC provider has been setup with a client ID/secret and the redirect URL of the dashboard.

Once the HTTP server starts, it will redirect unauthenticated users to the provider's login page to authenticate them. Upon successful authentication, the users' identity will be impersonated in any calls made to the Kubernetes API, as part of any action they take in the dashboard. At this point, the dashboard will fail to render correctly unless RBAC has been configured accordingly. The following manifests represent the minimal set of permissions needed to view applications, commits and profiles from the dashboard:

```yaml title="apps-reader.yaml"
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: apps-reader
  namespace: wego-system
rules:
- apiGroups: ["wego.weave.works"]
  resources: ["apps"]
  verbs: [ "get", "list"]
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
  namespace: default
rules:
- apiGroups: ["source.toolkit.fluxcd.io"]
  resources: ["helmrepositories"]
  verbs: ["get"]
  resourceNames: [ "weaveworks-charts"]
```

The following manifest represents the minimal set of permissions needed to add applications from the dashboard:

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

| Resource | API Group | Action | Description |
|-|-|-|-|
| `apps` | `wego.weave.works` | `list` | Required to list all applications |
| `apps` | `wego.weave.works` | `get` | Required to retrieve a single application |
| `gitrepositories` | `source.toolkit.fluxcd.io` | `get` | Required to retrieve a single application |
| `kustomizations` | `kustomize.toolkit.fluxcd.io` | `get` | Required to retrieve a single application |
| `gitrepositories` | `source.toolkit.fluxcd.io` | `update` | Required to sync an application |
| `helmrepositories` | `source.toolkit.fluxcd.io` | `update` | Required to sync an application |
| `kustomizations` | `kustomize.toolkit.fluxcd.io` | `update` | Required to sync an application |
| `secrets` |  | `get` | Required to read deploy key secret in order to retrieve the list of commits |
| `customresourcedefinitions` | `apiextensions.k8s.io` | `get` | Required to read custom resources of type `apps.wego.weave.works` when adding an application  |

## Future development
The GitOps Dashboard is under active development, watch this space for exciting new features.
