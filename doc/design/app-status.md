### Status of my Application

As a user, I expect to see the status of my application running across multiple clusters. When I go to my application list page I should be able to quickly see the "health" of each application in each cluster within a timely manner.

## Problem to be Solved

- Determine if there are unique statuses for helm or kustomize.
- Come up with a way for the CLI to have parity with the UI status request.
- Determine how to get the status that will scale with larger clusters and multiple clusters.

## Outcome

### Helm/Kustomize unique status

I recommend using kstatus/flux to get all the neccesary statuses. We will need to only know what resources are being used and the statuses will then all be the same. Flux will return the status for automation like it currently does and all other status will be retreived using kstatus.

### CLI/UI parity

The application status is returned for all resources using https://github.com/kubernetes-sigs/cli-utils/tree/v0.26.1/pkg/kstatus. The UI is currently using this to generate the status shown in this graph: 
![Screen Shot 2021-12-09 at 1.32.11 PM.png](https://images.zenhubusercontent.com/60f5a90306cf76325d5accdc/be24ca61-a0dd-47d6-a847-240b38f422af)

*_Only the UI is using this currently._

The CLI currently looks like:
```
NAME                       READY   MESSAGE                                                         REVISION                                        SUSPENDED 
gitrepository/pod-info     True    Fetched revision: b7b8ec639555c425ff8a5d61174d974f733ab6b0      b7b8ec639555c425ff8a5d61174d974f733ab6b0        False    

NAME                          READY   MESSAGE                                 REVISION        SUSPENDED 
kustomization/pod-info        True    Release reconciliation succeeded        2.8.2           False    
```

The UI design will remain the same while the CLI design will need to be updated in one of two ways: 

We can write our own output with details that we get ourselves that could look like:
```
***Application details***:
  Branch:           main
  config_url:       ssh://git@github.com/sympatheticmoose/weave-gitops-multicluster-demo.git
  deployment_type:  kustomize
  Path:             ./
  source_type:      git
  URL:              ssh://git@github.com/sympatheticmoose/podinfo-deploy.git

***Reconciliation details***:
Last successful reconciliation: 2022-01-06 10:43:47 +0000 GMT

NAME                 	READY	MESSAGE                                                        	REVISION                                     	SUSPENDED 
gitrepository/podinfo	True 	Fetched revision: main/1e644628deb645c454f4db167e1041e95e2d0ba3	main/1e644628deb645c454f4db167e1041e95e2d0ba3	False    	

NAME                 	READY	MESSAGE                                                        	REVISION                                     	SUSPENDED 
kustomization/podinfo	True 	Applied revision: main/1e644628deb645c454f4db167e1041e95e2d0ba3	main/1e644628deb645c454f4db167e1041e95e2d0ba3	False  

***Deployment details***:
KIND              NAME                     STATUS
Deployment        backend                  ready
ReplicaSet        backend-6b944d8b         ready
Pod               backend-6b944d8b-tzxnl   ready
Deployment        frontend                 ready
Kustomization     podinfo-deploy           ready
```

The second option is to use output generated from flux using health checks. According to stefan flux uses kstatus as well to get the status for helm and kustomize and it can handle clusters of any size. Some links from stefan regarding how flux handles health checks:
https://fluxcd.io/docs/components/kustomize/kustomization/ //Docs on how to use health checks
https://github.com/fluxcd/pkg/blob/main/ssa/manager_wait.go //Code for polling using kstatus
https://github.com/fluxcd/flux2/blob/cec8b5336cb111b408e7c5546835d239dc804d62/pkg/status/status.go //Code for implementing kstatus cache

Pro/Con for getting the status ourselves:
Pros:
  - Same code to get status for both UI and CLI.
  - Easier to make it look how we want.
Cons:
  - We need to write all the code to get the status from kubernetes including implementing the polling/caches to avoid hitting rate limits.

Pro/Con for getting status using flux health checks.
Pros:
  - Use already existing code for polling/cache.
Cons:
  - Need to write separate code to get the status from flux output to send to UI.
  - Need to figure out how to setup flux health checks without the knowing what resources are being added.
  - We still have to implement another cache to save the last known status in weave-gitops.

Either way the code to get the status needs to be the same for both the cli and the UI. We need to pull out the code in server.go that is getting the status into its own package.
Getting kstatus to work for helm/kustomized will require weave-gitops to provide a list of resources either added to the kustomization for use with flux or implented on our own using kstatus polling/cache. My recommendation is to copy what flux has done but implement it ourselves. This will provide weave-gitops with more freedom of how things look as well as being able to have the UI/CLI use mostly the same code. The UI will need a cache system to retrieve the last known status and while flux has this implemented I am not sure how we would get access to that cache without just doing it all ourselves.

The possible statuses returned from kstatus are:

InProgress: The actual state of the resource has not yet reached the desired state as specified in the resource manifest, i.e. the resource reconcile has not yet completed. Newly created resources will usually start with this status, although some resources like ConfigMaps are Current immediately.
Failed: The process of reconciling the actual state with the desired state has encountered and error or it has made insufficient progress.
Current: The actual state of the resource matches the desired state. The reconcile process is considered complete until there are changes to either the desired or the actual state.
Terminating: The resource is in the process of being deleted.
NotFound: The resource does not exist in the cluster.
Unknown: This is for situations when the library are unable to determine the status of a resource.

I have renamed current to ready to follow how flux returns the kustomize/source status

As for the list views for clusters and applications.
Clusters would look something like:
```
Cluster			Status		  Message
cluster1		Ready			  Fetched revision: main/2189165039567d90119e7680e989c052d6337c8d
cluster2		Failed		  Back-off pulling image "ghcr.io/weaveworks/wego-app:v0.6.0-12-g98ee6910"
```

Applications would look similar:
```
Applications			Status			Last Message
app1			        Ready			  Fetched revision: main/2189165039567d90119e7680e989c052d6337c8d
app2			        Failed			Back-off pulling image "ghcr.io/weaveworks/wego-app:v0.6.0-12-g98ee6910"
```
If any resource in the cluster/application are in what we consider failed (failed, terminating, notfound, unknown) it would return that resources status as well as the last message from it. There is still some discussion around if it will just return the first resource that isnt ready or if we want to have it go all the way through and look for what we consider the most important status to report. Example: failed might be more important to see than terminating.

### Scaling

As for scaling, getting status at large volumes I created an eks cluster with this yaml and ran podinfo with 200 replicas.
```
---
apiVersion: eksctl.io/v1alpha5
kind: ClusterConfig
metadata:
  name: jt-cluster
  region: us-west-2

# This section is required
iam:
  withOIDC: true
  serviceAccounts:
  - metadata:
      name: wego-service-account # Altering this will require a corresponding change in a later command
      namespace: wego-system
    roleOnly: true
    attachPolicy:
      Version: "2012-10-17"
      Statement:
      - Effect: Allow
        Action:
        - "aws-marketplace:RegisterUsage"
        Resource: '*'

managedNodeGroups:
- name: ng1
  instanceType: m5.large
  desiredCapacity: 3

```
It took a little less than 2 seconds to get all the info needed to generate the graph. This was with our current implentation using kstatus with no polling or cache. Flux has already solved this issue and have confirmed that it works with clusters of any size. Depending on the decision for the above option we either do it ourselves or use what flux already has.

As for how it works with CLI/UI:
CLI:
 When `gitops ui run` is used it will get the status for all resources in all clusters and using the kstatus cache it will save the last known status. The user will then be able to navigate the UI getting the status as needed from the cache in the backend.

UI: 
 This will work much like flux currently does. When the get status command is run it will always need to get the status for whatever resources are requested instead of getting it from a cache. Example `gitops get clusters` gets all resources for all clusters and applications. `gitops get apps` gets all resources for all applications.








