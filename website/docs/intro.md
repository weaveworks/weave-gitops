---
title: Introduction
sidebar_position: 0
hide_title: true
---
# Weave GitOps

Weave GitOps is a powerful extension to [Flux](https://fluxcd.io), a leading GitOps engine and CNCF project, which provides insights into your application deployments, and makes continuous delivery with GitOps easier to adopt and scale across your teams.

The web UI surfaces key information to help application operators easily discover and resolve issues. The intuitive interface provides a guided experience to build understanding and simplify getting started for new users; they can easily discover the relationship between Flux objects and navigate to deeper levels of information as required.

Weave GitOps is an open source project sponsored by [Weaveworks](https://weave.works) - the GitOps company, and original creators of [Flux](https://fluxcd.io).

## Why adopt GitOps?
> "GitOps is the best thing since configuration as code. Git changed how we collaborate, but declarative configuration is the key to dealing with infrastructure at scale, and sets the stage for the next generation of management tools"

<cite>- Kelsey Hightower, Staff Developer Advocate, Google.</cite><br/><br/>

Adopting GitOps can bring a number of key benefits:
- Faster and more frequent deployments
- Easy recovery from failures
- Improved security and auditability

To learn more about GitOps, check out these resources:

- [GitOps for absolute beginners](https://go.weave.works/WebContent-EB-GitOps-for-Beginners.html) - eBook from Weaveworks
- [Guide to GitOps](https://www.weave.works/technologies/gitops/) - from Weaveworks
- [OpenGitOps](https://opengitops.dev/) - CNCF Sandbox project aiming to define a vendor-neutral, principle-led meaning of GitOps.
- [gitops.tech](https://www.gitops.tech/) - supported by Innoq

## Getting Started

See [Installation](/docs/installation) and [Getting Started](/docs/getting-started)

## Features

- **Applications view** - allows you to quickly understand the state of your deployments across a cluster at a glance. It shows summary information from `kustomization` and `helmrelease` objects. 
- **Sources view** - shows the status of resources which are synchronizing content from where you have declared the desired state of your system, for example Git repositories. This shows summary information from `gitrepository`, `helmrepository` and `bucket` objects.
- **Flux Runtime view** - provides status on the GitOps engine continuously reconciling your desired and live state. It shows your installed GitOps Toolkit Controllers and their version.
- Drill down into more detailed information on any given Flux resource.
- Uncover relationships between resources and quickly navigate between them.
- Understand how workloads are reconciled through a directional graph.
- View Kubernetes events relating to a given object to understand issues and changes.
- Secure access to the dashboard through the ability to integrate with an OIDC provider (such as Dex) or through a configurable cluster user.
- Fully integrates with [Flux](https://fluxcd.io/docs/) as the GitOps engine to provide:
  - Continuous Delivery through GitOps for apps and infrastructure
  - Support for GitHub, GitLab, Bitbucket, and even use s3-compatible buckets as a source; all major container registries; and all CI workflow providers.
  - A secure, pull-based mechanism, operating with least amount of privileges, and adhering to Kubernetes security policies.
  - Compatible with any conformant [Kubernetes version](https://fluxcd.io/docs/installation/#prerequisites) and common ecosystem technologies such as Helm, Kustomize, RBAC, Prometheus, OPA, Kyverno, etc.
  - Multitenancy, multiple git repositories, multiple clusters
  - Alerts and notifications


