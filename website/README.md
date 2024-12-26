# Weave GitOps Product Website

This website is built using [Docusaurus 2](https://docusaurus.io/), a modern static website generator.
It requires Node.js installed on your system. Install or Update your current setup for Node.js from [here](https://nodejs.org/en/download/)
Install yarn from [here](https://classic.yarnpkg.com/en/docs/install)

The website is live [here](https://docs.gitops.weaveworks.org/).

## Developing the docs

### Start a local development server

1. Change into the `website` directory: `cd website`.

1. Install dependencies: `yarn install`.

1. Start the docs server.

    > If you are using newer version of Node (17+):
    > ```bash
    > export NODE_OPTIONS=--openssl-legacy-provider
    > ```

    ```bash
    GA_KEY=fake ALGOLIA_API_KEY=fake yarn start
    ```

    This will open a browser window. Your changes will be automatically loaded by
    the server.

    *to view latest local changes, update the url with `next`* e.g. `http://localhost:3000/docs/next/intro-weave-gitops/`

### Making changes to the docs

Below are some things to be aware of when adding content, such as where we store
things and how we manage the sidebar. For general info on how to build a
Docusaurus site, check [their docs](https://docusaurus.io/).

Generally if you need to add new content to the documentation website you need to add changes to the `/docs` folder.

#### Directory layout

- The docs are kept in `docs/`
- Any images or assets are kept in `static/`
- Custom CSS is kept in `src/css`
- Non-versioned (ie non-product) docs are kept in `src/pages`
- Older versions of docs are kept in `versioned_docs`
- React components used by files under `docs/` are kept in `docs/components`

#### General notes on style and content

Not all the below will apply, take them as hints.

- An intro explaining why they would want this thing
- A summary of what the tutorial will achieve
- Prerequisites
- Notes on assumed knowledge
- Screenshots (ones you can actually read)
- Emojis
- Colourful [admonitions](https://docusaurus.io/docs/markdown-features/admonitions)
- Short paragraphs (no big wall of text)
- Bullet points
- Getting too long? Break down into multiple pages
- Enterprise tags (if required)

#### How to add a new page to a sidebar

Sidebars are managed in the `sidebars.js` file.
More information on how to add/modify sidebars can be found [here](https://docusaurus.io/docs/sidebar)

#### How to do internal links

Please find guidance from Docusaurus [here](https://docusaurus.io/docs/markdown-features/links)
and [here](https://docusaurus.io/docs/versioning#link-docs-by-file-paths).

#### Making changes to older versions
**NOTE** DO NOT edit old versions in `versioned_docs/` folders unless its absolutely required.

Most of the time you will not need to touch the contents of `versioned_docs/`
but occasionally you may need to backfill fixes or missing feature
documentation.
In that case, once you have made the changes to the `main` docs under `docs/`,
copy the changes over to the correct version in `versioned_docs/`. If you need
to change the sidebar at that point, change the correct version in
`versioned_sidebars`.

#### How to change the navbar

The navbar is controlled in `docusaurus.config.js`. It is important to note that
this is not (and cannot be) versioned, meaning that if you change the navbar
that change will be present for all versions.

#### How to format yaml

When formatting `yaml` code, use two spaces (not tabs) per indentation level. For example, the following yaml:

```yaml
---
apiVersion: v1
kind: ServiceAccount
metadata:
		name: demo-01
		namespace: default

```

should be

```yaml
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: demo-01
  namespace: default
```

In our eyes it may appear the same, but yaml parsers may complain.

#### How to add yaml/json/console output without making it painful

Our product comes with a lot of `yaml`, and we need to document that. However,
too much can saturate a page and make it hard to read. It can also be easy to
miss bits of text lying in between large chunks of `yaml`. The solution is to
collapse large `yaml`, `json` or console output sections with `<details>` tags.

Smaller chunks are usually fine. But for anything more that 10 lines it is
neater to wrap it.

~~~markdown
<details>
<summary>Expand to see this awesome HelmRepository</summary>

Don't forget to leave a line break below the opening tag...
```yaml
---
woohoo:
  such: yaml
```
and before the closing one.

</details>
~~~

<details>
<summary>Expand to see this awesome HelmRepository</summary>

Don't forget to leave a line break below the opening tag...
```
---
woohoo:
  such: yaml
```
and before the closing one.

</details>

#### How to add screenshots

Please, please, zoom in before you take the shot. I promise they don't need to
see the whole page, just the bit you are talking about is fine.

### After you have finished your changes

Check that the docs build with `GA_KEY=fakekey yarn build`.

The build will also be run on the PR, but it is convenient to run it before you
open one.

## Deploying the docs

Do not attempt to deploy the docs manually.

Deployment happens automatically (once the tests pass) upon merging to the
`main` branch: see [.github/workflows/docs.yaml](.github/workflows/docs.yaml) for config.

## Terminology

This section lists commonly used terms in the docs in an effort to promote consistency.

- `Leaf clusters`: used in Weave GitOps Enterprise to describe any CAPI or non-CAPI clusters that appear in the `Clusters` section of the UI. Typically, these clusters are running application workloads.
