## Weave GitOps :heart:s your contributions

Thank you for taking the time to contribute to Weave Gitops!

This guide **is a continuous work in progress**, we are always welcome to ideas on how to improve the experience working with the project. 

## Requirements/tools

This is a list of the tools you may need to install:

* [go](https://go.dev) -- Primary compiler for the CLI.
* [npm](https://www.npmjs.com/) -- Package manager for UI components.
* [ginkgo](https://onsi.github.io/ginkgo/) -- A go testing framework.
* [docker](https://www.docker.com/) -- Used for generating containers & testing kubernetes set-ups.
* [golangci-lint](https://github.com/golangci/golangci-lint/) -- A go linter.
* [buf](https://buf.build/) -- To generate the protobufs used by the API.
* [reflex](https://github.com/cespare/reflex) -- A file watcher.
* [kind](https://kind.sigs.k8s.io/) -- Run kubernetes clusters in docker for testing.
* [lcov](https://github.com/linux-test-project/lcov) -- Used for code coverage.
* [flux](https://fluxcd.io/) -- Continuous delivery system for kubernetes that weave-gitops enriches.

Some other tools are installed automatically by the makefile for you:

* [go-acc](https://github.com/ory/go-acc) -- Calculates code coverage for go.
* [gcov2lcov](https://github.com/jandelgado/gcov2lcov) -- Converts output from go-acc to a format lcov understands.

And some tools that are installed by the `tools/download-deps.sh` script:

* [envtest](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/envtest) -- Run a kubernetes control plane locally for testing.
* [tilt](https://tilt.dev/) -- Automatically build and deploy to a local cluster.

## CLI/API development

To set up a development environment for the gitops CLI

### To use an existing environment

1. Install go v1.17
2. Install [buf](https://github.com/bufbuild/buf)
3. Run `make all` to install dependencies and build binaries and assets
4. Start a `kind` cluster like so: `KIND_CLUSTER_NAME=<some name> ./tools/kind-with-registry.sh`
5. Run `flux install -n flux-system`
6. Start the in-cluster API replacement job (powered by [http://tilt.dev](tilt.dev)) with `make cluster-dev`
7. `make` or `make unit-tests` to ensure everything built correctly.
8. Navigate to http://localhost:9001 in your browser. The login is `dev` with the password `dev`.

### To use a bootstrapped, ready made environment

1. Install go v1.17
2. Install [buf](https://github.com/bufbuild/buf)
3. Run `make all` to install dependencies and build binaries and assets
4. Run `make cluster-dev` which should install and bring up everything and then start `tilt` to take over monitoring

### Cluster Dev Tips

- You may need to turn off your `kustomize-controller` to prevent it from reconciling your "GitOps RunTime" and over-writing the `wego-app` deployment.
- Setting the system kustomization to `suspend: true` in the config repo will also keep `kustomize-controller` from fighting with `tilt`. You may need to kill a failing pod after suspending the kustomization.

### Unit testing

We are using [Ginko](https://onsi.github.io/ginkgo/) for our unit tests. To execute the all the unit tests, run `make unit-tests`.

To run a single test, you will need to set the KUBEBUILDER_ASSESTS environment variable to point to the directory containing our mock K8s objects.

```bash
export KUBEBUILDER_ASSETS=$(git rev-parse --show-toplevel)/tools/bin/envtest
go test github.com/weaveworks/weave-gitops/pkg/kube
```

or

```bash
export KUBEBUILDER_ASSETS=$(git rev-parse --show-toplevel)/tools/bin/envtest
cd pkg/kube
go test
```

#### Executing a subset of tests

Ginkgo allows you to run a subset of Describe/Context/It specs. See [Focused Specs](https://onsi.github.io/ginkgo/#focused-specs) for more information

### Setup golangci-lint in your editor

Link for golangci-lint editor integration: https://golangci-lint.run/usage/integrations/

For VSCode, use these editor configuration flags:

```json
    "go.lintFlags": [
        "--fast",
    ],
```

## UI Development

To set up a development environment for the UI

1. Install go v1.17
2. Install Node.js version 16.13.2
3. Make sure your `$GOPATH` is added to your `$PATH` in your bashrc or zshrc file, then install reflex for automated server builds: go get github.com/cespare/reflex
4. Go through the Weave GitOps getting started docs here: https://docs.gitops.weave.works/docs/getting-started/
5. Run `make node_modules`. NOTE: Running `npm install` could leave you unable to pass our ui-tests. If you're getting an error about a git diff in your package.lock, run `rm -rf node_modules && make node_modules`.
6. Make sure GitOps is installed on a fresh kind cluster for this repo by running `kind delete cluster`, `kind create cluster`, and finally `gitops install`.
7. To start up the HTTP server with automated re-compliation, run `make api-dev`
8. Run `npm start` to start the frontend dev server (with hot-reloading)

Lint frontend code with `make ui-lint` - using Prettier (https://prettier.io/) will get you on the right track!

Run frontend tests with `make ui-test` - update CSS snapshots with `npm run test -- -u`

Check dependency vulnerabilities with `make ui-audit`

To avoid invalidating JWT tokens on every server restart set the `GITOPS_JWT_ENCRYPTION_SECRET` env variable in your shell to use a static encryption secret. Else, a random encryption secret will be used that will change on every server (or pod) restart, thus invalidating any JWTs that were created with the old secret.

### Recommended Snippets

To create a new styled React component (with typescript):

```json
{
  "Export Default React Component": {
    "prefix": "tsx",
    "body": [
      "import * as React from 'react';",
      "import styled from 'styled-components'",
      "",
      "type Props = {",
      "  className?: string",
      "}",
      "",
      "function ${1:} ({ className }: Props) {",
      "  return (",
      "    <div className={className}>",
      "      ${0}",
      "    </div>",
      "  );",
      "}",
      "",
      "export default styled(${1:}).attrs({ className: ${1:}.name })``"
    ],
    "description": "Create a default-exported, styled React Component."
  }
}
```

## ⚠️ THE FOLLOWING CONTENT MAY BE OUT OF DATE ⚠️

## Building the binaries
To build the `gitops` binary locally you can run `make gitops`. This will create a `gitops`
binary in the `bin/` directory. Similarly running `make gitops-server` will create a
`gitops-server` binary in the same directory.


## Testing
We have 3 layers of testing:
- Unit testing
- Integration testing
- Acceptance testing

We aim to follow the [testing pyramids best practises](https://martinfowler.com/articles/practical-test-pyramid.html)
with the focus on a large number of units tests, fewer integration tests and even fewer acceptance tests.


### Running unit tests
You can run `make unit-tests` from the top of the repository to run all of the unit tests.
As part of running the unit tests we will install any required dependencies for building
the code. See [dependencies](tools/dependencies.toml) for more information on what's installed.


### Running integration tests
The integration tests can be run using `make integration-tests`. They require some
extra environment variables to run:

- `GITHUB_ORG`- The name of the org that test repositories can be created in.
If you're a Weaveworks employee you can use `weaveworks-gitops-test`
- `GITHUB_TOKEN`- Used to create, push and delete repositories. Please ensure
the token has sufficient permissions, as deleting repositories is not part
of the standard set of permissions a token is given.

If you wish to run the full suite, you will also need the equivalent for gitlab:
- `GITLAB_TOKEN`
- `GITLAB_ORG`

However if you only want to test it against one of the providers, use [ginkgos focus](
https://onsi.github.io/ginkgo/#focused-specs) feature to run the tests for one provider.
Example:
```
cat test/integration/server/add_test.go
...
 FContext("GitHub", func() {
```

### Running acceptance tests
The acceptance tests can be run using `make acceptance-tests`. They require the same
environment variables as the integration tests, with the following in addition:

- `WEGO_BIN_PATH`- The path to the `gitops` binary the tests should use. To test a locally
built binary which is created from `make gitops`, you want to set it to `WEGO_BIN_PATH=$PWD/bin/gitops`.
- `IS_TEST_ENV`- If this environment variable is set to anything, it indicates to the `gitops` binary
that it should deploy images using the `latest` tag, instead of the latest version.

#### Testing local changes with acceptance tests
As of right now, it's not easy to test your local code changes with acceptance tests.
Until [#1158](https://github.com/weaveworks/weave-gitops/issues/1158) is resolved you must:

1. Manually build and push the `wego-app` docker image to your own registry. Example:
`docker build -f Dockerfile -t aclevernameww/wego-app:latest . && docker push aclevernameww/wego-app:latest`
2. Update the `manifests/wego-app/deployment.yaml.tpl` to use your registry. Example:
`image: aclevernameww/wego-app:{{.Version}})`
3. Rebuild the local gitops binary via `make gitops`
4. Ensure `WEGO_BIN_PATH` is set to the local binary and `IS_TEST_ENV=true`

As a result, when the acceptance tests run `gitops install`, they should deploy your custom-built docker image with the desired code changes.

