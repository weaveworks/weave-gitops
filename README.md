# weave-gitops

Weave GitOps

[![Coverage Status](https://coveralls.io/repos/github/weaveworks/weave-gitops/badge.svg?branch=main)](https://coveralls.io/github/weaveworks/weave-gitops?branch=main)
![Test status](https://github.com/weaveworks/weave-gitops/actions/workflows/test.yml/badge.svg)
[![LICENSE](https://img.shields.io/github/license/weaveworks/weave-gitops)](https://github.com/weaveworks/weave-gitops/blob/master/LICENSE)
[![Contributors](https://img.shields.io/github/contributors/weaveworks/weave-gitops)](https://github.com/weaveworks/weave-gitops/graphs/contributors)
[![Release](https://img.shields.io/github/v/release/weaveworks/weave-gitops?include_prereleases)](https://github.com/weaveworks/weave-gitops/releases/latest)
[![FOSSA Status](https://app.fossa.com/api/projects/custom%2B19155%2Fgithub.com%2Fweaveworks%2Fweave-gitops.svg?type=shield)](https://app.fossa.com/projects/custom%2B19155%2Fgithub.com%2Fweaveworks%2Fweave-gitops?ref=badge_shield)

## Overview

Weave GitOps enables an effective GitOps workflow for continuous delivery of applications into Kubernetes clusters.
It is based on [CNCF Flux](https://fluxcd.io), a leading GitOps engine.

### Early access

_Weave GitOps is in early stages and iterating. Not all capabilities are available yet, and the CLI commands and other aspects may change. Please be aware this is not production ready yet. We would appreciate feedback and contributions of all kinds at this stage._

## Getting Started

### CLI Installation

Mac / Linux

```console
curl -L "https://github.com/weaveworks/weave-gitops/releases/download/v0.2.5/wego-$(uname)-$(uname -m)" -o wego
chmod +x wego
sudo mv ./wego /usr/local/bin/wego
wego version
```

Please see the [getting started guide](https://docs.gitops.weave.works/docs/getting-started).

## CLI Reference

```console
Weave GitOps

Usage:
  wego [command]

Available Commands:
  app         Add or status application
  flux        Use flux commands
  help        Help about any command
  install     Install or upgrade Wego
  version     Display wego version

Flags:
  -h, --help               help for wego
      --namespace string   gitops runtime namespace (default "wego-system")
  -v, --verbose            Enable verbose output

Use "wego [command] --help" for more information about a command.
```

For more information please see the [docs](https://docs.gitops.weave.works/docs/cli-reference)

## CLI/API development

To set up a development environment for the CLI

1. Install go v1.16
2. Install [buf](https://github.com/bufbuild/buf)
3. make

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

1. Install go v1.16
2. Install Node.js version 14.15.1
3. Install reflex for automated server builds: go get github.com/cespare/reflex
4. Run `npm install`
5. To start up the HTTP server with automated re-compliation, run `make ui-dev`
6. Run `npm start` to start the frontend dev server (with hot-reloading)

Lint frontend code with `make ui-lint`

Run frontend tests with `make ui-test`

Check dependency vulnerabilities with `make ui-audit`

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
      "export default styled(${1:})``"
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

[![FOSSA Status](https://app.fossa.com/api/projects/custom%2B19155%2Fgithub.com%2Fweaveworks%2Fweave-gitops.svg?type=large)](https://app.fossa.com/projects/custom%2B19155%2Fgithub.com%2Fweaveworks%2Fweave-gitops?ref=badge_large)
