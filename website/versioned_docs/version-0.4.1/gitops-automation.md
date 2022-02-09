---
sidebar_position: 7
---

# GitOps Automation Configuration

Weave GitOps has a key concept which is the "GitOps Automation".
This is the configuration that defines the flow that happens when an update happens in Git. This flow _reconciles_ the workload into the target.

## What is the GitOps Automation Configuration?

The GitOps Automation configuration consists of two types of configuration: `apps` and `targets`.
Fundamentally, GitOps is connecting `apps` with `targets` by automatically deploying the app into the target.

## Where does it live?

Weave GitOps supports three places to store this configuration:

1. In the same repository as your workload definition (e.g. in the same repository containing your k8s YAMLs)
2. In a separate repository that can cover different applications.
3. Only configured into the cluster (not stored)

The default behaviour is #1, which is to store the GitOps Automation configuration in the same repository as your workload YAMLs are stored.
This is the best option for a team that manages their own app and infrastructure and wants a simple, easy deployment of Weave GitOps.

Option 2 is designed for scenarios where there are multiple applications managed by a single team (maybe a platform ops team or a team that owns more than one app).

Option 3 is a basic approach that is useful if you are just learning about GitOps and don't want to be able to change the automation flow.

## What files are there

The `.wego` folder contains the following structure:

```
.
├── apps
│   └── appname
│       └── app.yaml
└── targets
    └── clustername
        └── appname
            └── appname-gitops-runtime.yaml
```

### `app.yaml`

The `app.yaml` looks like this:

```yaml
apiVersion: wego.weave.works/v1alpha1
kind: Application
metadata:
  name: nginx
spec:
  path: ./
  url: ssh://git@github.com/org/nginx.git
```

It defines the name of the application and the git url location of the repository.

### `app-gitops-runtime.yaml`

This file defines the flux runtime flow to deploy the application into a specific target.
For example, by default this uses the [flux kustomize](https://fluxcd.io/docs/components/kustomize/kustomization/) support
to deploy your application manifests into the cluster. Other options include managing helm charts.

```yaml
---
apiVersion: kustomize.toolkit.fluxcd.io/v1beta1
kind: Kustomization
metadata:
  name: nginx-fork-dot
  namespace: wego-system
spec:
  interval: 5m0s
  path: ./
  prune: true
  sourceRef:
    kind: GitRepository
    name: nginx-fork
  validation: client
```
