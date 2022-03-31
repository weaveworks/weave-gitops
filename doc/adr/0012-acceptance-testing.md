# 0012. Acceptance Testing in Weave GitOps Core - Immediate Action

## Status

Proposed

## Problem

As part of the Core Reloaded Project we have removed the existing acceptance tests within the Weave GitOps Core repo. With the removal of a large number of `gitops` CLI commands, this is a good opportunity to evaluate the testing strategy for the short-mid term. This decision aims to codify discussions around the short/mid-term goals of acceptance testing (or lack thereof) within the GitOps Core product. 

Clear the Way aims to provide a new foundation for the future development of the gitops core product, and therefore much of the existing CLI functionality has been removed and a new approach taken to how we interact with Flux. This resulted in a lack of clarity regarding the purpose and direction of the existing acceptance test, which along with a number of issues with the tests themselves lead to the decision to remove all existing acceptance tests.

## Additional Context

Before investigating possible ways forward, it was important to understand what we do and don’t want to achieve with acceptance tests.

### Do

- Validate functionality from end-to-end or across multiple components. Testing individual components is best kept within the bounds of other testing methods.
- Validate interface contracts and integration entry/exit point boundaries.
- Where necessary, validate core user interface functionality.

### Don’t

- Test functionality of external components or third party services. E.g. Flux, Kubernetes.
- Within reason, tests should not rely on third party services *outside of our control* - e.g. cloud services. (Within the test itself, not infrastructure).

As a general rule acceptance tests should be the smallest number of tests within the overall test suite. It’s easy for them to become flaky and difficult to maintain, especially with the considerable setup and teardown requirements.

Acceptance Test requirements:

1. Should be able to run locally, preferably with a single command.
2. Where possible, tests should run in parallel and in isolation (for some definition of isolation).
3. Failures should be consistent and reproducible (as far as is practically possible).

## Decision

As part of the Clear the Way initiative we will not be rewriting or reimplementing any acceptance or end-to-end test. Rather, additional time was invested in integration and unit tests. As the product is expected to evolve more over time, especially in the near future, building a new set of acceptance tests did not seem like a good use of time. Instead, a set of standards or strategies for acceptance / end-to-end test will be written detailing how we will approach them in the future.

## Consequences

This decision isn't intended to define a long term direction in regards to acceptance tests in Weave GitOps Core. At the current stage of the product there is little justification for a heavy investment in new acceptance testing implementations.
