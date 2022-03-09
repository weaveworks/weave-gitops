# Weave Gitops Release Process

The current release process for weave gitops is fairly straightforward. You need to:
- Create the actual release
- Update the [website](/website) with documentation for the new version
- Update the `CLI Installation` section of the `README.md` in the `weave-gitops` repository to reference the new version
- Add a record of the new version in the checkpoint system

# Creating the release
- In the terminal on main branch run `./tools/tag-release.sh <arg>`
  - Args for tag-release.sh are:
    - -d Dry run
    - -M for a major release candidate
    - -m for a minor release candidate
    - -p for a patch release candidate
    - -c for creating a new release candidate build
    - -r for a full release
- A release candidate is needed to do a full release. First create a release candidate (example patch release candidate `./tools/tag-release.sh -p`)
- Update `package.json` and `package-lock.json` . These changes must be on `main` before creating a release.
  - Update the package.json `version` field to reflect the new version.
  - Run `npm ci` to update the file `package-lock.json`.
- After a release candidate is created and the package files are updated, a full release can be made with `./tools/tag-release.sh -r`

- _note: a dry run can be run to show what tag will be created by adding `-d` example: `./tools/tag-release.sh -p -d`_

The go-releaser will spin for a bit, generating a changelog and artifacts.

# Updating the website
- Approve and merge the auto-generated PR if the release is declared satisfactory; otherwise, close the PR

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
