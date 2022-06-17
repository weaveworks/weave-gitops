# Helm chart reference
<!-- The contents of this file is generated directly from the chart's values.yaml, please make any edits there -->

This is a reference of all the configurable values in weave gitops's
helm chart. This is intended for customizing your installation after
you've gone through the [getting started](../getting-started.mdx) guide.

This reference was generated for the chart version 2.0.10 which installs weave gitops v0.8.1.

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| adminUser.create | bool | `false` | Whether the local admin user should be created. If you use this make sure you add it to `rbac.impersonationResourceNames`. |
| adminUser.createClusterRole | bool | `true` | Specifies whether the clusterRole & binding to the admin user should be created. Will be created only if `adminUser.create` is enabled. Without this, the adminUser will only be able to see resources in the `flux-system` namespace. |
| adminUser.createSecret | bool | `true` | Whether we should create the secret for the local adminUser. Will be created only if `adminUser.create` is enabled. Without this, we'll still set up the roles and permissions, but the secret with username and password has to be provided separately. |
| adminUser.passwordHash | string | `nil` | Set the password for local admin user. Requires `adminUser.create` and `adminUser.createSecret` This needs to have been hashed using the bcrypt algorithm. E.g. `go install github.com/bitnami/bcrypt-cli@v1.0.2 && bcrypt-cli <<< $PASSWORD` |
| adminUser.username | string | `"gitops-test-user"` | Set username for local admin user, these will be stored in a secret in k8s. Requires `adminUser.create` and `adminUser.createSecret`. |
| affinity | object | `{}` |  |
| fullnameOverride | string | `""` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.repository | string | `"ghcr.io/weaveworks/wego-app"` |  |
| image.tag | string | `"v0.8.1"` |  |
| imagePullSecrets | list | `[]` |  |
| ingress.annotations | object | `{}` |  |
| ingress.className | string | `""` |  |
| ingress.enabled | bool | `false` |  |
| ingress.hosts | string | `nil` |  |
| ingress.tls | list | `[]` |  |
| logLevel | string | `"info"` | What log level to output. Valid levels are 'debug', 'info', 'warn' and 'error' |
| nameOverride | string | `""` |  |
| nodeSelector | object | `{}` |  |
| podAnnotations | object | `{}` |  |
| podSecurityContext | object | `{}` |  |
| rbac.additionalRules | list | `[]` | If non-empty, these additional rules will be appended to the RBAC role and the cluster role. for example, additionalRules: - apiGroups: ["infra.contrib.fluxcd.io"]   resources: ["terraforms"]   verbs: [ "get", "list", "patch" ] |
| rbac.create | bool | `true` | Specifies whether the clusterRole & binding to the service account should be created |
| rbac.impersonationResourceNames | list | `[]` | If non-empty, this limits the resources that the service account can impersonate. This applies to both users and groups, e.g. `['user1@corporation.com', 'user2@corporation.com', 'operations']` |
| rbac.impersonationResources | list | `["users","groups"]` | Limit the type of principal that can be impersonated |
| rbac.viewSecretsResourceNames | list | `["cluster-user-auth","oidc-auth"]` | If non-empty, this limits the secrets that can be accessed by the service account to the specified ones, e.g. `['weave-gitops-enterprise-credentials']` |
| replicaCount | int | `1` |  |
| resources | object | `{}` |  |
| securityContext | object | `{}` |  |
| serverTLS.enable | bool | `false` | Enable TLS termination in gitops itself. If you enable this, you need to create a secret, and specify the secretName. Another option is to create an ingress. |
| serverTLS.secretName | string | `"my-secret-tls"` | Specify the tls secret name. This type of secrets have a key called `tls.crt` and `tls.key` containing their corresponding values in  base64 format. See https://kubernetes.io/docs/concepts/configuration/secret/#tls-secrets for more details and examples |
| service.annotations | object | `{}` |  |
| service.create | bool | `true` |  |
| service.port | int | `9001` |  |
| service.type | string | `"ClusterIP"` |  |
| serviceAccount.annotations | object | `{}` | Annotations to add to the service account |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `""` | The name of the service account to use. If not set and create is true, a name is generated using the fullname template |
| tolerations | list | `[]` |  |
