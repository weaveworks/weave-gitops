# Weave Gitops Release Process
We have two types of releases: stable, and pre-releases. These change
what steps are included.

- A pre-release version has a dash in it. Typically, this is of the
  form `v1.2.3-rc.4`, but as long as there's a dash in the version,
  it's a pre-release.
- A stable release does not have a dash in it. These numbers simply
  take the form `v1.2.3`.

There's two categories of things that get updated:
- Artifacts: docker images, binaries and javascript library
- Things that reference artifacts: links in the README, the versioned
 manual, a helm chart and a homebrew tap.

All releases release all artifacts, but only a stable release releases
things that reference artifacts - this is how a pre-release doesn't
get pushed to end-users.

To release a new version of Weave Gitops, you need to:
- Decide on an appropriate release number depending on if you want a
  pre-release or stable release. Do include a leading `v`. See [the
  releases page](https://github.com/weaveworks/weave-gitops/releases)
  for previous releases.
- Go to [the action to prepare a release](https://github.com/weaveworks/weave-gitops/actions/workflows/prepare-release.yaml)
  and click "Run workflow". In the popup, enter the version number you
  want and kick it off.
- Wait for the action to finish (1-5 minutes, depending on if it's an
  RC or not).
- Wait for CI, and review the PR.
  - For a pre-release, currently only a javascript version number
    should be updated.
  - If it's a stable release, there's also updates to our helm chart,
    documentation and README. Don't forget to check out the docs
    staging site (there's a link in the list of github status checks
    called "Doc site preview")
- If everything looks good, approve the PR - do *not* merge or things
  won't be published in the right order. This immediately kicks off the
  release job.
- Wait for the action to finish (~20 minutes), at which point the PR
  will be merged automatically.
- If it's a stable release, add a record of the new version in the
  checkpoint system.

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
