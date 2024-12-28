---
title: Install Enterprise in Air-gapped Environments

toc_max_heading_level: 4
---

# Install Enterprise in Air-gapped Environments ~ENTERPRISE~

From [wikipedia](https://en.wikipedia.org/wiki/Air_gap_(networking))

>An air gap, air wall, air gapping or disconnected network is a network security measure employed on one or more computers
to ensure that a secure computer network is physically isolated from unsecured networks, such as the public Internet or an unsecured local area network...

This document guides on how to install Weave GitOps Enterprise (WGE) in a restricted environment.

# Before You Start

There are multiple restrictions that could happen within an air-gapped environment. This guide assumes that you have egress network
restrictions. In order to install WGE, the required artifacts must be loaded
from a private registry. This guide helps you with the task to identity the Helm charts
and container images required to install WGE and to load them into your private registry.

It also assumes that you could prepare the installation from a proxy host. A proxy host is defined here
as a computer that is able to access to both the public and private network. It could take different shapes,
for example, it could be a bastion host, a corp laptop, etc.

Access to both public and private network is required during the airgap installation but not simultaneously.
It is expected to have an online stage to gather the artifacts first, and an offline stage later,
to load the artifacts in the private network.

Finally, we aim to provide an end to end example to use it as a guidance more than a recipe. Feel free to adapt the details
that do not fit within your context.

# Install WGE

There are different variations of the following stages and conditions. We consider that installing 
WGE in an air-gapped environment could follow the following stages.

1. Set up a WGE install environment.
2. Collect artifacts and publish to a private registry.
3. Install WGE in the air-gapped environment.

## Set up a WGE install environment

The main goal of this stage is to recreate a local WGE within your context, to collect
the container images and Helm charts, that will be required in your private registry for the offline installation.

A three-step setup is followed.

1. Setup a proxy host
2. Setup a private registry
3. Install WGE

### Setup a proxy host

There are many possible configurations for this host. This guide will assume that the host has installed the following:

- [docker](https://www.docker.com/) as container runtime.
- [kubectl and kind](https://kubernetes.io/docs/tasks/tools)
- [helm](https://helm.sh/docs/intro/install/)
- [skopeo](https://github.com/containers/skopeo) to manage container images
- [flux](https://fluxcd.io/flux/cmd/) to boostrap Flux in the environment.
- [clusterctl](https://cluster-api.sigs.k8s.io/user/quick-start.html#install-clusterctl) to replicate the cluster management
capabilities.

#### Create a Kind Cluster

Create a kind cluster with registry following [this guide](https://kind.sigs.k8s.io/docs/user/local-registry/)

#### Install Flux

You could just use `flux install` to install Flux into your kind cluster

#### Set up a Helm repo

We are going to install [ChartMuseum](https://chartmuseum.com/) via Flux.

Remember to also install helm plugin
[cm-push](https://github.com/chartmuseum/helm-push).

??? example "Expand to see installation yaml"

    ```yaml
    ---
    apiVersion: source.toolkit.fluxcd.io/v1beta2
    kind: HelmRepository
    metadata:
    name: chartmuseum
    namespace: flux-system
    spec:
    interval: 10m
    url: https://chartmuseum.github.io/charts
    ---
    apiVersion: helm.toolkit.fluxcd.io/v2beta1
    kind: HelmRelease
    metadata:
    name: chartmuseum
    namespace: flux-system
    spec:
    chart:
        spec:
        chart: chartmuseum
        sourceRef:
            kind: HelmRepository
            name: chartmuseum
            namespace: flux-system
    interval: 10m0s
    timeout: 10m0s
    releaseName: helm-repo
    install:
        crds: CreateReplace
        remediation:
        retries: 3
    values:
        env:
        open:
            DISABLE_API: "false"
            AUTH_ANONYMOUS_GET: "true"
    ```


Set up access from your host.

```bash
#expose kubernetes svc
kubectl -n flux-system port-forward svc/helm-repo-chartmuseum 8080:8080 &

#add hostname
sudo -- sh -c "echo 127.0.0.1 helm-repo-chartmuseum >> /etc/hosts"

```
Test that you could reach it.
```bash
#add repo to helm
helm repo add private http://helm-repo-chartmuseum:8080

#test that works
helm repo update private
```

At this stage you have already a private registry for container images and helm charts.

### Install WGE

This step is to gather the artifacts and images in your local environment to push to the private registry.

####  Cluster API

This would vary depending on the provider, given that we target a offline environment, most likely we are in
a private cloud environment, so we will be using [liquidmetal](https://weaveworks-liquidmetal.github.io/site/docs/tutorial-basics/capi/).

Export these environment variables to configure your CAPI experience. Adjust them to your context.

```shell
export CAPI_BASE_PATH=/tmp/capi
export CERT_MANAGER_VERSION=v1.9.1
export CAPI_VERSION=v1.3.0
export CAPMVM_VERSION=v0.7.0
export EXP_CLUSTER_RESOURCE_SET=true
export CONTROL_PLANE_MACHINE_COUNT=1
export WORKER_MACHINE_COUNT=1
export CONTROL_PLANE_VIP="192.168.100.9"
export HOST_ENDPOINT="192.168.1.130:9090"
```

Execute the following script to generate `clusterctl` config file.

```shell
cat << EOF > clusterctl.yaml
cert-manager:
  url: "$CAPI_BASE_PATH/cert-manager/$CERT_MANAGER_VERSION/cert-manager.yaml"

providers:
  - name: "microvm"
    url: "$CAPI_BASE_PATH/infrastructure-microvm/$CAPMVM_VERSION/infrastructure-components.yaml"
    type: "InfrastructureProvider"
  - name: "cluster-api"
    url: "$CAPI_BASE_PATH/cluster-api/$CAPI_VERSION/core-components.yaml"
    type: "CoreProvider"
  - name: "kubeadm"
    url: "$CAPI_BASE_PATH/bootstrap-kubeadm/$CAPI_VERSION/bootstrap-components.yaml"
    type: "BootstrapProvider"
  - name: "kubeadm"
    url: "$CAPI_BASE_PATH/control-plane-kubeadm/$CAPI_VERSION/control-plane-components.yaml"
    type: "ControlPlaneProvider"
EOF
```
Execute `make` using the following makefile to intialise CAPI in your cluster:

??? example "Expand to see Makefile contents"

    ```makefile
    .PHONY := capi

    capi: capi-init capi-cluster

    capi-init: cert-manager cluster-api bootstrap-kubeadm control-plane-kubeadm microvm clusterctl-init

    cert-manager:
        mkdir -p  $(CAPI_BASE_PATH)/cert-manager/$(CERT_MANAGER_VERSION)
        curl -L https://github.com/cert-manager/cert-manager/releases/download/$(CERT_MANAGER_VERSION)/cert-manager.yaml --output $(CAPI_BASE_PATH)/cert-manager/$(CERT_MANAGER_VERSION)/cert-manager.yaml

    cluster-api:
        mkdir -p  $(CAPI_BASE_PATH)/cluster-api/$(CAPI_VERSION)
        curl -L https://github.com/kubernetes-sigs/cluster-api/releases/download/$(CAPI_VERSION)/core-components.yaml --output $(CAPI_BASE_PATH)/cluster-api/$(CAPI_VERSION)/core-components.yaml
        curl -L https://github.com/kubernetes-sigs/cluster-api/releases/download/$(CAPI_VERSION)/metadata.yaml --output $(CAPI_BASE_PATH)/cluster-api/$(CAPI_VERSION)/metadata.yaml

    bootstrap-kubeadm:
        mkdir -p  $(CAPI_BASE_PATH)/bootstrap-kubeadm/$(CAPI_VERSION)
        curl -L https://github.com/kubernetes-sigs/cluster-api/releases/download/$(CAPI_VERSION)/bootstrap-components.yaml --output $(CAPI_BASE_PATH)/bootstrap-kubeadm/$(CAPI_VERSION)/bootstrap-components.yaml
        curl -L https://github.com/kubernetes-sigs/cluster-api/releases/download/$(CAPI_VERSION)/metadata.yaml --output $(CAPI_BASE_PATH)/bootstrap-kubeadm/$(CAPI_VERSION)/metadata.yaml

    control-plane-kubeadm:
        mkdir -p  $(CAPI_BASE_PATH)/control-plane-kubeadm/$(CAPI_VERSION)
        curl -L https://github.com/kubernetes-sigs/cluster-api/releases/download/$(CAPI_VERSION)/control-plane-components.yaml --output $(CAPI_BASE_PATH)/control-plane-kubeadm/$(CAPI_VERSION)/control-plane-components.yaml
        curl -L https://github.com/kubernetes-sigs/cluster-api/releases/download/$(CAPI_VERSION)/metadata.yaml --output $(CAPI_BASE_PATH)/control-plane-kubeadm/$(CAPI_VERSION)/metadata.yaml

    microvm:
        mkdir -p  $(CAPI_BASE_PATH)/infrastructure-microvm/$(CAPMVM_VERSION)
        curl -L https://github.com/weaveworks-liquidmetal/cluster-api-provider-microvm/releases/download/$(CAPMVM_VERSION)/infrastructure-components.yaml --output $(CAPI_BASE_PATH)/infrastructure-microvm/$(CAPMVM_VERSION)/infrastructure-components.yaml
        curl -L https://github.com/weaveworks-liquidmetal/cluster-api-provider-microvm/releases/download/$(CAPMVM_VERSION)/cluster-template-cilium.yaml --output $(CAPI_BASE_PATH)/infrastructure-microvm/$(CAPMVM_VERSION)/cluster-template-cilium.yaml
        curl -L https://github.com/weaveworks-liquidmetal/cluster-api-provider-microvm/releases/download/$(CAPMVM_VERSION)/metadata.yaml --output $(CAPI_BASE_PATH)/infrastructure-microvm/$(CAPMVM_VERSION)/metadata.yaml

    clusterctl-init:
        clusterctl init --wait-providers -v 4 --config clusterctl.yaml --infrastructure microvm

    capi-cluster:
        clusterctl generate cluster --config clusterctl.yaml -i microvm:$(CAPMVM_VERSION) -f cilium lm-demo | kubectl apply -f -
    ```


#### Deploying the Terraform Controller

Apply the following example manifest to deploy the Terraform Controller:

??? example "Expand to see file contents"

    ```yaml
    apiVersion: source.toolkit.fluxcd.io/v1beta2
    kind: HelmRepository
    metadata:
    name: tf-controller
    namespace: flux-system
    spec:
    interval: 10m
    url: https://weaveworks.github.io/tf-controller/
    ---
    apiVersion: helm.toolkit.fluxcd.io/v2beta1
    kind: HelmRelease
    metadata:
    name: tf-controller
    namespace: flux-system
    spec:
    chart:
        spec:
        chart: tf-controller
        version: "0.9.2"
        sourceRef:
            kind: HelmRepository
            name: tf-controller
            namespace: flux-system
    interval: 10m0s
    install:
        crds: CreateReplace
        remediation:
        retries: 3
    upgrade:
        crds: CreateReplace
    ```


#### WGE

Update the following manifest to your context.

??? example "Expand to see file contents"

    ```yaml
    ---
    apiVersion: v1
    data:
    deploy-key: <changeme>
    entitlement: <changeme>
    password: <changeme>
    username: <changeme>
    kind: Secret
    metadata:
    labels:
        kustomize.toolkit.fluxcd.io/name: shared-secrets
        kustomize.toolkit.fluxcd.io/namespace: flux-system
    name: weave-gitops-enterprise-credentials
    namespace: flux-system
    type: Opaque
    ---
    apiVersion: v1
    data:
    password: <changeme>
    username: <changeme>
    kind: Secret
    metadata:
    labels:
        kustomize.toolkit.fluxcd.io/name: enterprise
        kustomize.toolkit.fluxcd.io/namespace: flux-system
    name: cluster-user-auth
    namespace: flux-system
    type: Opaque
    ---
    apiVersion: source.toolkit.fluxcd.io/v1beta2
    kind: HelmRepository
    metadata:
    name: weave-gitops-enterprise-charts
    namespace: flux-system
    spec:
    interval: 10m
    secretRef:
        name: weave-gitops-enterprise-credentials
    url: https://charts.dev.wkp.weave.works/releases/charts-v3
    ---
    apiVersion: helm.toolkit.fluxcd.io/v2beta1
    kind: HelmRelease
    metadata:
    name: weave-gitops-enterprise
    namespace: flux-system
    spec:
    chart:
        spec:
        chart: mccp
        version: "0.10.2"
        sourceRef:
            kind: HelmRepository
            name: weave-gitops-enterprise-charts
            namespace: flux-system
    interval: 10m0s
    install:
        crds: CreateReplace
        remediation:
        retries: 3
    upgrade:
        crds: CreateReplace
    values:
        global:
        capiEnabled: true
        enablePipelines: true
        enableTerraformUI: true
        clusterBootstrapController:
        enabled: true
        cluster-controller:
        controllerManager:
            kubeRbacProxy:
            image:
                repository: gcr.io/kubebuilder/kube-rbac-proxy
                tag: v0.8.0
            manager:
            image:
                repository: docker.io/weaveworks/cluster-controller
                tag: v1.4.1
        policy-agent:
        enabled: true
        image: weaveworks/policy-agent
        pipeline-controller:
        controller:
            manager:
            image:
                repository: ghcr.io/weaveworks/pipeline-controller
        images:
        clustersService: docker.io/weaveworks/weave-gitops-enterprise-clusters-service:v0.10.2
        uiServer: docker.io/weaveworks/weave-gitops-enterprise-ui-server:v0.10.2
        clusterBootstrapController: weaveworks/cluster-bootstrap-controller:v0.4.0
    ```


At this stage you should have a local management cluster with Weave GitOps Enterprise installed.

```bash
➜ kubectl get pods -A
NAMESPACE                           NAME                                                              READY   STATUS    RESTARTS      AGE
...
flux-system                         weave-gitops-enterprise-cluster-controller-6f8c69dc8-tq994        2/2     Running   5 (12h ago)   13h
flux-system                         weave-gitops-enterprise-mccp-cluster-bootstrap-controller-cxd9c   2/2     Running   0             13h
flux-system                         weave-gitops-enterprise-mccp-cluster-service-8485f5f956-pdtxw     1/1     Running   0             12h
flux-system                         weave-gitops-enterprise-pipeline-controller-85b76d95bd-2sw7v      1/1     Running   0             13h
...
```

You can observe the installed Helm Charts with `kubectl`:

```bash
kubectl get helmcharts.source.toolkit.fluxcd.io
NAME                                  CHART           VERSION   SOURCE KIND      SOURCE NAME                      AGE   READY   STATUS
flux-system-cert-manager              cert-manager    0.0.7     HelmRepository   weaveworks-charts                13h   True    pulled 'cert-manager' chart with version '0.0.7'
flux-system-tf-controller             tf-controller   0.9.2     HelmRepository   tf-controller                    13h   True    pulled 'tf-controller' chart with version '0.9.2'
flux-system-weave-gitops-enterprise   mccp            v0.10.2   HelmRepository   weave-gitops-enterprise-charts   13h   True    pulled 'mccp' chart with version '0.10.2'
```

As well as the container images:

```bash

kubectl get pods --all-namespaces -o jsonpath="{.items[*].spec['containers','initContainers'][*].image}" |tr -s '[[:space:]]' '\n' \
| sort | uniq | grep -vE 'kindest|etcd|coredns'

docker.io/prom/prometheus:v2.34.0
docker.io/weaveworks/cluster-controller:v1.4.1
docker.io/weaveworks/weave-gitops-enterprise-clusters-service:v0.10.2
docker.io/weaveworks/weave-gitops-enterprise-ui-server:v0.10.2
ghcr.io/fluxcd/flagger-loadtester:0.22.0
ghcr.io/fluxcd/flagger:1.21.0
ghcr.io/fluxcd/helm-controller:v0.23.1
ghcr.io/fluxcd/kustomize-controller:v0.27.1
ghcr.io/fluxcd/notification-controller:v0.25.2
...
```

## Collect and Publish Artifacts

This section guides you to push installed artifacts to your private registry.
Here's a Makefile to help you with each stage:

??? example "Expand to see Makefile contents"

    ```makefile
        .PHONY := all

        #set these variable with your custom configuration
        PRIVATE_HELM_REPO_NAME=private
        REGISTRY=localhost:5001
        WGE_VERSION=0.10.2

        WGE=mccp-$(WGE_VERSION)
        WGE_CHART=$(WGE).tgz

        all: images charts

        charts: pull-charts push-charts

        images:
            kubectl get pods --all-namespaces -o jsonpath="{.items[*].spec['containers','initContainers'][*].image}" \
            |tr -s '[[:space:]]' '\n' | sort | uniq | grep -vE 'kindest|kube-(.*)|etcd|coredns' | xargs -L 1 -I {} ./image-sync.sh {} $(REGISTRY)
            kubectl get microvmmachinetemplates --all-namespaces -o jsonpath="{.items[*].spec.template.spec.kernel.image}"|tr -s '[[:space:]]' '\n' \
            | sort | uniq | xargs -L 1 -I {} ./image-sync.sh {} $(REGISTRY)

        pull-charts:
            curl -L https://s3.us-east-1.amazonaws.com/weaveworks-wkp/releases/charts-v3/$(WGE_CHART) --output  $(WGE_CHART)

        push-charts:
            helm cm-push -f $(WGE_CHART) $(PRIVATE_HELM_REPO_NAME)
    ```

The `image-sync.sh` referenced in the `images` target of the the above Makefile
is similar to:

```shell
skopeo copy docker://$1 docker://$2/$1 --preserve-digests --multi-arch=all
```

>[Skopeo](https://github.com/containers/skopeo) allows you to configure a range a security features to meet your requirements.
For example, configuring trust policies before pulling or signing containers before making them available in your private network.
Feel free to adapt the previous script to meet your security needs.

1. Configure the environment variables to your context.
2. Execute `make` to automatically sync Helm charts and container images.

```bash
➜  resources git:(docs-airgap-install) ✗ make
kubectl get microvmmachinetemplates --all-namespaces -o jsonpath="{.items[*].spec.template.spec.kernel.image}"|tr -s '[[:space:]]' '\n' \
	| sort | uniq | xargs -L 1 -I {} ./image-pull-push.sh {} docker-registry:5000

5.10.77: Pulling from weaveworks-liquidmetal/flintlock-kernel
Digest: sha256:5ef5f3f5b42a75fdb69cdd8d65f5929430f086621e61f00694f53fe351b5d466
Status: Image is up to date for ghcr.io/weaveworks-liquidmetal/flintlock-kernel:5.10.77
ghcr.io/weaveworks-liquidmetal/flintlock-kernel:5.10.77
...5.10.77: digest: sha256:5ef5f3f5b42a75fdb69cdd8d65f5929430f086621e61f00694f53fe351b5d466 size: 739
```

## Airgap Install

### Weave GitOps Enterprise
At this stage you have in your private registry both the Helm charts and container images required to install Weave GitOps
Enterprise. Now you are ready to install WGE from your private registry.

Follow the instructions to install WGE with the following considerations:

1. Adjust Helm Releases `spec.chart.spec.sourceRef` to tell Flux to pull Helm charts from your Helm repo.
2. Adjust Helm Releases `spec.values` to use the container images from your private registry.

An example of how it would look for Weave GitOps Enterprise is shown below.

??? example "Expand to view example WGE manifest"

    ```yaml title="weave-gitops-enterprise.yaml" 
    ---
    apiVersion: source.toolkit.fluxcd.io/v1beta2
    kind: HelmRepository
    metadata:
    name: weave-gitops-enterprise-charts
    namespace: flux-system
    spec:
    interval: 1m
    url: http://helm-repo-chartmuseum:8080
    ---
    apiVersion: helm.toolkit.fluxcd.io/v2beta1
    kind: HelmRelease
    metadata:
    name: weave-gitops-enterprise
    namespace: flux-system
    spec:
    chart:
        spec:
        chart: mccp
        version: "0.10.2"
        sourceRef:
            kind: HelmRepository
            name: weave-gitops-enterprise-charts
            namespace: flux-system
    interval: 1m0s
    install:
        crds: CreateReplace
        remediation:
        retries: 3
    upgrade:
        crds: CreateReplace
    values:
        global:
        capiEnabled: true
        enablePipelines: true
        enableTerraformUI: true
        clusterBootstrapController:
        enabled: true
        #images changed
        cluster-controller:
        controllerManager:
            kubeRbacProxy:
            image:
                repository: localhost:5001/gcr.io/kubebuilder/kube-rbac-proxy
                tag: v0.8.0
            manager:
            image:
                repository: localhost:5001/docker.io/weaveworks/cluster-controller
                tag: v1.4.1
        policy-agent:
        enabled: true
        image: localhost:5001/weaveworks/policy-agent
        pipeline-controller:
        controller:
            manager:
            image:
                repository: localhost:5001/ghcr.io/weaveworks/pipeline-controller
        images:
        clustersService: localhost:5001/docker.io/weaveworks/weave-gitops-enterprise-clusters-service:v0.10.2
        uiServer: localhost:5001/docker.io/weaveworks/weave-gitops-enterprise-ui-server:v0.10.2
        clusterBootstrapController: localhost:5001/weaveworks/cluster-bootstrap-controller:v0.4.0
    ```

###  Cluster API

Indicate in the Cluster API configuration file `clusterctl.yaml` that you want to use images from the private repo
by leveraging [image overrides](https://cluster-api.sigs.k8s.io/clusterctl/configuration.html#image-overrides).

```yaml
images:
  all:
    repository: localhost:5001/registry.k8s.io/cluster-api
  infrastructure-microvm:
    repository: localhost:5001/ghcr.io/weaveworks-liquidmetal
```
Then execute `make clusterctl-init` to init capi using your private registry.
