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

Feature:
  Scenario: can build non-release branches without changing the flow nor overhead
    Given a non release pr
    When build
    Then CI workflow  passes
    And Release process build passes
    Then I could merge

  Scenario: can build release branches with guardrails
    Given a release pr
    When build
    Then CI workflow  passes
    And cannot merge it
    When human approves it
    Then release process starts
    And cannot merge it
    When release process ends
    Then can merge it
```
**Scenario A: non release PR**


**Scenario B:  release PR**

