# 7. API service

Date: 2021-08-10

## Status

Proposed

## Context

As basic processes, `wego` interacts with the kubernetes api and the github api. Currently, this interaction occurs on the client side. So, for example, for 2 users to access the graphical interface, they both need to have at least `kubectl` and permissions on the `.kubeconfig` file to access the cluster. The proposal is to run this interaction logic directly in the cluster so that in the near future it allows the creation of more powerful features for `wego`. One of those features could be the exposition of the graphical interface in a simpler way. It would also allow people to write their own tools for wego.

## Decision

The proposal consists of extracting the API related to applications, which we will call `wego-api`, execute it within the cluster through using deployment and expose it through a service. The wego-api will be the only entity that will be able to interact directly with the Kubernetes API. In other words, the wego-api will expose the necessary options to interact with wego applications, abstracting the logic that interacts with the Kubernetes API.

To illustrate, you could say that the `wego` command line could tell the wego-api "Add this application" and it, in turn, would tell Kubernetes "Apply these manifest files. "

For security reasons, permissions will be implemented through RBAC so that the wego-api only has access to the resources it needs.

With this proposal, the use of `kubectl` would still be required, however, it would be an intermediate step to eventually do
a 100% migration into the cluster.

Specifically, the proposed changes are as follows:

### Create wego-api docker image and update authentication method.

For the wego-api to run within the cluster, a docker image will need to exist that can run it.

The part that needs to be run in this image is the HTTP server which is initialized in the `wego ui run` command.

To continue communicating with the Kubernetes API, it will be required to change the authentication method to one that allows communication from within the cluster called in-cluster client configuration.

Flux will be added to this docker image as well.

`wego version` will also show which container registry tag of `wego-api` it will use.

### Generate RBAC-based permissions to the wego-api and apply them when installing the GitOps runtime.

Since the wego-api needs to access resources within the cluster, RBAC permissions will be applied so that there are no security gaps. In consideration that the functionality of the wego-api will be specified only in actions within the namespace where the GitOps runtime is executed, the Role and RoleBinding resources will be used. Since the scope of this type of resource is by namespace.

The resources that will be required are:

- ServiceAccount
- Deployment
- Role
- RoleBinding
- Service

The specific permissions for the Role resource will be:

```
resources: ["apps.wego.weave.works"]
verbs: ["*"]

resources: ["secret"]
verbs: ["*"]

resources: ["kustomizations.kustomize.toolkit.fluxcd.io"]
verbs: ["*"]

resources: ["helmreleases.helm.toolkit.fluxcd.io"]
verbs: ["*"]

resources: ["helmrepositories.source.toolkit.fluxcd.io"]
verbs: ["*"]

resources: ["gitrepositories.source.toolkit.fluxcd.io"]
verbs: ["*"]
```

These resources will be applied when installing the GitOps engine.

### Enable access to the wego-api from the client-side.

To enable access to the `wego-api` from the client machine, we will use `kubectl` by running this command: `kubectl -n NAMESPACE port-forward service/wego-api :8283`. An example of the output would be:
```
Forwarding from 127.0.0.1:63701 -> 8283
Forwarding from [::1]:63701 -> 8283
```

The path to access `wego-api` will be `http://127.0.0.1:63701`. UI will be accessed by using this base path.

Initially the only command using `wego-api` will be `wego ui run` which will run the port-forward command in the background, keep it running and open the browser using that same local address, as the UI web server lives in `wego-api`.

By using this command there is security issue we need to solve. The issue is that `wego-api` could be accessed by any other program on the local machine. To fix this issue there will be an extra authorization layer that will be addressed in another ADR.

The flag `--wego-api-port` will be added to allow the user expose `wego-api` in the local port of their choice.

### Automate the build of docker images from the `wego-api`.

For production environments the following items are required, a container repository (e.g. docker), a token with permissions to push the image, and public access to those images.

When a `wego` release is made, a Github action will be triggered that will make the corresponding release in the wego-api's Github docker repository using the same tag used for the` wego` release. This Github action will be a dependency to the wego binary release we already have, to avoid the issue like of wego not finding the right `wego-api` image. We will notify via slack message to the team channel if any error occurs.

At runtime, we will use the version LD flag to tell `wego` which tag to use in the `wego-api` deployment.

To build `wego-api` we will use a Dockerfile with two-stage builds where possible, so that the runtime container has a very minimal surface area.

### Use source to build locally the wego-api docker image for development mode.

To avoid wasting time publishing / downloading to a docker registry. We want to improve the developer experience. Currently, there are many tools on the market to achieve this task, to mention a few there are `garden.io` and` tilt`.

The recommendation is to use `garden.io` because they use yaml files to set the configuration which is something the team is already used to. Rather than `tilt` that relies on python which would enforce learning that language.

### Reuse the local build of the wego-api inside Github actions for testing.

Since creating a pull request in `wego` triggers several tests. It will also be required to use the construction system of the previous point to create an image that is completely directed for the tests.

To ensure that the construction of images of the wego-api in Github actions is as fast as possible, a mechanism will be implemented that allows sharing the context of the docker builder. One option that seems simple to implement is https://github.com/docker/build-push-action.

## Consequences

To avoid dealing with the complexity of authorizing commands that use the kubectl implementation, migrating the rest of the commands, in addition to `wego ui run`, will not be considered as part of this ADR. Since that implementation is being planned to be completely removed in the near future.

`kubectl proxy` was also considered to expose `wego-api` but it also opens the kubernetes api, and we want to have less exposure.

We could add kubectl and flux binaries to wego via embedding but that would increase the binary size, and the build time dramatically.
