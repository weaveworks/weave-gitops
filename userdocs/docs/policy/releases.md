---
title: Profile Releases

---



# Profile Releases ~ENTERPRISE~


## v0.6.5

### Highlights

- **Agent**
  - Add support for mutating violating resource.

### Dependency Versions

- Policy Agent v2.2.0

### Policy Library Compatibility

Compatible with Policy Library versions:

- v1.2.0

Needs this [migration steps](./policy-set.md#migration-from-v2beta1-to-v2beta2) to be compatible with the following versions:

- v1.1.0
- v1.0.0
- v0.4.0


## v0.6.4

### Highlights
- **Agent**
  - Add PolicyConfig CRD to make it possible to customize policy configuration per namespaces, applications or resources
  - Add mode field to policy set and add policy modes to its status
  - Add policy modes to labels to support filtering
  - Support backward compatibility for policy version v2beta1

### Dependency Versions

- Policy Agent v2.0.0

### Policy Library Compatibility

Compatible with Policy Library versions:

- v1.2.0

Needs this [migration steps](./policy-set.md#migration-from-v2beta1-to-v2beta2) to be compatible with the following versions:

- v1.1.0
- v1.0.0
- v0.4.0


## v0.6.3

### Highlights
- **Agent**
  - Reference flux objects in violations events instead of the original resource object to be able to list specific flux application violations

### Dependency Versions

- policy-agent 1.2.1

### Policy Library Compatibility

- v0.4.0
- v1.0.0
- v1.1.0

## v0.6.2

### Highlights
- **Agent**
  - Add Terraform mode to allow validating terraform plans
  - Support targeting kubernetes HPA resources

### Dependency Versions

- policy-agent 1.2.0

### Policy Library Compatibility

- v0.4.0
- v1.0.0
- v1.1.0

While both v.0.4.0 and v1.0.0 are compatible with the agent. Only v1.1.0 includes the modification needed to make Controller Minimum Replica Count policy with with `horizontalpodautoscalers`

## v0.6.1

### Highlights
- **Agent**
  - Make the audit interval configurable through `config.audit.interval`. It defaults to 24 hours.
  - Add support for targeting certain flux resources (kustomizations, helmreleases and ocirepositories) in the admission mode.
- **Profile**
  - Add the ability to use an existing GitSource instead of creating a new one.


### Dependency Versions

- policy-agent 1.1.0

### Policy Library Compatibility

- v0.4.0
- v1.0.0

## v0.6.0

### Highlights
- **Agent**
  - Configure the agent through a configuration file instead of arguments.
  - Allow defining different validation sinks for audit and admission modes.
  - Add the PolicySet CRD to the hem chart.
- **Profile**
  - Disable the default policy source.

### Dependency Versions

- policy-agent 1.0.0

### Policy Library Compatibility

- v0.4.0
- v1.0.0
