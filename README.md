# weave-gitops

Weave GitOps core

[![Coverage Status](https://coveralls.io/repos/github/weaveworks/weave-gitops/badge.svg?branch=main)](https://coveralls.io/github/weaveworks/weave-gitops?branch=main)
![Test status](https://github.com/weaveworks/weave-gitops/actions/workflows/test.yml/badge.svg)
[![LICENSE](https://img.shields.io/github/license/weaveworks/weave-gitops)](https://github.com/weaveworks/weave-gitops/blob/master/LICENSE)
[![Contributors](https://img.shields.io/github/contributors/weaveworks/weave-gitops)](https://github.com/weaveworks/weave-gitops/graphs/contributors)
[![Release](https://img.shields.io/github/v/release/weaveworks/weave-gitops?include_prereleases)](https://github.com/weaveworks/weave-gitops/releases/latest)

## Overview
Weave GitOps enables an effective GitOps workflow for continuous delivery. It is based on [CNCF Flux](https://fluxcd.io), a leading GitOps engine. Weave GitOps 
provides pre-defined workflows and configuration around Flux to create effective GitOps flows. 

Weave GitOps is in pre-release development and not all capabilities are available yet.

## Getting Started

### Installation

Mac Intel
```console
curl -L https://github.com/weaveworks/weave-gitops/releases/download/v0.0.2/weave-gitops_0.0.2_darwin_amd64.tar.gz -o wego.tar.gz
tar -xf wego.tar.gz
chmod +x wego
sudo mv ./wego /usr/local/bin/wego
wego version
```

Please see the [getting started guide](https://docs.gitops.weave.works/docs/getting-started).

## CLI Reference

```console
$ wego 
Weave GitOps

Usage:
  wego [command]

Available Commands:
  add         Add a workload repository to a wego cluster
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

## CLI development

To set up a development environment for the CLI

1. Install go v1.16
2. make 


## UI Development

To set up a development environment for the UI

1. Install go v1.16
2. Install Node.js version 14.15.1
3. Install reflex for automated server builds: go get github.com/cespare/reflex
4. Run `npm install`
5. To start up the HTTP server with automated re-compliation, run `make ui-dev`
6. Run `npm start` to start the frontend dev server (with hot-reloading)
