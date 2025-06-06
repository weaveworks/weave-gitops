## gitops completion bash

Generate the autocompletion script for bash

### Synopsis

Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:

```
source <(gitops completion bash)
```

To load completions for every new session, execute once:

#### Linux:

```
gitops completion bash > /etc/bash_completion.d/gitops
```

#### macOS:

```
gitops completion bash > $(brew --prefix)/etc/bash_completion.d/gitops
```

You will need to start a new shell for this setup to take effect.


```
gitops completion bash
```

### Options

```
  -h, --help              help for bash
      --no-descriptions   disable completion descriptions
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

* [gitops completion](gitops_completion.md)	 - Generate the autocompletion script for the specified shell

###### Auto generated by spf13/cobra on 25-Oct-2023
