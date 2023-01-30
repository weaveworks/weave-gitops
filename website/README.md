# Website

This website is built using [Docusaurus 2](https://docusaurus.io/), a modern static website generator.
It requires Node.js installed on your system. Install or Update your current setup for Node.js from [here](https://nodejs.org/en/download/)
Install yarn from [here](https://classic.yarnpkg.com/en/docs/install)

Following commands can be run within the `website` folder.

## Installation

```console
yarn install
```

This will install all the dependencies required for the documentation site.

## Local Development

Set a fake Google Analytics API key to pass validation errors:

```shell
export GA_KEY=fakekey
```

```console
yarn start
```

This command starts a local development server and opens up a browser window. To preview your changes as you edit the files, you can run a local development server that will serve your website and reflect the latest changes.

If you are using newer version of Node (17+):

```bash
export NODE_OPTIONS=--openssl-legacy-provider
```

## Build

```console
yarn build
```

This command generates static content into the `build` directory and can be served using any static contents hosting service.
During development once you're happy with your changes, you should run this command to build the site locally to make sure all your changes are good.
This command can be bit slow and may take some time to build the site. 

## Deployment

Deployment happens automatically (once the tests pass) upon merging to the default branch: see [.github/workflows/docs.yml](.github/workflows/docs.yml) for config.
