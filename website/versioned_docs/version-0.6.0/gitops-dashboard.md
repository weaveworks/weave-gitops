---
sidebar_position: 4
---

# GitOps Dashboard

Weave GitOps provides a web UI to help you quickly understand your Application deployments and perform common operations, such as adding a new Application to be deployed to your cluster. The `gitops` binary contains an HTTP server that can be used to start this browser interface as per the instructions below:

To run the dashboard:

```shell
$ gitops ui run
INFO[0000] Opening browser at http://0.0.0.0:9001/
INFO[0000] Serving on port 9001
Opening in existing browser session.
```

Your browser should open to the Weave GitOps UI:

![Weave GitOps UI](/img/wego_ui.png)

## What information can I discover about my Applications?

Applications being managed by Weave GitOps are displayed in a list. Clicking the name of an Application allows you to view more details including:
- It's name, deployment type (Kustomize or Helm), URL for the source repository being synchronized to the cluster and the specific Path within the repository where we are looking for deployment manifests.
- A reconciliation graph detailing the on-cluster components which have been created as a result of the deployment.
- Information from Flux regarding the state of the reconciliation
- A list of the 10 most recent commits to the source git repository helping you to verify which change has been applied to the cluster. This includes a hyperlink back to your git provider for each commit.

## Future development
The GitOps Dashboard is under active development, watch this space for exciting new features.
