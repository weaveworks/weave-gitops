# Weave Gitops Release Process

The current release process for weave gitops is fairly straightforward. You need to:
- Create the actual release
- Update the `weave-gitops-docs` repository with documentation for the new version
- Update the `CLI Installation` section of the `README.md` in the `weave-gitops` repository to reference the new version

# Creating the release
- Update the package.json `version` field to reflect the new version. This change must be on `main` before creating a release.
- Go to the `Releases` page for the weave-gitops repository
- Click on `Draft a New Release`
- Fill in the `Tag Version` field with the new version (format: `vN.N.N` or `vN.N.N-rc.N` for a pre-release). We have configured go-releaser to implicitly make a version ending in `-rc.N` a pre-release and one without the `rc-N` a full release (but that can be changed after the fact by editing the release if so desired).
- Fill in the `Release Title` with the same version from the `Tag Version`
- Click on `Publish Release`

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

That's it!
