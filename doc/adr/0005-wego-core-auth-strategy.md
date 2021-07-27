# 5. Weave GitOps Core Auth Strategy

Date: 2021-07-22

## Status

Proposal

## Context

Weave GitOps needs to be able to do read and write operations against three different "back ends":

- The git repository for an Application
- The repository "host" (Github, Gitlab, Bitbucket, etc), known as a Git Provider, for an Application
- The Kubernetes cluster where we want to run the Application. (**Note: this doc does not seek to address K8s authn/authz. That will need to be a separate ADR)**

Each of these back ends have different authn/authz requirements:

### Git Repository

The operations for the git repository are the typically the ones done with the `git` CLI, ie: `git clone`, `git commit`, `git push`. We assume that git providers will allow for repository-specific keys to be added; GitHub and Gitlab calls these "deploy keys", other providers might call them "access keys". These are typically ssh key-pairs.

These deploy keys are outside the scope of a git repo, so we will need a higher level of access in the form of a Git Provider access token to add deploy keys programmatically.

### Git Provider

Operations such as creating pull requests (or "merge requests" in the case of Gitlab), adding deploy keys, and retrieving user data will need to be completed via HTTP requests. Each major Git Provider has some form of OAuth flow:

- [Github Device Flow](https://docs.github.com/en/developers/apps/building-github-apps/identifying-and-authorizing-users-for-github-apps#device-flow)
  - This approach doesn't require a `CLIENT_SECRET` or callback page
  - We can do these steps from a browser if we feel that is the best UX
- [Gitlab Proof Key for Code Exchange (PKCE) Flow](https://docs.gitlab.com/ee/api/oauth2.html#authorization-code-with-proof-key-for-code-exchange-pkce)
  - Note that the documentation explicitly calls out that PKCE is optimal for client-side apps without a public cloud server (our use case)
  - The documentation example specifies a `CLIENT_SECRET`, but that may be a documentation bug. A `CLIENT_SECRET` should not be necessary
- [Bitbucket Implicit Grant](https://developer.atlassian.com/cloud/bitbucket/oauth-2/#2--implicit-grant--4-2-)
  - No `CLIENT_SECRET` required here

### Kubernetes Cluster

All operations will need to go through an API server that is installed on the cluster via the `wego gitops install` command. This API server will run with a service account that provides the ability to do the following:

- Create `Namespace` resources
- Create `Application`, `Source`, and `Kustomize|Helm` resources
- Read Kubernetes objects in the `wego-system` namespace

### Security Concerns

Each of the Git Providers listed above support some sort of "personal access token" that allow for "impersonation" style authorization. These tokens never expire, and provide "sudo" access to the user's Github account.

For this reason, we want to utilize the more short-lived OAuth tokens to avoid exposing a personal access token. We also don't want to persist these tokens longer than we have to.

In addition, flows that require a `CLIENT_SECRET` will require us to embed it in the binary, which leaves us open to exposing it via decompilation.

### UX Concerns

Given that the Weave GitOps UI and API server may be running on a user's cluster, we may not always know the OAuth callback URL ahead of time. For this reason, we will need to use OAuth flows that do not require a static `redirect_uri`. The flows listed above in the Git Providers section fit that constraint.

We will need to ensure that the `redirect_uri` is configurable by the user via environment variables. This will allow users to run and expose the UI/API servers on whichever endpoint they choose.

## Design

Since most modern Git Providers will support a form of browser-based OAuth, **we will utilize the browser for both the CLI and UI authentication with Git Providers**.

In the case of the CLI, we can utilize a short-lived browser session that will recieve the OAuth callback and pass the token to the CLI.

**Note that we can skip the OAuth portion of the flow if a user provides a `--token=<some value>` flag to the CLI command:**

![CLI Auth Diagram](cli_auth.svg)

UI auth will work in a similar way, with a more straight-forward set of steps:

![UI Auth Diagram](ui_auth.svg)

TLDR:

1. The user does OAuth with their Git Provider
2. We use the resulting access_token to push a Deploy Key to the repository (giving us `git` permissions)
3. We create a pull request for the repository via the Git Provider HTTP API

For browser security, we will convert the Git Provider OAuth token to a JSON Web Token (JWT) to protect against Cross-site Scripting (XSS) attacks. The encrypted JWT will allow a malicious script to authenticate with the Weave GitOps API only, whereas passing the unencrypted OAuth token to the browser would allow a malicious script to authenticate with the Github API.

Additionally, we do not plan on adding third-party scripts to the Weave GitOps UI to minimize the surface area for XSS attacks. This does NOT, however, account for NPM modules or other dependencies that we add to our app at build time.

## Prior Art

This authentication approach is inspired by other CLI tools that have very smooth user experiences. For example, the `gcloud` CLI for Google Cloud Platform utilizes the browser to authenciate the user:

```
$ gcloud auth login

Your browser has been opened to visit:

https://accounts.google.com/o/oauth2/auth?response_type=code&client_id=1234.apps.googleusercontent.com&redirect_uri=http%3A%2F%2Flocalhost%3A8085%2F&scope=...

Opening in existing browser session.

```

In the browser:

![gcloud auth UI](gcloud_auth.png)

<hr />

GitHub's own CLI works in a similar way, and allows users to pass in their own token:

[https://cli.github.com/manual/gh_auth_login](https://cli.github.com/manual/gh_auth_login)

```
Authenticate with a GitHub host.

The default authentication mode is a web-based browser flow.

Alternatively, pass in a token on standard input by using --with-token. The minimum required scopes for the token are: "repo", "read:org".
```

## Decision

Implement the flow(s) defined in the "Design" section

## Consequences

This document does not seek to provide any mapping between Kubernetes users and Git Provider users. That may be out of scope for the Weave GitOps Core edition, or may be implemented later.
