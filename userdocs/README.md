# Writing and publishing user docs

The user docs are written in [MkDocs](https://www.mkdocs.org/) using [Material for MkDocs](https://squidfunk.github.io/mkdocs-material/). Install `mkdocs` by following instructions [here](https://squidfunk.github.io/mkdocs-material/getting-started/)

## Commands

`cd userdocs` then you can run:

* `mkdocs serve` - Start the live-reloading docs server.
* `mkdocs build` - Build the documentation site.
* `mkdocs -h` - Print help message and exit.

## Project layout

    mkdocs.yml    # The configuration file.
    docs/
        index.md  # The documentation homepage.
        ...       # Other markdown pages, images and other files.
