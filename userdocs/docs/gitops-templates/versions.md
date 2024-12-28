---
title: Version Information

---



# Version Information ~ENTERPRISE~

There are now multiple published versions of the template CRD.

## Migration notes

### `v1alpha1` to `v1alpha2`

When manually migrating a template from `v1alpha1` to `v1alpha2` (for example in git) you will need to:
1. Update the `apiVersion`:
    1. for `GitopsTemplate` update the apiVersion to `templates.weave.works/v1alpha2`
    1. for `CAPITemplate` update the apiVersion to `capi.weave.works/v1alpha2`
1. Move the `spec.resourcetemplates` field to `spec.resourcetemplates[0].content`
1. Either leave the `spec.resourcetemplates[0].path` field empty or give it a sensible value.

If you experience issues with the path not being recognised when Flux reconciles
the new template versions, try manually applying the new template to the cluster directly with:
1. Run `kubectl apply -f capi-template.yaml`
1. Run `flux reconcile kustomization --with-source flux-system` **twice**.

## Conversion Webhook

As of Weave Gitops Enterprise 0.28.0 the conversion webhook has been removed.

This removed the need for cert-manager to be installed, but you will now have to convert any `v1alpha1` templates to `v1alpha2` manually in git.

## `v1alpha2` (default) notes

This version changes the type of `spec.resourcetemplates` from a list of objects to a list of files with a `path` and `content`:

Example:
```yaml
spec:
  resourcetemplates:
    - path: "clusters/{{ .params.CLUSTER_NAME }}.yaml"
      content:
        - apiVersion: cluster.x-k8s.io/v1alpha3
          kind: Cluster
          metadata:
            name: "{{ .params.CLUSTER_NAME }}"
          path: "clusters/{{ .params.CLUSTER_NAME }}.yaml"
```

## `v1alpha1` notes

The original version of the template. This version no longer works with Weave Gitops Enterprise 0.28.0 and above.

It uses `spec.resourcetemplates` as a list of resources to render.

Example:
```yaml
spec:
  resourcetemplates:
    - apiVersion: cluster.x-k8s.io/v1alpha3
      kind: Cluster
      metadata:
        name: "{{ .params.CLUSTER_NAME }}"
```
