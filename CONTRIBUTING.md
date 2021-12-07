# WIP: Internal contributing guide

## Weave Gitops :heart:s your contributions

Thank you for taking the time to contribute to Weave Gitops!

This guide **is a work in progress** but aims to cover all aspects of how to
interact with the project and how to get involved in development as smoothly as possible.

If we have missed anything you think should be included, or if anything is not
clear, we also accept contributions to this contribution doc :smile:.


## Testing
We have 3 layers of testing:
- Unit testing
- Integration testing
- Acceptance testing

We aim to follow the [testing pyramids best practises](https://martinfowler.com/articles/practical-test-pyramid.html)
with the focus on a large number of units tests, fewer integration tests and even fewer acceptance tests.


### Running unit tests
You can run `make unit-tests` from the top of the repository to run all of the unit tests.
As part of running the unit tests we will install any required dependencies for building
the code. See [dependencies](tools/dependencies.toml) for more information on whats installed.


### Running integration tests
The integration tests can be run using `make integration-tests`. They require some
extra environment variables to run:

- `GITHUB_ORG`- The name of the org that test repositories can be created in.
if your a weaveworks employee you can use `weaveworks-gitops-test`
- `GITHUB_TOKEN`- Used to create, push and delete repoitories. Please ensure
the token has sufficient permissions, as deleting repositories is not part
of the standard set of permissions a token is given.

If you wish to run the full suite, you will also need the equivalent for gitlab:
- `GITLAB_TOKEN`
- `GITLAB_ORG`

However if you only want to test it against one of the providers, you can focus in
the test suite manually

### Running acceptance tests
The acceptance tests can be run using `make acceptance-tests`. They require the same
environment variables as the integration tests, with the following in addition:

- `WEGO_BIN_PATH`- The path to the `gitops` binary the tests should use. If you
are want to test a locally built binary (created from `make bin`) then you can
set it to `WEGO_BIN_PATH=$PWD/bin/gitops`
- `IS_TEST_ENV`- If this value is set to anything, it indicates to the `gitops` binary
that it should deploy images using the `latest` tag, instead of the latest version.

#### Testing locally changes in acceptance tests
As of right now its not easy to test your local code changes pass the acceptance test.
Until [#1158](https://github.com/weaveworks/weave-gitops/issues/1158) is resolved you must:

1. Manually build and push the `wego-app` docker image to your own registry. Example:
`docker build -f Dockerfile -t aclevernameww/wego-app:latest . && docker push aclevernameww/wego-app:latest`
2. Update the `manifests/wego-app/deployment.yaml.tpl` to use your registry. Example:
`image: aclevernameww/wego-app:{{.Version}})
3. Rebuild the local gitops binary via `make bin`
4. Ensure `WEGO_BIN_PATH` is set to the local binary and `IS_TEST_ENV=true`

This should result in the `gitops install` run during the acceptance tests deploying your custom built docker images with
the desired code changes.

