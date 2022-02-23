# 6. When to use wego, weave-gitops, and Weave GitOps

Date: 2021-07-26

## Status

Accepted

## Context

Our code, documentation, and Kubernetes object naming has a mixture of wego, weave-gitops, and Weave GitOps.  This ADR guides when and where you should use one over the other.

## Decision

### Weave GitOps
Use in all user-facing documentation, except when presenting actual CLI commands the user can execute.  It should be spelled out and follow this capitalization `Weave GitOps`.

Including:
* go docs for functions, packages, and variables
* Online documentation
* Blogs

### wego
* The name of the CLI binary.
* The API group will be `wego.weave.works`
* The default Kubernetes namespace will be `wego-system`
* When naming Weave GitOps objects in Kubernetes, they will have a `wego-` prefix
* Code variables - developer choice
* Code comments - developer choice, except for public facing docs
* Release artifacts if it consists of only the wego binary for the os and architecture.  e.g., `wego-darwin-x86_64`

### weave-gitops
* Name of code repository for Weave GitOps core
* **new** The release packages will be prefixed with `weave-gitops` when they comprise more than just the wego binary
* **new** When we have additional distribution packages, they will use `weave-gitops`. e.g., `brew install weave-gitops`


## Consequences

Using the shorter version, wego, as a prefix for naming objects, reduces the number of characters required and makes it obvious what the thing is part of Weave GitOps.
