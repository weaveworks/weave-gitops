# 11. CLI verb noun

Date: 2021-12-03

## Status

Accepted

Supercedes [3. CLI using noun verb format](0003-cli-using-noun-verb-format.md)

## Context

After creating a few CLI commands, we reevaluated whether noun-verb was the
right approach.  We decided that matching existing tools in the space was more
important than a strict noun-verb syntax.

## Decision

Update the Weave GitOps CLI to have a verb-noun syntax.  E.g., `gitops add app`.

## Consequences

* Need to update the existing commands (complete)
* If and when we need to add a new entity, i.e., plugin, with different verbs,
we will need to decide between adding a new verb and handling that with legacy
entities or take an approach similar to `kubectl` where the new object is
handled with noun-verb syntax to reduce the change to existing entities.
