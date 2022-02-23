---
name: Bug report
about: Create a report to help us improve
title: ''
labels: 'bug'
assignees: ''

---

**Describe the bug**
A clear and concise description of what the bug is.

**What is the severity of the bug**
Please add a label for this ticket indicating the severity of this bug.

| Label | Description |
| --- | --- |
| severity/Critical | Weave GitOps is crashing or experiencing data loss, and there is no known workaround |
| severity/Major | Weave Gitops functionality is broken, there is a workaround, and the workaround requires significant effort |
| severity/Minor | Weave Gitops functionality is broken, but there is a workaround |
| severity/Low | Doesn’t affect primary flow/functionality but would be good to fix |



**Environment**
 - gitops: [e.g. v0.1.0]
 - kubernetes:  [e.g. 1.20.4]
    - [ ] KinD - _version_]
    - [ ] k3s - _version_
    - [ ] cloud [e.g., EKS, AKS]  _version_
    - [ ] other - _name_ _version_
 - Browser + version: [e.g. chrome 74, safari 12, firefox 87]
 -
**Affects versions**

**To Reproduce**
Steps to reproduce the behavior:
1. Go to '...'
2. Click on '....'
3. Scroll down to '....'
4. See error

**Expected behavior**
A clear and concise description of what you expected to happen.

**Config and Logs**
If applicable, add logs to help explain your problem. _please compress the output before attaching_
- [ ] gitops config
- [ ] `gitops` cli call output (and parameters)
- [ ] gitops-controller logs
- [ ] Events from wego-* namespaces
- [ ] `kubectl cluster-info dump`
- [ ] Prometheus alerts
- [ ] Flux logs

**Screenshots**
If applicable, add screenshots to help explain your problem.

**Additional context**
Add any other context about the problem here.
