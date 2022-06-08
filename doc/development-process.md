# Development Process

Depending on your area of focus we're providing guidance of each kind of development setup you might want.

Step one, you **must** have the following things installed:

* [go](https://go.dev) v1.17 -- Primary development language.
* [docker](https://www.docker.com/) -- Used for generating containers & testing kubernetes set-ups.
* [helm](https://helm.sh/docs/intro/install/) v3. -- Package manager for kubernetes
* [kind](https://kind.sigs.k8s.io/docs/user/quick-start#installation) -- Run kubernetes clusters in docker for testing.
* [flux](https://fluxcd.io/docs/get-started/) -- Continuous delivery system for kubernetes that weave-gitops enriches.
* [tilt](https://tilt.dev/) -- Automatically build and deploy to a local cluster.

Now you have those things installed, lets go!

## Quick start development setup

Help me get going as fast as possible!

Create a kind cluster. We'll do this so we can have a local cluster on our machine

`$ kind create cluster --name=<some name>`

Then we install flux on it.

`$ flux install`

Then we bring up tilt.

`$ tilt up`

Tilt will tell you the url, or hit the space bar and it'll automagically open a browser for you which will show your local cluster and the resources running. You should see 3 things, Tiltfile, dev-weave-gitops, uncategorised.

Goto login with username: dev, password:dev at http://localhost:9001/sign_in

When you make changes in the code with this setup, you have to wait for k8s to redeploy the updated pod, it isn't lightening instant fast but **it works** to get you going.

Woop! It's working. It'll be empty cos you haven't created any flux objects. You can jump to [
quickly add a sample workload to UI](#quickly-add-a-sample-workload-to-ui).


## Frontend optimised development setup

Help me I just want super fast frontend development!

Step one, make sure you've installed the tools listed at the very top of this file.

Step two, now you **must** **ALSO** have the following installed.

* [node](https://nodejs.org/en/) v16 - Install Node.js

Now you have those things installed, lets go!

## Quick start frontend development setup

Create a kind cluster. We'll do this so we can have a local cluster on our machine

`$ kind create cluster --name=<some name>`

Then we install flux on it.

`$ flux install`

Then we bring up tilt, without auto-restart enabled (see [the FAQ
entry below](#the-server-keeps-restarting-and-its-annoying)).

`$ MANUAL_MODE=true tilt up`

Then we make our node_modules, this will take a little while.

`$ make node_modules`

Then we fire up our frontend server.

`$ npm start`

Goto login with username: dev, password:dev at http://localhost:4567/sign_in

Running this setup is what enables JavaScript hot reloading, it does websocket pushing of your code, so you don't even need to refresh!

Woop! It's working. It'll be empty cos you haven't created any flux objects. You can jump to [quickly add a sample workload to UI](#quickly-add-a-sample-workload-to-ui).


## Quickly add a sample workload to UI

To help see some objects in the UI lets create some sample sources, run the following:

```
$ flux create source git podinfo \
--url=https://github.com/stefanprodan/podinfo \
--branch=master \
--interval=30m
```

and then this one:

```
$ flux create kustomization podinfo \
--target-namespace=flux-system \
--source=podinfo \
--path="./kustomize" \
--prune=true \
--interval=5m
```

Boom! You'll see our newly created flux objects in the UI.

We use create rather than a flux bootstrap to create because we don't want our tiltfile and flux to start reconciling over each other.

## Other Frontend focussed commands

Lint frontend code with `make ui-lint` - using Prettier (https://prettier.io/) will get you on the right track!

Run frontend tests with `make ui-test`

Update CSS snapshots with `npm run test -- -u`

Check dependency vulnerabilities with `make ui-audit`


## Running tests

You can, at the moment, run two kinds of tests.

Our unit tests:

`$ make unit-tests`

Our frontend tests:

`$ make ui-tests`

We're re-working our integration tests. Coming back soon.


## Other optimisations for development setup

Mostly for backend development, if you want to not wait for k8s to reload the pod, you can go into more detail and install locally a bunch of the dependencies and there's a flag `FAST_AND_FURIOUSER` which you can pass to our tiltfile which will look for your local resources.

Only do this if you know what you're doing or are happy to spend timing learning how to dig around in glitchy figuring out go things on your operating system. Funtimes ahead :)

Step one, make sure you've installed the tools listed at the very top of this file.

Step two, install dependencies and build binaries and assets onto your machine.

`$ make all`

Then let's create a cluster.

`$ kind create cluster --name=<some name>`

Then we install flux on it.

`$ flux install`

Then we bring up tilt, passing it the flag `FAST_AND_FURIOUSER`. This tells our tiltfile we want to use our local resources.

`$ FAST_AND_FURIOUSER=true tilt up`

Goto login with username: dev, password:dev at http://localhost:9001/sign_in

Woop! It's working? Probably. Maybe. Or you may need to faff around with resolving build steps when you did `make all`.
Good luck!

## Tips & tricks

### The server keeps restarting and it's annoying

Tilt has a feature where it automatically restarts the pod whenever
you save a changed file. This might give you a few seconds of nothing
working, over and over.

Depending on your setup, that might be more annoying than helpful -
for instance, if you're doing frontend development outside of docker,
then the frontend is already being restarted on its own, so anything
tilt does is just getting in your way. If you disable auto-restart,
then any time you want to re-deploy the k8s pod, you have to open the
tilt UI and click on the spinny icon.

`tilt up` starts tilt with auto-reload enabled. To disable, instead
start it with `MANUAL_MODE=true tilt up`.
