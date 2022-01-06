### Status of my Application

As a user, I expect to see the status of my application running across multiple clusters. When I go to my application list page I should be able to quickly see the "health" of each application in each cluster within a timely manner.

## Problem to be Solved

- Come up with a way for the CLI to have parity with the UI status request.
- Determine how to get the status that will scale with larger clusters and multiple clusters.
- Determine if there are unique statuses for helm or kustomize.

## Outcome

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

The UI will remain the same while the CLI will need to be updated like so to match: 
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
Application       josh-pod-info            ready
Deployment        backend                  ready
ReplicaSet        backend-6b944d8b         ready
Pod               backend-6b944d8b-tzxnl   ready
Deployment        frontend                 ready
Kustomization     podinfo-deploy           ready
```

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
It took a little less than 2 seconds to get all the info needed to generate the graph. If scaling is needed in the future kstatus has the ability to collect and save the last reported status for a resource https://github.com/kubernetes-sigs/cli-utils/blob/v0.26.1/pkg/kstatus/polling/collector/collector.go. We could save that information in memory or if need be a redis cluster. I believe that scaling may not be needed at this point and will not be hard to implement in the future using the tools we currently have.


Kstatus does not work for helm. The UI currently does not support a helm application. The way the server code handles the status though should make it easy to include helm. Instead of using status.Compute to get the status for helm we could use the information returned when s.kube.list is called. GetReconciledObjects somewhat handles helm but it needs to have code added to fully support getting the status for helm.



