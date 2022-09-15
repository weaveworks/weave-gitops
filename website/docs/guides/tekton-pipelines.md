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

## General Tekton Flow
The overall flow for a Tekton pipeline is relatively the same for all use cases and it will look something like this:

EventListerer -> TriggerTemplate -> PipelineRun -> Pipeline -> Task

This will be the flow we will be creating for this demo.  More often then not we usually know what we want our pipeline to accomplish before we know how we want to trigger it.  So with that in mind we are going to start at the end with our Task definitions and work our way back to the EventListener.

## Create CI Pipeline
### Define CI Tasks
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
    - name: chart-path
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
        helm package $(params.chart-path) --version $(params.version) --app-version $(params.version)
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

### Define CI Pipeline
Now that we have our Tasks defined it is time to group them together into a [Pipeline](https://tekton.dev/docs/pipelines/pipelines).

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
    - name: chart-path
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
        - name: chart-path
          value: $(params.chart-path)
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

### Create CI Triggers
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
    - name: destination-git-full-name
    - name: image-name
    - name: path-to-image-context
    - name: tag
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
            value: $(tt.params.path-to-image-context)
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

### Create CI EventListener
Now we are going to create an EventListener that we'll configure to listen to Github events.  By default the EventListener is not exposed outside of the cluster.  To overcome this we will add a custom Ingress resource so that our listener is reachable from Github.

> Make sure to replace all values marked `<REPLACE ME>` with your true values.  You may also need to add additional configuration to your Ingress config to fit your environment needs.

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
  secretToken: "1234567" # This value will be set as the webhook secret in GitHub
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
        # use TriggerBinding to pass values to TriggerTemplate
        - ref: github-tag-binding

        # static values passed to TriggerTemplate
        - name: destination-git-url
          value: <REPLACE ME>
        - name: destination-git-full-name
          value: <REPLACE ME>
        - name: image-name
          value: <REPLACE ME>
        - name: path-to-image-context
          value: <REPLACE ME>
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
  - host: <REPLACE ME>
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

After this is created you'll need to create a webhook on your GitHub repo to call the EventListener url.  Details on how to create a webhook are availble in the [GitHub Docs](https://docs.github.com/en/developers/webhooks-and-events/webhooks/creating-webhooks).  Make sure to set the content type to `application/json`, set the secret to the value you set earlier for the `tekton-github-secret` k8s secret, and for this walkthrough you only need to trigger on `push` events.

## Create CD Pipeline
At this point we have our chart being built and released to our Helm repository automatically.  Now we are going to take it a step further and deploy the new chart to our environments automatically as well.  But lets say we have 3 environments we want to deploy to.  Dev, Staging and Prod for example.  You might not want to deploy to each of those environments at the same time.  You would probably rather deploy to Dev first, verify everything is work, then deploy to Staging, and then deploy to Prod.  This is where the fun begins.  For this we are going to use the Flux [Notification Controller](https://fluxcd.io/flux/components/notification) to fire alerts based on our HelmRelease events.  With these alerts, we'll be able to automatically promote the chart to the various enviroments after it was successfully deployed to the previous one.  Lets take a look at what this will look like.

### Define CD Tasks
For these tasks we do not need to build any new charts, images or binaries.  But we do need to update some configurations and move around some resource definition files so that Flux can deploy the correct files for each environment.

```yaml
apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: move-files
  namespace: default
spec:
  workspaces:
    - name: source
    - name: destination
  params:
    - name: source-files
      type: array
      default:
        - "*"
    - name: source-dir
      type: string
      default: ""
    - name: destination-dir
      type: string
      default: ""
    - name: destination-branch
      type: string
      default: main
    - name: commit-message
      type: string
      default: "commit from tekton pipeline"
    - name: GIT_USER_EMAIL
      type: string
      default: <>
    - name: GIT_USER_NAME
      type: string
      default: tekton-automation
  steps:
    - name: copy-files-to-destination
      image: alpine
      args: ["$(params.source-files[*])"]
      workingDir: $(workspaces.source.path)/$(params.source-dir)
      script: |
        #!/usr/bin/env sh
        set -e
        destination="$(workspaces.destination.path)/$(params.destination-dir)"
        mkdir -p "$destination"
        cp -r ${@} "$destination"
    - name: commit-to-destination
      image: alpine/git
      workingDir: $(workspaces.destination.path)
      script: |
        #!/usr/bin/env sh
        set -e
        git config --global user.email "$(params.GIT_USER_EMAIL)"
        git config --global user.name "$(params.GIT_USER_NAME)"
        git config --global --add safe.directory /workspace/destination
        git fetch origin
        git checkout -B $(params.destination-branch)
        git add .
        git commit -m "$(params.commit-message)"
        git push -f -u origin $(params.destination-branch)
---
apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: yq
  namespace: default
spec:
  workspaces:
    - name: source
  params:
    - name: args
      description: yq command arguments
      type: array
  steps:
    - name: yq
      image: mikefarah/yq:4
      securityContext:
        runAsUser: 0
      workingDir: $(workspaces.source.path)
      command: ["yq"]
      args: ["$(params.args[*])"]
```

### Create CD Pipeline
At this point everything you need to release should have built and test in the CI pipeline.  The CD pipeline should be more about updating the configuration needed for a new release.  For this example we want to promote a successfully deployed version from one environment to the next.  To do that we need to update target environment config to at least reflect the new chart version.  But depending on the flow, you may need to update other values as well.  Here is what it might look like to have a pipeline that also needs to update the namespace as part of the deployment.

```yaml
apiVersion: tekton.dev/v1beta1
kind: Pipeline
metadata:
  name: wego-pipeline-environment-promotion
  namespace: flux-system
spec:
  params:
    - name: revision
      type: string
    - name: source-git-url
      type: string
    - name: source-git-revision
      type: string
      default: main
    - name: source-git-directory
      type: string
      default: ""
    - name: source-git-files
      type: array
      default:
        - "*"
    - name: destination-git-url
      type: string
    - name: destination-git-full-name
      type: string
    - name: destination-git-directory
      type: string
      default: ""
    - name: destination-git-revision
      type: string
      default: main
    - name: promotion-namespace
      type: string
  workspaces:
    - name: shared-data
  tasks:
    - name: clone-source
      taskRef:
        name: git-clone
        kind: ClusterTask
      params:
        - name: url
          value: $(params.source-git-url)
        - name: revision
          value: $(params.source-git-revision)
      workspaces:
        - name: output
          workspace: shared-data
          subPath: source 
    - name: clone-destination
      taskRef:
        name: git-clone
        kind: ClusterTask
      params:
        - name: url
          value: $(params.destination-git-url)
        - name: revision
          value: $(params.destination-git-revision)
      workspaces:
        - name: output
          workspace: shared-data
          subPath: destination
    - name: set-chart-version
      taskRef:
        name: yq
      runAfter:
        - clone-source
        - clone-destination
      params:
        - name: args
          value:
            - '.spec.chart.spec.version="$(params.revision)"'
            - -i
            - ./$(params.source-git-directory)/release.yaml
      workspaces:
        - name: source
          workspace: shared-data
          subPath: source
    - name: set-chart-namespace
      taskRef:
        name: yq
      runAfter:
        - set-chart-version
      params:
        - name: args
          value:
            - '.metadata.namespace="$(params.promotion-namespace)"'
            - -i
            - ./$(params.source-git-directory)/release.yaml
      workspaces:
        - name: source
          workspace: shared-data
          subPath: source
    - name: set-chart-repository-namespace
      taskRef:
        name: yq
      runAfter:
        - set-chart-namespace
      params:
        - name: args
          value:
            - '.metadata.namespace="$(params.promotion-namespace)"'
            - -i
            - ./$(params.source-git-directory)/repository.yaml
      workspaces:
        - name: source
          workspace: shared-data
          subPath: source
    - name: copy-files-to-destination
      taskRef:
        name: move-files
      runAfter:
        - set-chart-version
        - set-chart-namespace
        - set-chart-repository-namespace
      params:
        - name: source-files
          value: ["$(params.source-git-files[*])"]
        - name: source-dir
          value: $(params.source-git-directory)
        - name: destination-dir
          value: $(params.destination-git-directory)
        - name: destination-branch
          value: $(params.revision)-$(params.destination-git-directory)
      workspaces:
        - name: source
          workspace: shared-data
          subPath: source
        - name: destination
          workspace: shared-data
          subPath: destination
    #
    # automatically create a pull-request for the release.
    # if you are deploying to an environment that does not
    # require review before deploying you can remove this step
    #
    - name: open-pr
      taskRef:
        name: github-open-pr
        kind: ClusterTask
      runAfter:
        - copy-files-to-destination
      params:
        - name: REPO_FULL_NAME
          value: $(params.destination-git-full-name)
        - name: HEAD
          value: $(params.revision)-$(params.destination-git-directory)
        - name: BASE
          value: $(params.destination-git-revision)
        - name: TITLE
          value: release $(params.revision) to $(params.promotion-namespace)
        - name: BODY
          value: release $(params.revision) to $(params.promotion-namespace)
        - name: GITHUB_TOKEN_SECRET_NAME
          value: github-api-token
```

### Create CD Triggers
The CD triggers will be function nearly identically as the CI triggers.  The only noticiable difference is sense we are using a different event source, we need to update the TriggerBinding to parse out the correct values.  Lets use a combination of static and event payload values to properly pass the correct values to our PipelineRun.

> Make sure to replace all values marked `<REPLACE ME>` with your true values.

```yaml
apiVersion: triggers.tekton.dev/v1beta1
kind: TriggerTemplate
metadata:
  name: wego-pipeline-environment-promotion-template
  namespace: flux-system
spec:
  params:
    - name: revision
    - name: source-git-url
    - name: source-git-directory
    - name: destination-git-url
    - name: destination-git-full-name
    - name: destination-git-directory
    - name: promotion-namespace
  resourcetemplates:
    - apiVersion: tekton.dev/v1beta1
      kind: PipelineRun
      metadata:
        generateName: wego-pipeline-environment-promotion-run-
      spec:
        serviceAccountName: build-bot
        pipelineRef:
          name: wego-pipeline-environment-promotion
        params:
          - name: revision
            value: $(tt.params.revision)
          - name: source-git-url
            value: $(tt.params.source-git-url)
          - name: source-git-directory
            value: $(tt.params.source-git-directory)
          - name: destination-git-url
            value: $(tt.params.destination-git-url)
          - name: destination-git-full-name
            value: $(tt.params.destination-git-full-name)
          - name: destination-git-directory
            value: $(tt.params.destination-git-directory)
          - name: promotion-namespace
            value: $(tt.params.promotion-namespace)
        workspaces:
          - name: shared-data
            volumeClaimTemplate:
              spec:
                accessModes:
                  - ReadWriteOnce
                resources:
                  requests:
                    storage: 1Gi
---
apiVersion: triggers.tekton.dev/v1beta1
kind: TriggerBinding
metadata:
  name: wego-pipeline-environment-promotion-binding
  namespace: flux-system
spec:
  params:
    - name: source-git-url
      value: <REPLACE ME>
    - name: destination-git-url
      value: <REPLACE ME>
    - name: destination-git-full-name
      value: <REPLACE ME>
    - name: revision
      value: $(body.metadata.revision)
    - name: kind
      value: $(body.involvedObject.kind)
    - name: name
      value: $(body.involvedObject.name)
    - name: namespace
      value: $(body.involvedObject.namespace)
```

### Create CD EventListener
You may notice there is no ingress configured for the Event Listener this time.  That is intentional.  When we configure the Flux notification Provider, we can leverage the internal service url instead.

> If your environment restricts interal service traffic then you will need to create an ingress config similar to the one created for the CI pipeline

```yaml
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: wego-pipeline-environment-promotion-trigger-sa
  namespace: flux-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: wego-pipeline-environment-promotion-trigger-binding
  namespace: flux-system
subjects:
- kind: ServiceAccount
  name: wego-pipeline-environment-promotion-trigger-sa
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: tekton-triggers-eventlistener-roles
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: wego-pipeline-environment-promotion-trigger-clusterbinding
subjects:
- kind: ServiceAccount
  name: wego-pipeline-environment-promotion-trigger-sa
  namespace: flux-system
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: tekton-triggers-eventlistener-clusterroles
---
apiVersion: triggers.tekton.dev/v1beta1
kind: EventListener
metadata:
  name: wego-pipeline-environment-promotion
  namespace: flux-system
spec:
  triggers:
    - name: dev-deployment
      interceptors:
        - name: "ignore failed releases" # we only want to promote the deployment if the release was successful
          ref:
            name: cel
          params:
            - name: filter
              value: "body.severity != 'error'"
        - name: "only listen to dev events" # on successful dev deployments, promote to staging
          ref:
            name: cel
          params:
            - name: filter
              value: "body.involvedObject.namespace == 'dev'"
      bindings:
        - ref: wego-pipeline-environment-promotion-binding
        - name: source-git-directory
          value: dev/helm
        - name: destination-git-directory
          value: staging/helm
        - name: promotion-namespace
          value: staging
      template:
        ref: wego-pipeline-environment-promotion-template
    - name: staging-deployment
      interceptors:
        - name: "ignore failed releases"
          ref:
            name: cel
          params:
            - name: filter
              value: "body.severity != 'error'"
        - name: "only listen to staging events" # on successful staging deployments, promote to prod
          ref:
            name: cel
          params:
            - name: filter
              value: "body.involvedObject.namespace == 'staging'"
      bindings:
        - ref: wego-pipeline-environment-promotion-binding
        - name: source-git-directory
          value: staging/helm
        - name: destination-git-directory
          value: prod/helm
        - name: promotion-namespace
          value: prod
      template:
        ref: wego-pipeline-environment-promotion-template
  resources:
    kubernetesResource:
      spec:
        template:
          spec:
            serviceAccountName: wego-pipeline-environment-promotion-trigger-sa
            containers:
              - resources:
                  requests:
                    memory: 64Mi
                    cpu: 250m
                  limits:
                    memory: 64Mi
                    cpu: 250m
```

### Create Flux Alert
With Flux managing our deployments, we can leverage Flux events to trigger the pipeline runs.

```yaml
apiVersion: notification.toolkit.fluxcd.io/v1beta1
kind: Provider
metadata:
  name: el-wego-pipeline-environment-promotion
  namespace: flux-system
spec:
  type: generic
  address: http://el-wego-pipeline-environment-promotion.flux-system.svc:8080 # Tekton Event Listener url
---
apiVersion: notification.toolkit.fluxcd.io/v1beta1
kind: Alert
metadata:
  name: tekton-pipelines-alerts
  namespace: flux-system
spec:
  providerRef: 
    name: el-wego-pipeline-environment-promotion
  eventSeverity: info
  eventSources:
    - kind: HelmRelease
      namespace: dev
      name: demo
    - kind: HelmRelease
      namespace: staging
      name: demo
    - kind: HelmRelease
      namespace: prod
      name: demo
  exclusionList:  # filter out events that should not trigger a pipeline run
    - ".*upgrade.*has.*started"
    - ".*not.*ready"
    - "^Dependencies.*"
```

## Testing
At this point you should have all the pieces in place for the pipeline to trigger and run successfully.  So lets try it out.  Create a new tag on your chart's repository and push it.

```bash
git tag v0.0.1
git push --tags
```

The Tekton pipeline should have started.  If you installed the Tekton Dashboard, navigate to the UI and you can monitor the pipeline progress.  If you installed the Tekton-CLI, run `tkn pipelinerun logs -f` to monitor the logs of each task.  If you do not have either installed, you can watch the pods being created and monitor the logs of each one.  But I would recommend using either the Tekton-CLI or Dashboard instead.

Once the pipeline is complete you should see the new tagged version of your chart available in your Helm repository.
