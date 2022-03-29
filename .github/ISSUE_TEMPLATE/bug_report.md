---
name: Bug report
about: Create a report to help us improve
title: ''
labels: 'bug'
assignees: ''

---

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

- [ ] `severity/Critical`: Weave GitOps is crashing or experiencing data loss, the UI is inaccessible or a key feature is unusable. There is no known workaround
- [ ] `severity/Major`: Weave Gitops functionality is broken, there is a workaround, but the workaround requires significant effort
- [ ] `severity/Minor`: Weave Gitops functionality is broken, but there is a fairly straightforward workaround
- [ ] `severity/Low`: Doesn’t affect primary flow/functionality but would be good to fix

**Environment**
 - gitops: [e.g. v0.1.0]
 - How you deployed the Weave GitOps server: [e.g. Tilt, Helm Chart, etc]
 - kubernetes:  [e.g. 1.20.4]
    - [ ] KinD - _version_]
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
