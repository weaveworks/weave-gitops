# weave-gitops

Weave GitOps core

[![Coverage Status](https://coveralls.io/repos/github/weaveworks/weave-gitops/badge.svg?branch=main)](https://coveralls.io/github/weaveworks/weave-gitops?branch=main)
![Test status](https://github.com/weaveworks/weave-gitops/actions/workflows/test.yml/badge.svg)
[![LICENSE](https://img.shields.io/github/license/weaveworks/weave-gitops)](https://github.com/weaveworks/weave-gitops/blob/master/LICENSE)
[![Contributors](https://img.shields.io/github/contributors/weaveworks/weave-gitops)](https://github.com/weaveworks/weave-gitops/graphs/contributors)
[![Release](https://img.shields.io/github/v/release/weaveworks/weave-gitops?include_prereleases)](https://github.com/weaveworks/weave-gitops/releases/latest)

## UI Development

To set up a development environment for the UI

1. Install go v1.16
2. Install Node.js version 14.15.1
3. Install reflex for automated server builds: go get github.com/cespare/reflex
4. Run `npm install`
5. To start up the HTTP server with automated re-compliation, run `make ui-dev`
6. Run `npm start` to start the frontend dev server (with hot-reloading)
