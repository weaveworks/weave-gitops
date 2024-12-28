---
title: PolicyConfig

---

# PolicyConfig ~ENTERPRISE~

## Goal

Users sometimes need to enforce the same policy(s) with different configurations (parameters) for different targets (workspaces, namespaces, applications, or resources).
The `PolicyConfig` CRD allows us to do that without duplicating policies by overriding policy parameters of multiple policies for a specific target.

## Schema

The PolicyConfig CRD consists of two sections 1) `match` used to specify the target of this PolicyConfig and 2) `config` used to specify the policy parameters that will override the orginal policy parameters.

??? example "Expand to see a PolicyConfig example"

    ```yaml
    apiVersion: pac.weave.works/v2beta2
    kind: PolicyConfig      # policy config resource kind
    metadata:
    name: my-config       # policy config name
    spec:
    match:                # matches (targets of the policy config)
        workspaces:         # add one or more name workspaces
        - team-a
        - team-b
    config:               # config for policies [one or more]
        weave.policies.containers-minimum-replica-count:
        parameters:
            replica_count: 3
    ```

Each PolicyConfig CR can target either workspaces, namespaces, applications or resources. Targeting the same target explicitly in multiple PolicyConfigs is not allowed, ie: you can't use the same namespace in several PolicyConfigs which target namespaces.

To target workspaces:

```yaml
match:
    workspaces:
    - team-a
    - team-b
```

To target namespaces:

```yaml
match:
    namespaces:
    - dev
    - prod
```

To target applications:

```yaml
match:
    apps:            # add one or more apps [HelmRelease, Kustomization]
    - kind: HelmRelease
    name: my-app            # app name
    namespace: flux-system  # app namespace [if empty will match in any namespace]
```

To target resources:

```yaml
match:
    resources:       # add one or more resources [Deployment, ReplicaSet, ..]
    - kind: Deployment
    name: my-deployment     # resource name
    namespace: default      # resource namespace [if empty will match in any namespace]
```

Each PolicyConfig can override the parameters of one or more policies:

```yaml
config:               # config for policies [one or more]
    weave.policies.containers-minimum-replica-count: # the id of the policy
    parameters:
        replica_count: 3
        owner: owner-4
    weave.policies.containers-running-in-privileged-mode:
    parameters:
        privilege: true
```

## Overlapping Targets

While it's not possible to create PolicyConfigs that explicitly target the same targets, it can happen implicitly ex: by targeting a namespace in a PolicyConfig and targeting an application that exists in this namespace in another.
Whenever targets overlap, the narrower the scope of the PolicyConfig, the more precedence it has. Accordingly in the previous example, the configuration of the PolicyConfig targeting the application will have precedence over the PolicyConfig targeting the namespace.

Those are the possible targets from lowest to highest precedence:

- PolicyConfig which targets a workspace.
- PolicyConfig which targets a namespace.
- PolicyConfig which targets an application in all namespaces.
- PolicyConfig which targets an application in a certain namespace.
- PolicyConfig which targets a kubernetes resource in all namespaces.
- PolicyConfig which targets a kubernetes resource in a specific namespace.

**Note**:

- All configs are applied from low priority to high priority while taking into consideration the common parameters between configs.
- Each config only affects the parameters defined in it.

### Example

We have a Kustomization application `app-a` and deployment `deployment-1` which is part of this application.

??? example "Expand to see manifests"

    ```yaml
    apiVersion: pac.weave.works/v2beta2
    kind: PolicyConfig
    metadata:
    name: my-config-1
    spec:
    match:
        namespaces:
        - flux-system
    config:
        weave.policies.containers-minimum-replica-count:
        parameters:
            replica_count: 2
            owner: owner-1
    ---
    apiVersion: pac.weave.works/v2beta2
    kind: PolicyConfig
    metadata:
    name: my-config-2
    spec:
    match:
        apps:
        - kind: Kustomization
        name: app-a
    config:
        weave.policies.containers-minimum-replica-count:
        parameters:
            replica_count: 3
    ---
    apiVersion: pac.weave.works/v2beta2
    kind: PolicyConfig
    metadata:
    name: my-config-3
    spec:
    match:
        apps:
        - kind: Kustomization
        name: app-a
        namespace: flux-system
    config:
        weave.policies.containers-minimum-replica-count:
        parameters:
            replica_count: 4
    ---
    apiVersion: pac.weave.works/v2beta2
    kind: PolicyConfig
    metadata:
    name: my-config-4
    spec:
    match:
        resources:
        - kind: Deployment
        name: deployment-1
    config:
        weave.policies.containers-minimum-replica-count:
        parameters:
            replica_count: 5
            owner: owner-4
    ---

    apiVersion: pac.weave.works/v2beta2
    kind: PolicyConfig
    metadata:
    name: my-config-5
    spec:
    match:
        resources:
        - kind: Deployment
        name: deployment-1
        namespace: flux-system
    config:
        weave.policies.containers-minimum-replica-count:
        parameters:
            replica_count: 6
    ```

**In the above example when you apply the 5 configurations**...

- `app-a` will be affected by `my-config-5`. It will be applied on the policies defined in it, which will affect deployment `deployment-1` in namespace `flux-system` as it matches the kind, name and namespace.

!!! note
    Deploying `deployment-1` in another namespace other than `flux-system` won't be affected by this configuration

Final config values will be as follows:

```yaml
    config:
    weave.policies.containers-minimum-replica-count:
        parameters:
        replica_count: 6 # from my-config-5
        owner: owner-4   # from my-config-4
```

- _Deployment `deployment-1` in namespace `flux-system`, `replica_count` must be `>= 6`_
- _Also it will be affected by `my-config-4` for `owner` configuration parameter `owner: owner-4`_

**In the above example when you apply `my-config-1`, `my-config-2`, `my-config-3` and `my-config-4`**

- `my-config-4` will be applied on the policies defined in it which will affect deployment `deployment-1` in all namespaces as it matches the kind and name only.

Final config values will be as follows:

```yaml
    config:
    weave.policies.containers-minimum-replica-count:
        parameters:
        replica_count: 5  # from my-config-4
        owner: owner-4    # from my-config-4
```

- _Deployment `deployment-1` in all namespaces `replica_count` must be `>= 5`_
- _Also it will be affected by `my-config-4` for `owner` configuration parameter `owner: owner-4`_

**In the previous example when you apply `my-config-1`, `my-config-2` and `my-config-3`**

- `my-config-3` will be applied on the policies defined in it which will affect application `app-a` and all the resources in it in namespace `flux-system` as it matches the kind, name and namespace.

!!! note
    Deploying `app-a` in another namespace other than `flux-system` won't be affected by this configuration

Final config values will be the follows:

```yaml
    config:
    weave.policies.containers-minimum-replica-count:
        parameters:
        replica_count: 4    # from my-config-3
        owner: owner-1      # from my-config-1
```

- _Application `app-a` and all the resources in it in namespaces `flux-system`, `replica_count` must be `>= 4`_
- _Also it will be affected by `my-config-1` for `owner` configuration parameter `owner: owner-1`_

**In the above example when you apply `my-config-1` and `my-config-2`**

- `my-config-2` will be applied on the policies defined in it which will affect application `app-a` and all the resources in it in all namespaces as it matches the kind and name only.

Final config values will be as follows:

```yaml
    config:
    weave.policies.containers-minimum-replica-count:
        parameters:
        replica_count: 3   # from my-config-2
        owner: owner-1     # from my-config-1
```

- _Application `app-a` and all the resources in all namespaces, `replica_count` must be `>= 3`_
- _Also it will be affected by `my-config-1` for `owner` configuration parameter `owner: owner-1`_

**In the above example when you apply `my-config-1`**

- `my-config-1` will be applied on the policies defined in it. which will affect the namespace `flux-system` with all applications and resources in it as it matches by namespace only.

Final config values will be as follows:

```yaml
    config:
    weave.policies.containers-minimum-replica-count:
        parameters:
        replica_count: 2  # from my-config-1
        owner: owner-1    # from my-config-1
```

- _Any application or resource in namespace `flux-system`, `replica_count` must be `>= 2`_
- _Also it will be affected by `my-config-1` for `owner` configuration parameter `owner: owner-1`_
