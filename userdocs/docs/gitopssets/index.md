---
title: Introduction

---



# GitOpsSets ~ENTERPRISE~

!!! warning

    **This feature is in alpha and certain aspects will change**

    We're very excited for people to use this feature.
    However, please note that some changes will be made to the API and behavior,
    particularly to enhance security by implementing impersonation for more
    fine-grained control over how the generated resources are applied.

## Introduction

GitOpsSets enable Platform Operators to have a single definition for an application for multiple environments and a fleet of clusters. A single definition can be used to generate the environment and cluster-specific configuration.

As an example, we can take an application that needs to be deployed to various environments (Dev, Test, Prod) built by a fleet of clusters. Each of those environments + clusters requires a specialized configuration powering the same Application. With GitOpsSets and the generators you just declare the template you want to use, the selector that will match the cluster of the inventory, and where to get the special configuration. 

GitOpsSets will create out of the single resource all the objects and Flux primitives that are required to successfully deploy this application. An operation that required the editing of hundreds of files can now be done with a single command. 

**The initial generators that are coming with the preview release are:**

- [List Generator](templating-from-generators.md#list-generator): The simplest generator. Provide a list of Key/Value pairs that you want to feed the template with.
- [Git Generator](templating-from-generators.md#gitrepository-generator): Enables you to extract a set of files (environment-specific configurations) from a Flux GitRepository and make their contents available to the templates. This lets you have config in *app-dev.json*, *app-staging.json*, and *app-production.json*, for example.
- [Matrix Generator](templating-from-generators.md#matrix-generator): Combine slices of generators into the desired compounded input.
- [Pull request Generator](templating-from-generators.md#pullrequests-generator): Automatically discover open pull requests within a repository to generate a new deployment.
- [API Client Generator](templating-from-generators.md#apiclient-generator): Poll an HTTP endpoint and parse the result as the generated values.
- [OCI Repository](templating-from-generators.md#ocirepository-generator)
- [Cluster](templating-from-generators.md#cluster-generator)
- [ImagePolicy](templating-from-generators.md#imagepolicy-generator)
- [Config](templating-from-generators.md#config-generator)

## Use Cases

- Single application definition for different environments (EU-West, North America, Germany)
- Deployment of a single definition across fleet of clusters matching any cluster based on a label (Production) 
- Separation of concerns between teams (teams managing different artifacts flowing into a single definition via generators)
