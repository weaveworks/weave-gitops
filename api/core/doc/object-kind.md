Kind is used as the primary identifier rather than GroupVersionKind

Supported kinds are:

 - some of the base Kubernetes groups including core (e.g. `ConfigMap`) and apps (e.g. `Deployment`)
 - All the of the Flux Custom Resource kinds e.g. (`GitRepository`, `HelmRelease`, `Kustomization`, `ImageAutomation`)
 - If using enterprise, `GitOpsSet` and `AutomatedClusterDiscovery` are also available
