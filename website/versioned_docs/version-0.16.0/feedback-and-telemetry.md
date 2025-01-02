---
title: Feedback and Telemetry
hide_title: true
---

## Feedback

We ❤️ your comments and suggestions as we look to make successfully adopting a cloud-native approach, to application deployment on Kubernetes with GitOps, easier and easier. There are a number of ways you can reach out:

- Raise an [issue](https://github.com/weaveworks/weave-gitops/issues)
- Invite yourself to the <a href="https://slack.weave.works/" target="_blank">Weave Users Slack</a>.
- Chat to us on the [#weave-gitops](https://weave-community.slack.com/messages/weave-gitops/) slack channel.
- Set up time with one of our team: [Charles](https://calendly.com/casibbald/30min) - Community Support
- Come along to one of our [events](https://www.meetup.com/Weave-User-Group/)

## Anonymous Aggregate User Behavior Analytics

Weaveworks is utilizing [Pendo](https://www.pendo.io/), a product-analytics app,  to gather anonymous user behavior analytics for both Weave GitOps and Weave GitOps Enterprise. We use this data so we can understand what you love about Weave GitOps, and areas we can improve.

Weave GitOps OSS users will be notified when you create the dashboard for the first time via **gitops create dashboard** or when you use **gitops run** for the first time and decide to install the dashboard via that functionality. Analytics will not be enabled until <u>after</u> this notification so that you can opt out before sending analytics data.

For Weave GitOps Enterprise users, this functionality is turned on by default. Further below we go into more detail about how you can control this functionality.

### Why are we collecting this data?

We want to ensure that we are designing the best features, addressing the most pressing bugs, and prioritizing our roadmap appropriately for our users. Collecting analytics on our users’ behaviors gives us valuable insights and allows us to conduct analyses on user behavior within the product. This is important for us so we can make informed decisions- based on how, where and when our users use Weave GitOps - and prioritize what is most important to users like you.

#### For example:

We’d like to understand the usage of the graph and dependency tabs within the dashboard. If users are utilizing this feature, we would like to understand the value and how we can improve that feature. However, if users aren’t using it, we can conduct research to understand why and either fix it, or come to the conclusion that it really doesn’t serve any utility and focus our efforts on more valuable features.

### How long is the collected data stored?

Weave GitOps’s anonymous user and event data has a 24 month retention policy. The default value for data retention in Pendo is 7 years. For more information on Pendo’s data storage policies, [click here](https://support.pendo.io/hc/en-us/articles/360051268732-Subscription-Data-Retention-Limit).

### What are we collecting?

Weave GitOps gathers data on how the CLI and Web UI are used. There is no way for us or Pendo to connect our IDs to individual users or sites.

For the CLI, we gather usage data  on:
- The specific sub command itself - e.g. `gitops get bcrypt-hash`
- The name of the flags used, without the value (e.g. `--password`, but not the value)
- A random string used as an anonymous user ID, stored on your machine
- - **Note: <u>We have no way of tracking individual users.</u>** We can only distinguish between user counts and event counts
- Whether the user has installed the Enterprise or open-source version of the CLI
- A value of `app=cli`, to know it’s a CLI metric

For the Web UI, we gather usage data  on:
- Your browser, version, and user agent
- The domain name of your server
- Every page interaction, and the time each page is left open
- All button interactions
- The complete URL of every page, including which resource you look at, and searches done
- We can push new content into your browser, to add questions, guides, or more data points
- We send a unique user hash, based on your user name
- - **Note: <u>We are not able to cross-reference unique users</u>** between here and anywhere else - not even your command line - but it gives us the ability to distinguish between user counts and event counts.
- Finally, we include a persistent ID representing your cluster, based on a hash of your `kube-system` namespace uuid
- - **Note: <u>There is no way for us to track individual clusters</u>** using this, but it gives us the ability to distinguish between cluster counts and event counts.

### When is the data collected and where is it sent?

Weave GitOps CLI analytics are sent at startup. The dashboard analytics are sent through its execution. Both CLI and Dashboard analytics are sent to Pendo over HTTPS.

### How?

The CLI code is viewable in pkg/analytics. It will ignore any errors, e.g. if you don’t have any network connection.

The dashboard setup code is viewable in ui/components/Pendo.tsx - this will fetch a 3rd party javascript from Pendo’s servers.

### Opting out

All the data collected, analytics, and feedback are for the sole purpose of creating better product experience for you and your teams. We would really appreciate it if you left the analytics on as it helps us prioritize which features to build next and what features to improve. However, if you do want to opt out of Weave GitOps’s analytics you can opt out of CLI and/or Dashboard analytics.

#### CLI

We have created a command to make it easy to turn analytics on or off for the CLI.

**To disable analytics:**
*gitops set config analytics false*

**To enable analytics:**
*gitops set config analytics true*

#### Dashboard

You need to update your helm release to remove `WEAVE_GITOPS_FEATURE_TELEMETRY` from the `envVars` value.
