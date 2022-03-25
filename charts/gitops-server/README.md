# Weave Gitops Helm Chart

This is the [Weave Gitops](https://github.com/weaveworks/weave-gitops) [Helm](https://helm.sh) chart.

It installs the weave-gitops-server component as a 1-replica deployment.

Optionally it will also install:
* Service Account
* ClusterRoleBinding (to the service account) and ClusterRole with the
  permissions required to run Gitops.
* Service, this is optional as you may want to limit access to the UI to via
  port-forwarding
* Ingress
* Test User -- A test user with hard-coded username & password with minimal
  permissions

This chart assumes kubernetes > 1.17

## Security

The role that this chart creates includes 2 main 'blocks' of permissions; that
should be treated separately and carefully:

* `impersonate` This is how the gitops-server gathers data to display in the UI,
  it impersonates the user, determined by OIDC/plain auth. This means that
  a user's permissions in the UI will reflect their permissions in the cluster
* `get`, `list`, `watch` on `helmrepositories` and `secrets`. These permissions
  are required by the profiles system.

### Impersonate

When deploying gitops-server it is recommended to limit the types of resource
and specific resources that the service account can impersonate. e.g.
```yaml
rbac:
  create: true
  impersonationResources: ["groups"]
  impersonationResourceNames: ["gitops-reader"]
```

Using groups is the recommended way of doing this as it means that you don't
have to enumerate all users in a group.

### Get helmrepositories

This permissions are scoped to enable the profiles functionality of gitops-server
and should not need to change.

### Test User

This user should not be used, it is intended for development and testing
purposes and relies on static credentials in a secret.
