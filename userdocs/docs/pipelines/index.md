---
title: Introduction

---


# Pipelines ~ENTERPRISE~

!!! warning
    **This feature is in alpha and certain aspects will change**
    We're very excited for people to use this feature.
    However, please note that changes in the API, behaviour and security will evolve.
    The feature is suitable to use in controlled testing environments.

Weave GitOps Enterprise Pipelines enables teams to increase the velocity, stability, and security of software systems via automated deployment pipelines. It provides insights into new application versions that are being rolled out across clusters and environments, which allows you to implement security guardrails and track metrics to assess if the application is working as desired. In instances of failures, the change is abandoned with an automatic rollout of the older version.

With Pipelines, you define a release pipeline for a given application as a custom resource. The pipeline can comprise any number of environments through which an application is expected to be deployed. Push a change to your application in your dev environment, for example, and watch the update roll out across staging and production environments all from a single PR (or an external process like Jenkins)—with Weave GitOps Enterprise orchestrating everything. 

Designed with flexibility in mind, Pipelines can be easily integrated within your existing CI setup—for example, CircleCI, Jenkins, Tekton, or GitHub Actions.

## Benefits to Developers

The Pipelines feature:
- reduces toil and errors when setting up a new pipeline or reproducing previous pipelines through YAML constructs
- saves time and overhead with automated code rollout from one environment to another, with minimal intervention from the Ops team
- enables users to observe code progression and track application versions through different environments from the Weave GitOps UI
- streamlines code deployment from one environment to another, and minimizes friction between application development and Ops teams
- enables you to easily define which Helm charts are part of the environments you create—saving lots of time through automated package management

Now that you know what delivery pipelines can do for you, follow the [guide to get started](../pipelines-getting-started).
