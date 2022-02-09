apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: wego-helm-watcher-role
rules:
  - apiGroups:
      - source.toolkit.fluxcd.io
    resources:
      - helmrepositories
    verbs:
      - get
      - list
      - watch
      - patch
      - update
  - apiGroups:
      - source.toolkit.fluxcd.io
    resources:
      - helmrepositories/finalizers
    verbs:
      - create
      - delete
      - get
      - patch
      - update
  - apiGroups:
      - source.toolkit.fluxcd.io
    resources:
      - helmrepositories/status
    verbs:
      - get
