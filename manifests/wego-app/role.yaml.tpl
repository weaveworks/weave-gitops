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
    verbs: [ "*" ]
  - apiGroups: ["kustomize.toolkit.fluxcd.io"]
    resources: [ "kustomizations" ]
    verbs: [ "*" ]
  - apiGroups: ["helm.toolkit.fluxcd.io"]
    resources: [ "helmreleases" ]
    verbs: [ "*" ]
  - apiGroups: ["source.toolkit.fluxcd.io"]
    resources: [ "helmrepositories" ]
    verbs: [ "*" ]
  - apiGroups: ["source.toolkit.fluxcd.io"]
    resources: [ "gitrepositories" ]
    verbs: [ "*" ]
