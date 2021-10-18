# Weave Gitops Release Process

The current release process for weave gitops is fairly straightforward. You need to:
- Create the actual release
- Update the `weave-gitops-docs` repository with documentation for the new version
- Update the `CLI Installation` section of the `README.md` in the `weave-gitops` repository to reference the new version
- Add a record of the new version in the checkpoint system

# Creating the release
- Update `package.json` and `package-lock.json` . These changes must be on `main` before creating a release.
  - Update the package.json `version` field to reflect the new version.
  - Run `npm ci` to update the file `package-lock.json`.
- In the terminal on main branch run `./tools/tag-release.sh <arg>` 
  - Args for tag-release.sh are:
    - -d Dry run
    - -M for a major release candidate"
    - -m for a minor release candidate"
    - -p for a patch release candidate"
    - -r for a full release"
- A release candidate is needed to do a full release. First create a release candidate (example patch release candidate `./tools/tag-release.sh -p`)
- After a release candidate is created a full release can be made with `./tools/tag-release.sh -r`

The go-releaser will spin for a bit, generating a changelog and artifacts.

# Updating weave-gitops-docs
Assuming any changes to user-visible behavior have already been documented for the new release (this should happen when the changes go into `weave-gitops`), you just need to create a new documentation version:
- Create a branch from main
- Checkout the new branch
- Update `docs/installation.md`:
  - Put the new version in the `curl` command
  - Run `wego version` in a downloaded copy of the new release binary and update the `You should see:` section with the output
- Create a new set of versioned documentation:
  - Run:

```console
npm run docusaurus docs:version N.N.N
```

where the `N.N.N` matches the `N.N.N` from the weave gitops version.

- Add and commit the new files (which will be a new directory containing copies of the files in main at the current point in time).
- Create a PR and merge when approved

# Updating the README
- Once the release is available, change the version in the `curl` command shown in the `CLI Installation` section of the `README.md` in the weave-gitops repository
- Create a PR and merge when approved

# Record the new version 
- Add a record in the [checkpoint system](https://checkpoint-api.weave.works/admin) to inform users of the new version.  The CLI checks for a more recent version and informs the user where to download it based on this data.
  - Record must match this template:
     ```
    Name: weave-gitops
    Version: N.N.N
    Release date: (current date in UTC. i.e.: 2021-09-08 12:41:00 )
    Download URL: https://github.com/weaveworks/weave-gitops/releases/tag/vN.N.N
    Changelog URL: https://github.com/weaveworks/weave-gitops/releases/tag/vN.N.N
    Project Website: https://www.weave.works/product/gitops-core/
    ```
  - _note: A Weaveworks employee must perform this step_

That's it!
