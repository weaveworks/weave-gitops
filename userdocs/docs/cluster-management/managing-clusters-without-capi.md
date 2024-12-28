---
title: Managing Clusters Without Cluster API

---

import CodeBlock from "@theme/CodeBlock";
import BrowserOnly from "@docusaurus/BrowserOnly";

# Managing Clusters Without Cluster API ~ENTERPRISE~

You do **not** need Cluster API to add your Kubernetes cluster to Weave GitOps Enterprise. The only thing you need is a secret containing a valid `kubeconfig`.

=== "Existing kubeconfig"

    ## Adding kubeconfig to Your Management Cluster

    If you already have a `kubeconfig` stored in a secret in your management cluster, continue with the "Create a `GitopsCluster`" step below.

    If you have a kubeconfig, but it is not yet stored in your management cluster, load it into the cluster using this command:

    ```
    kubectl create secret generic demo-01-kubeconfig \
    --from-file=value=./demo-01-kubeconfig
    ```

=== "How to create a kubeconfig for a ServiceAccount"

    Here's how to create a kubeconfig secret.

    1. Create a new service account on the remote cluster:

    ```yaml
    apiVersion: v1
    kind: ServiceAccount
    metadata:
    name: demo-01
    namespace: default
    ```

    2. Add RBAC permissions for the service account:

    ??? example "Expand to see role manifests"

        ```yaml
            ---
            apiVersion: rbac.authorization.k8s.io/v1
            kind: ClusterRoleBinding
            metadata:
                name: impersonate-user-groups
            subjects:
                - kind: ServiceAccount
                    name: wgesa
                    namespace: default
            roleRef:
                kind: ClusterRole
                name: user-groups-impersonator
                apiGroup: rbac.authorization.k8s.io
            ---
            apiVersion: rbac.authorization.k8s.io/v1
            kind: ClusterRole
            metadata:
                name: user-groups-impersonator
            rules:
                - apiGroups: [""]
                    resources: ["users", "groups"]
                    verbs: ["impersonate"]
                - apiGroups: [""]
                    resources: ["namespaces"]
                    verbs: ["get", "list"]
        ```

    This will allow WGE to introspect the cluster for available namespaces.

    Once we know what namespaces are available we can test whether the logged in user can access them via impersonation.

    3. Retrieve the token from the service account. First, run this command to get the list of secrets of the service accounts:

    ```bash
            kubectl get secrets --field-selector type=kubernetes.io/service-account-token
            NAME                      TYPE                                  DATA   AGE
            default-token-lsjz4       kubernetes.io/service-account-token   3      13d
            demo-01-token-gqz7p       kubernetes.io/service-account-token   3      99m
    ```

    (`demo-01-token-gqz7p` is the secret that holds the token for `demo-01` service account.)

    Then, run the following command to get the service account token:

    ```bash
    TOKEN=$(kubectl get secret demo-01-token-gqz7p -o jsonpath={.data.token} | base64 -d)
    ```

    4. Create a kubeconfig secret. We'll use a helper script to generate the kubeconfig, and then save it into `static-kubeconfig.sh`:

    ??? example "Expand to see script"

        ```bash title="static-kubeconfig.sh"
            #!/bin/bash
            if [[ -z "$CLUSTER_NAME" ]]; then
                    echo "Ensure CLUSTER_NAME has been set"
                    exit 1
            fi
            if [[ -z "$CA_CERTIFICATE" ]]; then
                    echo "Ensure CA_CERTIFICATE has been set to the path of the CA certificate"
                    exit 1
            fi
            if [[ -z "$ENDPOINT" ]]; then
                    echo "Ensure ENDPOINT has been set"
                    exit 1
            fi
            if [[ -z "$TOKEN" ]]; then
                    echo "Ensure TOKEN has been set"
                    exit 1
            fi
            export CLUSTER_CA_CERTIFICATE=$(cat "$CA_CERTIFICATE" | base64)
            envsubst <<EOF
            apiVersion: v1
            kind: Config
            clusters:
            - name: $CLUSTER_NAME
                cluster:
                    server: https://$ENDPOINT
                    certificate-authority-data: $CLUSTER_CA_CERTIFICATE
            users:
            - name: $CLUSTER_NAME
                user:
                    token: $TOKEN
            contexts:
            - name: $CLUSTER_NAME
                context:
                    cluster: $CLUSTER_NAME
                    user: $CLUSTER_NAME
            current-context: $CLUSTER_NAME
            EOF
        ```

    5. Obtain the cluster certificate (CA). How you do this depends on your cluster.

    - **AKS**: Visit the [Azure user docs](https://learn.microsoft.com/en-us/azure/aks/certificate-rotation) for more information.
    - **EKS**: Visit the [EKS docs](https://docs.aws.amazon.com/eks/latest/userguide/cert-signing.html) for more information.
    - **GKE**: You can view the CA on the GCP Console: Cluster->Details->Endpoint->”Show cluster certificate”.

    You'll need to copy the contents of the certificate into the `ca.crt` file used below.

    ```bash
    CLUSTER_NAME=demo-01 \
    CA_CERTIFICATE=ca.crt \
    ENDPOINT=<control-plane-ip-address> \
    TOKEN=<token> ./static-kubeconfig.sh > demo-01-kubeconfig
    ```

    6. Update the following fields:

    - CLUSTER_NAME: insert the name of your cluster—i.e., `demo-01`
    - ENDPOINT: add the API server endpoint i.e. `34.218.72.31`
    - CA_CERTIFICATE: include the path to the CA certificate file of the cluster
    - TOKEN: add the token of the service account retrieved in the previous step

    7. Finally, create a secret for the generated kubeconfig in the WGE management cluster:

    ```bash
    kubectl create secret generic demo-01-kubeconfig \
    --from-file=value=./demo-01-kubeconfig
    ```

## Add a Cluster Bootstrap Config

This step ensures that Flux gets installed into your cluster. Create a cluster bootstrap config as follows:

```bash
 kubectl create secret generic my-pat --from-literal GITHUB_TOKEN=$GITHUB_TOKEN
```

import CapiGitopsCDC from "!!raw-loader!./assets/bootstrap/capi-gitops-cluster-bootstrap-config.yaml";

Download the config with:

<BrowserOnly>
  {() => (
    <CodeBlock className="language-bash">
      curl -o
      clusters/management/capi/bootstrap/capi-gitops-cluster-bootstrap-config.yaml{" "}
      {window.location.protocol}
      //{window.location.host}
      {
        require("./assets/bootstrap/capi-gitops-cluster-bootstrap-config.yaml")
          .default
      }
    </CodeBlock>
  )}
</BrowserOnly>

Then update the `GITHUB_USER` variable to point to your repository

??? example "Expand to see full yaml"
    <CodeBlock
    title="clusters/management/capi/boostrap/capi-gitops-cluster-bootstrap-config.yaml"
    className="language-yaml"
    >
    {CapiGitopsCDC}
    </CodeBlock>

## Connect a Cluster

To connect your cluster, you need to add some common RBAC rules into the `clusters/bases` folder. When a cluster is provisioned, by default it will reconcile all the manifests in `./clusters/<cluster-namespace>/<cluster-name>` and `./clusters/bases`.

To display Applications and Sources in the UI, we need to give the logged-in user the permission to inspect the new cluster. Adding common RBAC rules to `./clusters/bases/rbac` is an easy way to configure this.

import WegoAdmin from "!!raw-loader!./assets/rbac/wego-admin.yaml";

<BrowserOnly>
  {() => (
    <CodeBlock className="language-bash">
      curl -o clusters/bases/rbac/wego-admin.yaml {window.location.protocol}//
      {window.location.host}
      {require("./assets/rbac/wego-admin.yaml").default}
    </CodeBlock>
  )}
</BrowserOnly>

??? example "Expand to see full template yaml"
    <CodeBlock
    title="clusters/bases/rbac/wego-admin.yaml"
    className="language-yaml"
    >
    {WegoAdmin}
    </CodeBlock>

## Create a `GitopsCluster`

When a `GitopsCluster` appears in the cluster, the Cluster Bootstrap Controller will install Flux on it and by default start reconciling the `./clusters/demo-01` path in your management cluster's Git repository:

```yaml title="./clusters/management/clusters/demo-01.yaml"
apiVersion: gitops.weave.works/v1alpha1
kind: GitopsCluster
metadata:
  name: demo-01
  namespace: default
  # Signals that this cluster should be bootstrapped.
  labels:
    weave.works/capi: bootstrap
spec:
  secretRef:
    name: demo-01-kubeconfig
```

To use the Weave GitOps Enterprise user interface (UI) to inspect the Applications and Sources running on the new cluster, you'll need permissions. We took care of this above when we stored your RBAC rules in `./clusters/bases`. In the following step, we'll create a kustomization to add these common resources onto our new cluster:

```yaml title="./clusters/demo-01/clusters-bases-kustomization.yaml"
apiVersion: kustomize.toolkit.fluxcd.io/v1beta2
kind: Kustomization
metadata:
  creationTimestamp: null
  name: clusters-bases-kustomization
  namespace: flux-system
spec:
  interval: 10m0s
  path: clusters/bases
  prune: true
  sourceRef:
    kind: GitRepository
    name: flux-system
```

Save these two files in your Git repository, then commit and push.

Once Flux has reconciled the cluster, you can inspect your Flux resources via the UI!

## Debugging Tip: Checking that Your kubeconfig Secret Is in Your Cluster

To test that your kubeconfig secret is correctly set up, apply the following manifest and check the logs after the job completes:

??? example "Expand to see manifest"

    ```yaml
        ---
        apiVersion: batch/v1
        kind: Job
        metadata:
        name: kubectl
        spec:
        ttlSecondsAfterFinished: 30
        template:
            spec:
            containers:
                - name: kubectl
                image: bitnami/kubectl
                args:
                    [
                    "get",
                    "pods",
                    "-n",
                    "kube-system",
                    "--kubeconfig",
                    "/etc/kubeconfig/value",
                    ]
                volumeMounts:
                    - name: kubeconfig
                    mountPath: "/etc/kubeconfig"
                    readOnly: true
            restartPolicy: Never
            volumes:
                - name: kubeconfig
                secret:
                    secretName: demo-01-kubeconfig
                    optional: false
    ```

In the manifest above, `demo-01-kubeconfig` is the name of the secret that contains the kubeconfig for the remote cluster.

---

## Additional Resources

Other documentation that you might find useful:

- [Authentication strategies](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#authentication-strategies)
  - [X509 client certificates](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#x509-client-certs): can be used across different namespaces
  - [Service account tokens](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#service-account-tokens): limited to a single namespace
- [Kubernetes authentication 101 (CNCF blog post)](https://www.cncf.io/blog/2020/07/31/kubernetes-rbac-101-authentication/)
- [Kubernetes authentication (Magalix blog post)](https://www.magalix.com/blog/kubernetes-authentication)
