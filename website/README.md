# Website

This website is built using [Docusaurus 2](https://docusaurus.io/), a modern static website generator.

## Installation

```console
yarn install
```

## Local Development

Set a fake Google Analytics API key to pass validation errors:

```shell
export GA_KEY=fakekey
```

```console
yarn start
```

This command starts a local development server and opens up a browser window. Most changes are reflected live without having to restart the server.

If you are using newer version of Node (17+):

```bash
export NODE_OPTIONS=--openssl-legacy-provider
```

## Build

```console
yarn build
```

This command generates static content into the `build` directory and can be served using any static contents hosting service.

## Deployment

Deployment happens automatically (once the tests pass) upon merging to the default branch: see [.github/workflows/docs.yml](.github/workflows/docs.yml) for config.
