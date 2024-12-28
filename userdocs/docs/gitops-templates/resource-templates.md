---
title: Resource Templates

---



# Resource templates ~ENTERPRISE~

Resource templates are used to create Kubernetes resources. They are defined in the `spec.resourcetemplates` section of the template.

### The `content` key

The `content` key is used to define a list of resources:

```yaml
spec:
  resourcetemplates:
    - content:
        - apiVersion: v1
          kind: Namespace
          metadata:
            name: nginx
        - apiVersion: v1
          kind: Namespace
          metadata:
            name: cert-manager
```

### The `raw` key

The `raw` key is used to define a raw string that will written to the specified path.

This can be useful to preserve comments or formatting in the rendered resource.

```yaml
spec:
  resourcetemplates:
    - path: "helm-release.yaml"
      raw: |
        apiVersion: helm.toolkit.fluxcd.io/v2beta1
        kind: HelmRelease
        metadata:
          name: podinfo
          namespace: prod-github
        spec:
          interval: 1m
          chart:
            spec:
              chart: podinfo
              version: "6.0.0" # {"$promotion": "flux-system:podinfo-github:prod"}
              sourceRef:
                kind: HelmRepository
                name: podinfo
              interval: 1m
```

!!! info
    - The `raw` key is not compatible with the `content` key. Only one of the two can be used.
    - The `raw` key data must still be a valid kubernetes unstructured object.
