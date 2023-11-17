---
title: Getting Started

---

# Getting Started ~ENTERPRISE~
Enabling the Weave Policy Engine features in Weave GitOps is done by running the policy agent on the cluster. This section gives an overview of the policy ecosystem and the steps required for installing and running the policy agent on leaf clusters.

## The Policy Ecosystem

The policy ecosystem consists of several moving parts. The two primary components are the [Policy Agent](./weave-policy-profile.md#policy-agent-configuration) and the [Policy CRs](./policy.md). The agent runs in several [modes](./weave-policy-profile.md#agent-modes), and uses the Policy CRs to perform validations on different resources. The results of those validations can be written to different [sinks](./weave-policy-profile.md#policy-validation-sinks).

There are two other optional components: the [PolicySet](./policy-set.md), and the [PolicyConfig](./policy-configuration.md). The PolicySet can be used to filter policies for a specific mode, while the PolicyConfig can be used to override policy parameters during the validation of a certain resource.

![Policy Ecosystem](../img/policy-ecosystem.png)

## Installation Pre-requisites

### Weave GitOps
You need to have a running instance of Weave GitOps with at least one CAPI provider installed to provision Kubernetes clusters. See [Weave GitOps Installation](https://docs.gitops.weave.works/docs/installation/) page for more details about installing Weave GitOps.

### Policy Library
For the policy agent to work, it will need a source for the [policies](./policy.md) that it will enforce in the cluster. Enterprise customers should request access to fork our policy library into their local repositories. Our policy library includes an extensive list of policy CRs that cover a multitude of security and compliance benchmarks.

## Install the Policy Agent

To install the policy agent on a leaf cluster, you should select the `weave-policy-agent` from the profiles dropdown in the `Create Cluster` page.

![Policy Profile](../img/weave-policy-profile.png)

You should then configure the `values.yaml` to pull the policies from your repo into the cluster. This is done by configuring the `policySource` section. If your policy library repo is private, you will also need to reference the `Secret` that contains the repo credentials. This is usually the secret you created while bootstrapping Flux on the management cluster and is copied to your leaf cluster during creation.

??? example "Expand to see an example that creates a new git source"

    ```yaml
    policySource:
    enabled: true
    url: ssh://git@github.com/weaveworks/policy-library # This should be the url of the forked repo
    tag: v1.0.0
    path: ./  # Could be a path to the policies dir or a kustomization.yaml file
    secretRef: my-pat # the name of the secret containing the repo credentials
    ```

??? example "Expand to see an example that uses an existing git source"

    ```yaml
    policySource:
    enabled: true
    sourceRef: # Specify the name for an existing GitSource reference
        kind: GitRepository
        name: policy-library
        namespace: flux-system
    ```

You can find more about other policy profile configurations [here](../weave-policy-profile/).

## Policies in UI
After the leaf cluster is provisioned and the profile is installed, you should now see the policies listed in the Policies tab in Weave GitOps UI.

![Policies](../img/weave-policies.png)

Now you have a provisioned cluster with these policies enforced by the policy agent.

> By default, the policy profile is set up to enforce policies at deployment time using admission controller, which results in blocking any deployment that violates the enforced policies.

## Prevent Violating Changes
Now let's try to deploy a Kubernetes deployment that violates the `Container Image Pull Policy` which is one of the enforced policies.
This policy is violated when the container's `imagePullPolicy` is not set to `Always`.

??? example "Expand for an example of a violating deployment"

    ```yaml
    apiVersion: apps/v1
    kind: Deployment
    metadata:
    name: nginx-deployment
    labels:
        app: nginx
    spec:
    replicas: 3
    selector:
        matchLabels:
        app: nginx
    template:
        metadata:
        labels:
            app: nginx
        spec:
        containers:
        - name: nginx
            image: nginx:1.14.2
            imagePullPolicy: IfNotPresent
            ports:
            - containerPort: 80
    ```

Once you apply it, the policy agent will deny this request and show a violation message, and accordingly the deployment will not be created.

## Violations Logs in UI
You can go to the `Violations Log` in Weave GitOps UI to view the policy violations of all the connected clusters, and dive into the details of each violation.

This view shows only the violations resulting from the [admission](./weave-policy-profile.md#admission) mode by configuring the [events sink](weave-policy-profile.md#policy-validation-sinks).

<strong>Violations Log</strong>

![Violations Logs](../img/violations-logs.png)

<strong>Violations Log Details</strong>

![Violation Log Details](../img/violations-log-detail.png)
