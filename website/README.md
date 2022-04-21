# Website

This website is built using [Docusaurus 2](https://docusaurus.io/), a modern static website generator.

## Installation

```console
yarn install
```

## Local Development

Set a fake Algolia API key to pass validation errors:

```shell
export ALGOLIA_API_KEY=fakekey
export GA_KEY=fakekey
```

```console
yarn start
```

This command starts a local development server and opens up a browser window. Most changes are reflected live without having to restart the server.

## Vscode Devcontainer for Docs
If you are using Vscode/docker and navigate to the website directory you may be prompted that there is an included container. This utilises [devcontainers](https://code.visualstudio.com/docs/remote/containers) to provide a clean environment with the tooling necessary to supplement the docsite. The devcontainer will automatically start a copy of the docs localhost:8080 and run in the background in parallel to the native dev build on port 3000.

## Build

```console
yarn build
```

This command generates static content into the `build` directory and can be served using any static contents hosting service.

## Deployment

Deployment happens automatically (once the tests pass) upon merging to the default branch: see [.github/workflows/docs.yml](.github/workflows/docs.yml) for config.
