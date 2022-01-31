# Application notifications

## Overview
An application can be comprised of 1 or more components. These components can come from one or more git sources. A user will want to define notification providers and events at the application level. We will leverage [Kubernetes common labels](https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/) on flux resources (source, kustomization, helm release). Users may want to define different providers for different environments of their apps (dev, QA, EU, US, etc.).

See [issue](https://github.com/weaveworks/weave-gitops/issues/1349) for complete details 

Events available in Flux: (from https://fluxcd.io/docs/guides/notifications/)

Info
* a Kubernetes object was created, updated or deleted
* heath checks are passing
* a dependency is delaying the execution
* an error occurs

Error
* any error encountered during the reconciliation process

## Terminology 
* **Flux Source** A source in the Flux sense
* **Flux Applier** A flux kustomization or helm release
* **App Instance** An application + environment running on a cluster.  There can be many app instances on a cluster.

## Target Outcome/Requirements
* Single Alert defintion per application
* Application environment can modify/overwrite Alert definition
* Single Alert instance per app instance
* Alert instance lifecycle is tied with the app instance lifecycle
* Flux Providers can be shared between Alert Instances.  
* When an app instance is deployed, its Alert Instance (if defined) should also be deployed
* Alerts must be included in the Application status

## Proposal
Create Alerter and Provider that are application aware, dynamic, and fit into the existing flux notification system.

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

### Operator based on secret resources
Create an operator that watches the cluster for annotated secrets. When one is found, create corresponding Alters and Providers.

_Working prototype_

#### Pros
- User is only required to create a secret; the rest happens automatically based on annotations
- Alerts and Providers follow the secret lifecycle. i.e., when the secret goes away, so does the Alerts/Providers

#### Cons
- Users may not have access to the secret manifests to annotate
- Difficult to replicate all the fields available in Alerts/Providers as annotations
- Alerts/Providers not stored in git as they are dynamically generated 
- Another component running in the users' system

### Operator based on flux primitives
Tweak to the _operator based on secrets_ alternative. Instead of tying the Alert and Provider generation to the secret, tie it to the flux source object. Once we have the flux source object, we can determine the downstream consumers and create the appropriate Alerts and Providers based on that.

#### Pros
- The user creates a Flux applier with annotations, the operator generates the Alerts/Providers
- Alerts and Providers follow the flux applier lifecycle.  

#### Cons
- Difficult to replicate all the fields available in Alerts/Providers as annotations
- Alerts/Providers not stored in git as they are dynamically generated 
- Another component running in the users' system

### Sidecar with flux applier
Using a mutating admission controller injects a sidecar to any matching flux applier pod that will run along with the flux applier and use it to manage alerts and providers for this particular flux applier.

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

### Application provider
Create an application provider that uses the involved object for the event, gathers the application labels, and writes structured logging records. These structured logs can then be collected and used in tools like grafana with loki to graph activity by application
#### Pros
- Application-aware
- Log data can be accumulated in logging solutions defined by the customer
#### Cons
- None

---

|Alternative|Application Aware|Artifacts in Git|Dynamic|Security|Scale|Speed to deliver|
|---|---|---|---|---|---|---|
|Existing flux tooling||*||||High|
|wego API(s) to generate|*|*||||Med|
|Operator based on secret resources|*||*|||Med|
|Operator based on flux primitives|*||*|||Med|
|Sidecar with flux applier|*||*|||Low|
|Application Provider Dispatcher|*|*||||Med|
|Application Alerter|*|*|*|||Med|
|Application Provider|*|*|*|||Med|
