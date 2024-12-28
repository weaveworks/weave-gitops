---
title: Quickstart

---

# Quickstart GitOps Templates ~ENTERPRISE~

`Quickstart` templates are [`GitOpsTemplate`s](https://docs.gitops.weave.works/docs/gitops-templates/templates/)
that you could use when getting started with [Weave Gitops Enterprise](../enterprise/index.md)
It aims to provide a simplified basic experience.

## Getting Started

The templates exist as a Helm Chart in the [weave-gitops-quickstart](https://github.com/weaveworks/weave-gitops-quickstart)
github repo.

To get started, add the following `HelmRelease` object to your Weave GitOps Enterprise
configuration repo for your management cluster.

??? example "Expand to view"

    ```yaml
    ---
    apiVersion: source.toolkit.fluxcd.io/v1beta2
    kind: GitRepository
    metadata:
    name: weave-gitops-quickstart
    namespace: flux-system
    spec:
    interval: 10m0s
    ref:
        branch: main
    url: https://github.com/weaveworks/weave-gitops-quickstart
    ---
    apiVersion: helm.toolkit.fluxcd.io/v2beta1
    kind: HelmRelease
    metadata:
    name: quickstart-templates
    namespace: flux-system
    spec:
    chart:
        spec:
        chart: "quickstart-templates"
        version: ">=0.1.0"
        sourceRef:
            kind: GitRepository
            name: weave-gitops-quickstart
            namespace: flux-system
    interval: 10m0s
    ```

Commit and merge the above file. Once the `HelmRelease` has been successfully
deployed to your cluster, navigate to your Weave GitOps UI Dashboard. You will
see that the `templates` Chart is now deployed to your cluster.

![quickstart templates deployed](../img/quickstart-templates-deployed.png)

If you click on the `Templates` tab in the sidebar, you will see the Quickstart
templates are now available for use:

![quickstart templates view](../img/quickstart-templates-view.png)

## Available Templates

The following [pipeline](../pipelines/pipelines-templates.md) templates have
been made available on your Weave GitOps Enterprise instance:

- `pipeline-view`: A template to create a sample pipeline to visualize a
	`HelmRelease` application delivered to dev, test and prod environments.
- `pipeline-promotion-resources`: A template to create the Flux Notification
	Controller resources required for promoting applications via pipelines.
- `pipeline-view-promote-by-cluster`: A template to create pipelines for hard
	tenancy when applications are isolated by cluster.
- `pipeline-view-promote-by-namespace`: A template to create pipelines for soft
	tenancy when applications are isolated by namespace.

## Using `GitOpsTemplate`s as a Platform Engineer

The above Quickstart templates are designed to provide a practical getting started
experience. We encourage Platform Operators to start off with these templates
within their team to ramp up on using Weave GitOps.

If the need arises later, operators can always expand on these templates to
develop their own set of self-service capabilities.

## Using `GitOpsTemplate`s as an Application Developer

As a developer using Weave GitOps Enterprise, use the templates to explore
GitOps's capabilities. For example, to create a pipeline for your application:
use the above template provided by your Operations team to create required
resources. Once they have been added to your GitOps repository, you can adapt
the rendered resources to meet your needs.

!!! tip "Want to contribute?"
    The Quickstart templates are maintained by the Weave Gitops team. If you would
    like to make alterations, suggest fixes, or even contribute a new template which
    you find cool, just head to the [repo](https://github.com/weaveworks/weave-gitops-quickstart)
    and open a new issue or PR!
