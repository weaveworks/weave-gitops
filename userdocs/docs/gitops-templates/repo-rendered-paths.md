---
title: Rendered Template Paths

---



# Rendered Template Paths ~ENTERPRISE~

Template authors can configure the eventual locatation of the rendered template
in the user's GitOps repository.

This allows for more control over where different resources in the template are rendered.

## Configuring Paths

The path for rendered resources is configured via the
`spec.resourcetemplates[].path` field.

!!! note "Important to note"
    - The path is relative to the repository root
    - The path can be templated using params

??? example "Expand to see example"

    ```yaml
    spec:
    resourcetemplates:
        // highlight-next-line
        - path: clusters/${CLUSTER_NAME}/definition/cluster.yaml
        content:
            - apiVersion: cluster.x-k8s.io/v1alpha4
            kind: Cluster
            metadata:
                name: ${CLUSTER_NAME}
            ...
            - apiVersion: infrastructure.cluster.x-k8s.io/v1alpha4
            kind: AWSCluster
            metadata:
                name: ${CLUSTER_NAME}
            ...
        // highlight-next-line
        - path: clusters/${CLUSTER_NAME}/workloads/helmreleases.yaml
        content:
            - apiVersion: helm.toolkit.fluxcd.io/v2beta1
            kind: HelmRelease
            metadata:
                name: ${CLUSTER_NAME}-nginx
            ...
            - apiVersion: helm.toolkit.fluxcd.io/v2beta1
            kind: HelmRelease
            metadata:
                name: ${CLUSTER_NAME}-cert-manager
            ...
    ```
    
### Configuring paths for `charts`

The `spec.charts.helmRepositoryTemplate.path` and `spec.charts.items[].template.path` fields can be used to specify the paths of these resources:

Example

```yaml
spec:
  charts:
    helmRepositoryTemplate:
      // highlight-next-line
      path: clusters/${CLUSTER_NAME}/workloads/helm-repo.yaml
    items:
      - chart: cert-manager
        version: 0.0.8
        template:
          // highlight-next-line
          path: clusters/${CLUSTER_NAME}/workloads/cert-manager.yaml
```


## Default Paths

If the `spec.resourcetemplates[].path` is omitted, a default path for the
rendered template is calculated.

In this case some of the submitted params are used. Users **must** provide one of the following parameters:
- `CLUSTER_NAME`
- `RESOURCE_NAME`

To ensure users supply these values, set the parameters to `required` in the the
template definition:

```yaml
spec:
  params:
    - name: RESOURCE_NAME
      required: true
    # or
    - name: CLUSTER_NAME
      required: true
```

!!! warning "Important"
    The **kustomization** feature and the `add-common-bases` annotation feature **always** use a calculated default path.
    If you are using these features one of `CLUSTER_NAME` or `RESOURCE_NAME`
    **must** be provided, even if you specify a `path` for all the other resources in the template.

The default path for a template has a few components:
- From the params: `CLUSTER_NAME` or `RESOURCE_NAME`, **required**.
- From the params: `NAMESPACE`, default: `default`
- From `values.yaml` for the Weave GitOps Enterprise `mccp` chart: `values.config.capi.repositoryPath`, default: `clusters/management/clusters`

These are composed to create the path:
`${repositoryPath}/${NAMESPACE}/${CLUSTER_OR_RESOURCE_NAME}.yaml`

Using the default values and supplying `CLUSTER_NAME` as `my-cluster` will result in the path:
`clusters/management/clusters/default/my-cluster.yaml`
