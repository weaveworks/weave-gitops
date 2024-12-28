---
title: Introduction to Weave GitOps Enterprise

---

# Weave GitOps Enterprise ~ENTERPRISE~

!!! tip "Ready for more GitOps?"
    To purchase an entitlement to Weave GitOps Enterprise, please contact [sales@weave.works](mailto:sales@weave.works).

Weave GitOps Enterprise provides ops teams with an easy way to assess the
health of multiple clusters in a single place. It shows cluster information such as
Kubernetes version and number of nodes and provides details about the GitOps operations
on those clusters, such as Git repositories and recent commits. Additionally, it
aggregates Prometheus alerts to assist with troubleshooting.

If you have already purchased your entitlement, head to the [installation page](./install-enterprise.md).

## Feature Breakdown

In addition to the features in the OSS edition, Weave GitOps Enterprise
offers the following capabilities, taking your delivery from simple Continuous Delivery to Internal Developer Platform:

### :sailboat: [Cluster Fleet Management](../cluster-management/index.md)
Weave GitOps Enterprise (WGE) simplifies cluster lifecycle management at scale—even massive scale. Through pull requests, which make every action recorded and auditable, WGE makes it possible for teams to create, update, and delete clusters across entire fleets. WGE further simplifies the process by providing both a user interface (UI) and a command line interface (CLI) for teams to interact with and manage clusters on-prem, across clouds, and in hybrid environments. WGE works with [Terraform](https://www.weave.works/blog/extending-gitops-beyond-kubernetes-with-terraform), [Crossplane](https://www.weave.works/blog/gitops-goes-beyond-kubernetes-with-weave-gitops-upbound-s-universal-crossplane), and any Cluster API provider.

![WGE dashboard with cluster view](/img/wge-dashboard-dark-mode.png)

### :closed_lock_with_key: [Trusted Application Delivery](../policy/index.md)
Add policy as code to GitOps pipelines and enforce security and compliance,
application resilience and coding standards from source to production.
Validate policy conformance at every step in the software delivery pipeline:
commit, build, deploy and run time.

### :truck: [Progressive Delivery](../progressive-delivery/progressive-delivery-flagger-install.md)
Deploy into production environments safely using canary, blue/green deployment, and A/B
strategies. Simple, single-file configuration defines success rollback. Measure Service Level Objectives (SLOs)
using observability metrics from Prometheus, Datadog, New Relic, and others.

### :infinity: [CD Pipelines](../pipelines/index.md)
Rollout new software from development to production.
Environment rollouts that work with your existing CI system.

### :factory_worker: [Team Workspaces](../workspaces/index.md)
Allow DevOps teams to work seamlessly together with multi-tenancy,
total RBAC control, and policy enforcement, with integration to enterprise IAM.

### :point_up_2: [Self-Service Templates and Profiles](../gitops-templates/index.md)
Component profiles enable teams to deploy standard services quickly,
consistently and reliably. Teams can curate the profiles that are available
within their estate ensuring there is consistency everywhere. Using GitOps
it's easy to guarantee the latest, secure versions of any component are
deployed in all production systems.

### :sparkling_heart: Health Status and Compliance Dashboards
Gain a single view of the health and state of the cluster and its workloads.
Monitor deployments and alert on policy violations across apps and clusters.

### :compass: Kubernetes Anywhere
Reduce complexity with GitOps and install across all major target environments
including support for on-premise, edge, hybrid, and multi-cloud Kubernetes clusters.

### :bell: [Critical 24/7 Support](/help-and-support/)
Your business and workloads operate around the clock, and so do we.
Whenever you have a problem, our experts are there to help. We’ve got your back!
