apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: read-profiles
subjects:
  - kind: ServiceAccount
    name: profiles-server-service-account
    namespace: {{.Namespace}}
roleRef:
  kind: Role
  name: profiles-reader
  apiGroup: rbac.authorization.k8s.io
