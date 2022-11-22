---
title: Feedback and Telemetry
sidebar_position: 8
hide_title: true
---

## Anonymous Aggregate User Behavior Analytics



### Why are we collecting this data?

We want to ensure that we are designing the best features, addressing the most pressing bugs, and prioritizing our roadmap appropriately for our users. Collecting analytics on our users’ behaviors gives us valuable insights and allows us to conduct analyses on user behavior within the product. This is important for us so we can make informed decisions- based on how, where and when our users use Weave GitOps - and prioritize what is most important to users like you.

#### For example:

### How long is the collected data stored?


### What are we collecting?


### When is the data collected and where is it sent?

### How?

### Opting out

#### CLI

#### Dashboard



## Feedback

We ❤️ your comments and suggestions as we look to make successfully adopting a cloud-native approach, to application deployment on Kubernetes with GitOps, easier and easier. There are a number of ways you can reach out:

- Raise an [issue](https://github.com/weaveworks/weave-gitops/issues)
- Invite yourself to the <a href="https://slack.weave.works/" target="_blank">Weave Users Slack</a>.
- Chat to us on the [#weave-gitops](https://weave-community.slack.com/messages/weave-gitops/) slack channel.
- Set up time with one of our team: [David](https://calendly.com/david-harris-weaveworks) - Product Manager (UK) or [James](https://calendly.com/james-weave-works/product-interview) - Product Director (US - East Coast)
- Come along to one of our [events](https://www.meetup.com/Weave-User-Group/)

## Telemetry

To help us understand how we can improve your experience with Weave GitOps, and prioritise enhancements, we would like to collect anonymised usage data. Currently, only the `gitops` CLI has any notion of telemetry, however we would like to expand this to Weave GitOps in the future.

### gitops CLI

No personally identifiable information is collected, we use [https://github.com/weaveworks/go-checkpoint](https://github.com/weaveworks/go-checkpoint) an implementation based on [https://checkpoint.hashicorp.com/](https://checkpoint.hashicorp.com/) to notify users of newly available updates, as well as collecting basic CLI metrics, up to 2 verbs, without any flags or user provided information.

For example the command: `gitops add cluster --from-template <template-name> --set key=val --dry-run`
would report the following: `gitops add cluster` alongside:

- OS/Arch - for example, darwin
- Version of gitops - for example, 0.6.2-RC1
- Whether the version of gitops is a release candidate or full release, yes/no
- A signature, when possible to derive from system uuid, to determine a non-identifiable (based on all other data) unique user.

You can opt-out at any time by issuing:

```cli
export CHECKPOINT_DISABLE=1
```

Weaveworks privacy policy is available [here](https://www.weave.works/weaveworks-privacy-policy/).
