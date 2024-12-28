---
title: PolicySet

---



# PolicySet ~ENTERPRISE~

This is an optional custom resource that is used to select a group of policies to work in specific [modes](./weave-policy-profile.md#agent-modes).

In each mode, the agent will list all the PolicySets of this mode and check which policies match any of those policysets, then validate the resources against them.

If there are no PolicySets found for a certain mode, all policies will be applied during this mode.

> Note: [Tenant Policies](./policy.md#tenant-policy) is always active in the [Admission](#admission) mode, event if it is not selected in the `admission` policysets

**Example**
```yaml
apiVersion: pac.weave.works/v2beta2
kind: PolicySet
metadata:
  name: my-policy-set
spec:
  mode: admission
  filters:
    ids:
      - weave.policies.containers-minimum-replica-count
    categories:
      - security
    severities:
      - high
      - medium
    standards:
      - pci-dss
    tags:
      - tag-1
```

PolicySets can be created for any of the three modes supported by the agent: `admission`, `audit`, and `tfAdmission`.


## Grouping Policies

Policies can be grouped by their ids, categories, severities, standards and tags

The policy will be applied if any of the filters are matched.


## Migration from v2beta1 to v2beta2

### New fields
- New required field `spec.mode` is added. PolicySets should be updated to set the mode

Previously the agent was configured with which policysets to use in each mode. Now we removed this argument from the agent's configuration and add the mode to the Policyset itself.

#### Example of the agent configuration in versions older than v2.0.0

```yaml
# config.yaml
admission:
   enabled: true
   policySet: admission-policy-set
   sinks:
      filesystemSink:
         fileName: admission.txt
```

#### Example of current PolicySet with mode field

```yaml
apiVersion: pac.weave.works/v2beta2
kind: PolicySet
metadata:
  name: admission-policy-set
spec:
  mode: admission
  filters:
    ids:
      - weave.policies.containers-minimum-replica-count
```


### Updated fields
- Field `spec.name` became optional.

### Deprecate fields
- Field `spec.id` is deprecated.
