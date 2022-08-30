---
title: Configuring OIDC with Dex and GitHub
sidebar_position: 3
---

In this guide we will show you how to enable users to login to the Weave GitOps dashboard by authenticating with their GitHub account.

This example uses [Dex][tool-dex] and its GitHub connector, and assumes Weave GitOps has already been installed on a Kubernetes clusters.

### Pre-requisites
- A Kubernetes cluster such as [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/) cluster running a 
[Flux-supported version of Kubernetes](https://fluxcd.io/docs/installation/#prerequisites)
- Weave GitOps is [installed](../installation.mdx) and [TLS has been enabled](../configuration/tls.md).

## What is Dex?

[Dex][tool-dex] is an identity service that uses [OpenID Connect][oidc] to
drive authentication for other apps.  

Alternative solutions for identity and access management exist such as [Keycloak](https://www.keycloak.org/).

[tool-dex]: https://dexidp.io/
[oidc]: https://openid.net/connect/

## Create Dex namespace

Create a namespace where Dex will be installed:

```yaml
---
apiVersion: v1
kind: Namespace
metadata:
  name: dex
```

## Add credentials

There are a [lot of options][dex-connectors] available with Dex, in this guide we will
use the [GitHub connector][dex-github].

We can get a GitHub ClientID and Client secret by creating a
[new OAuth appliation][github-oauth].

![GitHub OAuth configuration](/img/guides/setting-up-dex/github-oauth-application.png)

```bash
kubectl create secret generic github-client \
  --namespace=dex \
  --from-literal=client-id=${GITHUB_CLIENT_ID} \
  --from-literal=client-secret=${GITHUB_CLIENT_SECRET}
```

[dex-connectors]: https://dexidp.io/docs/connectors/
[dex-github]: https://dexidp.io/docs/connectors/github/
[github-oauth]: https://docs.github.com/en/developers/apps/building-oauth-apps/creating-an-oauth-app

## Deploy Dex

As we did before, we can use `HelmRepository` and `HelmRelease` objects to let
Flux deploy everything.

```yaml
---
apiVersion: source.toolkit.fluxcd.io/v1beta1
kind: HelmRepository
metadata:
  name: dex
  namespace: dex
spec:
  interval: 1m
  url: https://charts.dexidp.io
---
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  name: dex
  namespace: dex
spec:
  interval: 5m
  chart:
    spec:
      chart: dex
      version: 0.6.5 
      sourceRef:
        kind: HelmRepository
        name: dex
        namespace: dex
      interval: 1m
  values:
    image:
      tag: v2.31.0
    envVars:
    - name: GITHUB_CLIENT_ID
      valueFrom:
        secretKeyRef:
          name: github-client
          key: client-id
    - name: GITHUB_CLIENT_SECRET
      valueFrom:
        secretKeyRef:
          name: github-client
          key: client-secret
    config:
      # Set it to a valid URL
      issuer: https://dex.dev.example.tld

      # See https://dexidp.io/docs/storage/ for more options
      storage:
        type: memory

      staticClients:
      - name: 'Weave GitOps Core'
        id: weave-gitops
        secret: AiAImuXKhoI5ApvKWF988txjZ+6rG3S7o6X5En
        redirectURIs:
        - 'https://localhost:9001/oauth2/callback'
        - 'https://0.0.0.0:9001/oauth2/callback'
        - 'http://0.0.0.0:9001/oauth2/callback'
        - 'http://localhost:4567/oauth2/callback'
        - 'https://localhost:4567/oauth2/callback'
        - 'http://localhost:3000/oauth2/callback'

      connectors:
      - type: github
        id: github
        name: GitHub
        config:
          clientID: $GITHUB_CLIENT_ID
          clientSecret: $GITHUB_CLIENT_SECRET
          redirectURI: https://dex.dev.example.tld/callback
          orgs:
          - name: weaveworks
            teams:
            - team-a
            - team-b
            - QA
          - name: ww-test-org
    ingress:
      enabled: true
      className: nginx
      annotations:
        cert-manager.io/cluster-issuer: letsencrypt-prod
      hosts:
        - host: dex.dev.example.tld
          paths:
          - path: /
            pathType: ImplementationSpecific
      tls:
        - hosts:
          - dex.dev.example.tld
          secretName: dex-dev-example-tld
```

:::note SSL certificate without cert manager
If we don't want to use cert manager, we can remove the related annotation and
use our predefined secret in the `tls` section.
:::

An important part of the configuration is the `orgs` field on the GitHub
connector.

```yaml
orgs:
- name: weaveworks
  teams:
  - team-a
  - team-b
  - QA
```

Here we can define groups under a GitHub organisation. In this example the
GitHub organisation is `weaveworks` and all members of the `team-a`,
`team-b`, and `QA` teams can authenticate. Group membership will be added to
the user.

Based on these groups, we can bind roles to groups:

```yaml
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: wego-test-user-read-resources
  namespace: flux-system
subjects:
  - kind: Group
    name: weaveworks:QA
    namespace: flux-system
roleRef:
  kind: Role
  name: wego-admin-role
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: wego-admin-role
  namespace: flux-system
rules:
  - apiGroups: [""]
    resources: ["secrets", "pods" ]
    verbs: [ "get", "list" ]
  - apiGroups: ["apps"]
    resources: [ "deployments", "replicasets"]
    verbs: [ "get", "list" ]
  - apiGroups: ["kustomize.toolkit.fluxcd.io"]
    resources: [ "kustomizations" ]
    verbs: [ "get", "list", "patch" ]
  - apiGroups: ["helm.toolkit.fluxcd.io"]
    resources: [ "helmreleases" ]
    verbs: [ "get", "list", "patch" ]
  - apiGroups: ["source.toolkit.fluxcd.io"]
    resources: ["buckets", "helmcharts", "gitrepositories", "helmrepositories", "ocirepositories"]
    verbs: ["get", "list", "patch"]
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["get", "watch", "list"]
```

The same way we can bind cluster roles to a group:

```yaml
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: weaveworks:team-a
subjects:
- kind: Group
  name: weaveworks:team-a
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: cluster-admin
  apiGroup: rbac.authorization.k8s.io
```

### Set up static user

For static user, add `staticPasswords` to the `config`:

```yaml
spec:
  values:
    config:
      staticPasswords:
      - email: "admin@example.tld"
        hash: "$2a$10$2b2cU8CPhOTaGrs1HRQuAueS7JTT5ZHsHSzYiFPm1leZck7Mc8T4W"
        username: "admin"
        userID: "08a8684b-db88-4b73-90a9-3cd1661f5466"
```

Static user password can be generated with `htpasswd`:

```bash
echo password | htpasswd -BinC 10 admin | cut -d: -f2
```
## OIDC login

Using the "Login with OIDC Provider" button:

![Login page](/img/guides/setting-up-dex/oidc-login.png)

We have to authorize the GitHub OAuth application:

![GitHub OAuth page](/img/guides/setting-up-dex/github-auth.png)

After that, grant access to Dex:

![Dex grant access](/img/guides/setting-up-dex/dex-auth.png)

Now we are logged in with our GitHub user and we can see all resources we have
access to:

![UI logged in](/img/guides/setting-up-dex/ui-logged-in.png)
