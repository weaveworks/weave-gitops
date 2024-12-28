---
title: Displaying Custom Metadata
---
# Displaying Custom Metadata

Weave GitOps lets you add annotations with custom metadata to your
Flux automations and sources, and they will be displayed in the main UI.

For example, you might use this to add links to dashboards, issue
systems, or documentation and comments that you wish to be directly visible in
the GitOps UI.

We will use the `podinfo` application that we installed in the [getting
started guide](../open-source/deploy-oss.md) as an example. Open up the
podinfo kustomization and add annotations to it so it looks like this:

```yaml title="./clusters/my-cluster/podinfo-kustomization.yaml"
---
apiVersion: kustomize.toolkit.fluxcd.io/v1beta2
kind: Kustomization
metadata:
  name: podinfo
  namespace: flux-system
// highlight-start
  annotations:
    metadata.weave.works/description: |
      Podinfo is a tiny web application made with Go that showcases best practices of running microservices in Kubernetes.
      Podinfo is used by CNCF projects like Flux and Flagger for end-to-end testing and workshops.
    metadata.weave.works/grafana-dashboard: https://grafana.my-org.example.com/d/podinfo-dashboard
// highlight-end
spec:
  interval: 5m0s
  path: ./kustomize
  prune: true
  sourceRef:
    kind: GitRepository
    name: podinfo
  targetNamespace: flux-system
```

Close the file and commit and push your changes.

Back in your GitOps dashboard, navigate to the 'Applications' tab and select the
`podinfo` kustomization. At the bottom of the 'Details' section you will see the
new 'Metadata' entries:

![Application detail view showing custom metadata](/img/metadata-display.png)

!!! warning "Restrictions"
    * The annotation key **must** start with the domain
    `metadata.weave.works`. Any other annotations will be ignored.
    * The key that will be displayed is whatever you put after the
    domain, title cased, and with dashes replaced with spaces. Above,
    `metadata.weave.works/grafana-dashboard` was displayed as "Grafana Dashboard".
    * The value can either be a link, or can be plain text. Newlines in
    plain text will be respected.
    * The key is subject to certain limitations that kubernetes imposes on
    annotations, including:
        - it must be shorter than 63 characters (not including
    the domain)
        - it must be an English alphanumeric character, or one of `-._`.
        - See the [kubernetes documentation](https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/#syntax-and-character-set)
    for the full list of restrictions.
