---
title: Step 3 - Deploy an Application

---

# Step 3: Deploy an Application

Now that you have a feel for how to navigate the dashboard, let's deploy a new
application. In this section we will use [podinfo](https://github.com/stefanprodan/podinfo) as our sample web application.

## Deploying podinfo

1. Clone or navigate back to your Git repository where you have bootstrapped Flux. For example:

   ```
   git clone https://github.com/$GITHUB_USER/fleet-infra
   cd fleet-infra
   ```

1. Create a `GitRepository` Source for podinfo. This will allow you to use different authentication methods for different repositories.

   ```
   flux create source git podinfo \
     --url=https://github.com/stefanprodan/podinfo \
     --branch=master \
     --interval=30s \
     --export > ./clusters/management/podinfo-source.yaml
   ```

More information about `GitRepository` is available [here](https://fluxcd.io/flux/components/source/gitrepositories/). 

If you get stuck here, try the `ls` command to list your files and directories. If that doesn’t work, try `ls -l ./clusters`.

1. Commit and push the `podinfo-source` to your `fleet-infra` repository

   ```
   git add -A && git commit -m "Add podinfo source"
   git push
   ```

1. Create a `kustomization` to build and apply the podinfo manifest

   ```
   flux create kustomization podinfo \
     --target-namespace=flux-system \
     --source=podinfo \
     --path="./kustomize" \
     --prune=true \
     --interval=5m \
     --export > ./clusters/management/podinfo-kustomization.yaml
   ```

1. Commit and push the `podinfo-kustomization` to your `fleet-infra` repository

   ```
   git add -A && git commit -m "Add podinfo kustomization"
   git push
   ```

## View the Application in Weave GitOps

Flux will detect the updated `fleet-infra` and add podinfo. Navigate back to the [dashboard](http://localhost:9001/) to make sure that the podinfo application appears.

![Applications summary view showing Flux System, Weave GitOps and Podinfo](/img/dashboard-applications-with-podinfo.png)

Click on podinfo to find details about the deployment. There should be two pods available.

![Applications details view for podinfo showing 2 pods](/img/dashboard-podinfo-details.png)

!!! info
    Podinfo comes with a HorizontalPodAutoscaler, which uses the `metrics-server`.
    We don't use the `metrics-server` in this tutorial, but note that it's the reason why HorizontalPodAutoscaler will report as `Not ready` in your dashboard. We recommend ignoring the warning.

## Customize podinfo

To customize a deployment from a repository you don’t control, you can use Flux in-line patches. The following example shows how to use in-line patches to change the podinfo deployment.

1. Add the `patches` section as shown below to the field spec of your `podinfo-kustomization.yaml` file so it looks like this:

??? example "Expand to see Kustomization patches"

    ```yaml title="./clusters/management/podinfo-kustomization.yaml"
    ---
    apiVersion: kustomize.toolkit.fluxcd.io/v1beta2
    kind: Kustomization
    metadata:
        name: podinfo
        namespace: flux-system
    spec:
        interval: 60m0s
        path: ./kustomize
        prune: true
        sourceRef:
        kind: GitRepository
        name: podinfo
        targetNamespace: flux-system
    // highlight-start
        patches:
        - patch: |-
            apiVersion: autoscaling/v2beta2
            kind: HorizontalPodAutoscaler
            metadata:
                name: podinfo
            spec:
                minReplicas: 3
            target:
            name: podinfo
            kind: HorizontalPodAutoscaler
    // highlight-end
    ```

	 </details>

1. Commit and push the `podinfo-kustomization.yaml` changes:

   ```
   git add -A && git commit -m "Increase podinfo minimum replicas"
   git push
   ```

3. Navigate back to the dashboard. You should see a newly created pod:

   ![Applications details view for podinfo showing 3 pods](/img/dashboard-podinfo-updated.png)


## Suspend updates

Suspending updates to a kustomization allows you to directly edit objects applied from a kustomization, without your changes being reverted by the state in Git.

To suspend updates for a kustomization, from the details page, click on the suspend button at the top, and you should see it be suspended:

![Podinfo details showing Podinfo suspended](/img/dashboard-podinfo-details-suspended.png)

This shows in the applications view with a yellow warning status indicating it is now suspended

![Applications summary view showing Podinfo suspended](/img/dashboard-podinfo-suspended.png)

To resume updates, go back to the details page, click the resume button, and after a few seconds reconsolidation will continue.

## Delete Podinfo 

To delete Podinfo in the GitOps way, run this command from the root of your working directory:

```
  rm ./clusters/management/podinfo-kustomization.yaml
  rm ./clusters/management/podinfo-source.yaml
  git add -A && git commit -m "Remove podinfo kustomization and source"
  git push
```

## Complete!

Congratulations 🎉🎉🎉

You've now completed the getting started guide. We welcome any and all [feedback](/feedback-and-telemetry), so please let us know how we could have made your experience better.
