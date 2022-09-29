## gitops create dashboard

Create a HelmRepository and HelmRelease to deploy Weave GitOps

### Synopsis

Create a HelmRepository and HelmRelease to deploy Weave GitOps

```
gitops create dashboard [flags]
```

### Examples

```

# Create a HelmRepository and HelmRelease to deploy Weave GitOps
gitops create dashboard ww-gitops \
  --password=$PASSWORD \
  --export > ./clusters/my-cluster/weave-gitops-dashboard.yaml
		
```

### Options

```
      --context string   The name of the kubeconfig context to use
  -h, --help             help for dashboard
```

### Options inherited from parent commands

```
  -e, --endpoint WEAVE_GITOPS_ENTERPRISE_API_URL   The Weave GitOps Enterprise HTTP API endpoint can be set with WEAVE_GITOPS_ENTERPRISE_API_URL environment variable
      --export                                     Export in YAML format to stdout.
      --insecure-skip-tls-verify                   If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --kubeconfig string                          Paths to a kubeconfig. Only required if out-of-cluster.
      --namespace string                           The namespace scope for this operation (default "flux-system")
  -p, --password WEAVE_GITOPS_PASSWORD             The Weave GitOps Enterprise password for authentication can be set with WEAVE_GITOPS_PASSWORD environment variable
      --timeout duration                           The timeout for operations during resource creation. (default 3m0s)
  -u, --username WEAVE_GITOPS_USERNAME             The Weave GitOps Enterprise username for authentication can be set with WEAVE_GITOPS_USERNAME environment variable
```

### SEE ALSO

* [gitops create](gitops_create.md)	 - Creates a resource

