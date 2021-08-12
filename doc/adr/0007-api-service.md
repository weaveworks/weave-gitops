# 7. API service

Date: 2021-08-10

## Status

Proposed

## Context

As basic processes, `wego` interacts with the kubernetes api and the github api. Currently, this interaction occurs on the client side. So, for example, for 2 users to access the graphical interface, they both need to have at least `kubectl` and permissions on the` .kubeconfig` file to access the cluster. The proposal is to run this interaction logic directly in the cluster so that in the near future it allows the creation of more powerful features for `wego`. One of those features could be the exposition of the graphical interface in a simpler way. It would also allow people to write their own tools for wego.

## Decision

The proposal consists of extracting the API related to applications, which we will call `API-service`, execute it within the cluster through using deployment and expose it through a service. The API-service will be the only entity that will be able to interact directly with the Kubernetes API. In other words, the API-service will expose the necessary options to interact with wego applications, abstracting the logic that interacts with the Kubernetes API.

To illustrate, you could say that the `wego` command line could tell the API-service "Add this application" and it, in turn, would tell Kubernetes "Apply these manifest files. "

For security reasons, permissions will be implemented through RBAC so that the API-service only has access to the resources it needs.

With this proposal, the use of CLI options would still be required, however, it would be an intermediate step to eventually do
a 100% migration into the cluster.

Specifically, the proposed changes are as follows:

### Migrate the remaining functionality of the implementation in `KubeClient` to`KubeHTTP`

The methods that have not been migrated are:

- Apply
- Delete
- DeleteByName
- LabelExistsInCluster

### Create API-service docker image and update authentication method.

For the API-service to run within the cluster, a docker image will need to exist that can run it.

The part that needs to be run in this image is the HTTP server which is initialized in the `wego ui run` command.

To continue communicating with the Kubernetes API, it will be required to change the authentication method to one that allows communication from within the cluster called in-cluster client configuration.

### Generate RBAC-based permissions to the API-service and apply them when installing the GitOps runtime.

Since the API-service needs to access resources within the cluster, RBAC permissions will be applied so that there are no security gaps. In consideration that the functionality of the API-service will be specified only in actions within the namespace where the GitOps runtime is executed, the Role and RoleBinding resources will be used. Since the scope of this type of resource is by namespace. 

The resources that will be required are:

- ServiceAccount
- Deployment
- Role
- RoleBinding
- Service

The specific permissions for the Role resource will be:

```
resources: ["apps.wego.weave.works"]
verbs: ["get","create","list"]

resources: ["secret"]
verbs: ["get","create"]

resources: ["kustomizations.kustomize.toolkit.fluxcd.io"]
verbs: ["get","create"]

resources: ["helmreleases.helm.toolkit.fluxcd.io"]
verbs: ["get","create"]

resources: ["helmrepositories.source.toolkit.fluxcd.io"]
verbs: ["get","create"]
```

These resources will be applied when installing the GitOps engine.

### Enable access to the API-service from the client-side.

The `port-forward` command will be executed using `kubectl` to enable communication between the `wego` CLI and the API-service. This command will expose a dynamic port on the client machine. When executing this instruction, the dynamic port will be shown in the standard output and it will have to be extracted to build the address of the API-service. Which will be `localhost:DYNAMIC_PORT`. This process of exposing the API-service will be executed for each command that is executed from CLI. With the exception of the `wego ui run` command that will keep the API-service exposure execution in the background for the duration of that command session, so that the UI can access the resource.

### Automate the build of docker images from the `API-service`.

For production environments, it will be needed a docker repository, a token with permissions to push the image, and public access to those images.

When a `wego` release is made, a Github action will be triggered that will make the corresponding release in the API-service's docker repository using the same tag used for the` wego` release.

At runtime, we will use the version LD flag to tell `wego` which tag to use in the `API-service` deployment. 

### Use source to build locally the API-service docker image for development mode.

Part of the mechanism from the previous point will be used so that it can also be applied in development mode. And thus avoid wasting time publishing / downloading to a docker registry. Currently, there are many tools on the market to achieve this task, to mention a few there are `garden.io` and` tilt`.

### Reuse the local build of the API-service inside Github actions for testing.

Since creating a pull request in `wego` triggers several tests. It will also be required to use the construction system of the previous point to create an image that is completely directed for the tests.

To ensure that the construction of images of the API-service in Github actions is as fast as possible, a mechanism will be implemented that allows sharing the context of the docker builder. One option that seems simple to implement is https://github.com/docker/build-push-action.

## Consequences

One consequence of running the API-service within the cluster is that we will need to think about how to interact with flux. One option could be to have flux installed in the docker image of the API-service or the other would be to make direct use of the Kubernetes resources that flux manages to obtain the same results of the commands we have used so far.
