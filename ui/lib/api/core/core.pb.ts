/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "../../fetch.pb"
import * as GoogleProtobufAny from "../../google/protobuf/any.pb"
import * as Gitops_coreV1Types from "./types.pb"
export type GetInventoryRequest = {
  kind?: string
  name?: string
  namespace?: string
  clusterName?: string
  withChildren?: boolean
}

export type GetInventoryResponse = {
  entries?: Gitops_coreV1Types.InventoryEntry[]
}

export type PolicyValidation = {
  id?: string
  message?: string
  clusterId?: string
  category?: string
  severity?: string
  createdAt?: string
  entity?: string
  namespace?: string
  violatingEntity?: string
  description?: string
  howToSolve?: string
  name?: string
  clusterName?: string
  occurrences?: PolicyValidationOccurrence[]
  policyId?: string
  parameters?: PolicyValidationParam[]
}

export type ListPolicyValidationsRequest = {
  clusterName?: string
  pagination?: Pagination
  application?: string
  namespace?: string
  kind?: string
  policyId?: string
  validationType?: string
}

export type ListPolicyValidationsResponse = {
  violations?: PolicyValidation[]
  total?: number
  nextPageToken?: string
  errors?: ListError[]
}

export type GetPolicyValidationRequest = {
  validationId?: string
  clusterName?: string
  validationType?: string
}

export type GetPolicyValidationResponse = {
  validation?: PolicyValidation
}

export type PolicyValidationOccurrence = {
  message?: string
}

export type PolicyValidationParam = {
  name?: string
  type?: string
  value?: GoogleProtobufAny.Any
  required?: boolean
  configRef?: string
}

export type PolicyParamRepeatedString = {
  value?: string[]
}

export type Pagination = {
  pageSize?: number
  pageToken?: string
}

export type ListError = {
  clusterName?: string
  namespace?: string
  message?: string
}

export type ListFluxRuntimeObjectsRequest = {
  namespace?: string
  clusterName?: string
}

export type ListFluxRuntimeObjectsResponse = {
  deployments?: Gitops_coreV1Types.Deployment[]
  errors?: ListError[]
}

export type ListFluxCrdsRequest = {
  clusterName?: string
}

export type ListFluxCrdsResponse = {
  crds?: Gitops_coreV1Types.Crd[]
  errors?: ListError[]
}

export type GetObjectRequest = {
  name?: string
  namespace?: string
  kind?: string
  clusterName?: string
}

export type GetObjectResponse = {
  object?: Gitops_coreV1Types.Object
}

export type ListObjectsRequest = {
  namespace?: string
  kind?: string
  clusterName?: string
  labels?: {[key: string]: string}
}

export type NamespaceList = {
  namespaces?: string[]
}

export type ListObjectsResponse = {
  objects?: Gitops_coreV1Types.Object[]
  errors?: ListError[]
  searchedNamespaces?: {[key: string]: NamespaceList}
}

export type GetReconciledObjectsRequest = {
  automationName?: string
  namespace?: string
  automationKind?: string
  kinds?: Gitops_coreV1Types.GroupVersionKind[]
  clusterName?: string
}

export type GetReconciledObjectsResponse = {
  objects?: Gitops_coreV1Types.Object[]
}

export type GetChildObjectsRequest = {
  groupVersionKind?: Gitops_coreV1Types.GroupVersionKind
  namespace?: string
  parentUid?: string
  clusterName?: string
}

export type GetChildObjectsResponse = {
  objects?: Gitops_coreV1Types.Object[]
}

export type GetFluxNamespaceRequest = {
}

export type GetFluxNamespaceResponse = {
  name?: string
}

export type ListNamespacesRequest = {
}

export type ListNamespacesResponse = {
  namespaces?: Gitops_coreV1Types.Namespace[]
}

export type ListEventsRequest = {
  involvedObject?: Gitops_coreV1Types.ObjectRef
}

export type ListEventsResponse = {
  events?: Gitops_coreV1Types.Event[]
}

export type SyncFluxObjectRequest = {
  objects?: Gitops_coreV1Types.ObjectRef[]
  withSource?: boolean
}

export type SyncFluxObjectResponse = {
}

export type GetVersionRequest = {
}

export type GetVersionResponse = {
  semver?: string
  commit?: string
  branch?: string
  buildTime?: string
  fluxVersion?: string
  kubeVersion?: string
}

export type GetFeatureFlagsRequest = {
}

export type GetFeatureFlagsResponse = {
  flags?: {[key: string]: string}
}

export type ToggleSuspendResourceRequest = {
  objects?: Gitops_coreV1Types.ObjectRef[]
  suspend?: boolean
}

export type ToggleSuspendResourceResponse = {
}

export type GetSessionLogsRequest = {
  sessionNamespace?: string
  sessionId?: string
  token?: string
  logSourceFilter?: string
  logLevelFilter?: string
}

export type LogEntry = {
  timestamp?: string
  source?: string
  level?: string
  message?: string
  sortingKey?: string
}

export type GetSessionLogsResponse = {
  logs?: LogEntry[]
  nextToken?: string
  error?: string
  logSources?: string[]
}

export type IsCRDAvailableRequest = {
  name?: string
}

export type IsCRDAvailableResponse = {
  clusters?: {[key: string]: boolean}
}

export type ListPoliciesRequest = {
  clusterName?: string
  pagination?: Pagination
}

export type ListPoliciesResponse = {
  policies?: PolicyObj[]
  total?: number
  nextPageToken?: string
  errors?: ListError[]
}

export type GetPolicyRequest = {
  policyName?: string
  clusterName?: string
}

export type GetPolicyResponse = {
  policy?: PolicyObj
  clusterName?: string
}

export type PolicyObj = {
  name?: string
  id?: string
  code?: string
  description?: string
  howToSolve?: string
  category?: string
  tags?: string[]
  severity?: string
  standards?: PolicyStandard[]
  gitCommit?: string
  parameters?: PolicyParam[]
  targets?: PolicyTargets
  createdAt?: string
  clusterName?: string
  tenant?: string
  modes?: string[]
}

export type PolicyStandard = {
  id?: string
  controls?: string[]
}

export type PolicyParam = {
  name?: string
  type?: string
  value?: GoogleProtobufAny.Any
  required?: boolean
}

export type PolicyTargets = {
  kinds?: string[]
  labels?: PolicyTargetLabel[]
  namespaces?: string[]
}

export type PolicyTargetLabel = {
  values?: {[key: string]: string}
}

export class Core {
  static GetObject(req: GetObjectRequest, initReq?: fm.InitReq): Promise<GetObjectResponse> {
    return fm.fetchReq<GetObjectRequest, GetObjectResponse>(`/v1/object/${req["name"]}?${fm.renderURLSearchParams(req, ["name"])}`, {...initReq, method: "GET"})
  }
  static ListObjects(req: ListObjectsRequest, initReq?: fm.InitReq): Promise<ListObjectsResponse> {
    return fm.fetchReq<ListObjectsRequest, ListObjectsResponse>(`/v1/objects`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ListFluxRuntimeObjects(req: ListFluxRuntimeObjectsRequest, initReq?: fm.InitReq): Promise<ListFluxRuntimeObjectsResponse> {
    return fm.fetchReq<ListFluxRuntimeObjectsRequest, ListFluxRuntimeObjectsResponse>(`/v1/flux_runtime_objects?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static ListFluxCrds(req: ListFluxCrdsRequest, initReq?: fm.InitReq): Promise<ListFluxCrdsResponse> {
    return fm.fetchReq<ListFluxCrdsRequest, ListFluxCrdsResponse>(`/v1/flux_crds?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static GetReconciledObjects(req: GetReconciledObjectsRequest, initReq?: fm.InitReq): Promise<GetReconciledObjectsResponse> {
    return fm.fetchReq<GetReconciledObjectsRequest, GetReconciledObjectsResponse>(`/v1/reconciled_objects`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static GetChildObjects(req: GetChildObjectsRequest, initReq?: fm.InitReq): Promise<GetChildObjectsResponse> {
    return fm.fetchReq<GetChildObjectsRequest, GetChildObjectsResponse>(`/v1/child_objects`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static GetFluxNamespace(req: GetFluxNamespaceRequest, initReq?: fm.InitReq): Promise<GetFluxNamespaceResponse> {
    return fm.fetchReq<GetFluxNamespaceRequest, GetFluxNamespaceResponse>(`/v1/namespace/flux`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ListNamespaces(req: ListNamespacesRequest, initReq?: fm.InitReq): Promise<ListNamespacesResponse> {
    return fm.fetchReq<ListNamespacesRequest, ListNamespacesResponse>(`/v1/namespaces?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static ListEvents(req: ListEventsRequest, initReq?: fm.InitReq): Promise<ListEventsResponse> {
    return fm.fetchReq<ListEventsRequest, ListEventsResponse>(`/v1/events?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static SyncFluxObject(req: SyncFluxObjectRequest, initReq?: fm.InitReq): Promise<SyncFluxObjectResponse> {
    return fm.fetchReq<SyncFluxObjectRequest, SyncFluxObjectResponse>(`/v1/sync`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static GetVersion(req: GetVersionRequest, initReq?: fm.InitReq): Promise<GetVersionResponse> {
    return fm.fetchReq<GetVersionRequest, GetVersionResponse>(`/v1/version?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static GetFeatureFlags(req: GetFeatureFlagsRequest, initReq?: fm.InitReq): Promise<GetFeatureFlagsResponse> {
    return fm.fetchReq<GetFeatureFlagsRequest, GetFeatureFlagsResponse>(`/v1/featureflags?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static ToggleSuspendResource(req: ToggleSuspendResourceRequest, initReq?: fm.InitReq): Promise<ToggleSuspendResourceResponse> {
    return fm.fetchReq<ToggleSuspendResourceRequest, ToggleSuspendResourceResponse>(`/v1/suspend`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static GetSessionLogs(req: GetSessionLogsRequest, initReq?: fm.InitReq): Promise<GetSessionLogsResponse> {
    return fm.fetchReq<GetSessionLogsRequest, GetSessionLogsResponse>(`/v1/session_logs`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static IsCRDAvailable(req: IsCRDAvailableRequest, initReq?: fm.InitReq): Promise<IsCRDAvailableResponse> {
    return fm.fetchReq<IsCRDAvailableRequest, IsCRDAvailableResponse>(`/v1/crd/is_available?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static GetInventory(req: GetInventoryRequest, initReq?: fm.InitReq): Promise<GetInventoryResponse> {
    return fm.fetchReq<GetInventoryRequest, GetInventoryResponse>(`/v1/inventory?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static ListPolicies(req: ListPoliciesRequest, initReq?: fm.InitReq): Promise<ListPoliciesResponse> {
    return fm.fetchReq<ListPoliciesRequest, ListPoliciesResponse>(`/v1/policies?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static GetPolicy(req: GetPolicyRequest, initReq?: fm.InitReq): Promise<GetPolicyResponse> {
    return fm.fetchReq<GetPolicyRequest, GetPolicyResponse>(`/v1/policies/${req["policyName"]}?${fm.renderURLSearchParams(req, ["policyName"])}`, {...initReq, method: "GET"})
  }
  static ListPolicyValidations(req: ListPolicyValidationsRequest, initReq?: fm.InitReq): Promise<ListPolicyValidationsResponse> {
    return fm.fetchReq<ListPolicyValidationsRequest, ListPolicyValidationsResponse>(`/v1/policyvalidations`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static GetPolicyValidation(req: GetPolicyValidationRequest, initReq?: fm.InitReq): Promise<GetPolicyValidationResponse> {
    return fm.fetchReq<GetPolicyValidationRequest, GetPolicyValidationResponse>(`/v1/policyvalidations/${req["validationId"]}?${fm.renderURLSearchParams(req, ["validationId"])}`, {...initReq, method: "GET"})
  }
}