# Contributing

Weave Gitops is a [Apache-2.0](LICENSE) project. This is an open source product with a community
led by volunteers interested in the brilliant software originally created by Weaveworks. :heart:

We welcome improvements to reporting issues and documentation as well as to code.

## Developer Certificate of Origin

By submitting any contributions to this repository as an individual or on behalf of a corporation, you agree to the [Developer Certificate of Origin](DCO).

## Understanding how to run development process

The [internal guide](doc/development-process.md) **is a work in progress** but aims to cover all aspects of how to
interact with the project and how to get involved in development as smoothly as possible.

## Acceptance Policy

These things will make a PR more likely to be accepted:

- should address a single concern or task
- avoid bundling unrelated changes
- a well-described requirement
- tests for new code
- tests for old code!
- new code and tests follow the conventions in old code and tests
- a good PR title and description (see below)
- all code must abide by [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- names should abide by [What's in a name](https://talks.golang.org/2014/names.slide#1)
- code must build on both Linux and Darwin, via plain `go build`
- code should have appropriate test coverage and tests should be written
  to work with `go test`

In general, we will merge a PR once at least one maintainer has endorsed it. For substantial changes, more people may become involved, and you might get asked to resubmit the PR or divide the changes into more than one PR.

## Format of the pull request

This project enforces that your pull request title match the
[Conventional Commits specification](https://conventionalcommits.org).
When the PR is merged, the full PR changeset will be squashed into a single
commit on the default branch using the PR title as the commit message.
While this approach might surprise some, this is done to simplify the release
process including generating (good) release notes.

The commit message on the default branch will automatically include a reference
back to the PR, so please explain what and why in the PR description.

Each pull request should address a single concern or task.
Avoid bundling unrelated changes into one PR, as it can make the review process
more challenging and less efficient. Keeping PRs focused helps reviewers
understand the purpose and context of the changes.
