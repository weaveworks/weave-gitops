---
title: Multi Tenancy

---

# Multi Tenancy ~ENTERPRISE~

Multi tenancy provides users with the ability to define boundaries to multiple engineering teams working on a single cluster. Through a simple interface it adds permissions to the necessary Kubernetes resources to make it easy for customers to manage their multiple tenants.

WGE multi tenancy expands on the multi tenancy feature provided by `flux`. In addition to creating the necessary Kubernetes tenancy resources that `flux` adds, multi tenancy in WGE also adds the following:
- Defining tenancy using a single yaml file that serves as a source of truth for the organization
- Makes use of WGE policy features to enforce non Kubernetes native permissions

## Prerequisites

- [`gitops` command line tool](../references/cli-reference/gitops.md)
- [Tenancy File](#tenancy-file) (optional)
- [Policies](../policy/index.md) (optional)

## How it works

`gitops` command line tool is responsible for creating the multi tenancy resources. The tool is distributed as part of WGE offering. It reads the definitions of a yaml file and can either apply the necessary changes directly to the cluster or output it to stdout so it can be saved into a file and pushed to a repo to be reconciled by `flux`.

To make use of the policy features, [policy agent](../policy/index.md) needs to be installed in the necessary cluster(s). 

### Tenancy file

Below is an example of a tenancy file:

??? example "Expand to view"

    ```yaml title="tenancy.yaml"
    ---
    tenants:
    - name: first-tenant
        namespaces:
        - first-ns
    - name: second-tenant
        namespaces:
        - second-test-ns
        - second-dev-ns
        allowedRepositories:
        - kind: GitRepository
        url: https://github.com/testorg/testrepo
        - kind: GitRepository
        url: https://github.com/testorg/testinfo
        - kind: Bucket
        url: minio.example.com
        - kind: HelmRepository
        url: https://testorg.github.io/testrepo
        allowedClusters:
        - kubeConfig: cluster-1-kubeconfig
        - kubeConfig: cluster-2-kubeconfig
        teamRBAC:
        groupNames:
        - foo-group
        - bar-group
        rules:
            - apiGroups:
                - ''
            resources:
                - 'namespaces'
                - 'pods'
            verbs:
                - 'list'
                - 'get'
        deploymentRBAC:
        bindRoles:
            - name: foo-role
            kind: Role
        rules:
            - apiGroups:
                - ''
            resources:
                - 'namespaces'
                - 'pods'
            verbs:
                - 'list'
                - 'get'
    serviceAccount:
    name: "reconcilerServiceAccount"
    ```


The file above defines two tenants: `first-tenant` and `second-tenant` as follows:

- `namespaces`: describes which namespaces should be part of the tenant. Meaning that users who are part of the tenant would have access on those namespaces.
- `allowedRepositories`: limits the `flux` repositories sources that can be used in the tenant's namespaces. This is done through policies and thus requires `policy-agent` to be deployed on the cluster which will stop these sources from being deployed if they aren't allowed as part of the tenant. IT consists of:
  - `kind`: the `flux` source kind. Can be: `GitRepository`, `Bucket` and `HelmRepository`.
  - `url`: the URL for that source.
- `allowedClusters`: limits which secrets containing cluster configuraton can be used. It stops WGE `GitopsCluster` and flux `Kustomization` from being deployed if they point to a secret not in the list, essentially giving control on which cluster can be added to a multi-cluster setup. Requires `policy-agent`.
  - `kubeConfig`: name of the secret that can be used for this tenant.
- `teamRBAC`: Generate Roles and Rolebindings for a list of `groupNames`. This allows you to easily give an OIDC group access to a tenant's resources. When the Weave Gitops Enterprise UI is configured with your OIDC provider, tenants can log in and view the status of the resources they have been granted access to.
- `deploymentRBAC`: generate Roles and Rolebindings for a service account. Can additionally bind to an existing Roles/ClusterRoles. Would use the global service account if specified in the tenants file, otherwise it will use the created service account which takes the tenant name. If not specified a Rolebinding would be created that binds to `cluster-admin` ClusterRole.

Global options:

- `serviceAccount`: Override the name of the generated `ServiceAccount` for all tenants. This allows you to easily use the flux controllers' [`--default-service-account`](https://github.com/fluxcd/flux2-multi-tenancy#enforce-tenant-isolation) feature. Tenants do not need to make sure they correctly specify the `serviceAccount` when using `Kustomization` or `HelmRelease` resources. The kustomization-controller and helm-controller will instead look for the `default-service-account` in the namespace being reconciled to and use that. Just configure `serviceAccount.name` and `--default-service-account` to the same value.

### Gitops create tenants command

The command creates the necessary resources to apply multi tenancy on the user's cluster. To use the command to apply the resources directly the user needs to have the necessary configuration to connect to the desired cluster.
The command considers the tenancy file as a source of truth and will change the cluster state to match what is currently described in the file.

For more control on a specific tenant a tenancy file should be used, the command allows the creation of the base resources that defines a tenancy through the arguments:

```bash
gitops create tenants --name test-tenant --namespace test-ns1 --namespace test-ns2
```

??? example "Expand to view command output"

    ```bash
    namespace/test-ns1 created
    test-ns1/serviceaccount/test-tenant created
    test-ns1/rolebinding.rbac.authorization.k8s.io/test-tenant-service-account-cluster-admin created
    namespace/test-ns2 created
    test-ns2/serviceaccount/test-tenant created
    test-ns2/rolebinding.rbac.authorization.k8s.io/test-tenant-service-account-cluster-admin created
    policy.pac.weave.works/weave.policies.tenancy.test-tenant-allowed-application-deploy created
    ```


The above will create the namespaces and permissions through a `ServiceAccount` with the same name as the tenant, `test-tenant` in the case of the above example, in each required namespace.
The same can be done through a file as follows:

```yaml
tenants:
  - name: test-tenant
    namespaces:
    - test-ns1
    - test-ns2
```

```bash
gitops create tenants --from-file tenants.yaml
```

??? example "Expand to view command output"

    ```bash
    namespace/test-ns1 created
    test-ns1/serviceaccount/test-tenant created
    test-ns1/rolebinding.rbac.authorization.k8s.io/test-tenant-service-account-cluster-admin created
    namespace/test-ns2 created
    test-ns2/serviceaccount/test-tenant created
    test-ns2/rolebinding.rbac.authorization.k8s.io/test-tenant-service-account-cluster-admin created
    policy.pac.weave.works/weave.policies.tenancy.test-tenant-allowed-application-deploy created
    ```


To check the resources that would be deployed first use the `export` flag:

```bash
gitops create tenants --from-file tenants.yaml --export
```

??? example "Expand to view command output"

    ```bash
    apiVersion: v1
    kind: Namespace
    metadata:
    creationTimestamp: null
    labels:
        toolkit.fluxcd.io/tenant: test-tenant
    name: test-ns1
    spec: {}
    status: {}
    ---
    apiVersion: v1
    kind: ServiceAccount
    metadata:
    creationTimestamp: null
    labels:
        toolkit.fluxcd.io/tenant: test-tenant
    name: test-tenant
    namespace: test-ns1
    ---
    apiVersion: rbac.authorization.k8s.io/v1
    kind: RoleBinding
    metadata:
    creationTimestamp: null
    labels:
        toolkit.fluxcd.io/tenant: test-tenant
    name: test-tenant-service-account-cluster-admin
    namespace: test-ns1
    roleRef:
    apiGroup: rbac.authorization.k8s.io
    kind: ClusterRole
    name: cluster-admin
    subjects:
    - kind: ServiceAccount
    name: test-tenant
    namespace: test-ns1
    ---
    apiVersion: v1
    kind: Namespace
    metadata:
    creationTimestamp: null
    labels:
        toolkit.fluxcd.io/tenant: test-tenant
    name: test-ns2
    spec: {}
    status: {}
    ---
    apiVersion: v1
    kind: ServiceAccount
    metadata:
    creationTimestamp: null
    labels:
        toolkit.fluxcd.io/tenant: test-tenant
    name: test-tenant
    namespace: test-ns2
    ---
    apiVersion: rbac.authorization.k8s.io/v1
    kind: RoleBinding
    metadata:
    creationTimestamp: null
    labels:
        toolkit.fluxcd.io/tenant: test-tenant
    name: test-tenant-service-account-cluster-admin
    namespace: test-ns2
    roleRef:
    apiGroup: rbac.authorization.k8s.io
    kind: ClusterRole
    name: cluster-admin
    subjects:
    - kind: ServiceAccount
    name: test-tenant
    namespace: test-ns2
    ---
    apiVersion: pac.weave.works/v2beta2
    kind: Policy
    metadata:
    creationTimestamp: null
    labels:
        toolkit.fluxcd.io/tenant: test-tenant
    name: weave.policies.tenancy.test-tenant-allowed-application-deploy
    spec:
    category: weave.categories.tenancy
    code: |
        package weave.tenancy.allowed_application_deploy

        controller_input := input.review.object
        violation[result] {
            namespaces := input.parameters.namespaces
            targetNamespace := controller_input.spec.targetNamespace
            not contains_array(targetNamespace, namespaces)
            result = {
            "issue detected": true,
            "msg": sprintf("using target namespace %v is not allowed", [targetNamespace]),
            }
        }
        violation[result] {
            serviceAccountName := controller_input.spec.serviceAccountName
            serviceAccountName != input.parameters.service_account_name
            result = {
            "issue detected": true,
            "msg": sprintf("using service account name %v is not allowed", [serviceAccountName]),
            }
        }
        contains_array(item, items) {
            items[_] = item
        }
    description: Determines which helm release and kustomization can be used in a tenant
    how_to_solve: ""
    id: weave.policies.tenancy.test-tenant-allowed-application-deploy
    name: test-tenant allowed application deploy
    parameters:
    - name: namespaces
        required: false
        type: array
        value:
        - test-ns1
        - test-ns2
    - name: service_account_name
        required: false
        type: string
        value: test-tenant
    provider: kubernetes
    severity: high
    standards: []
    tags:
    - tenancy
    targets:
        kinds:
        - HelmRelease
        - Kustomization
        labels: []
        namespaces:
        - test-ns1
        - test-ns2
    status: {}
    ---
    ```


Applying the resources through the command line is not usually recommended. For WGE the recommended way is to commit the result of the `create tenants` command to source control and let `flux` handle deployment. To achieve that you can save the result of the `export` to a file:

```bash
gitops create tenants --from-file tenants.yaml --export > clusters/management/tenants.yaml 
```
