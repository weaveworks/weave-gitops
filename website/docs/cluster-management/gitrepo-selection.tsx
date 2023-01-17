# Git repositories and resources

During the pull request creation, to authenticate using Git, you will need to select the git repository where the pull request will be created.  

To set up a default git repository, the `weave.works/repo-role: default` annotation can be added to the git repository object.

Depending on the action performed on the resource (creation/deletion/editing), the default git repository will be set as (from highest to lowest priority):

- the repository used to initially create the resource found in the `templates.weave.works/create-request` annotation (in the case of editing or deleting of resources) or
```yaml
        metadata: {
          annotations: {
            "templates.weave.works/create-request": "{...\"parameter_values\":{...\"url\":\"https://github.com/weave-example-org/weave-demo\"}",
          },
        }
```

- the repository containing the `weave.works/repo-role: default` annotation or 
```yaml
        metadata: {
          annotations: {
            "weave.works/repo-role": "default",
          },
        }
```

- the flux-system repository 
```yaml
          metadata: {
            name: "flux-system",
            namespace: "flux-system",
          }
 ```
- the first repository in the list of git repositories that the user has access to

In the case of deletion and editing, if the resource repository is found amongst the git repositories that the user has access to, it will be preselected and the selection will be disabled. If it is not found, the user will be able to choose a new repository.

Some link formats (e.g git@github.com:org/repo.git) are not supported by `flux`. To add a link in the https format to your git repository object (which will override the initial url that the git repository is using), you can use the `weave.works/repo-https-url` annotation.

In the case of tenants, the git repository to be used will need to be added to one of the tenant namespaces together with the `weave.works/repo-role: default` added to it.

The above also applies to the creation of applications.
