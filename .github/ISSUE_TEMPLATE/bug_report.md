---
name: Bug report
about: Create a report to help us improve
title: ''
labels: 'bug'
assignees: ''

---

<!--
ATTENTION WEAVEWORKS EMPLOYEES!!
If you are NOT using kind/some other cluster of your own,
check out the extra docs for submitting bug reports here:
https://github.com/weaveworks/weave-gitops-clusters/blob/main/gke-cluster/README.md#reporting-bugs-found-the-gke-cluster

Please read them :pray:
-->

**Describe the bug**
<!--
A clear and concise description of what the bug is.
-->

**Is this a UI bug or a server bug?**
- [ ] UI
- [ ] Server

**What is the severity of the bug**
<!--
Please select a label for this ticket indicating the severity of this bug.
A maintainer will add the label to the issue, taking your suggestion into account.
-->

- [ ] `severity/Critical`: Weave GitOps is crashing, the UI is inaccessible or a key feature is unusable. There is no known workaround.
- [ ] `severity/Major`: Weave Gitops functionality or a key feature is broken. There is a workaround, but the workaround requires significant effort.
- [ ] `severity/Minor`: A non-key feature or functionality is broken. There is a fairly straightforward workaround.
- [ ] `severity/Low`: Doesn’t affect primary flow/functionality but would be good to fix.

**Environment**
 - Weave-Gitops Version, commit sha or image tag: [e.g. `v0.1.0`, `3499ba2a`, or `1651069883-3499ba2a2c78a7c04c47c56c58eb77eede4b21f4`]
    + This can be found by going to `/v1/version` on your server. 
 - How you deployed the Weave GitOps server: [e.g. Tilt, Helm Chart, etc]
 - Kubernetes version:  [e.g. 1.20.4]
 - Where are you running your cluster?
    - [ ] KinD - _version_
    - [ ] k3s - _version_
    - [ ] cloud [e.g., EKS, AKS]  _version_
    - [ ] other - _name_ _version_
 - Browser + version: [e.g. chrome 74, safari 12, firefox 87]

**To Reproduce**
Steps to reproduce the behavior:
<!--
Eg:
1. Go to '...'
2. Click on '....'
3. Scroll down to '....'
4. See error
-->

**Expected behavior**
<!--
A clear and concise description of what you expected to happen.
-->

**Config and Logs**
If applicable, add logs to help explain your problem. _please compress the output before attaching_
- [ ] The yaml of the object you are viewing
- [ ] Logs from the `wego-app` pod
- [ ] Events from `flux-system` namespace (Or the namespace you deployed flux and/or Weave GitOps)
- [ ] `kubectl cluster-info dump`
- [ ] Prometheus alerts
- [ ] Flux logs

**Screenshots**
<!--
If applicable, add screenshots to help explain your problem.
-->

**Additional context**
<!--
Add any other context about the problem here.
-->
