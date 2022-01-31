---
sidebar_position: 8
---

# Deploying the Sock Shop Application

In this example, we'll show how easy it is to deploy a real world application using Weave GitOps. The _Sock Shop_ is a well known microservices application that is widely used in demonstration and testing of microservice environments such as Kubernetes. We'll actually see two different ways of deploying the _Sock Shop_:

- as a plain set of Kubernetes manifests
- as a helm chart

# Prerequisites

In order to deploy the _Sock Shop_, you need to first deploy Weave GitOps to a Kubernetes cluster. If you'd like to test this out locally, you can set up a [kind](https://kind.sigs.k8s.io/) cluster by following the instructions at the link. Regardless of which cluster you'd like to use, you can install Weave GitOps by first making sure your default kubeconfig points to the chosen cluster and then running `gitops install --config-repo <configuration repository>`. The configuration repository is a Git repository that will hold the resource definitions required to manage your applications via GitOps. Please note that these examples are being run with the `GITOPS_TOKEN` environment variable set to a valid GitHub Personal Access Token (PAT) possessing `repo` access. If that were not the case, you would see extra user authentication steps in the output.

```console
gitops install --config-repo ssh://git@github.com/example/external.git

✚ generating manifests
✔ manifests build completed
► installing components in wego-system namespace
CustomResourceDefinition/alerts.notification.toolkit.fluxcd.io created
CustomResourceDefinition/buckets.source.toolkit.fluxcd.io created
CustomResourceDefinition/gitrepositories.source.toolkit.fluxcd.io created
CustomResourceDefinition/helmcharts.source.toolkit.fluxcd.io created
CustomResourceDefinition/helmreleases.helm.toolkit.fluxcd.io created
CustomResourceDefinition/helmrepositories.source.toolkit.fluxcd.io created
CustomResourceDefinition/imagepolicies.image.toolkit.fluxcd.io created
CustomResourceDefinition/imagerepositories.image.toolkit.fluxcd.io created
CustomResourceDefinition/imageupdateautomations.image.toolkit.fluxcd.io created
CustomResourceDefinition/kustomizations.kustomize.toolkit.fluxcd.io created
CustomResourceDefinition/providers.notification.toolkit.fluxcd.io created
CustomResourceDefinition/receivers.notification.toolkit.fluxcd.io created
Namespace/wego-system created
ServiceAccount/wego-system/helm-controller created
ServiceAccount/wego-system/image-automation-controller created
ServiceAccount/wego-system/image-reflector-controller created
ServiceAccount/wego-system/kustomize-controller created
ServiceAccount/wego-system/notification-controller created
ServiceAccount/wego-system/source-controller created
ClusterRole/crd-controller-wego-system created
ClusterRoleBinding/cluster-reconciler-wego-system created
ClusterRoleBinding/crd-controller-wego-system created
Service/wego-system/notification-controller created
Service/wego-system/source-controller created
Service/wego-system/webhook-receiver created
Deployment/wego-system/helm-controller created
Deployment/wego-system/image-automation-controller created
Deployment/wego-system/image-reflector-controller created
Deployment/wego-system/kustomize-controller created
Deployment/wego-system/notification-controller created
Deployment/wego-system/source-controller created
NetworkPolicy/wego-system/allow-egress created
NetworkPolicy/wego-system/allow-scraping created
NetworkPolicy/wego-system/allow-webhooks created
◎ verifying installation
✔ helm-controller: deployment ready
✔ image-automation-controller: deployment ready
✔ image-reflector-controller: deployment ready
✔ kustomize-controller: deployment ready
✔ notification-controller: deployment ready
✔ source-controller: deployment ready
✔ install finished
Deploy key generated and uploaded to git provider
► Writing manifests to disk
► Committing and pushing gitops updates for application
► Pushing app changes to repository
► Applying manifests to the cluster
arete: /tmp/sock-shop>
```

Once you see `► Applying manifests to the cluster`, your cluster is ready to go with Weave GitOps.

## Deploying with Weave GitOps

Once you have a cluster running Weave GitOps, it's simple to deploy an application like [_Sock Shop_](https://github.com/microservices-demo/microservices-demo).

To deploy the _Sock Shop_, we need to use `gitops add app`. `gitops add app` will store the GitOps automation support for your application in the `.weave-gitops` directory of the configuration repository you specified at install time. The definition of your application can be stored either in a separate repository or in the configuration repository itself (for a simple all-in-one configuration). If you want to store the application resources in the configuration repository, you only need to specify the `--url` flag which will be used for both application and configuration resources; however, this assumes that the application repository URL was passed to `gitops install`. If you want the application resources to be stored separately, you need to specify both `--url` and `--config-repo` parameters. The `--url` parameter should be the URL of the repository containing the application definition and the `--config-repo` parameter must be the URL that was used in `gitops install`.

First, let's fork the _Sock Shop_ repository. You can simply go to the [repository](https://github.com/microservices-demo/microservices-demo) in `GitHub` and select `Fork`.

Now, we can add the Sock Shop application to the configuration repository so it can be managed through GitOps:

```console
> gitops add app --url ssh://git@github.com/example/microservices-demo.git --path ./deploy/kubernetes/manifests --config-repo 
ssh://git@github.com/example/external.git --auto-merge
```

```console
Adding application:

Name: microservices-demo
URL: ssh://git@github.com/example/microservices-demo.git
Path: ./deploy/kubernetes/manifests
Branch: master
Type: kustomize

◎ Checking cluster status
✔ GitOps installed
✚ Generating application spec manifest
✚ Generating GitOps automation manifests
► Adding application "microservices-demo" to cluster "kind-kind" and repository
► Committing and pushing gitops updates for application
► Pushing app changes to repository
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
gitops add app \                                                 # (1)
   --url ssh://git@github.com/example/microservices-demo.git \   # (2)
   --path ./deploy/kubernetes/manifests \                        # (3)
   --config-repo ssh://git@github.com/example/external.git    # (4)
   --auto-merge                                                  # (5)`
```

1. Add an application to a cluster under the control of Weave GitOps
2. The application is defined in the GitHub repository at the specified URL
3. Only the manifests at the specified path within the repository are part of the application
4. Store the management manifests in a separate configuration repository within GitHub; the `config-repo` parameter says where to store management manifests. The default location (if no `config-repo` is specified) is to place them in the `.weave-gitops` directory within the application repository itself. An actual URL value causes them to be stored in the repository referenced by the URL
5. Don't create a pull request for the management manifests; push them directly to the upstream repository

### Using Helm Charts

The application can also be deployed via a helm chart. Applications defined in helm charts can be deployed from either helm repositories or git repositories. In the case of the _Sock Shop_ application, a helm chart is included in the `GitHub` repository. We only need to make minor changes to the command we used above to switch to a helm chart, but using a helm chart for _Sock Shop_ requires the target namespace to exist before deploying. By default, the chart would be deployed into the `wego-system` namespace (since we know it exists), but we'd like to put it in the `sock-shop` namespace. So, before we run `gitops add app`, we'll run:

```console
kubectl create namespace sock-shop
namespace/sock-shop created
>
```

Then, we can run:

```console
> gitops add app --url ssh://git@github.com/example/microservices-demo.git --path ./deploy/kubernetes/helm-chart --config-repo ssh://git@github.com/example/external.git --deployment-type helm --helm-release-target-namespace sock-shop --auto-merge
Adding application:

Name: microservices-demo
URL: ssh://git@github.com/example/microservices-demo.git
Path: ./deploy/kubernetes/helm-chart
Branch: master
Type: helm

◎ Checking cluster status
✔ GitOps installed
✚ Generating application spec manifest
✚ Generating GitOps automation manifests
► Adding application "microservices-demo" to cluster "kind-kind" and repository
► Committing and pushing gitops updates for application
► Pushing app changes to repository
>
```

Examining this command, we see two new arguments:

```console
gitops add app \
--name microservices-demo
--url ssh://git@github.com/example/microservices-demo.git \
--path ./deploy/kubernetes/helm-chart \
--config-repo ssh://git@github.com/example/external.git
--deployment-type helm \                   # (1)
--helm-release-target-namespace sock-shop  # (2)
--auto-merge
```

1. Since we're pulling the chart from a git repository, we need to explicitly state that we're using a helm chart. If we were using a helm repository, we would use `--chart <chart name>` instead of `--path <path to application>` and the deployment type would be unambiguous
2. The application will be deployed in the namespace specified by `--helm-release-target-namespace`

You can check the status of the application by running the `gitops get app microservices-demo` command.

### Single Repository Usage

As we mentioned above, it's possible to have a single repository perform hold both the application and the configuration. If you place the application manifests in the configuration repository passed to `gitops install`, you can leave off the separate `--config-repo` parameter. In this case, we would either have had to pass the `microservices-demo` URL to `gitops install` or copy the application manifests into the `external` repository. Let's proceed as if we had initialized the cluster with: `gitops install --config-repo ssh://git@github.com/example/microservices-demo.git`.

```console
> gitops add app --url ssh://git@github.com/example/microservices-demo.git --path ./deploy/kubernetes/manifests --auto-merge
Adding application:

Name: microservices-demo
URL: ssh://git@github.com/example/microservices-demo.git
Path: ./deploy/kubernetes/manifests
Branch: master
Type: kustomize

◎ Checking cluster status
✔ GitOps installed
✚ Generating application spec manifest
✚ Generating GitOps automation manifests
► Adding application "microservices-demo" to cluster "kind-kind" and repository
► Committing and pushing gitops updates for application
► Pushing app changes to repository
>
```

So, it's just like the example above except we didn't have to call out the location of the configuration repository. Regardless of whether or not the application manifests are stored in the configuration repository, though, the configuration itself is stored in a special directory (`.weave-gitops`) at the top level of the configuration repository:

```console
> tree .weave-gitops
.weave-gitops
├── apps
│   └── microservices-demo
│       ├── app.yaml
│       ├── kustomization.yaml
│       ├── microservices-demo-gitops-deploy.yaml
│       └── microservices-demo-gitops-source.yaml
└── clusters
    └── kind-kind
        ├── system
        │   ├── flux-source-resource.yaml
        │   ├── flux-system-kustomization-resource.yaml
        │   ├── flux-user-kustomization-resource.yaml
        │   ├── gitops-runtime.yaml
        │   ├── wego-app.yaml
        │   └── wego-system.yaml
        └── user
            └── kustomization.yaml

6 directories, 11 files
```

In this case, the `apps` directory contains one app (microservices-demo). The `app.yaml` file looks like:

```yaml
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
  config_url: ssh://git@github.com/example/external.git
  deployment_type: kustomize
  path: ./deploy/kubernetes/manifests
  source_type: git
  url: ssh://git@github.com/example/microservices-demo.git
```

It describes the application and includes a label derived from the URL, path, and branch to prevent multiple applications from referencing the same source within git.

The `kustomization.yaml` file holds a list of application components that will be deployed:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
metadata:
  name: microservices-demo
  namespace: wego-system
resources:
  - app.yaml
  - microservices-demo-gitops-deploy.yaml
  - microservices-demo-gitops-source.yaml
```

The `microservices-demo-gitops-source.yaml` file tells flux the location (repository) containing the application. It has a special `ignore` section that skips `.weave-gitops` to support keeping an application in the configuration repository:

```yaml
apiVersion: source.toolkit.fluxcd.io/v1beta1
kind: GitRepository
metadata:
  name: microservices-demo
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
    branch: master
  url: https://github.com/example/microservices-demo.git
```

The `microservices-demo-gitops-deploy.yaml` file defines the path within the repository and the sync interval for the application:

```yaml
---
apiVersion: kustomize.toolkit.fluxcd.io/v1beta2
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
```

(This will look different in the case of a helm chart; it will hold a `HelmRelease` rather than a `Kustomization`)

Finally, the `clusters` directory has a subdirectory for each cluster defining which applications will run there. The `user/kustomization.yaml` file in a specific cluster directory has a `resources` section containing a list of applications references:

```yaml
resources:
  - ../../../apps/microservices-demo
```

### Using Pull Requests

We've reached the all-singing, all-dancing case now. This is the way most people will actually use Weave GitOps in a real environment. Whether you use the default application repository model or have a separate configuration repository, you can support reviewing and auditing changes to your GitOps resources via _Pull Requests_. (Also, as a practical matter, many people don't allow direct merges to their repositories without pull requests anyway)

In order to use pull requests for your GitOps resources, you simply need to leave off the `--auto-merge` flag we've been passing so far (in other words, using pull requests is the default). For example, if we run the previous command without `--auto-merge`, we see different output:

```console
> gitops add app --url ssh://git@github.com/example/microservices-demo.git --path ./deploy/kubernetes/manifests
Adding application:

Name: microservices-demo
URL: ssh://git@github.com/example/microservices-demo.git
Path: ./deploy/kubernetes/manifests
Branch: master
Type: kustomize

◎ Checking cluster status
✔ GitOps installed
✚ Generating application spec manifest
✚ Generating GitOps automation manifests
► Adding application "microservices-demo" to cluster "kind-kind" and repository
Pull Request created: https://github.com/example/external/pull/14

>
```

Note the line: `Pull Request created: https://github.com/example/external/pull/14`. If we were to go to that GitHub repository and merge the pull request, the app would then be deployed.

Hopefully, this example has given you a good understanding of how to deploy applications with Weave GitOps. Thanks for reading!
