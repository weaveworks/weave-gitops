/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "../applications/fetch.pb"
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

export class Core {
  static ListKustomizations(req: ListKustomizationsRequest, initReq?: fm.InitReq): Promise<ListKustomizationsResponse> {
    return fm.fetchReq<ListKustomizationsRequest, ListKustomizationsResponse>(`/v1/kustomizations?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static ListHelmReleases(req: ListHelmReleasesRequest, initReq?: fm.InitReq): Promise<ListHelmReleasesResponse> {
    return fm.fetchReq<ListHelmReleasesRequest, ListHelmReleasesResponse>(`/v1/helmreleases?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
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
}
