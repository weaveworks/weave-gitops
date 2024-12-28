---
title: Manual Approval for Progressive Delivery Deployments

---

# Manual Approval for Progressive Delivery Deployments ~ENTERPRISE~

To help you understand the state of progressive delivery updates to your applications, Weave GitOps Enterprise uses [Flagger](https://flagger.app)—part of the Flux family of open source projects. WGE's Delivery view shows all of your deployed `Canary` objects and rollout progress.

By default, Flagger automatically promotes a new version of an application whenever it passes the defined checks of an analysis phase. However, you can also configure [webhooks](https://docs.flagger.app/usage/webhooks) to enable manual approvals of rollout stages.

This guide shows you how to manually gate a progressive delivery promotion with Flagger by using the in-built load tester.

## Prerequisites
- Basic knowledge of [Flagger](https://flagger.app)
- An existing `Canary` object and target deployment
- Flagger's load tester [installed](https://docs.flagger.app/usage/webhooks#load-testing)

## Basic Introduction to Webhooks and Gating

You can configure Flagger to work with several types of hooks that will be called at 
given stages during a progressive delivery rollout. Some of these hooks allow you to manually 
gate whether a rollout proceeds at certain points:
- Before scaling up a new deployment and canary analysis begins with `confirm-rollout`. 
- Before increasing traffic weight with `confirm-traffic-increase`.
- Before promoting a new version after successful canary analysis with `confirm-promotion`.

Any URL can serve as a webhook target. It will approve if a `200 OK` status code is returned, and halt if `403 Forbidden`.

The webhook will receive a JSON payload that can be unmarshaled as
`CanaryWebhookPayload`:

```go
type CanaryWebhookPayload struct {
	// Name of the canary
	Name string `json:"name"`

	// Namespace of the canary
	Namespace string `json:"namespace"`

	// Phase of the canary analysis
	Phase CanaryPhase `json:"phase"`

	// Metadata (key-value pairs) for this webhook
	Metadata map[string]string `json:"metadata,omitempty"`
}
```

The [Flagger documentation](https://docs.flagger.app/usage/webhooks) provides more information about webhooks.

## Use Flagger's Load Tester to Manually Gate a Promotion

To enable manual approval of a promotion, configure the 
`confirm-promotion` webhook. This will call a particular gate provided through 
Flagger's load tester, and is an easy way to experiment using Flagger's included components. 

!!! tip
    We strongly recommend that you DO NOT USE the load tester for manual gating in a production environment. It lacks auth, so anyone with cluster access could open and close it. It also lacks storage, so all gates would close upon a restart. Instead, configure these webhooks for appropriate integration with a tool of your choice, such Jira, Slack, Jenkins, etc.

### Configure the `confirm-promotion` Webhook

In your canary object, add the following in the `analysis` section:

```yaml
  analysis:
    webhooks:
      - name: "ask for confirmation"
        type: confirm-promotion
        url: http://flagger-loadtester.test/gate/check
```

This gate is closed by default.

### Deploy a New Version of Your Application

Trigger a Canary rollout by updating your target deployment/daemonset—for 
example, by bumping the container image tag. A full list of ways to trigger 
a rollout is available [here](https://docs.flagger.app/faq#how-to-retry-a-failed-release).

Weave GitOps Enterprise (WGE)'s Applications > Delivery view enables you to watch the progression of a canary:

![Podinfo Canary progressing](/img/pd-table-progressing.png)

### Wait for the Canary Analysis to Complete

Once the canary analysis has successfully completed, Flagger will call the 
`confirm-promotion` webhook and change status to `WaitingPromotion`:

![Podinfo Canary showing Waiting Promotion - table view](/img/pd-table-waiting.png)

![Podinfo Canary showing Waiting Promotion - details view](/img/pd-details-waiting.png)

### Open the Gate

To open the gate and confirm that you approve promotion of the new 
version of your application, exec into the load tester 
container:

```
$ kubectl -n test exec -it flagger-loadtester-xxxx-xxxx sh

# to open
> curl -d '{"name": "app","namespace":"test"}' http://localhost:8080/gate/open
```

Flagger will now promote the canary version to the primary and 
complete the progressive delivery rollout. :tada:

![Podinfo Canary succeeded - full events history](/img/pd-events-gate-passed.png)

![Podinfo Canary succeeded - promoting](/img/pd-table-promoting.png)

![Podinfo Canary succeeded - promoted](/img/pd-table-succeeded.png)


To manually close the gate again, issue this command:

```
> curl -d '{"name": "app","namespace":"test"}' http://localhost:8080/gate/close
```

**References:**

* The [Official Flagger documentation](https://docs.flagger.app/usage/webhooks#manual-gating) informs this guide.
