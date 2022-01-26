# Application notifications

## Overview
An application can be comprised of 1 or more components. These components can come from one or more git sources. A user will want to define notification providers and events at the application level. We will leverage [Kubernetes common labels](https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/) on flux resources (source, kustomization, helm release). Users may want to define different providers for different environments of their apps (dev, QA, EU, US, etc.).

See [issue](https://github.com/weaveworks/weave-gitops/issues/1349) for complete details 

## Alternatives
### Existing flux tooling
Utilize existing tools `flux create alert` and either export it to the cluster git-repo or just have it generated directly in the cluster. Perform the same operations for providers. The user manages secrets used by the provider manually.

#### Pros
- Already available
- Ultimate flexibility

#### Cons
- Alerts and Providers are independent of the secrets, flux kustomizations/helm releases. i.e., when they are removed from the cluster, there isn't a need to have Alerts and Providers
- Not actively managing the set of alerts and providers
- Not application aware 

### wego API(s) to generate
Create APIs that can interrogate the cluster and generate the set of Alerts and Providers. Either API or user checks these into the cluster repo.

#### Pros
- Less manual work by the user, i.e., API "knows" what Alerts/Providers to generate

#### Cons
- Alerts and Providers are independent of the secrets, flux kustomizations/helm releases
- Potential of requiring the user to re-generate the Alerts and Providers when new flux kustomizations/helm releases are added/changed
- Not actively managing the set of alerts and providers
- Difficult to replicate all the fields available in Alerts/Providers as annotations

### Create an operator based on secret resources
Create an operator that watches the cluster for annotated secrets. When one is found, create corresponding Alters and Providers.

_Working prototype_

### Pros
- User is only required to create a secret; the rest happens automatically based on annotations
- Alerts and Providers follow the secret lifecycle. i.e., when the secret goes away, so does the Alerts/Providers

### Cons
- Users may not have access to the secret manifests to annotate
- Difficult to replicate all the fields available in Alerts/Providers as annotations
- Alerts/Providers not stored in git as they are dynamically generated 
- Another component running in the users' system

### Create an operator based on flux sources
Tweak to the _operator based on secrets_ alternative. Instead of tying the Alert and Provider generation to the secret, tie it to the flux source object. Once we have the flux source object, we can determine the downstream consumers and create the appropriate Alerts and Providers based on that.

### Pros
- The user creates a Flux kustomization/helm release with annotations, the operator generates the Alerts/Providers
- Alerts and Providers follow the flux kustomization/helm release lifecycle.  

### Cons
- Difficult to replicate all the fields available in Alerts/Providers as annotations
- Alerts/Providers not stored in git as they are dynamically generated 
- Another component running in the users' system

### Sidecar with flux kustomization/helm release
Using a mutating admission controller injects a sidecar to any matching flux kustomization/helm release pod that will run along with the flux kustomization/helm release and use it to manage alerts and providers for this particular flux kustomization/helm release.

#### Pros
- Dynamic and catches objects as they are added
- leverages existing flux providers
- Cool to write

#### Cons
- Potentially overkill
- Would folks trust/allow mutating admission controller

### Application Provider Dispatcher
Create an Application Provider to be the provider for all alerts. This dispatcher is configured with rules by the user to determine what alerts go where. These rules can be stored in git, managed application by application.  

#### Pros
- Application-aware alert routing
- leverages existing flux providers

#### Cons
- Sort of duplicates existing flux Alerts/Providers - only with more coupling

### Application Alerter
Create a new type of Flux Alerter that is application-aware. It watches all kustomization/helm events and filters them based on application. It augments the event with application data then utilizes the providers as they exist today. E.g., selecting all kustomizations with the label `app.kubernetes.io/name: billing` and watching for events from these.

#### Pros
- Application-aware
- leverages existing flux providers

#### Cons
- Could be done without creating a new type of alert
  - Could be accomplished using an existing Alert with eventSource field populated with all kustomization object names in use by the application
    - The downside is keeping this up to date as kustomizations/helm releases come and go 




