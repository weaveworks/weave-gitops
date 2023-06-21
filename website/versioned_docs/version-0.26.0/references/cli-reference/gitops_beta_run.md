## gitops beta run

Set up an interactive sync between your cluster and your local file system

### Synopsis

This will set up a sync between the cluster in your kubeconfig and the path that you specify on your local filesystem.  If you do not have Flux installed on the cluster then this will add it to the cluster automatically.  This is a requirement so we can sync the files successfully from your local system onto the cluster.  Flux will take care of producing the objects for you.

```
gitops beta run [flags]
```

### Examples

```

# Run the sync on the current working directory
gitops beta run . [flags]

# Run the sync against the dev overlay path
gitops beta run ./deploy/overlays/dev

# Run the sync on the dev directory and forward the port.
# Listen on port 8080 on localhost, forwarding to 5000 in a pod of the service app.
gitops beta run ./dev --port-forward port=8080:5000,resource=svc/app

# Run the sync on the dev directory with a specified root dir.
gitops beta run ./clusters/default/dev --root-dir ./clusters/default

# Run the sync on the podinfo demo.
git clone https://github.com/stefanprodan/podinfo
cd podinfo
gitops beta run ./deploy/overlays/dev --no-session --timeout 3m --port-forward namespace=dev,resource=svc/backend,port=9898:9898

# Run the sync on the podinfo demo in the session mode.
git clone https://github.com/stefanprodan/podinfo
cd podinfo
gitops beta run ./deploy/overlays/dev --timeout 3m --port-forward namespace=dev,resource=svc/backend,port=9898:9898

# Run the sync on the podinfo Helm chart, in the session mode. Please note that file Chart.yaml must exist in the directory.
git clone https://github.com/stefanprodan/podinfo
cd podinfo
gitops beta run ./charts/podinfo --timeout 3m --port-forward namespace=flux-system,resource=svc/run-dev-helm-podinfo,port=9898:9898
```

### Options

```
      --allow-k8s-context strings          The name of the KubeConfig context to explicitly allow.
      --components strings                 The Flux components to install. (default [source-controller,kustomize-controller,helm-controller,notification-controller])
      --components-extra strings           Additional Flux components to install, allowed values are image-reflector-controller,image-automation-controller.
      --context string                     The name of the kubeconfig context to use
      --dashboard-hashed-password string   GitOps Dashboard password in BCrypt hash format
      --dashboard-port string              GitOps Dashboard port (default "9001")
      --decryption-key-file string         Path to an age key file used for decrypting Secrets using SOPS.
      --disable-compression                If true, opt-out of response compression for all requests to the server
      --flux-version string                The version of Flux to install. (default "0.37.0")
  -h, --help                               help for run
      --no-bootstrap                       Disable bootstrapping at shutdown.
      --no-session                         Disable session management. If not specified, the session will be enabled by default.
      --port-forward string                Forward the port from a cluster's resource to your local machine i.e. 'port=8080:8080,resource=svc/app'.
      --root-dir string                    Specify the root directory to watch for changes. If not specified, the root of Git repository will be used.
      --session-name string                Specify the name of the session. If not specified, the name of the current branch and the last commit id will be used. (default "run-main-9da468fe-dirty")
      --session-namespace string           Specify the namespace of the session. (default "default")
      --skip-dashboard-install             Skip installation of the Dashboard. This also disables the prompt asking whether the Dashboard should be installed.
      --skip-resource-cleanup              Skip resource cleanup. If not specified, the GitOps Run resources will be deleted by default.
      --timeout duration                   The timeout for operations during GitOps Run. (default 5m0s)
```

### Options inherited from parent commands

```
  -e, --endpoint WEAVE_GITOPS_ENTERPRISE_API_URL   The Weave GitOps Enterprise HTTP API endpoint can be set with WEAVE_GITOPS_ENTERPRISE_API_URL environment variable
      --insecure-skip-tls-verify                   If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --kubeconfig string                          Paths to a kubeconfig. Only required if out-of-cluster.
  -n, --namespace string                           The namespace scope for this operation (default "flux-system")
  -p, --password WEAVE_GITOPS_PASSWORD             The Weave GitOps Enterprise password for authentication can be set with WEAVE_GITOPS_PASSWORD environment variable
  -u, --username WEAVE_GITOPS_USERNAME             The Weave GitOps Enterprise username for authentication can be set with WEAVE_GITOPS_USERNAME environment variable
```

### SEE ALSO

* [gitops beta](gitops_beta.md)	 - This component contains unstable or still-in-development functionality

