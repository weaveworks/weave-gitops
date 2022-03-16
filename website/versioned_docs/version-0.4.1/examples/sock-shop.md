---
sidebar_position: 8
---

# Deploying the Sock Shop Application

In this example, we'll show how easy it is to deploy a real world application using Weave GitOps. The *Sock Shop* is a well known microservices application that is widely used in demonstration and testing of microservice environment such as Kubernetes. We'll actually see two different ways of deploying the *Sock Shop*:
- as a plain set of Kubernetes manifests
- as a helm chart

# Prerequisites

In order to deploy the *Sock Shop*, you need to first deploy Weave GitOps to a Kubernetes cluster. If you'd like to test this out locally, you can set up a [kind](https://kind.sigs.k8s.io/) cluster by following the instructions at the link. Regardless of which cluster you'd like to use, you can install Weave GitOps by first making sure your default kubeconfig points to the chosen cluster and then running `gitops install`.

```console
> gitops install
✚ generating manifests
✔ manifests build completed
► installing components in wego-system namespace
◎ verifying installation
✔ source-controller: deployment ready
✔ kustomize-controller: deployment ready
✔ helm-controller: deployment ready
✔ notification-controller: deployment ready
✔ image-reflector-controller: deployment ready
✔ image-automation-controller: deployment ready
✔ install finished
arete: /tmp/sock-shop>
```

Once you see `install finished`, your cluster is ready to go with Weave GitOps.

## Simple Deployment
Once you have a cluster running Weave GitOps, it's simple to deploy an application like [*Sock Shop*](https://github.com/microservices-demo/microservices-demo).

To deploy the *Sock Shop*, we need to use `gitops add app`. `gitops add app` comes in three flavors depending on how you'd like your GitOps support to be managed. Just as GitOps in general lets you synchronize your application definitions with their in-cluster versions via `git` operations and pull requests, Weave GitOps lets you manage the resources that _do_ the synchronization in the same way. That might sound a bit "meta" but it allows you to update the management resources themselves via pull requests. You can change things like synchronization intervals or even upgrade the resources via git. Why _wouldn't_ you want to be able to do those things?

The three storage options for the management resources provided by Weave GitOps are:
- in a `.wego` subdirectory of the same repository as your application (for a simple all-in-one configuration)
- in a separate configuration repository (allowing you to collect all your management resources together)
- _only_ in the cluster (this is mostly intended for quick turnaround and testing)

First, let's fork the *Sock Shop* repository. You can simply go to the [repository](https://github.com/microservices-demo/microservices-demo) in `GitHub` and select `Fork`.

Now, if we just want to get the *Sock Shop* running in the simplest way possible, without modifying anything, we can run a single command:

```console
> gitops add app --url ssh://git@github.com/example/microservices-demo.git --path ./deploy/kubernetes/manifests --app-config-url NONE
◎ Checking cluster status
✔ GitOps installed
uploading deploy key
Deploy key generated and uploaded to git provider
Adding application:

Name: microservices-demo
URL: ssh://git@github.com/example/microservices-demo.git
Path: ./deploy/kubernetes/manifests
Branch: master
Type: kustomize

✚ Generating Source manifest
✚ Generating GitOps automation manifests
✚ Generating Application spec manifest
► Applying manifests to the cluster
>
```

Here we see all the pods running:

```console
> kubectl get pods -A
NAMESPACE            NAME                                           READY   STATUS    RESTARTS   AGE
kube-system          coredns-558bd4d5db-jgcf2                       1/1     Running   0          9d
kube-system          coredns-558bd4d5db-sht4v                       1/1     Running   0          9d
kube-system          etcd-kind-control-plane                        1/1     Running   0          9d
kube-system          kindnet-tdcd2                                  1/1     Running   0          9d
kube-system          kube-apiserver-kind-control-plane              1/1     Running   0          9d
kube-system          kube-controller-manager-kind-control-plane     1/1     Running   0          9d
kube-system          kube-proxy-mqvbc                               1/1     Running   0          9d
kube-system          kube-scheduler-kind-control-plane              1/1     Running   0          9d
local-path-storage   local-path-provisioner-547f784dff-mqgjc        1/1     Running   0          9d
sock-shop            carts-b4d4ffb5c-g82h6                          1/1     Running   0          9d
sock-shop            carts-db-6c6c68b747-xtlgk                      1/1     Running   0          9d
sock-shop            catalogue-759cc6b86-jk4gf                      1/1     Running   0          9d
sock-shop            catalogue-db-96f6f6b4c-865w4                   1/1     Running   0          9d
sock-shop            front-end-5c89db9f57-99vw6                     1/1     Running   0          9d
sock-shop            orders-7664c64d75-qlz9d                        1/1     Running   0          9d
sock-shop            orders-db-659949975f-fggdb                     1/1     Running   0          9d
sock-shop            payment-7bcdbf45c9-fhl8m                       1/1     Running   0          9d
sock-shop            queue-master-5f6d6d4796-cs5f6                  1/1     Running   0          9d
sock-shop            rabbitmq-5bcbb547d7-kfzmn                      2/2     Running   0          9d
sock-shop            session-db-7cf97f8d4f-bms4c                    1/1     Running   0          9d
sock-shop            shipping-7f7999ffb7-llkrw                      1/1     Running   0          9d
sock-shop            user-68df64db9c-7gcg2                          1/1     Running   0          9d
sock-shop            user-db-6df7444fc-7s6wp                        1/1     Running   0          9d
wego-system          helm-controller-6dcbff747f-sfp97               1/1     Running   0          9d
wego-system          image-automation-controller-75f784cfdc-wxwk9   1/1     Running   0          9d
wego-system          image-reflector-controller-67d6bdcb59-hg2cv    1/1     Running   0          9d
wego-system          kustomize-controller-5d47cf49fb-b6pmg          1/1     Running   0          9d
wego-system          notification-controller-7569f7c974-824p9       1/1     Running   0          9d
wego-system          source-controller-5b976b8dd6-gqrl7             1/1     Running   0          9d
>
```

We can expose the sock shop in our browser by:

```console
> kubectl port-forward service/front-end -n sock-shop 8080:80
Forwarding from 127.0.0.1:8080 -> 8079
Forwarding from [::1]:8080 -> 8079
```

and if we visit `http://localhost:8080`, we'l see:

![sock shop](/img/sock-shop.png)

Pretty simple! Now, let's go back and look at that command in more detail:

```console
gitops add app \                                                   # (1)
   --url ssh://git@github.com/example/microservices-demo.git \   # (2)
   --path ./deploy/kubernetes/manifests \                        # (3)
   --app-config-url NONE                                         # (4)`
```

1. Add an application to a cluster under the control of Weave GitOps
2. The application is defined in the GitHub repository at the specified URL
3. Only the manifests at the specified path within the repository are part of the application
4. Don't store the management manifests in GitHub; the `app-config-url` parameter says where to store management manifests. The default location (if no `app-config-url` is specified) is to place them in a hidden directory (`.wego`) within the application repository itself. An actual URL value causes them to be stored in the repository referenced by the URL. `NONE` means to apply them to the cluster but don't store them in `git` at all.

For a quick turnaround (during development or for testing) `NONE` does the trick. The application is being managed via GitOps, so any changes made in the application repository will be applied to the cluster when they are merged.

The application can also be deployed via a helm chart. Applications defined in helm charts can be deployed from either helm repositories or git repositories. In the case of the *Sock Shop* application, a helm chart is included in the `GitHub` repository. We only need to make minor changes to the command we used above to switch to a helm chart, but using a helm chart for *Sock Shop* requires the target namespace to exist before deploying. By default, the chart would be deployed into the `wego-system` namespace (since we know it exists), but we'd like to put it in the `sock-shop` namespace. So, before we run `gitops add app`, we'll run:

```console
kubectl create namespace sock-shop
namespace/sock-shop created
>
```

Then, we can run:

```console
> gitops add app --url ssh://git@github.com/example/microservices-demo.git --deployment-type helm --path ./deploy/kubernetes/helm-chart --helm-release-target-namespace sock-shop --app-config-url NONE
◎ Checking cluster status
✔ GitOps installed
uploading deploy key
Deploy key generated and uploaded to git provider
Adding application:

Name: microservices-demo
URL: ssh://git@github.com/example/microservices-demo.git
Path: ./deploy/kubernetes/helm-chart
Branch: master
Type: helm

✚ Generating Source manifest
✚ Generating GitOps automation manifests
✚ Generating Application spec manifest
► Applying manifests to the cluster
>
```

Examining this command, we see two new arguments:

```console
gitops add app \
--url ssh://git@github.com/example/microservices-demo.git \
--path ./deploy/kubernetes/helm-chart \
--app-config-url NONE \
--deployment-type helm \                   # (1)
--helm-release-target-namespace sock-shop  # (2)
```

1. Since we're pulling the chart from a git repository, we need to explicitly state that we're using a helm chart. If we were using a helm repository, we would use `--chart <chart name>` instead of `--path <path to application>` and the deployment type would be unambiguous
2. The application will be deployed in the namespace specified by `--helm-release-target-namespace`


## Deployment with Managed GitOps Resources
As mentioned above, Weave GitOps allows you to also manage your GitOps resources via GitOps. This has several advantages:
1. Increased control - you can alter parameters such as synchronization interval via updates to git
2. Upgradability - when the Weave GitOps upgrades can be managed via git updates
3. Reviewability and Auditability - Changes performed via git can be tracked; additionally, the default behavior of `gitops add app` is to create pull requests for upstream repositories which makes reviewing and auditing straightforward using the same tools used during development
4. Recoverability - if the cluster is restored after failure, the management resources can be restored from git

To run with managed GitOps resources, we need to tell Weave GitOps where to put them. The default behavior of putting them in a private directory within the application repository works fine for repositories under our control (but doesn't work if we want to use either a helm repository or an open source git repository). To use the default, we simply leave off the `--app-config-url NONE` parameter (or, equivalently, use `--app-config-url ''`). This will store the manifests for our GitOps resources in the `.wego` directory within our application repository.

### Deployment with GitOps Resources in Application Repository
```console
> gitops add app --url ssh://git@github.com/example/microservices-demo.git --path ./deploy/kubernetes/manifests --auto-merge
◎ Checking cluster status
✔ GitOps installed
uploading deploy key
Deploy key generated and uploaded to git provider
Adding application:

Name: microservices-demo
URL: ssh://git@github.com/example/microservices-demo.git
Path: ./deploy/kubernetes/manifests
Branch: master
Type: kustomize

✚ Generating Source manifest
✚ Generating GitOps automation manifests
✚ Generating Application spec manifest
► Cloning ssh://git@github.com/example/microservices-demo.git
► Writing manifests to disk
► Applying manifests to the cluster
► Committing and pushing gitops updates for application
► Pushing app changes to repository
>
```

After running this, not only is the *Sock Shop* running in the cluster, but we can observe and modify the GitOps resources for the application. Here's how they look in the application repository:

```console
> tree .wego
.wego
├── apps
│   └── microservices-demo
│       └── app.yaml
└── targets
    └── kind-kind
        └── microservices-demo
            ├── microservices-demo-gitops-deploy.yaml
            └── microservices-demo-gitops-source.yaml
```

The `apps` directory contains one app (microservices-demo). The `app.yaml` file looks like:

```console
---
apiVersion: wego.weave.works/v1alpha1
kind: Application
metadata:
  labels:
    wego.weave.works/app-identifier: wego-85414ad27cd476d497d715818deda0c6
  name: microservices-demo
  namespace: wego-system
spec:
  branch: master
  deployment_type: kustomize
  path: ./deploy/kubernetes/manifests
  source_type: git
  url: ssh://git@github.com/example/microservices-demo.git
```

It describes the application and includes a label derived from the URL, path, and branch to prevent multiple applications from referencing the same source within git.

The `targets` directory has a subdirectory for each cluster containing the GitOps resources for each application deployed to the cluster. The `source` manifest defines the repository to pull from:

```console
---
apiVersion: source.toolkit.fluxcd.io/v1beta1
kind: GitRepository
metadata:
  name: microservices-demo
  namespace: wego-system
spec:
  interval: 30s
  ref:
    branch: master
  url: https://github.com/example/microservices-demo.git
```

The `deploy` manifest defines how the specific application information will be pulled from the repository:

```console
---
apiVersion: kustomize.toolkit.fluxcd.io/v1beta1
kind: Kustomization
metadata:
  name: microservices-demo
  namespace: wego-system
spec:
  interval: 1m0s
  path: ./deploy/kubernetes/manifests
  prune: true
  sourceRef:
    kind: GitRepository
    name: microservices-demo
  validation: client
```

(This will look different in the case of a helm chart; it will hold a `HelmRelease` rather than a `Kustomization`)

### Deployment with GitOps Resources in Separate Configuration Repository
If you'd like to manage GitOps resources for a helm repository or a git repository not under your control, or you'd like to manage all of your GitOps resources together, you can store them in a separate configuration repository by passing a URL to `--app-config-url`:

```console
> gitops add app --url ssh://git@github.com/example/microservices-demo.git --path ./deploy/kubernetes/manifests --app-config-url ssh://git@github.com/example/external.git --auto-merge
◎ Checking cluster status
✔ GitOps installed
uploading deploy key
Deploy key generated and uploaded to git provider
uploading deploy key
Deploy key generated and uploaded to git provider
Adding application:

Name: microservices-demo
URL: ssh://git@github.com/example/microservices-demo.git
Path: ./deploy/kubernetes/manifests
Branch: master
Type: kustomize

✚ Generating Source manifest
✚ Generating GitOps automation manifests
✚ Generating Application spec manifest
► Writing manifests to disk
► Applying manifests to the cluster
► Committing and pushing gitops updates for application
✔ App is up to date
>
```

After running this command, the `external` repository will contain (assuming it was empty to start):

```console
> tree
.
├── apps
│   └── microservices-demo
│       └── app.yaml
├── README.md
└── targets
    └── kind-kind
        └── microservices-demo
            ├── microservices-demo-gitops-deploy.yaml
            └── microservices-demo-gitops-source.yaml
```

Ignoring the `README.md` file, the rest of the repository contents are the same as the contents of the `.wego` directory in the previous example. Since this repository is _not_ an application repository, though, the contents are at the top level.

### Using Pull Requests
We've reached the all-singing, all-dancing case now. This is the way most people will actually use Weave GitOps in a real environment. Whether you use the default application repository model or have a separate configuration repository, you can support reviewing and auditing changes to your GitOps resources via *Pull Requests*. (Also, as a practical matter, many people don't allow direct merges to their repositories without pull requests anyway)

In order to use pull requests for your GitOps resources, you simply need to leave off the `--auto-merge` flag we've been passing since we started storing our GitOps resources (In other words, using pull requests is the default). For example, if we run the previous command without `--auto-merge`, we see different output:

```console
> gitops add app --url ssh://git@github.com/example/microservices-demo.git --path ./deploy/kubernetes/manifests --app-config-url ssh://git@github.com/example/external.git
◎ Checking cluster status
✔ GitOps installed
uploading deploy key
Deploy key generated and uploaded to git provider
uploading deploy key
Deploy key generated and uploaded to git provider
Adding application:

Name: microservices-demo
URL: ssh://git@github.com/example/microservices-demo.git
Path: ./deploy/kubernetes/manifests
Branch: master
Type: kustomize

✚ Generating Source manifest
✚ Generating GitOps automation manifests
✚ Generating Application spec manifest
Pull Request created: https://github.com/example/external/pull/12

► Applying manifests to the cluster
► Committing and pushing gitops updates for application
✔ App is up to date
>
```

Note the line: `Pull Request created: https://github.com/example/external/pull/12`. If we were to go to that GitHub repository and merge the pull request, the app would then be deployed.

(The lines below the pull request line refer to updating the GitOps resources that _manage_ the GitOps resources which is a separate topic for another time)

Hopefully, this example has given you a good understanding of how to deploy applications with Weave GitOps. Thanks for reading!

