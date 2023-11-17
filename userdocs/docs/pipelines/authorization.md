---
title: Authorization

---


# Authorization ~ENTERPRISE~

!!! warning
    **This feature is in alpha and certain aspects will change**
    We're very excited for people to use this feature.
    However, please note that changes in the API, behaviour and security will evolve.
    The feature is suitable to use in controlled testing environments.

To view pipelines, users need read access to the `pipeline` resource and the underlying `application` resources. This sample configuration shows a recommended way to configure RBAC to provide such access. The `pipeline-reader` role and the `search-pipeline-reader`
role-binding allow a group `search-developer` to access pipeline resources within the `search` namespace.

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: pipeline-reader
rules:
  - apiGroups: [ "pipelines.weave.works" ]
    resources: [ "pipelines" ]
    verbs: [ "get", "list", "watch"]
  - apiGroups: ["helm.toolkit.fluxcd.io"]
    resources: [ "helmreleases" ]
    verbs: [ "get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: search-pipeline-reader
  namespace: search
subjects:
  - kind: Group
    name: search-developer
    apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: pipeline-reader
  apiGroup: rbac.authorization.k8s.io
```
