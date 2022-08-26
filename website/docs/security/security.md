---
title: Security
description: weave gitops security page. find information about vulnerabilities and others.
hide_title: true
---

# Weave Gitops Security

This document defines security reporting, handling, disclosure, and audit information for Weave Gitops.

## Security Process

### Report a Vulnerability

- To make a report please email the private security list at <security@weave.works> with the details.
  We ask that reporters act in good faith by not disclosing the issue to others.
- You may, but are not required to, encrypt your email to this list using the PGP keys of Security Team members, listed below.
- The Security Team will fix the issue as soon as possible and coordinate a release date with you.
- You will be able to choose if you want public acknowledgement of your effort and how you would like to be credited.

### Handling

- All reports are thoroughly investigated by the Security Team.
- Any vulnerability information shared with the Security Team will not be shared with others unless it is necessary to fix the issue.
  Information is shared only on a need to know basis.
- As the security issue moves through the identification and resolution process, the reporter will be notified.
- Additional questions about the vulnerability may also be asked of the reporter.

### Disclosures

Vulnerability disclosures announced publicly.
Disclosures will contain an overview, details about the vulnerability, a fix that will typically be an update, 
and optionally a workaround if one is available.

We will coordinate publishing disclosures and security releases in a way that is realistic and necessary for end users.
We prefer to fully disclose the vulnerability as soon as possible once a user mitigation is available.
Disclosures will always be published in a timely manner after a release is published that fixes the vulnerability.

## Advisories

Here is an overview of all our published security advisories.

### Weave Gitops OSS

Date | CVE | Tittle                                                                                                                                                   | Severity | Affected version(s) | 
---- | -- |----------------------------------------------------------------------------------------------------------------------------------------------------------|-----------| ------------------- | 
2022-06-23 | CVE-2022-31098| [Weave GitOps leaked cluster credentials into logs on connection errors](https://github.com/advisories/GHSA-xggc-qprg-x6mw) | Critical  | <= 0.8.1-rc.5| 


### Weave Gitops Enterprise

Date | CVE | Tittle                                                                                                                                                   | Severity | Affected version(s) | 
---- | -- |----------------------------------------------------------------------------------------------------------------------------------------------------------|-----------| ------------------- | 
2022-08-27 | TBA | [Malicious links can be crafted by users and shown in the UI](cve/enterprise/GHSA-378h-hpjp-gxfc) | Critical  | < v0.9.0-rc.5|

