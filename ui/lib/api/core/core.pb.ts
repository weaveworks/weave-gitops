/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "../../fetch.pb"
import * as Gitops_coreV1Types from "./types.pb"
export type ListKustomizationsRequest = {
  namespace?: string
}

export type ListKustomizationsResponse = {
  kustomizations?: Gitops_coreV1Types.Kustomization[]
}

export type ListHelmReleasesRequest = {
  namespace?: string
}

export type ListHelmReleasesResponse = {
  helmReleases?: Gitops_coreV1Types.HelmRelease[]
}

export type GetHelmReleaseRequest = {
  name?: string
  namespace?: string
  clusterName?: string
}

export type GetHelmReleaseResponse = {
  helmRelease?: Gitops_coreV1Types.HelmRelease
}

export type ListGitRepositoriesRequest = {
  namespace?: string
}

export type ListGitRepositoriesResponse = {
  gitRepositories?: Gitops_coreV1Types.GitRepository[]
}

export type ListHelmRepositoriesRequest = {
  namespace?: string
}

export type ListHelmRepositoriesResponse = {
  helmRepositories?: Gitops_coreV1Types.HelmRepository[]
}

export type ListBucketRequest = {
  namespace?: string
}

export type ListBucketsResponse = {
  buckets?: Gitops_coreV1Types.Bucket[]
}

export type ListFluxRuntimeObjectsRequest = {
  namespace?: string
  clusterName?: string
}

export type ListFluxRuntimeObjectsResponse = {
  deployments?: Gitops_coreV1Types.Deployment[]
}

export type ListHelmChartsRequest = {
  namespace?: string
}

export type ListHelmChartsResponse = {
  helmCharts?: Gitops_coreV1Types.HelmChart[]
}

export type GetKustomizationRequest = {
  name?: string
  namespace?: string
  clusterName?: string
}

export type GetKustomizationResponse = {
  kustomization?: Gitops_coreV1Types.Kustomization
}

export type GetReconciledObjectsRequest = {
  automationName?: string
  namespace?: string
  automationKind?: Gitops_coreV1Types.AutomationKind
  kinds?: Gitops_coreV1Types.GroupVersionKind[]
  clusterName?: string
}

export type GetReconciledObjectsResponse = {
  objects?: Gitops_coreV1Types.UnstructuredObject[]
}

export type GetChildObjectsRequest = {
  groupVersionKind?: Gitops_coreV1Types.GroupVersionKind
  namespace?: string
  parentUid?: string
  clusterName?: string
}

export type GetChildObjectsResponse = {
  objects?: Gitops_coreV1Types.UnstructuredObject[]
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

export type ListFluxEventsRequest = {
  namespace?: string
  involvedObject?: Gitops_coreV1Types.ObjectReference
}

export type ListFluxEventsResponse = {
  events?: Gitops_coreV1Types.Event[]
}

export class Core {
  static ListKustomizations(req: ListKustomizationsRequest, initReq?: fm.InitReq): Promise<ListKustomizationsResponse> {
    return fm.fetchReq<ListKustomizationsRequest, ListKustomizationsResponse>(`/v1/kustomizations?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static GetKustomization(req: GetKustomizationRequest, initReq?: fm.InitReq): Promise<GetKustomizationResponse> {
    return fm.fetchReq<GetKustomizationRequest, GetKustomizationResponse>(`/v1/kustomizations/${req["name"]}?${fm.renderURLSearchParams(req, ["name"])}`, {...initReq, method: "GET"})
  }
  static ListHelmReleases(req: ListHelmReleasesRequest, initReq?: fm.InitReq): Promise<ListHelmReleasesResponse> {
    return fm.fetchReq<ListHelmReleasesRequest, ListHelmReleasesResponse>(`/v1/helmreleases?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static GetHelmRelease(req: GetHelmReleaseRequest, initReq?: fm.InitReq): Promise<GetHelmReleaseResponse> {
    return fm.fetchReq<GetHelmReleaseRequest, GetHelmReleaseResponse>(`/v1/helmrelease/${req["name"]}?${fm.renderURLSearchParams(req, ["name"])}`, {...initReq, method: "GET"})
  }
  static ListGitRepositories(req: ListGitRepositoriesRequest, initReq?: fm.InitReq): Promise<ListGitRepositoriesResponse> {
    return fm.fetchReq<ListGitRepositoriesRequest, ListGitRepositoriesResponse>(`/v1/gitrepositories?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static ListHelmCharts(req: ListHelmChartsRequest, initReq?: fm.InitReq): Promise<ListHelmChartsResponse> {
    return fm.fetchReq<ListHelmChartsRequest, ListHelmChartsResponse>(`/v1/helmcharts?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static ListHelmRepositories(req: ListHelmRepositoriesRequest, initReq?: fm.InitReq): Promise<ListHelmRepositoriesResponse> {
    return fm.fetchReq<ListHelmRepositoriesRequest, ListHelmRepositoriesResponse>(`/v1/helmrepositories?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static ListBuckets(req: ListBucketRequest, initReq?: fm.InitReq): Promise<ListBucketsResponse> {
    return fm.fetchReq<ListBucketRequest, ListBucketsResponse>(`/v1/buckets?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static ListFluxRuntimeObjects(req: ListFluxRuntimeObjectsRequest, initReq?: fm.InitReq): Promise<ListFluxRuntimeObjectsResponse> {
    return fm.fetchReq<ListFluxRuntimeObjectsRequest, ListFluxRuntimeObjectsResponse>(`/v1/flux_runtime_objects?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
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
  static ListFluxEvents(req: ListFluxEventsRequest, initReq?: fm.InitReq): Promise<ListFluxEventsResponse> {
    return fm.fetchReq<ListFluxEventsRequest, ListFluxEventsResponse>(`/v1/events?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
}