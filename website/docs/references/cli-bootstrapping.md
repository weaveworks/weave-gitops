# CLI Bootstrapping Reference

-  `--kubeconfig`                   Paths to a kubeconfig. Only required if out-of-cluster.
-  `--bootstrap-flux`               chose whether you want to install flux in the generic way in case no flux installation detected
-  `-b`, `--branch`                 git branch for your flux repository (example: main)
-  `-i`, `--client-id`              OIDC client ID
-  `--client-secret`                OIDC client secret
-  `--components-extra`             extra components to be installed from (policy-agent, tf-controller)
-  `--discovery-url`                OIDC discovery URL
-  `--export`                       write to stdout the bootstrapping manifests without writing in the cluster or Git. It requires Flux to be bootstrapped.
-  `--git-password`                 git password/token used in https authentication type
-  `--git-username`                 git username used in https authentication type
-  `-h`, `--help`                   help for bootstrap
-  `-k`, `--private-key`            private key path. This key will be used to push the Weave GitOps Enterprise's resources to the default cluster repository
-  `-c`, `--private-key-password`   private key password. If the private key is encrypted using password
-  `-r`, `--repo-path`              git path for your flux repository (example: clusters/my-cluster)
-  `--repo-url`                     Git repo URL for your Flux repository. For supported URL examples see [here](https://fluxcd.io/flux/cmd/flux_bootstrap_git/)
-  `-s`, `--silent`                 chose the defaults with current provided information without asking any questions
-  `-v`, `--version`                version of Weave GitOps Enterprise (should be from the latest 3 versions)
-  `-p`, `--password`               The Weave GitOps Enterprise password for dashboard access