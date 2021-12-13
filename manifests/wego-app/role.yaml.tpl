apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: resources-reader
  namespace: {{.Namespace}}
rules:
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get","create"]
  - apiGroups: ["wego.weave.works"]
    resources: [ "apps" ]
    verbs: [ "get","create","list","delete","watch" ]
  - apiGroups: ["kustomize.toolkit.fluxcd.io"]
    resources: [ "kustomizations" ]
    verbs: [ "get","create","delete","list" ]
  - apiGroups: ["helm.toolkit.fluxcd.io"]
    resources: [ "helmreleases" ]
    verbs: [ "get","create","delete","list" ]
  - apiGroups: ["source.toolkit.fluxcd.io"]
    resources: [ "helmrepositories" ]
    verbs: [ "get","create","delete","list" ]
  - apiGroups: ["source.toolkit.fluxcd.io"]
    resources: [ "gitrepositories" ]
    verbs: [ "get","create","delete","list" ]
