apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: profiles-reader
rules:
  - apiGroups: [".source.toolkit.fluxcd.io"]
    resources: [ "helmrepositories" ]
    verbs: [ "get","list" ]
