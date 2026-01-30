# 0017. Defining user documentation categories 

## Status

Proposed

## Context

Weave GitOps user documentation could be found here https://docs.gitops.weave.works/docs/intro-weave-gitops/ 

It has navigation menu on the left hand side ![documentation-navigation.png](imgs%2Fdocumentation-navigation.png)
To help users finding relevant information in a natural way. 

User documentation is regularly added by prodeng with two possible scenarios:
- i am extending existing section so i dont need to do decisions in terms of navigation.
- i am adding new page that might require a new navigation item so i need to do decision on where to add it.

Given that we dont have explicit guidance on this topic, it is likely that two docs contributors might follow different 
considerations that might resul in inconsistent navigtion translating into user experience degration. 

In order to help mitigating this risk, we decide to raise this ADR that defines the principle ruling the navigation 
design and the categories to think about when adding documentation that requires navigation.

## Decision

### Design Principles

1. Navigation should intuitively follow the lifecycle of a platform engineer using the product.
2. The lifecycle the Platfrom engineer transitions is:
   - Platform engineer gets started with the product or day 1.
   - Platform engineer gets started with the product features or day 1.
   - After got started it gets into day 2 or operating the product. 
   - After the platform engineer looks into how to expand the product reach by integrating with other tools.

### Categories

Matching the lifecycle, we have the following categories

1. Introduction / Getting Started = Our current one
2. Capabilities: it includes cluster mangements, gitopssets, etc …
3. Operations -> this includes anything  ops: monitoring, logginc, etc
4. Integrations -> other systems that we integrate  -> this includes Backstage
5. Guides / Tutorials

### Categories in Practice

// to expand on why the existing documentation matches categories

## Consequences

Now we have some support to do aligned decisions regarding where to put new documentation that translates into new navigation items. 

