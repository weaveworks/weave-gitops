# Current Endpoints for Building the Reconciliation Graph

This shows the breakdown of building a reconcilation graph via requests.  The reconcilation_object is called and then we climb down the tree of `deployments` by calling the child_object endpoint for each `Deployment`.

## Request for reconciled_objects
```
{
    "automationName": "josh-pod-info",
    "automationNamespace": "wego-system",
    "kinds": [
        {
            "group": "",
            "kind": "Namespace",
            "version": "v1"
        },
        {
            "group": "",
            "kind": "Service",
            "version": "v1"
        },
        {
            "group": "apps",
            "kind": "Deployment",
            "version": "v1"
        },
        {
            "group": "autoscaling",
            "kind": "HorizontalPodAutoscaler",
            "version": "v2beta1"
        },
        {
            "group": "kustomize.toolkit.fluxcd.io",
            "kind": "Kustomization",
            "version": "v1beta2"
        },
        {
            "group": "source.toolkit.fluxcd.io",
            "kind": "GitRepository",
            "version": "v1beta1"
        },
        {
            "group": "wego.weave.works",
            "kind": "Application",
            "version": "v1alpha1"
        }
    ]
}
```

## Response for reconciled_objects
```
{
    "objects": [
        {
            "groupVersionKind": {
                "group": "",
                "kind": "Service",
                "version": "v1"
            },
            "name": "backend",
            "namespace": "test",
            "uid": "416ed5e7-cbff-4bde-b5a7-79ebe158ece5",
            "status": "Current"
        },
        {
            "groupVersionKind": {
                "group": "apps",
                "kind": "Deployment",
                "version": "v1"
            },
            "name": "backend",
            "namespace": "test",
            "uid": "9c560818-a4db-407d-8914-ff8f58763e5e",
            "status": "Current"
        },
        {
            "groupVersionKind": {
                "group": "apps",
                "kind": "Deployment",
                "version": "v1"
            },
            "name": "frontend",
            "namespace": "test",
            "uid": "a4451f79-c3a4-4da3-9c50-ce6488cd37d8",
            "status": "Current"
        },
        {
            "groupVersionKind": {
                "group": "autoscaling",
                "kind": "HorizontalPodAutoscaler",
                "version": "v2beta1"
            },
            "name": "backend",
            "namespace": "test",
            "uid": "ecd9742e-f3a0-4fb5-87e9-e888cd701bf8",
            "status": "Current"
        },
        {
            "groupVersionKind": {
                "group": "kustomize.toolkit.fluxcd.io",
                "kind": "Kustomization",
                "version": "v1beta2"
            },
            "name": "podinfo-deploy",
            "namespace": "wego-system",
            "uid": "626a1958-99ed-4c57-a2b9-95c357fde629",
            "status": "Current"
        },
        {
            "groupVersionKind": {
                "group": "source.toolkit.fluxcd.io",
                "kind": "GitRepository",
                "version": "v1beta1"
            },
            "name": "podinfo-deploy",
            "namespace": "wego-system",
            "uid": "181aeb04-2e4c-4a5c-8f11-9c8b27a20db7",
            "status": "Current"
        }
    ]
}
```

## Child Objects call

### Request 1 for child_objects
```
{
    "parentUid": "9c560818-a4db-407d-8914-ff8f58763e5e",
    "groupVersionKind": {
        "group": "apps",
        "version": "v1",
        "kind": "ReplicaSet",
        "children": [
            {
                "version": "v1",
                "kind": "Pod"
            }
        ]
    }
}
```

### Response 1 for child_objects
```
{
    "objects": [
        {
            "groupVersionKind": {
                "group": "apps",
                "kind": "ReplicaSet",
                "version": "v1"
            },
            "name": "backend-6b944d8b",
            "namespace": "test",
            "uid": "131966ae-2699-4747-b58b-4363795b7fd3",
            "status": "Current"
        }
    ]
}
```

### Request 2 for child_objects
**Note:** this is walking down the tree of the parent object above

```
{"parentUid":"131966ae-2699-4747-b58b-4363795b7fd3","groupVersionKind":{"version":"v1","kind":"Pod"}}
```

### Response 2 for child_objects
```
{
    "objects": [
        {
            "groupVersionKind": {
                "group": "",
                "kind": "Pod",
                "version": "v1"
            },
            "name": "backend-6b944d8b-tzxnl",
            "namespace": "test",
            "uid": "bc3913d9-6ce2-4c96-8171-c24b2937a054",
            "status": "Current"
        }
    ]
}
```

### Request 3 for child_objects
```
{
    "parentUid": "a4451f79-c3a4-4da3-9c50-ce6488cd37d8",
    "groupVersionKind": {
        "group": "apps",
        "version": "v1",
        "kind": "ReplicaSet",
        "children": [
            {
                "version": "v1",
                "kind": "Pod"
            }
        ]
    }
}
```

### Response 3 for child_objects
```
{
    "objects": [
        {
            "groupVersionKind": {
                "group": "apps",
                "kind": "ReplicaSet",
                "version": "v1"
            },
            "name": "frontend-5c64b4bdf5",
            "namespace": "test",
            "uid": "47bbeae1-6003-46e4-b093-580ece24c861",
            "status": "Current"
        }
    ]
}
```

### Request 4 for child_objects
```
{"parentUid":"47bbeae1-6003-46e4-b093-580ece24c861","groupVersionKind":{"version":"v1","kind":"Pod"}}
```

### Response 4 for child_objects
```
{
    "objects": [
        {
            "groupVersionKind": {
                "group": "",
                "kind": "Pod",
                "version": "v1"
            },
            "name": "frontend-5c64b4bdf5-4pskz",
            "namespace": "test",
            "uid": "093f84a0-5e43-4c82-88ee-6e723aff86db",
            "status": "Current"
        }
    ]
}
```
