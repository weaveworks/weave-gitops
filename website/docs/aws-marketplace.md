---
title: AWS Marketplace
sidebar_position: 6
hide_title: true
---

import TierLabel from "./_components/TierLabel";

<h1>
  {frontMatter.title} <TierLabel tiers="Core" />
</h1>


Weave GitOps is also available via the AWS Marketplace. 

# Deploy Weave GitOps on an EKS Cluster via Helm

The following steps will allow you to deploy the Weave GitOps product to an EKS cluster via a Helm Chart.

These instructions presume you already have installed [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl/),
[`eksctl`](https://github.com/weaveworks/eksctl), [`helm`](https://github.com/helm/helm) and
the [Helm S3 Plugin](https://github.com/hypnoglow/helm-s3).

## Step 1: Subscribe to Weave GitOps on the AWS Marketplace

To deploy the managed Weave GitOps solution, first subscribe to the product on [AWS Marketplace](https://aws.amazon.com/marketplace/pp/prodview-vkn2wejad2ix4).
**This subscription is only available for deployment on EKS versions 1.17-1.21.**

_Note: it may take ~20 minutes for your Subscription to become live and deployable._

## [Optional] Step 2: Create an EKS cluster

**If you already have an EKS cluster, you can skip ahead to Step 3.**

If you do not have a cluster on EKS, you can use [`eksctl`](https://github.com/weaveworks/eksctl) to create one.

Copy the contents of the sample file below into `cluster-config.yaml` and replace the placeholder values with your settings.
See the [`eksctl` documentation](https://eksctl.io/) for more configuration options.

```yaml
---
apiVersion: eksctl.io/v1alpha5
kind: ClusterConfig
metadata:
  name: CLUSTER_NAME # Change this
  region: REGION # Change this

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

# This section will create a single Managed nodegroup with one node.
# Edit or remove as desired.
managedNodeGroups:
- name: ng1
  instanceType: m5.large
  desiredCapacity: 1
```

Create the cluster:

```bash
eksctl create cluster -f cluster-config.yaml
```

## [Optional] Step 3: Update your EKS cluster

**If you created your cluster using the configuration file in Step 2, your cluster is
already configured correctly and you can skip ahead to Step 4.**

In order to use the Weave GitOps container product,
your cluster must be configured to run containers with the correct IAM Policies.

The recommended way to do this is via [IRSA](https://aws.amazon.com/blogs/opensource/introducing-fine-grained-iam-roles-service-accounts/).

Use this `eksctl` configuration below (replacing the placeholder values) to:
- Associate an OIDC provider
- Create the required service account ARN

Save the example below as `oidc-config.yaml`
```yaml
---
apiVersion: eksctl.io/v1alpha5
kind: ClusterConfig
metadata:
  name: CLUSTER_NAME # Change this
  region: REGION # Change this

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

```

```bash
eksctl utils associate-iam-oidc-provider -f oidc-config.yaml --approve
eksctl create iamserviceaccount -f oidc-config.yaml --approve
```

## Step 4: Fetch the Service Account Role ARN
First retrieve the ARN of the IAM role which you created for the `wego-service-account`:

```bash
# replace the placeholder values with your configuration
# if you changed the service account name from wego-service-account, update that in the command
export SA_ARN=$(eksctl get iamserviceaccount --cluster <cluster-name> --region <region> | awk '/wego-service-account/ {print $3}')

echo $SA_ARN
# should return
# arn:aws:iam::<account-id>:role/eksctl-<cluster-name>-addon-iamserviceaccount-xxx-Role1-1N41MLVQEWUOF
```

_This value will also be discoverable in your IAM console, and in the Outputs of the Cloud Formation
template which created it._

## Step 5: Install Weave GitOps

Copy the Chart URL from Usage Instructions, or download the file from the Deployment template to your workstation.

```bash
helm install wego <URL/PATH> \
  --set serviceAccountRole="$SA_ARN"

# if you changed the name of the service account
helm install wego <URL/PATH> \
  --set serviceAccountName='<name>' \
  --set serviceAccountRole="$SA_ARN"
```

## Step 6: Check your installation

Run the following from your workstation:

```bash
kubectl get pods -n wego-system
# you should see something like the following returned
wego-system          helm-controller-5b96d94c7f-tds9n                    1/1     Running   0          53s
wego-system          image-automation-controller-5cf75fd555-zqm89        1/1     Running   0          53s
wego-system          image-reflector-controller-6787985855-l4q4g         1/1     Running   0          53s
wego-system          kustomize-controller-8467b8b884-x2cpd               1/1     Running   0          53s
wego-system          notification-controller-55f94bc746-ggmwc            1/1     Running   0          53s
wego-system          source-controller-78bfb8576-stnr5                   1/1     Running   0          53s
wego-system          wego-metering-f7jqp                                 1/1     Running   0          53s
```

Your Weave GitOps installation is now ready!

