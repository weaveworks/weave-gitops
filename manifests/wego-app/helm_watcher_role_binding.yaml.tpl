apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: wego-helm-watcher-rolebinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: wego-helm-watcher-role
subjects:
  - kind: ServiceAccount
    name: wego-app-service-account
    namespace: {{ .Namespace }}
