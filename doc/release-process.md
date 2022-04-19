# Weave Gitops Release Process

To release a new version of Weave Gitops, you need to:
- Decide on a new release number. We use e.g. `v1.2.3-rc.4` for
  pre-releases, and e.g. `v1.2.3` for releases.
- Go to [the action to prepare a release](https://github.com/weaveworks/weave-gitops/actions/workflows/prepare-release.yaml)
  and click "Run workflow". In the popup, enter the version number you
  want and kick it off.
- Wait for the action to finish (1-5 minutes, depending on if it's an
  RC or not).
- Review the PR. Don't forget to check out the docs staging site!
- If everything looks good, approve the PR. This will kick off the
  release job.
- Wait for the action to finish (~20 minutes), at which point the PR
  will be merged automatically.
- Add a record of the new version in the checkpoint system

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

# Technical details
There's 2 jobs, prepare-release and release. prepare-release is only
triggered manually, release is triggered whenever the state of a
branch called `release/something` turns approved.

The prepare job always updates the javascript package version. It only
updates the docs and the helm chart if what we're releasing looks like
a stable release.

The release job tags the PR itself, builds binaries and images, and
then merges the PR. This means that we already have binaries and
images available by the time the docs and the helm chart is merged and
uploaded. However, that means that since the tag sits on a non-main
branch, we have to merge (no rebase, no squash), or the ref won't be
traceable form main.
