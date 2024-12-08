# Weave GitOps Core API
The API handles operations for Weave GitOps Core

## Version: 0.1

---
## Core

### /v1/child_objects

#### POST
##### Summary

GetChildObjects returns the children of a given object,
specified by a GroupVersionKind.
Not all Kubernets objects have children. For example, a Deployment
has a child ReplicaSet, but a Service has no child objects.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| body | body |  | Yes | [v1GetChildObjectsRequest](#v1getchildobjectsrequest) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1GetChildObjectsResponse](#v1getchildobjectsresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1/crd/is_available

#### GET
##### Summary

IsCRDAvailable returns with a hashmap where the keys are the names of
the clusters, and the value is a boolean indicating whether given CRD is
installed or not on that cluster.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| name | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1IsCRDAvailableResponse](#v1iscrdavailableresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1/events

#### GET
##### Summary

ListEvents returns with a list of events

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| involvedObject.kind | query |  | No | string |
| involvedObject.name | query |  | No | string |
| involvedObject.namespace | query |  | No | string |
| involvedObject.clusterName | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1ListEventsResponse](#v1listeventsresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1/featureflags

#### GET
##### Summary

GetFeatureFlags returns configuration information about the server

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1GetFeatureFlagsResponse](#v1getfeatureflagsresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1/flux_crds

#### GET
##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| clusterName | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1ListFluxCrdsResponse](#v1listfluxcrdsresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1/flux_runtime_objects

#### GET
##### Summary

ListFluxRuntimeObjects lists the flux runtime deployments from a cluster.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| namespace | query |  | No | string |
| clusterName | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1ListFluxRuntimeObjectsResponse](#v1listfluxruntimeobjectsresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1/inventory

#### GET
##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| kind | query |  | No | string |
| name | query |  | No | string |
| namespace | query |  | No | string |
| clusterName | query |  | No | string |
| withChildren | query |  | No | boolean |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1GetInventoryResponse](#v1getinventoryresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1/namespace/flux

#### POST
##### Summary

GetFluxNamespace returns with a namespace with a specific label.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| body | body |  | Yes | [v1GetFluxNamespaceRequest](#v1getfluxnamespacerequest) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1GetFluxNamespaceResponse](#v1getfluxnamespaceresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1/namespaces

#### GET
##### Summary

ListNamespaces returns with the list of available namespaces.

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1ListNamespacesResponse](#v1listnamespacesresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1/object/{name}

#### GET
##### Summary

GetObject gets data about a single primary object from a cluster.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| name | path |  | Yes | string |
| namespace | query |  | No | string |
| kind | query |  | No | string |
| clusterName | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1GetObjectResponse](#v1getobjectresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1/objects

#### POST
##### Summary

ListObjects gets data about primary objects.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| body | body |  | Yes | [v1ListObjectsRequest](#v1listobjectsrequest) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1ListObjectsResponse](#v1listobjectsresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1/policies

#### GET
##### Summary

ListPolicies list policies available on the cluster

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| clusterName | query |  | No | string |
| pagination.pageSize | query |  | No | integer |
| pagination.pageToken | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1ListPoliciesResponse](#v1listpoliciesresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1/policies/{policyName}

#### GET
##### Summary

GetPolicy gets a policy by name

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| policyName | path |  | Yes | string |
| clusterName | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1GetPolicyResponse](#v1getpolicyresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1/policyvalidations

#### POST
##### Summary

ListPolicyValidations lists policy validations

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| body | body |  | Yes | [v1ListPolicyValidationsRequest](#v1listpolicyvalidationsrequest) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1ListPolicyValidationsResponse](#v1listpolicyvalidationsresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1/policyvalidations/{validationId}

#### GET
##### Summary

GetPolicyValidation gets a policy validation by id

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| validationId | path |  | Yes | string |
| clusterName | query |  | No | string |
| validationType | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1GetPolicyValidationResponse](#v1getpolicyvalidationresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1/reconciled_objects

#### POST
##### Summary

GetReconciledObjects returns a list of objects that were created
as a result of reconciling a Flux automation.
This list is derived by looking at the Kustomization or HelmRelease
specified in the request body.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| body | body |  | Yes | [v1GetReconciledObjectsRequest](#v1getreconciledobjectsrequest) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1GetReconciledObjectsResponse](#v1getreconciledobjectsresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1/session_logs

#### POST
##### Summary

GetSessionLogs returns the logs for a given session

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| body | body |  | Yes | [v1GetSessionLogsRequest](#v1getsessionlogsrequest) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1GetSessionLogsResponse](#v1getsessionlogsresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1/suspend

#### POST
##### Summary

ToggleSuspendResource suspends or resumes a flux object.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| body | body |  | Yes | [v1ToggleSuspendResourceRequest](#v1togglesuspendresourcerequest) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1ToggleSuspendResourceResponse](#v1togglesuspendresourceresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1/sync

#### POST
##### Summary

SyncResource forces a reconciliation of a Flux resource

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| body | body |  | Yes | [v1SyncFluxObjectRequest](#v1syncfluxobjectrequest) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1SyncFluxObjectResponse](#v1syncfluxobjectresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1/version

#### GET
##### Summary

GetVersion returns version information about the server

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1GetVersionResponse](#v1getversionresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

---
### Models

#### CrdName

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| plural | string |  | No |
| group | string |  | No |

#### protobufAny

`Any` contains an arbitrary serialized protocol buffer message along with a
URL that describes the type of the serialized message.

Protobuf library provides support to pack/unpack Any values in the form
of utility functions or additional generated methods of the Any type.

Example 1: Pack and unpack a message in C++.

    Foo foo = ...;
    Any any;
    any.PackFrom(foo);
    ...
    if (any.UnpackTo(&foo)) {
      ...
    }

Example 2: Pack and unpack a message in Java.

    Foo foo = ...;
    Any any = Any.pack(foo);
    ...
    if (any.is(Foo.class)) {
      foo = any.unpack(Foo.class);
    }

 Example 3: Pack and unpack a message in Python.

    foo = Foo(...)
    any = Any()
    any.Pack(foo)
    ...
    if any.Is(Foo.DESCRIPTOR):
      any.Unpack(foo)
      ...

 Example 4: Pack and unpack a message in Go

     foo := &pb.Foo{...}
     any, err := anypb.New(foo)
     if err != nil {
       ...
     }
     ...
     foo := &pb.Foo{}
     if err := any.UnmarshalTo(foo); err != nil {
       ...
     }

The pack methods provided by protobuf library will by default use
'type.googleapis.com/full.type.name' as the type URL and the unpack
methods only use the fully qualified type name after the last '/'
in the type URL, for example "foo.bar.com/x/y.z" will yield type
name "y.z".

JSON
====
The JSON representation of an `Any` value uses the regular
representation of the deserialized, embedded message, with an
additional field `@type` which contains the type URL. Example:

    package google.profile;
    message Person {
      string first_name = 1;
      string last_name = 2;
    }

    {
      "@type": "type.googleapis.com/google.profile.Person",
      "firstName": <string>,
      "lastName": <string>
    }

If the embedded message type is well-known and has a custom JSON
representation, that representation will be embedded adding a field
`value` which holds the custom JSON in addition to the `@type`
field. Example (for message [google.protobuf.Duration][]):

    {
      "@type": "type.googleapis.com/google.protobuf.Duration",
      "value": "1.212s"
    }

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| @type | string | A URL/resource name that uniquely identifies the type of the serialized protocol buffer message. This string must contain at least one "/" character. The last segment of the URL's path must represent the fully qualified name of the type (as in `path/google.protobuf.Duration`). The name should be in a canonical form (e.g., leading "." is not accepted).  In practice, teams usually precompile into the binary all types that they expect it to use in the context of Any. However, for URLs which use the scheme `http`, `https`, or no scheme, one can optionally set up a type server that maps type URLs to message definitions as follows:  *If no scheme is provided, `https` is assumed.* An HTTP GET on the URL must yield a [google.protobuf.Type][]   value in binary format, or produce an error. * Applications are allowed to cache lookup results based on the   URL, or have them precompiled into a binary to avoid any   lookup. Therefore, binary compatibility needs to be preserved   on changes to types. (Use versioned type names to manage   breaking changes.)  Note: this functionality is not currently available in the official protobuf release, and it is not used for type URLs beginning with type.googleapis.com.  Schemes other than `http`, `https` (or the empty scheme) might be used with implementation specific semantics. | No |

#### rpcStatus

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| code | integer |  | No |
| message | string |  | No |
| details | [ [protobufAny](#protobufany) ] |  | No |

#### v1ClusterNamespaceList

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| clusterName | string |  | No |
| namespaces | [ string ] |  | No |

#### v1Condition

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| type | string |  | No |
| status | string |  | No |
| reason | string |  | No |
| message | string |  | No |
| timestamp | string |  | No |

#### v1Crd

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| name | [CrdName](#crdname) |  | No |
| version | string |  | No |
| kind | string |  | No |
| clusterName | string |  | No |
| uid | string |  | No |

#### v1Deployment

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| name | string |  | No |
| namespace | string |  | No |
| conditions | [ [v1Condition](#v1condition) ] |  | No |
| images | [ string ] |  | No |
| suspended | boolean |  | No |
| clusterName | string |  | No |
| uid | string |  | No |
| labels | object |  | No |

#### v1Event

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| type | string |  | No |
| reason | string |  | No |
| message | string |  | No |
| timestamp | string |  | No |
| component | string |  | No |
| host | string |  | No |
| name | string |  | No |
| uid | string |  | No |

#### v1GetChildObjectsRequest

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| groupVersionKind | [v1GroupVersionKind](#v1groupversionkind) |  | No |
| namespace | string |  | No |
| parentUid | string |  | No |
| clusterName | string |  | No |

#### v1GetChildObjectsResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| objects | [ [v1Object](#v1object) ] |  | No |

#### v1GetFeatureFlagsResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| flags | object |  | No |

#### v1GetFluxNamespaceRequest

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| v1GetFluxNamespaceRequest | object |  |  |

#### v1GetFluxNamespaceResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| name | string |  | No |

#### v1GetInventoryResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| entries | [ [v1InventoryEntry](#v1inventoryentry) ] |  | No |

#### v1GetObjectResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| object | [v1Object](#v1object) |  | No |

#### v1GetPolicyResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| policy | [v1PolicyObj](#v1policyobj) |  | No |
| clusterName | string |  | No |

#### v1GetPolicyValidationResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| validation | [v1PolicyValidation](#v1policyvalidation) |  | No |

#### v1GetReconciledObjectsRequest

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| automationName | string |  | No |
| namespace | string |  | No |
| automationKind | string |  | No |
| kinds | [ [v1GroupVersionKind](#v1groupversionkind) ] |  | No |
| clusterName | string |  | No |

#### v1GetReconciledObjectsResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| objects | [ [v1Object](#v1object) ] |  | No |

#### v1GetSessionLogsRequest

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| sessionNamespace | string |  | No |
| sessionId | string |  | No |
| token | string |  | No |
| logSourceFilter | string |  | No |
| logLevelFilter | string |  | No |

#### v1GetSessionLogsResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| logs | [ [v1LogEntry](#v1logentry) ] |  | No |
| nextToken | string |  | No |
| error | string |  | No |
| logSources | [ string ] |  | No |

#### v1GetVersionResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| semver | string |  | No |
| commit | string |  | No |
| branch | string |  | No |
| buildTime | string |  | No |
| kubeVersion | string |  | No |

#### v1GroupVersionKind

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| group | string |  | No |
| kind | string |  | No |
| version | string |  | No |

#### v1HealthStatus

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| status | string |  | No |
| message | string |  | No |

#### v1InventoryEntry

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| payload | string |  | No |
| tenant | string |  | No |
| clusterName | string |  | No |
| health | [v1HealthStatus](#v1healthstatus) |  | No |
| children | [ [v1InventoryEntry](#v1inventoryentry) ] |  | No |

#### v1IsCRDAvailableResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| clusters | object |  | No |

#### v1ListError

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| clusterName | string |  | No |
| namespace | string |  | No |
| message | string |  | No |

#### v1ListEventsResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| events | [ [v1Event](#v1event) ] |  | No |

#### v1ListFluxCrdsResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| crds | [ [v1Crd](#v1crd) ] |  | No |
| errors | [ [v1ListError](#v1listerror) ] |  | No |

#### v1ListFluxRuntimeObjectsResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| deployments | [ [v1Deployment](#v1deployment) ] |  | No |
| errors | [ [v1ListError](#v1listerror) ] |  | No |

#### v1ListNamespacesResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| namespaces | [ [v1Namespace](#v1namespace) ] |  | No |

#### v1ListObjectsRequest

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| namespace | string |  | No |
| kind | string |  | No |
| clusterName | string |  | No |
| labels | object |  | No |

#### v1ListObjectsResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| objects | [ [v1Object](#v1object) ] |  | No |
| errors | [ [v1ListError](#v1listerror) ] |  | No |
| searchedNamespaces | [ [v1ClusterNamespaceList](#v1clusternamespacelist) ] |  | No |

#### v1ListPoliciesResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| policies | [ [v1PolicyObj](#v1policyobj) ] |  | No |
| total | integer |  | No |
| nextPageToken | string |  | No |
| errors | [ [v1ListError](#v1listerror) ] |  | No |

#### v1ListPolicyValidationsRequest

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| clusterName | string |  | No |
| pagination | [v1Pagination](#v1pagination) |  | No |
| application | string |  | No |
| namespace | string |  | No |
| kind | string |  | No |
| policyId | string |  | No |
| validationType | string |  | No |

#### v1ListPolicyValidationsResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| violations | [ [v1PolicyValidation](#v1policyvalidation) ] |  | No |
| total | integer |  | No |
| nextPageToken | string |  | No |
| errors | [ [v1ListError](#v1listerror) ] |  | No |

#### v1LogEntry

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| timestamp | string |  | No |
| source | string |  | No |
| level | string |  | No |
| message | string |  | No |
| sortingKey | string |  | No |

#### v1Namespace

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| name | string |  | No |
| status | string |  | No |
| annotations | object |  | No |
| labels | object |  | No |
| clusterName | string |  | No |

#### v1Object

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| payload | string |  | No |
| clusterName | string |  | No |
| tenant | string |  | No |
| uid | string |  | No |
| inventory | [ [v1GroupVersionKind](#v1groupversionkind) ] |  | No |
| info | string |  | No |
| health | [v1HealthStatus](#v1healthstatus) |  | No |

#### v1ObjectRef

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| kind | string |  | No |
| name | string |  | No |
| namespace | string |  | No |
| clusterName | string |  | No |

#### v1Pagination

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| pageSize | integer |  | No |
| pageToken | string |  | No |

#### v1PolicyObj

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| name | string |  | No |
| id | string |  | No |
| code | string |  | No |
| description | string |  | No |
| howToSolve | string |  | No |
| category | string |  | No |
| tags | [ string ] |  | No |
| severity | string |  | No |
| standards | [ [v1PolicyStandard](#v1policystandard) ] |  | No |
| gitCommit | string |  | No |
| parameters | [ [v1PolicyParam](#v1policyparam) ] |  | No |
| targets | [v1PolicyTargets](#v1policytargets) |  | No |
| createdAt | string |  | No |
| clusterName | string |  | No |
| tenant | string |  | No |
| modes | [ string ] |  | No |

#### v1PolicyParam

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| name | string |  | No |
| type | string |  | No |
| value | [protobufAny](#protobufany) |  | No |
| required | boolean |  | No |

#### v1PolicyStandard

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| controls | [ string ] |  | No |

#### v1PolicyTargetLabel

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| values | object |  | No |

#### v1PolicyTargets

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| kinds | [ string ] |  | No |
| labels | [ [v1PolicyTargetLabel](#v1policytargetlabel) ] |  | No |
| namespaces | [ string ] |  | No |

#### v1PolicyValidation

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| message | string |  | No |
| clusterId | string |  | No |
| category | string |  | No |
| severity | string |  | No |
| createdAt | string |  | No |
| entity | string |  | No |
| entityKind | string |  | No |
| namespace | string |  | No |
| violatingEntity | string |  | No |
| description | string |  | No |
| howToSolve | string |  | No |
| name | string |  | No |
| clusterName | string |  | No |
| occurrences | [ [v1PolicyValidationOccurrence](#v1policyvalidationoccurrence) ] |  | No |
| policyId | string |  | No |
| parameters | [ [v1PolicyValidationParam](#v1policyvalidationparam) ] |  | No |

#### v1PolicyValidationOccurrence

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| message | string |  | No |

#### v1PolicyValidationParam

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| name | string |  | No |
| type | string |  | No |
| value | [protobufAny](#protobufany) |  | No |
| required | boolean |  | No |
| configRef | string |  | No |

#### v1SyncFluxObjectRequest

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| objects | [ [v1ObjectRef](#v1objectref) ] |  | No |
| withSource | boolean |  | No |

#### v1SyncFluxObjectResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| v1SyncFluxObjectResponse | object |  |  |

#### v1ToggleSuspendResourceRequest

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| objects | [ [v1ObjectRef](#v1objectref) ] |  | No |
| suspend | boolean |  | No |

#### v1ToggleSuspendResourceResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| v1ToggleSuspendResourceResponse | object |  |  |
