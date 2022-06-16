# 0014. Kubernetes Client API Fairness and Rate Limits

## Status

Proposed

## Problem

The UI part of Weave GitOps Core is an interactive Web application, which pulls the Kubernetes cluster periodically.
The QPS configuration of the Kubernetes Client API is for controlling the maximum query per second to the API server from a client, which is Weave GitOps Core in this case.
Burst configuration is the maximum burst for throttle. When QPS is not specified, the default value is 5 query per second, and the default Burst value is 10.

Before Kubernetes 1.23, QPS and Burst must be specified only from the client side.
From Kubernetes 1.24+, the server can configure these values for the client if the `APIPriorityAndFairness` feature gate is enabled for Kubernetes API Servers.

The current QPS in Weave GitOps Core is 1,000, which is very high compared to what Flux has (50).
The current Burst in Weave GitOps Core is 2,000, which is also very high compared to what Flux has (100).
Without properly setting these values, Weave GitOps Core may be unknowingly hammering the Kubernetes clusters and causing connection errors.

## Decision

The purpose of this decision is to
  1. Set a proper value of QPS for Weave GitOps Core to work with Kubernetes 1.23. (QPS=50)
  2. Set a proper value of Burst for Weave GitOps Core to work with Kubernetes 1.23. (Burst=100)
  3. Use the `APIPriorityAndFairness` feature if enabled for Kubernetes 1.24+.

## Consequences

Consequences should be minimal as QPS=50, and Burst=100 should be enough for the UI of Weave GitOps Core to still be responsive when working against Kubernetes <= 1.23.
For Kubernetes 1.24+, Weave GitOps Core's QPS and Burst will be capped from the server-side configuration if the feature is enabled.
An additional consequence is that there will be a need to monitor and optimize the number of query against the Kubernetes API servers.
