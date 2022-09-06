---
title: CI/CD with Weave GitOps and Tekton
sidebar_position: 8
---

In this guide we will show you how to create a full CI/CD solution using Weave GitOps and Tekton.

### Pre-requisites
- A Kubernetes cluster such as [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/) cluster running a 
[Flux-supported version of Kubernetes](https://fluxcd.io/docs/installation/#prerequisites)
- Weave GitOps is [installed](../installation.mdx)

## Install Tekton
In a location being reconciled by Flux, create a new `tekton` directory and add following `kustomization.yaml` file.
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- https://storage.googleapis.com/tekton-releases/pipeline/previous/v0.38.3/release.yaml
- https://storage.googleapis.com/tekton-releases/triggers/previous/v0.20.2/release.yaml
- https://storage.googleapis.com/tekton-releases/triggers/previous/v0.20.2/interceptors.yaml
```
This will install the core Tekton pipeline components as well as the trigger components necessary to kickoff pipeline runs in an automated fasion.

### Optional Tools
- [Tekton CLI](https://tekton.dev/docs/cli)
- [Tekton Dashboard](https://tekton.dev/docs/dashboard/install)
> You do not need these for the pipelines to run, but they are ***highly recommended*** and are very useful in viewing pipelines and debugging.

## Create CI Pipeline
### Define Tasks
To create the pipline, we want to first define our Tasks.  From the [Tekton Tasks docs](https://tekton.dev/docs/pipelines/tasks), "A Task is a collection of Steps that you define and arrange in a specific order of execution as part of your continuous integration flow."  Each Step references a container image that will to used to run the Step command.  This image needs to contain all the tools necessary for the Step to complete successfully.  Here is an example Task definition:

```yaml
apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: helm-release
  namespace: default
spec:
  workspaces:
    - name: source
    - name: gh-pages
  params:
    - name: chart-dir
      type: string
    - name: version
      type: string
    - name: repo-url
      type: string
    - name: release-message
      type: string
      default: release new helm chart
    - name: GIT_USER_EMAIL
      type: string
      default: <>
    - name: GIT_USER_NAME
      type: string
      default: tekton-automation
  steps:
    - name: package
      image: alpine/helm
      workingDir: $(workspaces.source.path)
      script: |
        #!/usr/bin/env sh
        helm package $(params.chart-dir) --version $(params.version) --app-version $(params.version)
    - name: index
      image: alpine/helm
      workingDir: $(workspaces.gh-pages.path)
      script: |
        #!/usr/bin/env sh
        cp $(workspaces.source.path)/*.tgz .
        helm repo index --url $(params.repo-url) --merge index.yaml .
    - name: release
      image: alpine/git
      workingDir: $(workspaces.gh-pages.path)
      script: |
        #!/usr/bin/env sh
        set -e
        git config --global user.email "$(params.GIT_USER_EMAIL)"
        git config --global user.name "$(params.GIT_USER_NAME)"
        git config --global --add safe.directory /workspace/gh-pages
        git add .
        git commit -m "$(params.release-message)"
        git push
```

This Task is used to create a new Helm Chart version and publish it to Helm Repository hosted on Github Pages.  It is made up of 3 Steps the will sequentially package, index, and release a new chart.

Tekton also provides some ready to use Tasks via the [Tekton Hub](https://hub.tekton.dev).  Before creating your own Task checkout the hub first to see if one of these Tasks will fit your needs.  Some good examples are the [git clone](https://hub.tekton.dev/tekton/task/git-clone) and [kaniko](https://hub.tekton.dev/tekton/task/kaniko) Tasks (both of which we will be using later in this example).

### Define Pipeline
Now that we have our Tasks defined it is time to group them together into a [Pipeline](https://tekton.dev/docs/pipelines/pipelines).

# TODO - update refs
```yaml
apiVersion: tekton.dev/v1beta1
kind: Pipeline
metadata:
  name: sample-release-pipeline
  namespace: default
spec:
  params:
    - name: git-url
      type: string
    - name: git-revision
      type: string
    - name: image-name
      type: string
    - name: tag
      type: string
      default: latest
    - name: path-to-image-context
      type: string
      default: ./
    - name: path-to-dockerfile
      type: string
      default: ./Dockerfile
    - name: gh-pages-branch
      type: string
      default: gh-pages
  workspaces:
    - name: shared-data
    - name: docker-credentials
  tasks:
    - name: clone-source
      taskRef:
        name: git-clone
      params:
        - name: url
          value: $(params.git-url)
        - name: revision
          value: $(params.git-revision)
      workspaces:
        - name: output
          workspace: shared-data
          subPath: source # keep source data separate from rest of data
    - name: clone-gh-pages
      taskRef:
        name: git-clone
      params:
        - name: url
          value: $(params.git-url)
        - name: revision
          value: $(params.gh-pages-branch)
      workspaces:
        - name: output
          workspace: shared-data
          subPath: gh-pages # keep gh-pages data separate from rest of data
    - name: build-image
      taskRef:
        name: kaniko
      params:
        - name: IMAGE
          value: $(params.image-name):$(params.tag)
        - name: CONTEXT
          value: $(params.path-to-image-context)
        - name: DOCKERFILE
          value: $(params.path-to-dockerfile)
      runAfter:
        - clone-source
      workspaces:
        - name: source
          workspace: shared-data
          subPath: source
        - name: dockerconfig
          workspace: docker-credentials
    - name: helm-release
      taskRef:
        name: helm-release
      params:
        - name: chart-dir
          value: charts/demo-chart # param values can be variablized or hard coded like this
        - name: version
          value: $(params.tag)
        - name: repo-url
          value: $(params.git-url)
      runAfter:
        - build-image
      workspaces:
        - name: source
          workspace: shared-data
          subPath: source
        - name: gh-pages
          workspace: shared-data
          subPath: gh-pages
```

### Create Triggers
Now that we have our Pipeline defined we want to create a [PipelineRun](https://tekton.dev/docs/pipelines/pipelineruns) to execute the Pipeline.  You can do this manually by creating a new PipelineRun resource, but that is not GitOpsy and can be prone to misconfiguration.  Instead we are going to use a combination of [TriggerTemplates](https://tekton.dev/docs/triggers/triggertemplates) and [EventListeners](https://tekton.dev/docs/triggers/eventlisteners) to kickoff the Pipeline automatically.  First lets look at TriggerTemplate.

```yaml
apiVersion: triggers.tekton.dev/v1beta1
kind: TriggerTemplate
metadata:
  name: sample-release-pipeline-trigger-template
  namespace: default
spec:
  params: # any params that need to be passed into the pipeline must also be declared here
    - name: git-url
    - name: git-revision
    - name: destination-git-url
      default: https://github.com/rparmer/tekton-pipeline-environments.git
    - name: destination-git-full-name
      default: rparmer/tekton-pipeline-environments
    - name: image-name
      default: ghcr.io/rparmer/tekton
    - name: tag
      default: latest
  resourcetemplates:
    - apiVersion: tekton.dev/v1beta1
      kind: PipelineRun
      metadata:
        generateName: sample-release-pipeline-run- # create unique name automatically
      spec:
        serviceAccountName: build-bot
        pipelineRef:
          name: sample-release-pipeline
        params:
          - name: source-git-url
            value: $(tt.params.git-url)
          - name: source-git-revision
            value: $(tt.params.git-revision)
          - name: destination-git-url
            value: $(tt.params.destination-git-url)
          - name: destination-git-full-name
            value: $(tt.params.destination-git-full-name)
          - name: image-name
            value: $(tt.params.image-name)
          - name: tag
            value: $(tt.params.tag)
          - name: path-to-image-context
            value: demo
        workspaces:
          - name: shared-data # pvc volume definition (used to share data between tasks)
            volumeClaimTemplate:
              spec:
                accessModes:
                  - ReadWriteOnce
                resources:
                  requests:
                    storage: 1Gi
          - name: docker-credentials # kubernetes secret reference
            secret:
              secretName: ghcr-credentials
---
apiVersion: triggers.tekton.dev/v1beta1
kind: TriggerBinding
metadata:
  name: github-tag-binding
spec:
  params:
    - name: tag
      value: $(extensions.tag) # extract value from EventListener cel overlay
    - name: git-revision
      value: $(body.head_commit.id) # extract value from event payload
    - name: git-url
      value: $(body.repository.clone_url)
```

### Create EventListener
Now we are going to create an EventListener that we'll configure to listen to Github events.  By default the EventListener is not exposed outside of the cluster.  To overcome this we will add a custom Ingress resource so that our listener is reachable from Github.

> Make sure to replace `<YOUR URL HERE>` with your true ingress url.  You may need to add additional configuration to match your environment needs.

```yaml
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: tekton-triggers-sa
  namespace: default
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: triggers-eventlistener-binding
  namespace: default
subjects:
- kind: ServiceAccount
  name: tekton-triggers-sa
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: tekton-triggers-eventlistener-roles
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: triggers-eventlistener-clusterbinding
subjects:
- kind: ServiceAccount
  name: tekton-triggers-sa
  namespace: default
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: tekton-triggers-eventlistener-clusterroles
---
apiVersion: v1
kind: Secret
metadata:
  name: tekton-github-secret
  namespace: default
type: Opaque
stringData:
  secretToken: "1234567"
---
apiVersion: triggers.tekton.dev/v1beta1
kind: EventListener
metadata:
  name: tekton-listener
  namespace: default
spec:
  triggers:
    - name: tag-push-events
      interceptors:
        - ref:
            name: github
          params:
            - name: secretRef
              value:
                secretName: tekton-github-secret
                secretKey: secretToken
            - name: eventTypes
              value: ["push"]
        - name: "only on tag creation"
          ref:
            name: cel
          params:
            - name: filter
              value: >
                body.ref.startsWith('refs/tags') &&
                body.created == true
            - name: overlays
              value:
                - key: tag
                  expression: "body.ref.split('/')[2]"
      bindings:
        - ref: github-tag-binding
      template:
        ref: sample-release-pipeline-trigger-template
  resources:
    kubernetesResource:
      spec:
        template:
          spec:
            serviceAccountName: tekton-triggers-sa
            containers:
              - resources:
                  requests:
                    memory: 64Mi
                    cpu: 250m
                  limits:
                    memory: 128Mi
                    cpu: 500m
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: tekton-el
  namespace: default
spec:
  rules:
  - host: <YOUR URL HERE>
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: el-tekton-listener
            port:
              name: http-listener
```

After this is created you'll need to create a webhook on your GitHub repo to call the EventListener url.  Details on how to create a webhook are availble in the [GitHub Docs](https://docs.github.com/en/developers/webhooks-and-events/webhooks/creating-webhooks).  Make sure to set the content type to `application/json` and for this walkthrough you only trigger on `push` events.

## Testing
At this point you should have all the pieces in place for the pipeline to trigger and run successfully.  So lets try it out.  Create a new tag for your chart and push it to the repository.

```bash
git tag v0.0.1
git push --tags
```

The Tekton pipeline should have started.  If you installed the Tekton Dashboard, navigate to the UI and you can monitor the pipeline progress.  If you installed the Tekton-CLI, run `tkn pipelinerun logs -f` to monitor the logs of each task.  If you do not have either installed, you can watch the pods being created and monitor the logs of each one.  But I would recommend using either the Tekton-CLI or Dashboard instead.

Once the pipeline is complete you should see the new tagged version of your chart available in your Helm repository.
