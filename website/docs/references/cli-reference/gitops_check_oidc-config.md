## gitops check oidc-config

Check an OIDC configuration for proper functionality.

### Synopsis

This command will send the user through an OIDC authorization code flow using the given OIDC configuration. This is helpful for verifying that a given configuration will work properly with Weave GitOps or for debugging issues. Without any provided flags it will read the configuration from a Secreton the cluster.

```
gitops check oidc-config [flags]
```

### Examples

```

# Check the OIDC configuration stored in the flux-system/oidc-auth Secret
gitops check oidc-config

# Check a different set of scopes
gitops check oidc-config --scopes=openid,groups

# Check a different username cliam
gitops check oidc-config --claim-username=sub

# Check configuration without fetching a Secret from the cluster
gitops check oidc-config --skip-secret --client-id=CID --client-secret=SEC --issuer-url=https://example.org
		
```

### Options

```
      --claim-username string   ID token claim to use as the user name.
      --client-id string        OIDC client ID
      --client-secret string    OIDC client secret
      --context string          The name of the kubeconfig context to use
      --disable-compression     If true, opt-out of response compression for all requests to the server
      --from-secret string      Get OIDC configuration from the given Secret resource (default "oidc-auth")
  -h, --help                    help for oidc-config
      --issuer-url string       OIDC issuer URL
      --scopes strings          OIDC scopes to request (default [openid,offline_access,email,groups])
      --skip-secret             Do not read OIDC configuration from a Kubernetes Secret but rely solely on the values from the given flags.
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

* [gitops check](gitops_check.md)	 - Validates flux compatibility

