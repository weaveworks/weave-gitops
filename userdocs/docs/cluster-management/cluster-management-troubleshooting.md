---
title: Cluster Management Troubleshooting
---

# Cluster Management Troubleshooting ~ENTERPRISE~

We'll use this page to help you move past common troublesome situations. 

## Git Repositories and Resources

To authenticate using Git during the pull request creation, you will need to select the Git repository where you'll create the pull request.  

Depending on the action performed on the resource (creation/deletion/editing), the default Git repository selected in the UI is determined in the following order:

1. the repository used to initially create the resource found in the `templates.weave.works/create-request` annotation (in the case of editing or deleting of resources)
  ```yaml
  metadata:
    annotations:
      templates.weave.works/create-request: "{...\"parameter_values\":{...\"url\":\"https://github.com/weave-example-org/weave-demo\"}"
  ```

2. the first repository found with a `weave.works/repo-role: default` annotation
  ```yaml
  metadata:
    annotations:
      weave.works/repo-role: default
  ```

3. the flux-system repository 
  ```yaml
  metadata:
    name: flux-system
    namespace: flux-system
  ```

4. the first repository in the list of Git repositories that the user has access to.

In the case of deletion and editing, if the resource repository is found amongst the Git repositories that the user has access to, it will be preselected and the selection will be disabled. If it is not found, you can choose a new repository.

In the case of tenants, we recommend adding the `weave.works/repo-role: default` to an appropriate Git repository.

### Overriding the Calculated Git Repository HTTPS URL

The system will try and automatically calculate the correct HTTPS API endpoint to create a pull request against. For example, if the Git repository URL is `ssh://git@github.com/org/repo.git`, the system will try and convert it to `https://github.com/org/repo.git`.

However, it is not always possible to accurately derive this URL. An override can be specified to set the correct URL instead. For example, the SSH URL may be `ssh://git@interal-ssh-server:2222/org/repo.git` and the correct HTTPS URL may be `https://gitlab.example.com/org/repo.git`. 

In this case, we set the override via the `weave.works/repo-https-url` annotation on the `GitRepository` object:

```yaml
apiVersion: source.toolkit.fluxcd.io/v1beta1
kind: GitRepository
metadata:
  name: repo
  namespace: flux-system
  annotations:
    // highlight-start
    weave.works/repo-https-url: https://gitlab.example.com/org/repo.git
    // highlight-end
spec:
  interval: 1m
  url: ssh://git@interal-ssh-server:2222/org/repo.git
```

The pull request will then be created against the correct HTTPS API.

The above also applies to application creation.
