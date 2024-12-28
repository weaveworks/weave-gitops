---
title: Enterprise Releases

---


# Releases ~ENTERPRISE~

!!! info
    This page details the changes for Weave GitOps Enterprise and its associated components. For Weave GitOps OSS, please see the release notes on [GitHub](https://github.com/weaveworks/weave-gitops/releases).

## v0.31.0
2023-08-31

## Highlights

- ConnectCluster Functionality: Adding the foundation functionality to support connecting leaf clusters via CLI `gitops connect cluster`.
- Explorer extends source rendering to include OCIRepository resources to be rendered as regular flux sources. 
- [UI Enhancement] Improved Top-Right Applied Status and Time Text for Applications and Terraform Details pages. 

## Dependency versions
- flux >v2.0.0
- weave-gitops v0.31.2
- cluster-controller v1.5.2
- cluster-bootstrap-controller v0.6.1
- (optional) pipeline-controller v0.21.0
- (optional) policy-agent v2.5.0
- (optional) gitopssets-controller v0.15.3

## v0.30.0
2023-08-17

## Highlights

### UI

- UI token refreshing! OIDC token refreshing is now handled by the UI, this avoids unintentionally making multiple token requests to the OIDC provider. This old behaviour sometimes triggered rate limiting in OIDC providers, causing errors.
- UI polish including removing duplicate error messages and more consistency in headers and font sizes.

### Policy

- View Policy Audit violations in policies page as a tab

### GitOpsSets

* ClusterGenerator - return labels as generic maps, allows for easily using them in params.
* Handle empty artifacts in directory processing, if a `GitRepository` or `OCIRepository` has no artifact, stop generating with an error.
* Update the ImagePolicy generator to add the image. 
* Ignore empty generators in the Matrix generator, fixing a panic if a generator produced an empty list.


## Dependency versions
- flux >v2.0.0
- weave-gitops v0.30.0
- cluster-controller v1.5.2
- cluster-bootstrap-controller v0.6.1
- (optional) pipeline-controller v0.21.0
- (optional) policy-agent v2.5.0
- (optional) gitopssets-controller v0.15.3

## v0.29.1
2023-08-04

!!! warning
    This release builds upon Weave GitOps v0.29.0 that has breaking changes from Flux v2.0.0. Please make
    sure that you read [these release notes](#v0290).

### Dependency versions
- flux >v2.0.0
- weave-gitops v0.29.0
- cluster-controller v1.5.2
- cluster-bootstrap-controller v0.6.0
- templates-controller v0.3.0
- (optional) pipeline-controller v0.21.0
- (optional) policy-agent v2.5.0
- (optional) gitopssets-controller v0.14.1

### 🚀 Enhancements

- PR: #3126 - Uses weave-gitops [v0.29.0](https://github.com/weaveworks/weave-gitops/releases/tag/v0.29.0) that as
major changes include:
  - Support for Flux v2.0.0
  - Suspend/resume/reconcile Image Repositories
  - Add Verified sources to Applications and Sources UI

## v0.29.0
2023-08-03

!!! danger

    ### ⚠️ Breaking changes

    We introduced a breaking change in this release by upgrading to Flux v2 APIs, notably `GitRepository` v1, `Kustomization` v1, and `Receiver` v1. This means that this version of Weave GitOps Enterprise is not compatible with previous versions of Flux v2, such as v0.41.x and earlier.

    ### ✍️ Action required

    Follow [Flux](https://github.com/fluxcd/flux2/releases/tag/v2.0.0) or [Weave GitOps](https://docs.gitops.weave.works/docs/guides/fluxga-upgrade/) to upgrade to Flux v2 GA before upgrading Weave GitOps Enterprise.

### Highlights

#### Flux
- Using Flux v2.0.0 APIs. Managing `GitRepository` v1, `Kustomization` v1, and `Receiver` v1 resources. See Breaking Changes.

#### Explorer
- Generates metrics for indexer write operations

### Dependency versions
- flux >v2.0.0
- weave-gitops v0.29.0-rc.1
- cluster-controller v1.5.2
- cluster-bootstrap-controller v0.6.0
- templates-controller v0.3.0
- (optional) pipeline-controller v0.21.0
- (optional) policy-agent v2.5.0
- (optional) gitopssets-controller v0.14.1

### 🚀 Enhancements

- PR: #3137 - Upgrade to Weave GitOps OSS v0.29.0-rc.1 and Flux v2.0.0 APIs
- PR: #3119 - Bump GitOpsSets to v0.14.0
- PR: #3134 - add RED metrics for indexer writes
- PR: #3098 - [UI] Cleanup forms across sections to ensure consistency
- PR: #3145 - Wge 3144 - create sops secrets uses v1 kustomizations api
- PR: #3146 - generate v1 kustomizations when adding apps
- PR: #3164 - Bump gitopssets-controller to v0.14.1

### 🔥 UI

- PR: #3120 - Add large info display of Applied Revision and Last Updated on Terraform detail page
- PR: #3138 - Fix checkboxes on terraform data table

## v0.28.0
2023-07-20

### Highlights

- This release drops the requirement to install cert-manager
- Extends external secrets creation form to allow selecting multiple properties or all properties

#### UI

- Better support for organising your clusters and templates in the UI via namespaces
- Better support for Azure and Bitbucket Repositories in the UI, you can now click through to Open Pull Requests from these providers
- Dark Mode support for Policy Config

#### Explorer

- Adds support for Kubernetes Events 

### Breaking Changes

- This version of Weave Gitops Enterprise drops support for `v1alpha1` of the `CAPITemplate` and `GitopsTemplate` CRDs. Please migrate to `v1alpha2` of these CRDs. See the [migration guide](../gitops-templates/versions.md)

### Dependency versions

- weave-gitops v0.28.0
- cluster-controller v1.5.2
- cluster-bootstrap-controller v0.6.0
- templates-controller v0.3.0
- (optional) pipeline-controller v0.21.0
- (optional) policy-agent v2.5.0
- (optional) gitopssets-controller v0.13.4

## v0.27.0
2023-07-07

### Highlights

#### Explorer

- Retries to make sure we're showing you the freshest data
- Exports more metrics to enhance observability

#### GitOpsSets

- Config generator enabled by default! Include values from ConfigMaps and Secrets in your GitOpsSets

#### UI

- Dark mode enhancements
- Consistent form styling

### Dependency versions

- weave-gitops v0.26.0
- cluster-controller v1.5.2
- cluster-bootstrap-controller v0.6.0
- templates-controller v0.2.0
- (optional) pipeline-controller v0.21.0
- (optional) policy-agent v2.5.0
- (optional) gitopssets-controller v0.13.4

## v0.26.0
2023-06-22

### Highlights

- Dark Mode is now available in WGE.
- Added Prometheus metrics for all API endpoints.

### Dependency versions

- weave-gitops v0.26.0
- cluster-controller v1.5.2
- cluster-bootstrap-controller v0.6.0
- templates-controller v0.2.0
- (optional) pipeline-controller v0.21.0
- (optional) policy-agent v2.5.0
- (optional) gitopssets-controller v0.13.2

## v0.25.0
2023-06-08

_Bug fixes_

### Dependency versions

- weave-gitops v0.25.1-rc.1
- cluster-controller v1.5.2
- cluster-bootstrap-controller v0.6.0
- templates-controller v0.2.0
- (optional) pipeline-controller v0.21.0
- (optional) policy-agent v2.4.0
- (optional) gitopssets-controller v0.12.0

## v0.24.0
2023-05-25

### Highlights

#### GitOpsSets

- GitOpsSets can now generate from [Flux Image Automation](https://fluxcd.io/flux/components/image/)'s `ImagePolicy`. This allows you to include the latest version of an image in your templates, for example to keep a `Deployment` up to date.
- Cross namespace support lands, create resources in multiple namespaces, they'll also be cleaned up properly via finalizers.
- The **Sync** button in the UI now works correctly

#### Profiles and Charts

- You can now filter out the versions that will be shown from a HelmRepository when installing a chart via annotations:
  - `"weave.works/helm-version-filter": "> 0.0.0"` to filter out rc releases
  - `"weave.works/helm-version-filter": "> 1.0.0"` to filter any pre 1.0 releases
  - `"weave.works/helm-version-filter": "> 3.0.0-0"` to filter any pre 3.0 releases but include rc releases

#### Explorer
- You could now navigate by filters and enabled full-text search.

### Breaking Changes

(none)

### Known issues

#### Explorer

- Full-text search with terms including special characters like dashes (-) returns more results than expected by exact match term. For example, searching by term "flux-system" is treated as two terms: "flux" & "system" so returns the results for the joint of them.  A fix for this will be part of the following releases.

### Dependency versions

- weave-gitops v0.24.0
- cluster-controller v1.5.2
- cluster-bootstrap-controller v0.6.0
- templates-controller v0.2.0
- (optional) pipeline-controller v0.21.0
- (optional) policy-agent v2.3.0
- (optional) gitopssets-controller v0.12.0

## v0.23.0
2023-05-12

### Highlights

#### Application Details

- Health status is now displayed for Kubernetes built-in resources.

#### Explorer
- You could [configure the service account](https://docs.gitops.weave.works/docs/explorer/configuration/#authentication-and-authorization-for-collecting) to use for collecting resources.

#### Templates

- You can now provide a _Markdown_ description of a template. This will be rendered at the top of the Template page allowing template authors to provide clear instructions to their users on how to best fill in the values and complete any other required tests and checks.
- Raw templates are more flexible and allow you to render resources which don't have an explicit `metadata.name` field.

#### Cluster details

- The cluster details page now shows a Cluster's Connectivity status, along with more details for _both_ GitopsClusters and CAPIClusters, including:
  - Conditions
  - Labels
  - Annotations

#### Explorer

- When enabled [useQueryServiceBackend](https://docs.gitops.weave.works/docs/explorer/configuration/#setup) navigation from Clusters UI to Applications UI is not possible as Explorer does not yet support filtering.

### Dependency versions

- weave-gitops v0.23.0
- cluster-controller v1.4.1
- cluster-bootstrap-controller v0.6.0
- templates-controller v0.2.0
- (optional) pipeline-controller v0.21.0
- (optional) policy-agent v2.3.0
- (optional) gitopssets-controller v0.11.0



## v0.22.0
2023-04-27


### Highlights

#### Explorer

- Explorer supports now Flux sources.
- Applications UI and Sources UI could be configured to use Explorer backend to improve UI experience.
- Explorer collector uses impersonation. Ensure you [configure collector](../explorer/configuration.md/#authentication-and-authorization-for-collecting) for your leaf clusters.

#### GitopsSets

- Now supports correctly templating numbers and object chunks

#### Cluster Bootstrapping

- Don't wait for ControlPlane readiness to sync secrets, this allows secrets to be sync'd related to CNI or other early stage processes

### Upgrade Notes (from the previous release)

- Explorer: you should configure [collector service account](https://docs.gitops.weave.works/docs/explorer/configuration/#authentication-and-authorization-for-collecting) in your leaf clusters.

### Known issues

- Clusters page horizontally scrolls too much and status becomes unreadable for some fields

### Dependency versions

- weave-gitops v0.22.0
- cluster-controller v1.4.1
- cluster-bootstrap-controller v0.6.0
- templates-controller v0.2.0
- (optional) pipeline-controller v0.20.0
- (optional) policy-agent v2.3.0
- (optional) gitopssets-controller v0.9.0

## v0.21.2
2023-04-13

### Highlights

- See your gitopssets on leaf clusters in the UI
- Fixed bug where gitopssets would not update ConfigMaps
- View Open Pull requests button in the UI now allows you to select any GitRepository

### Dependency versions

- weave-gitops v0.21.2
- cluster-controller v1.4.1
- cluster-bootstrap-controller v0.5.0
- templates-controller v0.1.4
- (optional) pipeline-controller v0.20.0
- (optional) policy-agent v2.3.0
- (optional) gitopssets-controller v0.8.0

## v0.20.0
2023-03-30

### Dependency versions

- weave-gitops v0.20.0
- cluster-controller v1.4.1
- cluster-bootstrap-controller v0.5.0
- templates-controller v0.1.4
- (optional) pipeline-controller v0.20.0
- (optional) policy-agent v2.3.0
- (optional) gitopssets-controller v0.7.0

## v0.19.0
2023-03-16

### Highlights

#### UI

- Gitopsssets come to the UI!

### Dependency versions

- weave-gitops v0.19.0
- cluster-controller v1.4.1
- cluster-bootstrap-controller v0.3.0
- templates-controller v0.1.4
- (optional) pipeline-controller v0.20.0
- (optional) policy-agent v2.3.0
- (optional) gitopssets-controller v0.6.0

## v0.18.0
2023-03-02
### Highlights

#### UI

- See the logged in user's OIDC groups in the UI via the new User Profile page
- Image Automation pages now show cluster information
- See details about the configured promotion strategy for a pipeline
- Log filtering by source and level for GitOps Run
- See all Policy Configs listed in the UI

#### GitopsSets

- New `cluster` generator allows you to interact with the Weave GitOps Cluster inventory. GitOps Clusters that are added and removed to the inventory are reflected by the generator. That can be used to target for example to manage applications across a fleet of clusters.
- Enhanced `gitRepository` generator can now scan directories and paths with the new `directory` option, which enables you to create for example dynamically Flux Kustomizations , based on your repository.
- New `apiClient` generator allows you to query and endpoint, and provide data for your template.
- Reconciliation metrics are now reported to the `/metrics` endpoint ready to be collected


### Dependency versions

- weave-gitops v0.18.0
- cluster-controller v1.4.1
- cluster-bootstrap-controller v0.3.0
- templates-controller v0.1.3
- (optional) pipeline-controller v0.20.0
- (optional) policy-agent v2.3.0
- (optional) gitopssets-controller v0.5.0

## v0.17.0
2023-02-16
### Highlights

This release contains dependency upgrades and bug fixes. For a larger list of updates, check out the [Weave GitOps v0.17.0](https://github.com/weaveworks/weave-gitops/releases/tag/v0.17.0) release.

## v0.16.0
2023-02-02
### Highlights

#### Create External Secrets via WGE UI
- It's becoming easier to create new a external secret CR through the UI instead of writing the whole CR yaml.
- The creation form will help users choose which cluster to deploy the External Secret to and which secret store to sync the secrets from.
- It's all done in the GitOps way.

#### Plan Button in Terraform
- Adding **Add Plan** button in the terraform plan page to enable users to re-plan changes made.

### Dependency versions

- weave-gitops v0.16.0
- cluster-controller v1.4.1
- cluster-bootstrap-controller v0.3.0
- templates-controller v0.1.2
- (optional) pipeline-controller v0.14.0
- (optional) policy-agent v2.2.0
- (optional) gitopssets-controller v0.2.0

### Breaking changes

No breaking changes

## v0.15.1
2023-01-19
### Highlights

#### Multi Repository support. Weave GitOps Enterprise adapts and scales to your repository structure
- Weave GitOps Enterprise, is now supporting via the WGE GUI the selection of the Git Repository. Enabling to scale and match the desired Git Repository structure.

#### GitOps Templates
- Supporting path for Profiles, enabling to set the path for profiles in the template to configure where in the directory the HelmRelease gets created.
- Enhanced Enterprise CLI support for GitOps Templates.
#### GitOps Templates CLI enhancements
- Support for profiles in templates via CLI
- ```gitops create template``` supporting ```--config``` allows you to read command line flags from a config file and ```--output-dir``` allows you to write files out to a directory instead of just stdout
#### GitOpsSets in preview
- GitOpsSets enable Platform Operators to have a single definition for an application for multiple environments and a fleet of clusters. A single definition can be used to generate the environment and cluster-specific configuration.
- GitOpsSets has been released as a feature in preview of WGE. The Preview phase helps us to actively collect feedback and use cases, iterating and improving the feature to reach a level of maturity before we call it stable. Please contact us via [email](mailto:david.stauffer@weave.works) or [slack](https://join.slack.com/t/weave-community/shared_invite/zt-1nrm7dc6b-QbCec62CJ7qj_OUOtuJbrw) if you want to get access to the preview.



### Minor fixes
#### OIDC
- Allows customising the requested scopes via config.oidc.customScopes: "email,groups,something_else"
- Token refreshing is now supported


### Dependency versions

- weave-gitops v0.15.0
- cluster-controller v1.4.1
- cluster-bootstrap-controller v0.3.0
- (optional) pipeline-controller v0.9.0
- (optional) policy-agent v2.2.0

### Breaking changes

No breaking changes

## v0.14.1
2023-01-05
### Highlights

#### Secrets management
- We are introducing new functionality into Weave GitOps Enterprise to help observe and manage secrets through external secrets operator (ESO). The new secrets UI will enable customers using ESO to observe and manage external secrets, as well as help them troubleshoot issues during their secrets creation and sync operations. In this release, we are including the ability to list all ExternalSecrets custom resources across multi-cluster environments. Users also will have the ability to navigate to each ExternalSecret and know the details of the secret, its sync status, and the last time this secret has been updated, as well as the latest events associated with the secret.

#### Pipelines
- Retry promotion on failure. Now if a promotion fails there is an automatic retry functionalty, you can configure the threshold and delay via the CLI.
- Promotion webhook rate limiting. We enable now the configuration of the rate limit for the promotion webhooks.

### Minor fixes
#### Workspaces
** [UI] "Tenant" ** is renamed to "Workspace" on details page.

** [UI] Use time.RFC3339 ** format for all timestamps of the workspaces tabs.

#### Other
** [UI] Error notification boundary ** does not allow user to navigate away from the page.

** [Gitops run] GitOps Run ** doesn't ask to install dashboard twice

### Dependency versions

- weave-gitops v0.14.1
- cluster-controller v1.4.1
- cluster-bootstrap-controller v0.3.0
- (optional) pipeline-controller v0.9.0
- (optional) policy-agent v2.2.0

### Breaking changes

No breaking changes

## v0.13.0
2022-12-22
### Highlights

#### GitOps Templates Path feature
- GitOps templates now provide the capability to write resources to multiple
	paths in the Git repository. This feature allows complex scenarios, like for
	example creating a self-service for an application that requires an RDS
	database. We’ve provided
	[documentation](../gitops-templates/repo-rendered-paths.md) which has a example.

```yaml
spec:
  resourcetemplates:
    - path: ./clusters/${CLUSTER_NAME}/definition/cluster.yaml
      content:
        - apiVersion: cluster.x-k8s.io/v1alpha4
          kind: Cluster
          metadata:
            name: ${CLUSTER_NAME}
          ...
        - apiVersion: infrastructure.cluster.x-k8s.io/v1alpha4
          kind: AWSCluster
          metadata:
            name: ${CLUSTER_NAME}
          ...
    - path: ./clusters/${CLUSTER_NAME}/workloads/helmreleases.yaml
      content:
        - apiVersion: helm.toolkit.fluxcd.io/v2beta1
          kind: HelmRelease
          metadata:
            name: ${CLUSTER_NAME}-nginx
          ...
        - apiVersion: helm.toolkit.fluxcd.io/v2beta1
          kind: HelmRelease
          metadata:
            name: ${CLUSTER_NAME}-cert-manager
          ...
```

#### Workspace UI
- Weave GitOps now provides a GUI for Workspaces.

#### Enhanced Terraform Table in UI
- Weave GitOps now provides more details on the Terraform inventory GUI page. Adding the type and identifier fields to the inventory table, plus filtering and a 'no data' message.

#### Keyboard shortcuts for "port forwards" on GitOps Run
- Weave GitOps now building and printing a list of set up port forwards.
- Weave GitOps now opening the selected port forward URL on key press. Listening for keypress is performed with the `github.com/mattn/go-tty` package (other options required pressing Enter after a keypress, this catches just a single numeric keypress) and opening URLs with the `github.com/pkg/browser` package.

#### Minor fixes
**[UI] Notifications** Fixed provider page showing a 404.

### Dependency versions

- weave-gitops v0.13.0
- cluster-controller v1.4.1
- cluster-bootstrap-controller v0.3.0
- (optional) pipeline-controller v0.8.0
- (optional) policy-agent v2.2.0

### Breaking changes

No breaking changes

## v0.12.0
2022-12-09

### Highlights

**We highly recommend users of v0.11.0 upgrade to this version as it includes fixes for a number of UI issues.**

#### GitOps Templates

- Support to specify Helm charts inside the CRD, instead of annotations. We’ve
	provided [documentation](../gitops-templates/profiles.md) which has a example.

```yaml
spec:
  charts:
    items:
      - chart: cert-manager
        version: v1.5.3
        editable: false
        required: true
        values:
          installCRDs: ${CERT_MANAGER_INSTALL_CRDS}
        targetNamespace: cert-manager
        layer: layer-1
        template:
          content:
            metadata:
              labels:
                app.kubernetes.io/name: cert-manager
            spec:
              retries: ${CERT_MANAGER_RETRY_COUNT}
```

- Ability to edit all fields now, including name/namespace

#### Authentication with OIDC support
Supporting custom OIDC groups claims for azure/okta integration
Support for OIDC custom username and group claims:

```yaml
config
  oidc:
    claimUsername: ""
    claimGroups: ""
```

#### Policy commit-time agent
- Support Azure DevOps and auto-remediation in commit-time enforcement.

#### Admin User- simpler RBAC
- Weave GitOps default admin user can now “read” all objects. Why is this important? As users are trying out Weave GitOps they will most likely try it out with some of their favorite Cloud Native tools such as Crossplane, Tekton, Istio, etc. This enables them to see all of those resources and explore the full power of Weave GitOps. We still do not recommend this user for “production-use” cases, and customers should always be pushed towards implementing OIDC with scoped roles.

#### Pipelines - adding Pipelines through Templates
- From the Pipelines view you can add new Pipelines in a way which leverages GitOpsTemplates, additionally - to help users configure these, we’ve provided [documentation](../pipelines/pipelines-templates.md) which has some samples.

#### Support for multiple Flux instances on a single cluster
- Support for running multiple flux instances in different namespaces on a single cluster for resource isolation.

#### Minor fixes

**Terraform CRD Error**
Users of the Terraform Controller will be pleased to know we’ve addressed the issue where an error would be displayed if it had not been installed on all connected clusters.

**Management cluster renaming**
If the name of the cluster where Weave GitOps Enterprise is installed, was changed from the default of management through the config.cluster.name parameter, certain workflows could fail such as fetching profiles, this has now been resolved.

### Dependency versions​
weave-gitops v0.12.0
cluster-controller v1.4.1
cluster-bootstrap-controller v0.3.0
(optional) pipeline-controller v0.0.11
(optional) policy-agent 2.1.1

### Known issues
- [UI] Notifications provider page shows a 404.

## v0.11.0
2022-11-25

### Highlights

#### GitOpsTemplates
- We are working towards unifying CAPI and GitOps Templates under a single umbrella. For those already using CAPITemplates, we will ensure a smooth transition is possible by making use of a conversion hooks. There are some breaking changes for GitOpsTemplates as part of this transitionary period, so be sure to check the guidance under [Breaking Changes](#breaking-changes).
- We now retain the ordering of parameters in the template instead of sorting them alphabetically. Providing to the author control in what sequence the parameters are rendered in the form and thus present a more logically grouped set of parameters to the end consumer.
- You can control what
	[delimiters](../gitops-templates/supported-langs.md#custom-delimiters) you
	want to use in a template. This provides flexibility for if you want to use
	the syntax for dynamic functions like the [helper functions](../gitops-templates/supported-langs.md#supported-functions-1) we support.

#### Pipelines
- This [feature](../pipelines/index.md) is now enabled by default when you install the Weave GitOps Enterprise Helm Chart. You can toggle this with the `enablePipelines` flag.
- GitOpsTemplates are a highly flexible way to create new resources - including Pipelines. We now provide a shortcut on the Pipelines table view to navigate you to Templates with the `weave.works/template-type=pipeline` label.

#### Telemetry
This release incorporates anonymous aggregate user behavior analytics to help us continuously improve the product. As an Enterprise customer, this is enabled by default. You can learn more about this [here](/feedback-and-telemetry#anonymous-aggregate-user-behavior-analytics).

### Dependency versions
- weave-gitops v0.11.0
- cluster-controller v1.4.1
- cluster-bootstrap-controller v0.3.0
- (optional) pipeline-controller v0.0.11
- (optional) policy-agent 2.1.1

### Breaking changes

#### GitOpsTemplates and CAPITemplates
We are making these changes to provide a unified and intuitive self-service experience within Weave GitOps Enterprise, removing misleading and potentially confusing terminology born from when only Clusters were backed by Templates.

**New API Group for the GitOpsTemplate CRD**
- old: `clustertemplates.weave.works`
- new: `templates.weave.works`

After upgrading Weave GitOps Enterprise which includes the updated CRD:
1. Update all your GitOpsTemplates in Git changing all occurrences of `apiVersion: clustertemplates.weave.works/v1alpha1` to `apiVersion: templates.weave.works/v1alpha1`.
2. Commit, push and reconcile. They should now be viewable in the Templates view again.
3. Clean up the old CRD. As it stands:
   - `kubectl get gitopstemplate -A` will be empty as it is pointing to the old `clustertemplates.weave.works` CRD.
   - `kubectl get gitopstemplate.templates.weave.works -A` will work
To fix the former of the commands, remove the old CRD (helm does not do this automatically for safety reasons):
   - `kubectl delete crd gitopstemplates.clustertemplates.weave.works`
   - You may have to wait up to 5 minutes for your local kubectl CRD cache to invalidate, then `kubectl get gitopstemplate -A` should be working as usual

**Template Profiles / Applications / Credentials sections are hidden by default**

For both `CAPITemplates` and `GitopsTemplates` the default visibility for all sections in a template has been set to `"false"`. To re-enable profiles or applications on a template you can tweak the annotations

```yaml
annotations:
  templates.weave.works/profiles-enabled: "true" # enable profiles
  templates.weave.works/kustomizations-enabled: "true" # enable applications
  templates.weave.works/credentials-enabled: "true" # enable CAPI credentials
```

**The default values for a profile are not fetched and included in a pull-request**

Prior to this release WGE would fetch the default values.yaml for every profile installed and include them in the `HelmReleases` in the Pull Request when rendering out the profiles of a template.

This was an expensive operation and occasionally led to timeouts.

The new behaviour is to omit the values and fall back to the defaults included in the helm-chart. This sacrifices some UX (being able to see all the defaults in the PR and tweak them) to improve performance. **There should not be any final behaviour changes to the installed charts**.

You can still view and tweak the `values.yaml` when selecting profiles to include on the "Create resource (cluster)" page. If changes are made here the updated values.yaml will be included.

## v0.10.2
2022-11-15

### Highlights
- Retain template parameter ordering.
- Allow configuration of the delimiters in templates.
- Add create a pipeline button.
- add missing support for policy version v2beta2 to tenancy cmd.

### Dependency versions
- weave-gitops v0.10.2
- cluster-controller v1.4.1
- cluster-bootstrap-controller v0.3.0
- (optional) policy-agent 2.1.1

## v0.10.1
2022-11-10

### Highlights

- Create non-cluster resources / Add Edit option to resources with create-request annotation
- bump pipeline-controller
- Parse annotations from template
- Add cost estimate message if available
- Adds support for showing policy modes and policy configs in the UI

- Show suspended status on pipelines detail
- YAML view for Pipelines
- Align and link logo

- Actually remove the watcher from the helm-watcher-cache
- UI 1817 disable create target name space if name space is flux system

- Adding edit capi cluster resource acceptance test
- Add preview acceptance test

### Dependency versions

- weave-gitops v0.10.1
- cluster-controller v1.4.1
- cluster-bootstrap-controller v0.3.0
- (optional) policy-agent 2.0.0


## v0.9.6
2022-10-17

### Highlights
- When adding applications, you can now preview the changes(PR) before creating a pull request
- You can now see included Cluster Profiles when previewing your Create Cluster PR
- Notifications are now available in the Notifications Page
- You can now automatically create namespace when adding applications

### Dependency versions

- weave-gitops v0.9.6
- cluster-controller v1.3.2
- cluster-bootstrap-controller v0.3.0
- (optional) policy-agent 1.2.1

## v0.9.5
2022-09-22

### Highlights
- **Tenancy**
  - `gitops create tenant` now supports `--prune` to remove old resources from the cluster if you're not using `--export` with GitOps.
  - `deploymentRBAC` section in `tenancy.yaml` allows you to specify the permissions given to the flux `Kustomizations` that will apply the resources from git to your tenants' namespaces in the cluster.
  - Support for `OCIRepository` sources when restricting/allowing the sources that can be applied into tenants' namespaces.
- **Templates**
  - Templates now support helm functions for simple transformations of values: `{{ .params.CLUSTER_NAME | upper }}`
  - Templates has moved to its own page in the UI, this is the first step in moving towards embracing them as a more generic feature, not just for cluster creation.
  - If a version is not specified in a **template profile annotation** it can be selected by the user.
  - A `namespace` can be specified in the **template profile annotation** that will be provided as the `HelmRelease`'s `targetNamespace` by default.
- **Bootstrapping**
  - A ClusterBootstrapConfig can now optionally be triggered when `phase="Provisioned"`, rather than `ControlPlaneReady=True` status.

### Dependency versions

- weave-gitops v0.9.5
- cluster-controller v1.3.2
- cluster-bootstrap-controller v0.3.0
- (optional) policy-agent 1.1.0

### Known issues

- [UI] Notifications page shows a 404 instead of the notification-controller's configuration.

### ⚠️ Breaking changes from v0.9.4

If using the policy-agent included in the weave-gitops-enterprise helm chart, the configuration should now be placed under the `config` key.

**old**
```yaml
policy-agent:
  enabled: true
  accountId: "my-account"
  clusterId: "my-cluster"
```

**new**
```yaml
policy-agent:
  enabled: true
  config:
    accountId: "my-account"
    clusterId: "my-cluster"
```
