---
title: Getting Started

---


# Getting Started with Pipelines ~ENTERPRISE~

!!! warning
    **This feature is in alpha and certain aspects will change**
    We're very excited for people to use this feature.
    However, please note that changes in the API, behaviour and security will evolve.
    The feature is suitable to use in controlled testing environments.

## Prerequisites

Before using Pipelines, please ensure that:
- You have Weave GitOps Enterprise installed on a cluster.
- You have configured Weave GitOps Enterprise [RBAC for Pipelines](../authorization).
- The Pipelines feature flag `enablePipelines` has been enabled. This flag is part of the [Weave GitOps Enterprise Helm chart values](https://docs.gitops.weave.works/docs/references/helm-reference/) and is enabled by default.
- Any leaf clusters running workloads that you need to visualise using Pipelines have been added to Weave GitOps Enterprise.
- You have [exposed the promotion webhook ](../promoting-applications/#expose-the-promotion-webhook) on the management cluster and leaf clusters can reach that webhook endpoint over the network.

## Define a Pipeline

A pipeline allows you to define the route your application is taking, so that you can get it to production. Three main concepts are at play:
- the `application` to deliver
- the `environments` that your app will go through on its way to production (general). An environment describes the different stages of a pipeline and consists of one or more deployment targets.
- the `deployment targets`, the clusters that each environment has. A deployment target consists of a namespace and a [`GitOpsCluster` reference](../cluster-management/managing-clusters-without-capi.md) and is used to specify where the application is running in your fleet. 

You can define a delivery pipeline using a `Pipeline` custom resource. An example of such a CR is shown here:

??? example "Expand to view"

    ```yaml
    ---
    apiVersion: pipelines.weave.works/v1alpha1
    kind: Pipeline
    metadata:
    name: podinfo-02
    namespace: flux-system
    spec:
    appRef:
        apiVersion: helm.toolkit.fluxcd.io/v2beta1
        kind: HelmRelease
        name: podinfo
    environments:
        - name: dev
        targets:
            - namespace: podinfo-02-dev
            clusterRef:
                kind: GitopsCluster
                name: dev
                namespace: flux-system
        - name: test
        targets:
            - namespace: podinfo-02-qa
            clusterRef:
                kind: GitopsCluster
                name: dev
                namespace: flux-system
            - namespace: podinfo-02-perf
            clusterRef:
                kind: GitopsCluster
                name: dev
                namespace: flux-system
        - name: prod
        targets:
            - namespace: podinfo-02-prod
            clusterRef:
                kind: GitopsCluster
                name: prod
                namespace: flux-system
    ```


In the example above, the `podinfo` application is delivered to a traditional pipeline composed of `dev`, `test`, and `prod` environments. In this case, the `test` environment consists of two deployment targets, `qa` and `perf`. This is to indicate that, although both targets are part of the same stage (testing), they can evolve separately and may run different versions of the application. Note that two clusters, `dev` and `prod`, are used for the environments; both 
are defined in the `flux-system` namespace.

For more details about the spec of a pipeline, [go here](spec/v1alpha1/pipeline.md).

## View Your List of Pipelines

Once Flux has reconciled your pipeline, you can navigate to the Pipelines view in the WGE UI to see the list of pipelines to which you have access.

![view pipelines](../img/view-pipelines.png)

For each pipeline, the WGE UI shows a simplified view with the application `Type` and `Environments` it goes through.

## View Pipeline Details

Once you have selected a pipeline from the list, navigate to its details view where you can see the current status of your application by environment and deployment target.

![view pipeline details](../img/view-pipeline-details.png)

