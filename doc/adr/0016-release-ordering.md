# 0016. Change release ordering to only publish GitHub release when container image is available 

## Status

Proposed

## Context

Current [release workflow](../../.github/workflows/release.yaml) ordering publishes a GitHub release before the artifacts are available. That flags users 
that there is a new release available when in reality the container image is not yet available. It has different undesired outcomes:
1. an upgrade following the notification will fail as the artifacts are not available.
2. given our current release failure handling, we release a patch version for releases that were not actually released. For example see in image below 
release 0.31.0, 0.31.1:

![release-failures-slack-notifications.png](imgs%2Frelease-failures-slack-notifications.png)


## Decision

Change release ordering to ensure that we only publish the GitHub release after the container image is published.


## Consequences

Weave Gitops users that upgrades after a GitHub release has been publishes will not fail the upgrade. It also, in case 
of issues in building container image, there won't be need to delete or patch releases as it won't be yet released.

There is a consideration/limitation to consider: the helm chart is published in [a workflow](../../.github/workflows/helm.yaml) that only 
is triggered once the PR is merged. Therefore, there is still a gap between when the GitHub Release is published  and all the artifacts are available. 
This gap  in terms of time is small to compare the one that this ADR addresses, but it also needs to be addressed.
