---
title: Feedback and Telemetry
sidebar_position: 8
hide_title: true
---

## Feedback

We ❤️ your comments and suggestions as we look to make successfully adopting a cloud-native approach, to application deployment on Kubernetes with GitOps, easier and easier. There are a number of ways you can reach out:

- Raise an [issue](https://github.com/weaveworks/weave-gitops/issues)
- Chat to us in [Slack](https://app.slack.com/client/T2NDH1D9D/C0248LVC719/thread/C2ND76PAA-1621532937.019800)
- Set up time with one of our team, [David](https://calendly.com/david-harris-weaveworks) - Product Manager (UK) or [James](https://calendly.com/james-weave-works/product-interview) - Engineering Manager (US - East Coast)
- Come along to one of our [events](https://www.meetup.com/Weave-User-Group/)

## Telemetry

To help us understand how we can improve your experience with Weave GitOps, and prioritise enhancements, we would like to collect anonymised usage data. Currently, only the `gitops` CLI has any notion of telemetry, however we would like to expand this to Weave GitOps in the future.

### gitops CLI
No personally identifiable information is collected, we use [https://github.com/weaveworks/go-checkpoint](https://github.com/weaveworks/go-checkpoint) an implementation based on [https://checkpoint.hashicorp.com/](https://checkpoint.hashicorp.com/) to notify users of newly available updates as well as basic CLI metrics, up to 2 verbs, without any flags or user provided information.

For example the command: `gitops add cluster --from-template <template-name> --set key=val --dry-run` 
Would report the following: `gitops add cluster` alongside:
- OS/Arch - for example, darwin
- Version of gitops - for example, 0.6.2-RC1
- Whether the version of gitops is a release candidate or full release, yes/no
- A signature, when possible to derive from system uuid, to determine a non-identifiable (based on all other data) unique user. 

You can opt-out at any time by issuing:

```
export CHECKPOINT_DISABLE=1
```

Weaveworks privacy policy is available [here](https://www.weave.works/weaveworks-privacy-policy/).