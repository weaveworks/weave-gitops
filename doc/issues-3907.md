# Improving OSS Release to avoid out of order issues  

https://github.com/weaveworks/weave-gitops/issues/3907

## Identified out of order paths

1. Release human merged release pr without CI job finish.
2. Release job steps ordering https://github.com/weaveworks/weave-gitops/blob/main/.github/workflows/release.yaml#L104-L107

## Release human merged release pr without CI job finish

We should be looking to avoid by design ending up in a human merging it. The events happening during the release process:

1. new release branch
2. releaser approves it
3. job releases starts
4. job release ends and branch is merged
5. release completed

What we want to avoid is the release by mistake to do anything after 2) with the following alternatives:

1. no human intervention has effect or not possible after 2
2. no human intervention required during the release process

###  no human intervention has effect or not possible after 2

We could achieve that by extending current `branch merge protection` to block on a new `release` status check:
- if not release branch does nothing -> noop 
- if release branch gets updated by the release job -> update happens after release 

Therefore, an attempt to merge a release would be blocked until the release does not happen


#### PoC 

Acceptance for this solution

```gherkin
Feature: can build non-release branches without changing the flow nor overhead
  Scenario: current flow  
    Given a non release pr
    When build
    Then CI workflow  passes
    And I could merge
```


```gherkin
Feature: can build release branches with guardrails

  Scenario: add a check to the PR that only gets passed when release goes on 
    Given a release pr
    And CI build check preventing merging
    And Release check preventing merging
    When build
    Then CI workflow passes so CI build check passes
    And cannot merge it cause release hasnt happened
    When human approves it
    Then release process starts
    And cannot merge it cause release hasnt happened
    When release process ends so release check passes
    Then can merge it
```






**Scenario B:  release PR**

