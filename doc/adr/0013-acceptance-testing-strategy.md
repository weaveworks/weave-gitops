# 0013. Acceptance Testing Strategy / Philosophy

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

The purpose of this decision is **not** to set a series of rules that must be followed in future testing development. Rather it attempts to capture a set of preferences with how the team will approach acceptance testing going forward.

When approaching the use of acceptance tests within the project, consider the following guidelines:

1. Avoid the use of User Interface (UI) testing tools such as Selenium and Cypress for as long as is practical, especially while we focus on product-market fit. 
2. Utilise alternative method for testing user interface functionality, such as the [jsdom](https://github.com/jsdom/jsdom) project to simulate UI events within a virtual DOM.
3. Prefer to make use of alternative testing methods where possible. This is not just limited to unit and integration tests, but also [contract testing](https://pactflow.io/blog/what-is-contract-testing/) and API stubbing.

When approaching the introduction of acceptance tests within the project **PR and release process**, consider the following guidelines:

1. Ensure tests adhere to the philosophy of a short feedback loop. Tests should not considerably increase the time to merge or time to fail.
2. Do not introduce non-blocking tests into the CI pipeline, these will often end up being ignored.
3. All tests should run as early as possible with the CI and release process. Therefore, avoid introducing testing steps that only occur late with the process - e.g. Pre-release.
4. Design and execute tests in a manner that preferences towards early failure and clear error messages.

## Consequences

These preferences are a departure to our previous philosophy of utilising acceptance tests. While consequences should be minimal, there is a risk that the introduction of acceptance testing into the Weave GitOps Core product will require a higher than normal amount of early investment time, as initial design will be important for the long term health of the tests. The same can be said for the introduction of testing methods not previously employed within the project, such as contract based testing.

Additionally, arguments could be made that there is an inherent risk associated with the lack of proper visual UI testing, such as selenium or cypress tests. The utilisation of projects such as jsdom can alleviate some of this risk, but the benefits of not investing too early in these testing methods can have payoff in velocity.
