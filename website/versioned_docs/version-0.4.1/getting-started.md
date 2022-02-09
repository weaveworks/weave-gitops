---
sidebar_position: 2
---

# Getting Started

Let's get an app deployed with **Weave GitOps**! We'll show you how to get a simple application running in a kind cluster, then make a change to the application in Git, and see it automatically update on the cluster.

You can also join our regular workshops where we go through these steps and help you along the way:
- [Join our user group](https://www.meetup.com/Weave-User-Group/)
---

## Pre-requisites

To follow along with this guide you will need:
1. A GitHub account
2. kubectl installed - [instructions](https://kubernetes.io/docs/tasks/tools/#kubectl)
3. For our Kubernetes cluster we use [kind](https://kind.sigs.k8s.io/docs/user/quick-start/) which requires [Docker](https://docs.docker.com/get-docker/).

## Install the Weave GitOps CLI

```console
curl -L "https://github.com/weaveworks/weave-gitops/releases/download/v0.4.1/gitops-$(uname)-$(uname -m)" -o gitops
chmod +x gitops
sudo mv ./gitops /usr/local/bin/gitops
gitops version
```

For complete installation instructions and general pre-requisites, see the  [Installation page](installation.md).

## Prepare your cluster

### 1) Create the cluster

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

### 2 - Install Weave GitOps onto the cluster

```console
gitops install
```

The install will take roughly 1-2 minutes depending on your system. 

Once complete, you will see:

```
✔ image-reflector-controller: deployment ready
✔ image-automation-controller: deployment ready
✔ source-controller: deployment ready
✔ kustomize-controller: deployment ready
✔ helm-controller: deployment ready
✔ notification-controller: deployment ready
✔ install finished
```
## Configure Weave GitOps to deploy your application

### 3 - Fork the Podinfo repository

For Weave GitOps to set up automation to reconcile the workload, we will need write access to the repository, so we will first fork the podinfo sample repository.

Go to [https://github.com/wego-example/podinfo-deploy](https://github.com/wego-example/podinfo-deploy).

![fork](/img/github-fork.png)

### 4 - Start the GitOps Dashboard web UI

```
gitops ui run
```

This will open the dashboard in your browser at [http://0.0.0.0:9001/](http://0.0.0.0:9001/).

You will see an empty Applications view as shown in the image below.

![GitOps Dashboard web UI - empty applications view](/img/dashboard-applications-empty.png)

### 5 - Add the podinfo application

Click **add application** in the top right of the screen to bring up the following form:

![GitOps Dashboard web UI - add applications form](/img/dashboard-add-application.png)

Next fill out the form with the required values:

- Name: **podinfo-deploy**
- Kubernetes Namespace: **wego-system**  
  -  **Leave as default**, this is where the automation objects for the application will be created.
- Source Repo URL: (**your-forked-podinfo-repo**)
  - For example: `https://github.com/sympatheticmoose/podinfo-deploy` or `git@github.com:sympatheticmoose/podinfo-deploy`
- Config Repo URL: (**leave blank**)  
  - It is **recommended** to use a separate repository as your config repo when deploying multiple applications, potentially to multiple clusters, however, for simplicity - you can also use the same repo as the `Source Repo URL`.
- Path: **./**  
  - This allows you to specify a particular folder with a repository, should the repo contain more than a single application's deployment manifests.
- Branch: (**should not require changing**)

Click **Submit** in the bottom right of the form. This will result in an error as we have not recently authenticated with GitHub, so we are not yet authorized to raise a pull request against your repository.

Click **Authenticate with GitHub** within the error message as shown below:

![Not authenticated error](/img/dashboard-add-application-auth-error.png)

You will be prompted with a screen to copy a code from, before visiting GitHub to authorize Weave GitOps to write to your repositories. Copy the code to your clipboard and click **AUTHORIZE GITHUB ACCESS**.

Paste the code into the screen as shown below:
![device flow activation](/img/github-device-flow-start.png)

You should then see this confirmation:

![device flow complete](/img/github-device-flow-complete.png)

Once authorization has completed, navigate back to the GitOps Dashboard.

The screen should update with a message that the Pull Request has been created, click **Open Pull Requests** to review the PR in GitHub.

![Pull request raised](/img/dashboard-add-application-PR-raised.png)

The Pull Request adds three files which will be deployed to your cluster:
- An **Application custom resource** with details about the deployment.
- A [**GitRepository**](https://fluxcd.io/docs/concepts/#sources) source which "defines the origin of a repository containing the desired state of the system and the requirements to obtain it". This includes the `interval` which is how frequently to check for available new versions.
- A [**Kustomization**](https://fluxcd.io/docs/concepts/#kustomization) which "represents a local set of Kubernetes resources (e.g. kustomize overlay) that Flux is supposed to reconcile in the cluster". This deploys the resources found under the specified path, reconciling between the cluster and the declared state in Git. Where "if you make any changes to the cluster using kubectl edit/patch/delete, they will be promptly reverted." based on the `interval` value. Note you can pause this reconciliation process using `gitops suspend <app-name>`.

Merge the Pull Request to start the deployment.

![Merge](/img/podinfo-pr-merge.png)

## View the running application

### 6 - See application details

As the workloads begin to be deployed, you can view the progress and check for key reconciliation status updates in the Application details page.

Navigate to the Applications view by clicking in the left menu bar, you should now see the `podinfo-deploy` application listed. Click the name of the Application to view the details page. You may need to refresh the page to view up to date status.

![Weave GitOps UI](/img/wego_ui.png)

From the Application details page you can see the reconciled objects on your cluster, which specific commit was last fetched from Git and which was last deployed onto the cluster. 

You can also see the 10 most recent commits to your repository to quickly understand the changes which have occurred, with hyperlinks back to GitHub so you can find more details or revert changes as necessary.

As you can see, you have successfully deployed the app!

### 7 - Access the running application

To access the `podinfo` UI you can set up a port forward into the pod.
```console
kubectl port-forward service/frontend 9898:9898 --namespace test
```
```
Forwarding from 127.0.0.1:9898 -> 9898
Forwarding from [::1]:9898 -> 9898
```

Now you can browse [http://localhost:9898](http://localhost:9898)

You should see something like:

![Podinfo](/img/podinfo-web.png)

Use CTRL+C to cancel the `kubectl port-forward` command to continue with your command prompt.

## GitOps reconciliation in action

The real aim of GitOps is not just to deploy once, but to continuously reconcile desired state in Git with live state in Kubernetes. So we will now show a change.

### 8 - Bad actor time, delete your application.

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
