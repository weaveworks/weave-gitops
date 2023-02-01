## gitops remove run

Remove GitOps Run sessions

### Synopsis

Remove GitOps Run sessions

```
gitops remove run [flags]
```

### Examples

```

# Remove the GitOps Run session "dev-1234" from the "flux-system" namespace
gitops remove run --namespace flux-system dev-1234

# Remove all GitOps Run sessions from the default namespace
gitops remove run --all-sessions

# Remove all GitOps Run sessions from the dev namespace
gitops remove run -n dev --all-sessions

```

### Options

```
      --all-sessions     Remove all GitOps Run sessions
      --context string   The name of the kubeconfig context to use
  -h, --help             help for run
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

* [gitops remove](gitops_remove.md)	 - Remove various components of Weave GitOps

