# Weave Gitops Release Process

## Versioning

We follow [semantic versioning](https://semver.org/) where
- Releases adding new features or changing existing ones increase the minor versions (0.11.0, 0.12.0, etc)
- Releases exclusively fixing bugs increase the patch version (0.11.1, 0.11.2)

## Releases

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

## Let's get started

To release a new version of Weave Gitops, you need to:
- Verify that there are no outstanding PRs that need to be merge in the weave-gitops-dev channel on Slack, then notify the channel that you are starting the release process, and to not merge anything into main until you say otherwise.
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
  - For a pre-release, most steps are skipped - there's updates to the
    javascript, and maybe some small documentation changes.
  - If it's a stable release, there's also updates to our helm chart,
    documentation and README. Don't forget to check out the docs
    staging site (there's a link in the list of github status checks
    called "Doc site preview")
- The PR cover message contains draft release notes. Edit the cover
  message to fill in or delete blocks as appropriate. Move as many PRs out of "Uncategorized" as you possibly can. 
- If everything looks good, approve the PR - do *not* merge or things
  won't be published in the right order. This immediately kicks off the
  release job.
- Cross you fingers and ask for a blessing from Mr. Kubernetes, then wait for the action to finish (~20 minutes), at which point the PR will be merged automatically.
- Notify weave-gitops-dev channel that PRs are now safe to merge

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

# Artifacts

Each stable release emits the artifacts listed below:

- [versioned instance of the website](https://github.com/weaveworks/weave-gitops/tree/main/website/versioned_docs)
- [New chart version](https://github.com/weaveworks/weave-gitops/pkgs/container/charts%2Fweave-gitops)
- [Git tag](https://github.com/weaveworks/weave-gitops/tags)
- [gitops binaries](https://github.com/weaveworks/weave-gitops/releases)
- updated [Homebrew Tap](https://github.com/weaveworks/homebrew-tap/blob/master/Formula/gitops.rb)
- [npm package](https://github.com/weaveworks/weave-gitops/pkgs/npm/weave-gitops)
- [weave-gitops container image](https://github.com/weaveworks/weave-gitops/pkgs/container/weave-gitops)
