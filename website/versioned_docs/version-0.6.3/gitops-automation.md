---
title: GitOps Automation Configuration
sidebar_position: 7
hide_title: true
---

import TierLabel from "./_components/TierLabel";

<h1>
  {frontMatter.title} <TierLabel tiers="All tiers" />
</h1>

Weave GitOps has a key concept which is the "GitOps Automation".
This is the configuration that defines the flow that happens when an update happens in Git. This flow _reconciles_ the workload into the target.

## What is the GitOps Automation Configuration?

The GitOps Automation configuration consists of two types of configuration: `apps` and `clusters`.
Fundamentally, GitOps is connecting `apps` with `clusters` by automatically deploying the app into the cluster.

## Where does it live?

The configuration is stored in the repository specified in `gitops install`. The application manifests can be stored in the same repository as the configuration or a separate repository.

## What files are there

The `.weave-gitops` folder contains the following structure:

```
.weave-gitops
├── apps
│   └── <app name>
│       ├── app.yaml
│       ├── kustomization.yaml
│       ├── <app name>-gitops-deploy.yaml
│       └── <app name>-gitops-source.yaml
└── clusters
    └── <cluster name>
        ├── system
        │   ├── flux-source-resource.yaml
        │   ├── flux-system-kustomization-resource.yaml
        │   ├── flux-user-kustomization-resource.yaml
        │   ├── gitops-runtime.yaml
        │   ├── wego-app.yaml
        │   └── wego-system.yaml
        └── user
            └── kustomization.yaml
```

### `app.yaml`

The `app.yaml` looks like this:

```yaml
---
apiVersion: wego.weave.works/v1alpha1
kind: Application
metadata:
  labels:
    wego.weave.works/app-identifier: wego-52cdbd4b6d1a20a934f101708a93cf10
  name: <app name>
  namespace: wego-system
spec:
  branch: main
  config_url: ssh://git@github.com/user/example.git
  deployment_type: kustomize
  path: ./
  source_type: git
  url: ssh://git@github.com/<yr-org-goes-here>/podinfo-deploy.git
```

It defines:
- the name of the application (name)
- the namespace of the application (namespace)
- the git URL location of the repository, or URL of the helm repository for a helm chart (url)
- the path and branch containing the application manifests within the repository (path, branch)
- the git URL location of the configuration repository (config\_url)
- whether the application will be read from a helm repository or a git repository (source\_type)
- how the application will be deployed -- via kustomize or a helm release (deployment\_type)

### `kustomization.yaml`

The [kustomization.yaml](https://kustomize.io/) file defines the set of resources that will be managed via GitOps for the application.

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
metadata:
  name: <app name>
  namespace: wego-system
resources:
- app.yaml
- <app name>-gitops-deploy.yaml
- <app name>-gitops-source.yaml
```

The content of each file mentioned in the `resources` section is synchronized with the cluster by `flux`. The `kustomization.yaml` file itself is associated with the cluster via a reference to the app directory in the `user` kustomization file for the cluster (`.weave-gitops/clusters/<cluster name>/user/kustomization.yaml`).

### `<app name>-gitops-source.yaml`

This file defines how the application will be read into the system. Current options are `GitRepository` and `HelmRepository`:

```yaml
apiVersion: source.toolkit.fluxcd.io/v1beta1
kind: GitRepository
metadata:
  name: <app name>
  namespace: wego-system
spec:
  ignore: |-
    .weave-gitops/
    .git/
    .gitignore
    .gitmodules
    .gitattributes
    *.jpg
    *.jpeg
    *.gif
    *.png
    *.wmv
    *.flv
    *.tar.gz
    *.zip
    .github/
    .circleci/
    .travis.yml
    .gitlab-ci.yml
    appveyor.yml
    .drone.yml
    cloudbuild.yaml
    codeship-services.yml
    codeship-steps.yml
    **/.goreleaser.yml
    **/.sops.yaml
    **/.flux.yaml
  interval: 30s
  ref:
    branch: main
  url: https://github.com/<yr-org-goes-here>/podinfo-deploy.git
```

### `<app name>-gitops-deploy.yaml`

This file defines the flux runtime flow to deploy the application into a specific target.
For example, by default this uses the [flux kustomize](https://fluxcd.io/docs/components/kustomize/kustomization/) support
to deploy your application manifests into the cluster. Other options include managing helm charts.

```yaml
---
apiVersion: kustomize.toolkit.fluxcd.io/v1beta2
kind: Kustomization
metadata:
  name: <app name>
  namespace: wego-system
spec:
  interval: 1m0s
  path: ./
  prune: true
  sourceRef:
    kind: GitRepository
    name: <app name>
```
