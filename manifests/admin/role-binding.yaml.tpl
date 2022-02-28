apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: user-read-resources
  namespace: {{.Namespace}}
subjects:
  - kind: User
    name: admin
    namespace: {{.Namespace}}
roleRef:
  kind: Role
  name: resources-reader
  apiGroup: rbac.authorization.k8s.io
