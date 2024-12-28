---
title: Introducing Weave GitOps

---

# Introducing Weave GitOps

> "GitOps is the best thing since configuration as code. Git changed how we collaborate, but declarative configuration is the key to dealing with infrastructure at scale, and sets the stage for the next generation of management tools"

<cite>- Kelsey Hightower, Staff Developer Advocate, Google.</cite>

Weave GitOps improves developer experience—simplifying the complexities and cognitive load of deploying and managing cloud native apps on Kubernetes so that teams can go faster. It’s a powerful extension of Flux, a leading GitOps engine and [Cloud Native Computing Foundation project](https://www.cncf.io/projects/). [Weaveworks](https://weave.works) are the creators of [Flux][flux].

Weave GitOps’ intuitive user interface surfaces key information to help application operators easily discover and resolve issues—simplifying and scaling adoption of GitOps and continuous delivery. The UI provides a guided experience that helps users to easily discover the relationships between Flux objects and build understanding while providing insights into application deployments.

Today Weave GitOps defaults are Flux, Kustomize, Helm, SOPS, and Kubernetes Cluster API. If you use Flux already, then you can easily add Weave GitOps to create a platform management overlay.

!!! tip
    Adopting GitOps can bring a number of key benefits—including faster and more frequent deployments, easy recovery from failures, and improved security and auditabiity. Check out our [GitOps for Absolute Beginners](https://go.weave.works/WebContent-EB-GitOps-for-Beginners.html) eBook and [Guide to GitOps](https://www.weave.works/technologies/gitops/) for more information.

## Getting Started

This user guide provides content that will help you to install and get started with our free and paid offerings:
- **Weave GitOps Open Source**: a simple, open source developer platform for people who don't have Kubernetes expertise but who want cloud native applications. It includes the UI and many other features that take your team beyond a simple CI/CD system. Experience how easy it is to enable GitOps and run your apps in a cluster. [Go here to install](./open-source/install-oss.md).
- **Weave GitOps Enterprise**: an [enterprise version](./enterprise/index.md) that adds automation and 100% verifiable trust to existing developer platforms, enabling faster and more frequent deployments with guardrails and golden paths for every app team. Note that Enterprise offers a more robust UI than what you'll find in our open source version. [Go here to install](./enterprise/install-enterprise.md).

!!! tip
    Want to learn more about how [Weave GitOps Enterprise](./enterprise/index.md) can help your team?
    Get in touch with sales@weave.works to discuss your needs.

Weave GitOps works on any Chromium-based browser (Chrome, Opera, Microsoft Edge), Safari, and Firefox. We only support the latest and prior two versions of these browsers.

To give Weave GitOps a test drive, we recommend checking out the Open Source version and its [UI](./open-source/ui-oss.md), then deploying an application. Let's take a closer look at the features it offers you, all for free.

### Weave GitOps Open Source Features

Like our Enterprise version, Weave GitOps Open Source fully integrates with [Flux](https://fluxcd.io/docs/) as the GitOps engine to provide:

- :infinity: Continuous Delivery through GitOps for apps and infrastructure.
- :jigsaw: Support for GitHub, GitLab, and Bitbucket; S3-compatible buckets as a source; all major container registries; and all CI workflow providers.
- :key: A secure, pull-based mechanism, operating with least amount of privileges, and adhering to Kubernetes security policies.
- :electric_plug: Compatibility with any conformant [Kubernetes version](https://fluxcd.io/docs/installation/#prerequisites) and common ecosystem technologies such as Helm, Kustomize, RBAC, Prometheus, OPA, Kyverno, etc.
- :office: Multitenancy, multiple Git repositories, multiple clusters.
- :exclamation: Alerts and notifications.

Some of the things you can do with it:

- :heavy_check_mark: Application Operations—manage and automate deployment pipelines for apps and more.
- :magic_wand: Easily have your own custom PaaS on cloud or on premise.
- :link: Coordinate Kubernetes rollouts with virtual machines, databases, and cloud services.
- :construction: Drill down into more detailed information on any given Flux resource.
- :mag: Uncover relationships between resources and quickly navigate between them.
- :thinking_face: Understand how workloads are reconciled through a directional graph.
- :goggles: View Kubernetes events relating to a given object to understand issues and changes.
- :no_pedestrians: Secure access to the dashboard through the ability to integrate with an OIDC provider (such as Dex).

OK, time to [install](./open-source/install-oss.md)!

[flux]: https://fluxcd.io
