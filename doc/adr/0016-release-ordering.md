# 0016. Change release ordering to only publish GitHub release when artifacts are available 

## Status

Proposed

## Context

Current release workflow ordering published a GitHub release before the artifacts are available. That flags users 
that they can use the new release however if they follow the signal of the GitHub release, they will fail the release 
as the container image won't be available at that time. 

## Decision

Change release ordering to ensure that we only publish the GitHub release after the container image is published.


## Consequences

Weave Gitops users that upgrades after a GitHub release has been publishes will not fail the upgrade. There is a consideration/limitation to this point: the helm chart is published in another workflow that only 
is triggered once the PR is merged. Therefore, there is still a gap between when the GitHub Release is published  and all the artifacts are available. 
This gap  in terms of time is small to compare the one that this ADR addresses, but it also needs to be addressed .
