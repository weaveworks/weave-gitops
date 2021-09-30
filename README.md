# weave-gitops

Weave GitOps

[![Coverage Status](https://coveralls.io/repos/github/weaveworks/weave-gitops/badge.svg?branch=main)](https://coveralls.io/github/weaveworks/weave-gitops?branch=main)
![Test status](https://github.com/weaveworks/weave-gitops/actions/workflows/test.yml/badge.svg)
[![LICENSE](https://img.shields.io/github/license/weaveworks/weave-gitops)](https://github.com/weaveworks/weave-gitops/blob/master/LICENSE)
[![Contributors](https://img.shields.io/github/contributors/weaveworks/weave-gitops)](https://github.com/weaveworks/weave-gitops/graphs/contributors)
[![Release](https://img.shields.io/github/v/release/weaveworks/weave-gitops?include_prereleases)](https://github.com/weaveworks/weave-gitops/releases/latest)
[![FOSSA Status](https://app.fossa.com/api/projects/custom%2B19155%2Fgithub.com%2Fweaveworks%2Fweave-gitops.svg?type=shield)](https://app.fossa.com/reports/005da7c4-1f10-4889-9432-8b97c2084e41)

## Overview

Weave GitOps enables an effective GitOps workflow for continuous delivery of applications into Kubernetes clusters.
It is based on [CNCF Flux](https://fluxcd.io), a leading GitOps engine.

## Getting Started

### CLI Installation

Mac / Linux

```console
curl --silent --location "https://github.com/weaveworks/weave-gitops/releases/download/v0.6.1/gitops-$(uname)-$(uname -m).tar.gz" | tar xz -C /tmp
sudo mv /tmp/gitops /usr/local/bin
gitops version
```

Alternatively, users can use Homebrew:

```console
brew tap weaveworks/tap
brew install weaveworks/tap/gitops
```

Please see the [getting started guide](https://docs.gitops.weave.works/docs/getting-started).

## CLI Reference

```console
Weave GitOps
Command line utility for managing Kubernetes applications via GitOps.

Usage:
  gitops [command]

Available Commands:
  app         Add or status application
  flux        Use flux commands
  help        Help about any command
  install     Install or upgrade Weave GitOps
  ui          Manages Weave GitOps UI
  uninstall   Uninstall Weave GitOps
  version     Display Weave GitOps version

Flags:
  -h, --help               Help for gitops
      --namespace string   The namespace scope for this operation (default "wego-system").
  -v, --verbose            Enable verbose output

Use "gitops [command] --help" for more information about a command.
```

For more information please see the [docs](https://docs.gitops.weave.works/docs/cli-reference)

## CLI/API development

To set up a development environment for the CLI

1. Install go v1.17
2. Install [buf](https://github.com/bufbuild/buf)
3. Run `make all` to install dependencies and build binaries and assets
4. Start a `kind` cluster like so: `KIND_CLUSTER_NAME=<some name> ./tools/kind-with-registry.sh`
5. Run `./bin/gitops install --config-repo=<repo url>`
6. Start the in-cluster API replacement job (powered by [http://tilt.dev](tilt.dev)) with `make cluster-dev`
7. make or make unit-tests to ensure everything built correctly.

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
2. Install Node.js version 14.15.1
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

## FAQ

Please see our Weave GitOps Core [FAQ](https://www.weave.works/faqs-for-weave-gitops-core/)

## Contribution

Need help or want to contribute? Please see the links below.

- Getting Started?
  - Follow our [Get Started guide](https://docs.gitops.weave.works/docs/getting-started) and give us feedback
- Need help?
  - Talk to us in the [#weave-gitops channel](https://app.slack.com/client/T2NDH1D9D/C0248LVC719/thread/C2ND76PAA-1621532937.019800) on Weaveworks Community Slack. [Invite yourself if you haven't joined yet.](https://slack.weave.works/)
- Have feature proposals or want to contribute?
  - Please create a [Github issue](https://github.com/weaveworks/weave-gitops/issues)

## License scan details

[![FOSSA Status](https://app.fossa.com/api/projects/custom%2B19155%2Fgithub.com%2Fweaveworks%2Fweave-gitops.svg?type=large)](https://app.fossa.com/reports/005da7c4-1f10-4889-9432-8b97c2084e41)
