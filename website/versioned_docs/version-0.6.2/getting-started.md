---
sidebar_position: 2
---

# Getting Started

**Weave GitOps** makes it easy to get started on your GitOps journey; deploying your Kubernetes workloads in a way which is:

- **Declarative** - desired state is expressed declaratively
- **Versioned and immutable** - including retaining complete version history
- **Pulled Automatically** - software agents automatically pull the desired state declarations from the source
- **Continuously reconciled** - software agents continuously observe actual system state and attempt to apply the desired state

_See GitOps Principles v1.0.0 - [OpenGitOps](https://opengitops.dev/)_

You can learn more about [What is GitOps?](https://www.weave.works/technologies/gitops/) and [Weave GitOps](https://www.weave.works/product/gitops-core/) at [weave.works](https://weave.works).

Now, let's get an app deployed with **Weave GitOps**! This guide will walk you through a complete end-to-end scenario from creating a fresh local Kubernetes cluster, to having an application workload being deployed and managed through GitOps.

We will:

- Create a new Kubernetes cluster
- Install the Weave GitOps CLI
- Deploy a sample application with Weave GitOps
- Show GitOps reconciliation by attempting a change to the app running on the cluster, then pushing a change to the application through git

### Workshops

You can also join our regular workshops where we go through these steps together and can help you along the way:

- [Join our user group](https://www.meetup.com/Weave-User-Group/)

---

## Pre-requisites

To follow along with this guide you will need:

1. A GitHub account
2. kubectl installed - [instructions](https://kubernetes.io/docs/tasks/tools/#kubectl)
3. For our Kubernetes cluster we use [kind](https://kind.sigs.k8s.io/docs/user/quick-start/) which requires [Docker](https://docs.docker.com/get-docker/).

## Install the Weave GitOps CLI

```console
curl --silent --location "https://github.com/weaveworks/weave-gitops/releases/download/v0.6.2/gitops-$(uname)-$(uname -m).tar.gz" | tar xz -C /tmp
sudo mv /tmp/gitops /usr/local/bin
gitops version
```

For complete installation instructions and general pre-requisites, see the [Installation page](installation.md).

## Create your GitHub Repositories

Weave GitOps can be used to deploy multiple applications, each with deployment manifests in their own separate repositories, to multiple Kubernetes clusters. We store the [GitOps Automation](gitops-automation.md) for one or more clusters in a single configuration repository for ease of management.

Whilst you can add the automation to any existing repository, including one with your application deployment manifests, we recommend using a new or empty repository for this purpose and our guide will take this approach.

### 1 - Create a configuration repository

This is where we will push GitOps Automation manifests. Storing in git allows for easy recovery and cluster bootstrapping.

From [GitHub](https://github.com) click "+" and create a new repository; be sure to initialize the repository. This guide assumes the name `gitops-config` is used.

### 2 - Fork the podinfo sample application repository

We will be making changes to the sample application to show GitOps reconciliation in action. So we will first fork the podinfo sample repository.

Go to [https://github.com/wego-example/podinfo-deploy](https://github.com/wego-example/podinfo-deploy) and fork the repository.

![fork](/img/github-fork.png)

`Podinfo` is a simple web application written in Go made up of a frontend and backend component; it is designed to showcase best practices of running microservices in Kubernetes. The full application source can be found [here](https://github.com/stefanprodan/podinfo).

```
.
├── README.md
├── backend
│   ├── deployment.yaml
│   ├── hpa.yaml
│   └── service.yaml
├── frontend
│   ├── deployment.yaml
│   └── service.yaml
└── namespace.yaml
2 directories, 7 files
```

## Prepare your cluster

### 3 - Create the cluster

Open a terminal and enter the following:

```console
kind create cluster
```

You should see output similar to:

```
Creating cluster "kind" ...
 ✓ Ensuring node image (kindest/node:v1.21.1) 🖼
 ✓ Preparing nodes 📦
 ✓ Writing configuration 📜
 ✓ Starting control-plane 🕹️
 ✓ Installing CNI 🔌
 ✓ Installing StorageClass 💾
Set kubectl context to "kind-kind"
You can now use your cluster with:
kubectl cluster-info --context kind-kind

Have a nice day! 👋

```

### 4 - Install Weave GitOps on the Kubernetes cluster

```console
gitops install --config-repo git@github.com:<username>/gitops-config
```

Run the install command specifying the location of your configuration repository created in step 1.

The install will take roughly 2 minutes depending on your system.

Once complete, you will see:

```
...
◎ verifying installation
✔ image-reflector-controller: deployment ready
✔ image-automation-controller: deployment ready
✔ source-controller: deployment ready
✔ kustomize-controller: deployment ready
✔ helm-controller: deployment ready
✔ notification-controller: deployment ready
✔ install finished
Deploy key generated and uploaded to git provider
► Writing manifests to disk
► Committing and pushing gitops updates for application
► Pushing app changes to repository
► Applying manifests to the cluster
```

This will commit a new `.weave-gitops` folder to your configuration repository with the following files to manage the Weave GitOps runtime on the specified cluster:

```
.
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
            └── .keep
```

- `flux-source-resource`: a [**GitRepository**](https://fluxcd.io/docs/concepts/#sources) source which "defines the origin of a repository containing the desired state of the system and the requirements to obtain it". This includes the `interval` which is how frequently to check for available new versions.
- `flux-system-kustomization-resource`: a [**Flux Kustomization**](https://fluxcd.io/docs/concepts/#kustomization) which "represents a local set of Kubernetes resources (e.g. kustomize overlay) that Flux is supposed to reconcile in the cluster". This deploys the resources found under the specified path, in this case the `/system` folder, reconciling between the cluster and the declared state in Git. Where "if you make any changes to the cluster using kubectl edit/patch/delete, they will be promptly reverted." based on the `interval` value.
- `flux-user-kustomization-resource`: another [**Flux Kustomization**](https://fluxcd.io/docs/concepts/#kustomization), this time for anything in the `/user` folder which later in this guide will include a reference to our sample application.
- `gitops-runtime`: which creates the `wego-system` namespace and deploys the Flux runtime.
- `wego-app`: which deploys our on-cluster web UI (not currently exposed).
- `wego-system`: which creates our Application Custom Resource Definition (CRD).

To learn more about these files, see our [GitOps Automation](gitops-automation.md).

## Configure Weave GitOps to deploy the podinfo application

### 5 - Start the GitOps Dashboard web UI

Weave GitOps provides a web UI to help manage the lifecycle management of your applications.

```
gitops ui run
```

Running the above command will open the dashboard in your browser at [http://0.0.0.0:9001/](http://0.0.0.0:9001/).

You will see an empty Applications view as shown in the image below.

![GitOps Dashboard web UI - empty applications view](/img/dashboard-applications-empty.png)

### 6 - Add the podinfo application

Click **add application** in the top right of the screen to bring up the following form:

![GitOps Dashboard web UI - add applications form](/img/dashboard-add-form-complete.png)

Next fill out the form with the required values:

- Name: **podinfo-deploy**
- Kubernetes Namespace: **wego-system**
  - **Leave as default**, this is where the GitOps Automation objects for the application will be deployed.
- Source Repo URL: (**\*git@github.com:\<username\>/podinfo-deploy**)
  - URL references can be in either the HTTPS `https://github.com/sympatheticmoose/podinfo-deploy` or SSH `git@github.com:sympatheticmoose/podinfo-deploy` format.
- Config Repo URL: (**git@github.com:\<username\>/gitops-config**)
  - Specify the same configuration repository from steps 1 and 4 (install).
- Path: **./**
  - **Leave as default**, this allows you to specify a particular folder with a repository, should the repo contain more than a single application's deployment manifests.
- Branch: (**should not require changing**)

Click **AUTHENTICATE WITH GITHUB** next to the `Source Repo URL` field.

You will be prompted with a screen to copy a code from, before visiting GitHub to authorize Weave GitOps with a short-lived token to write to your repositories. Copy the code to your clipboard and click **AUTHORIZE GITHUB ACCESS**.

Paste the code into the screen as shown below:
![device flow activation](/img/github-device-flow-start.png)

You should then see this confirmation:

![device flow complete](/img/github-device-flow-complete.png)

Once authorization has completed, navigate back to the GitOps Dashboard.

Click **Submit** in the bottom right of the form.

The screen should update with a message that the Pull Request has been created, click **Open Pull Requests** to review the PR in GitHub.

![Pull request raised](/img/dashboard-add-application-PR-raised.png)

The Pull Request adds five additional files under the `.weave-gitops` folder in your configuration repository. In a new `apps` top level folder the following four files are added:

- `app.yaml`: An **Application custom resource** with details about the deployment.
- `kustomization.yaml`: a [**Kubernetes Kustomization**](https://kubectl.docs.kubernetes.io/references/kustomize/glossary/#kustomization) which references the resources within this folder to deploy the workload.
- `podinfo-deploy-gitops-source`: a [**GitRepository**](https://fluxcd.io/docs/concepts/#sources) (described in step 4) for the application source repository.
- `podinfo-deploy-gitops-deploy`: a [**Flux Kustomization**](https://fluxcd.io/docs/concepts/#kustomization) (described in step 4) to deploy the application. Note you can pause the reconciliation process using `gitops suspend <app-name>` for i.e. debugging.

Then in the `../clusters/kind-kind/user` folder:

- Another [**Kubernetes Kustomization**](https://kubectl.docs.kubernetes.io/references/kustomize/glossary/#kustomization) which associates the `podinfo-deploy` application with your kind cluster as a target for its deployment.

So now our `.weave-gitops` folder looks like:

```
.
├── apps
│   └── podinfo-deploy
│       ├── app.yaml
│       ├── kustomization.yaml
│       ├── podinfo-deploy-gitops-deploy.yaml
│       └── podinfo-deploy-gitops-source.yaml
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
            ├── .keep
            └── kustomization.yaml
```

Merge the Pull Request to start the deployment.

![Merge](/img/podinfo-pr-merge.png)

## View the running application

### 7 - See application details

As the workloads begin to be deployed, you can view the progress and check for key reconciliation status updates in the Application details page.

Navigate to the Applications view by clicking in the left menu bar, you should now see the `podinfo-deploy` application listed. Click the name of the Application to view the details page.

![Weave GitOps UI](/img/wego_ui.png)

From the Application details page you can see the reconciled objects on your cluster, which specific commit was last fetched from Git and which was last deployed onto the cluster. The `sync` button allows you to force a reconciliation and bypass the `interval` values. `Remove app` will remove from GitHub repository and then the cluster as GitOps reconciliation takes effect.

You can also see the 10 most recent commits to your application source repository to quickly understand the changes which have occurred, with hyperlinks back to GitHub so you can find more details or revert changes as necessary.

As you can see, you have successfully deployed the app!

### 8 - Access the running application

To access the `podinfo` UI you can set up a port forward into the pod.

```console
kubectl port-forward service/frontend 9898:9898 --namespace test
```

```
Forwarding from 127.0.0.1:9898 -> 9898
Forwarding from [::1]:9898 -> 9898
```

Now you can browse [http://localhost:9898](http://localhost:9898).

You should see something like:

![Podinfo](/img/podinfo-web.png)

Use CTRL+C to cancel the `kubectl port-forward` command to continue with your command prompt.

## GitOps reconciliation in action

The real aim of GitOps is not just to deploy once, but to continuously reconcile desired state in Git with live state in Kubernetes. So we will now show a change.

### 9 - Bad actor time, delete your application.

Let's try deleting the `frontend` deployment and seeing what happens.

First set up a watch on the pods

```
kubectl get pods -n test --watch
```

Now delete the deployment for `frontend`

```
kubectl delete deployment/frontend -n test
```

You will see the pod terminate, but after a short while, a new deployment will be created and a new pod created as drift was detected between our desired state and live state.

```
NAME                        READY   STATUS    RESTARTS   AGE
backend-6b944d8b-hrg7f      1/1     Running   0          2m34s
frontend-64769bf658-b5nv6   1/1     Running   0          34s
frontend-64769bf658-b5nv6   1/1     Terminating   0          68s
frontend-64769bf658-b5nv6   0/1     Terminating   0          72s
frontend-64769bf658-b5nv6   0/1     Terminating   0          72s
frontend-64769bf658-b5nv6   0/1     Terminating   0          72s
frontend-64769bf658-mjjcs   0/1     Pending       0          0s
frontend-64769bf658-mjjcs   0/1     Pending       0          0s
frontend-64769bf658-mjjcs   0/1     ContainerCreating   0          0s
frontend-64769bf658-mjjcs   0/1     Running             0          2s
frontend-64769bf658-mjjcs   1/1     Running             0          11s
```

### 9 - Make a (desired) change to your application

Edit `frontend/deployment.yaml`

Change the `PODINFO_UI_COLOR` to grey:

```yaml
env:
  - name: PODINFO_UI_COLOR
    value: "#888888"
```

You can do this through the GitHub web interface (including via pull request ♥) or clone the repo locally and make the change as shown below:

```
git clone git@github:<username>/podinfo-deploy

# make changes in /frontend/deployment.yaml

git add .
git commit -m "change color"
git push
```

Now wait for the reconciliation to take place and you should see the pods recycle

```console
kubectl get pods --namespace test
```

```
NAME                       READY   STATUS              RESTARTS   AGE
backend-5cd878f8dd-rl64h   1/1     Running             0          33m
frontend-5c45876f-pnxrq    1/1     Running             0          6m51s
frontend-ff74574fc-7ntw4   0/1     ContainerCreating   0          1s
```

And a little later:

```
NAME                       READY   STATUS        RESTARTS   AGE
backend-5cd878f8dd-rl64h   1/1     Running       0          34m
frontend-5c45876f-pnxrq    0/1     Terminating   0          7m9s
frontend-ff74574fc-7ntw4   1/1     Running       0          19s
```

Notice that the backend has not changed and so the backend pod is not affected.

Restart `kubectl port-forward service/frontend 9898:9898 --namespace test` and you will see the color has changed. NB: If you use a real ingress then you wouldn't need to do this.

## Complete!

**Congratulations!** You have achieved a GitOps deployment using Weave GitOps. We hope you continue to be successful and are available for any questions or feedback in our [Slack](https://weave-community.slack.com/archives/C0248LVC719).

### Where next?

- Learn more about GitOps from the Weaveworks [Guide to GitOps](https://www.weave.works/technologies/gitops/)
- Read more about the [GitOps Automation](gitops-automation.md) objects used to reconcile workloads
- Get to know our [CLI](https://docs.gitops.weave.works/docs/cli-reference/gitops)
- Get started with Cluster management with Weave GitOps Enterprise [Weave GitOps Enterprise](https://docs.gitops.weave.works/docs/cluster-management/getting-started)
- Dive into [Flux](https://fluxcd.io/), the proven CNCF technology that provides the foundation for Weave GitOps, to learn about the configurations possible for your deployments
