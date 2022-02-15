/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "../applications/fetch.pb"
import * as Gitops_serverV1Flux from "./flux.pb"
import * as Gitops_serverV1Source from "./source.pb"
export type App = {
  namespace?: string
  name?: string
  description?: string
  displayName?: string
}

export type AddAppRequest = {
  namespace?: string
  name?: string
  description?: string
  displayName?: string
}

export type AddAppResponse = {
  success?: boolean
  app?: App
}

export type GetAppRequest = {
  namespace?: string
  appName?: string
}

export type GetAppResponse = {
  app?: App
}

export type ListAppRequest = {
  namespace?: string
}

export type ListAppResponse = {
  apps?: App[]
}

export type RemoveAppRequest = {
  namespace?: string
  name?: string
  autoMerge?: boolean
}

export type RemoveAppResponse = {
  success?: boolean
}

export class Apps {
  static AddApp(req: AddAppRequest, initReq?: fm.InitReq): Promise<AddAppResponse> {
    return fm.fetchReq<AddAppRequest, AddAppResponse>(`/v1/namespace/${req["namespace"]}/app`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static GetApp(req: GetAppRequest, initReq?: fm.InitReq): Promise<GetAppResponse> {
    return fm.fetchReq<GetAppRequest, GetAppResponse>(`/v1/namespace/${req["namespace"]}/app/${req["appName"]}?${fm.renderURLSearchParams(req, ["namespace", "appName"])}`, {...initReq, method: "GET"})
  }
  static ListApps(req: ListAppRequest, initReq?: fm.InitReq): Promise<ListAppResponse> {
    return fm.fetchReq<ListAppRequest, ListAppResponse>(`/v1/namespace/${req["namespace"]}/app?${fm.renderURLSearchParams(req, ["namespace"])}`, {...initReq, method: "GET"})
  }
  static RemoveApp(req: RemoveAppRequest, initReq?: fm.InitReq): Promise<RemoveAppResponse> {
    return fm.fetchReq<RemoveAppRequest, RemoveAppResponse>(`/v1/namespace/${req["namespace"]}/app/${req["name"]}`, {...initReq, method: "DELETE"})
  }
  static AddKustomization(req: Gitops_serverV1Flux.AddKustomizationReq, initReq?: fm.InitReq): Promise<Gitops_serverV1Flux.AddKustomizationRes> {
    return fm.fetchReq<Gitops_serverV1Flux.AddKustomizationReq, Gitops_serverV1Flux.AddKustomizationRes>(`/v1/namespace/${req["namespace"]}/app/${req["appName"]}/kustomization`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ListKustomizations(req: Gitops_serverV1Flux.ListKustomizationsReq, initReq?: fm.InitReq): Promise<Gitops_serverV1Flux.ListKustomizationsRes> {
    return fm.fetchReq<Gitops_serverV1Flux.ListKustomizationsReq, Gitops_serverV1Flux.ListKustomizationsRes>(`/v1/kustomizations?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static RemoveKustomizations(req: Gitops_serverV1Flux.RemoveKustomizationReq, initReq?: fm.InitReq): Promise<Gitops_serverV1Flux.RemoveKustomizationRes> {
    return fm.fetchReq<Gitops_serverV1Flux.RemoveKustomizationReq, Gitops_serverV1Flux.RemoveKustomizationRes>(`/v1/namespace/${req["namespace"]}/app/${req["appName"]}/kustomization/${req["kustomizationName"]}`, {...initReq, method: "DELETE"})
  }
  static AddGitRepository(req: Gitops_serverV1Source.AddGitRepositoryReq, initReq?: fm.InitReq): Promise<Gitops_serverV1Source.AddGitRepositoryRes> {
    return fm.fetchReq<Gitops_serverV1Source.AddGitRepositoryReq, Gitops_serverV1Source.AddGitRepositoryRes>(`/v1/namespace/${req["namespace"]}/app/${req["appName"]}/gitrepository`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ListGitRepositories(req: Gitops_serverV1Source.ListGitRepositoryReq, initReq?: fm.InitReq): Promise<Gitops_serverV1Source.ListGitRepositoryRes> {
    return fm.fetchReq<Gitops_serverV1Source.ListGitRepositoryReq, Gitops_serverV1Source.ListGitRepositoryRes>(`/v1/namespace/${req["namespace"]}/app/${req["appName"]}/gitrepository?${fm.renderURLSearchParams(req, ["namespace", "appName"])}`, {...initReq, method: "GET"})
  }
  static AddHelmRepository(req: Gitops_serverV1Source.AddHelmRepositoryReq, initReq?: fm.InitReq): Promise<Gitops_serverV1Source.AddHelmRepositoryRes> {
    return fm.fetchReq<Gitops_serverV1Source.AddHelmRepositoryReq, Gitops_serverV1Source.AddHelmRepositoryRes>(`/v1/namespace/${req["namespace"]}/app/${req["appName"]}/helmrepository`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ListHelmRepositories(req: Gitops_serverV1Source.ListHelmRepositoryReq, initReq?: fm.InitReq): Promise<Gitops_serverV1Source.ListHelmRepositoryRes> {
    return fm.fetchReq<Gitops_serverV1Source.ListHelmRepositoryReq, Gitops_serverV1Source.ListHelmRepositoryRes>(`/v1/namespace/${req["namespace"]}/app/${req["appName"]}/helmrepository?${fm.renderURLSearchParams(req, ["namespace", "appName"])}`, {...initReq, method: "GET"})
  }
  static AddHelmChart(req: Gitops_serverV1Source.AddHelmChartReq, initReq?: fm.InitReq): Promise<Gitops_serverV1Source.AddHelmChartRes> {
    return fm.fetchReq<Gitops_serverV1Source.AddHelmChartReq, Gitops_serverV1Source.AddHelmChartRes>(`/v1/namespace/${req["namespace"]}/app/${req["appName"]}/helmchart`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ListHelmCharts(req: Gitops_serverV1Source.ListHelmChartReq, initReq?: fm.InitReq): Promise<Gitops_serverV1Source.ListHelmChartRes> {
    return fm.fetchReq<Gitops_serverV1Source.ListHelmChartReq, Gitops_serverV1Source.ListHelmChartRes>(`/v1/namespace/${req["namespace"]}/app/${req["appName"]}/helmchart?${fm.renderURLSearchParams(req, ["namespace", "appName"])}`, {...initReq, method: "GET"})
  }
  static AddBucket(req: Gitops_serverV1Source.AddBucketReq, initReq?: fm.InitReq): Promise<Gitops_serverV1Source.AddBucketRes> {
    return fm.fetchReq<Gitops_serverV1Source.AddBucketReq, Gitops_serverV1Source.AddBucketRes>(`/v1/namespace/${req["namespace"]}/app/${req["appName"]}/bucket`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ListBuckets(req: Gitops_serverV1Source.ListBucketReq, initReq?: fm.InitReq): Promise<Gitops_serverV1Source.ListBucketRes> {
    return fm.fetchReq<Gitops_serverV1Source.ListBucketReq, Gitops_serverV1Source.ListBucketRes>(`/v1/namespace/${req["namespace"]}/app/${req["appName"]}/bucket?${fm.renderURLSearchParams(req, ["namespace", "appName"])}`, {...initReq, method: "GET"})
  }
  static AddHelmRelease(req: Gitops_serverV1Source.AddHelmReleaseReq, initReq?: fm.InitReq): Promise<Gitops_serverV1Source.AddHelmReleaseRes> {
    return fm.fetchReq<Gitops_serverV1Source.AddHelmReleaseReq, Gitops_serverV1Source.AddHelmReleaseRes>(`/v1/namespace/${req["namespace"]}/app/${req["appName"]}/helmrelease`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ListHelmReleases(req: Gitops_serverV1Source.ListHelmReleaseReq, initReq?: fm.InitReq): Promise<Gitops_serverV1Source.ListHelmReleaseRes> {
    return fm.fetchReq<Gitops_serverV1Source.ListHelmReleaseReq, Gitops_serverV1Source.ListHelmReleaseRes>(`/v1/namespace/${req["namespace"]}/app/${req["appName"]}/helmrelease?${fm.renderURLSearchParams(req, ["namespace", "appName"])}`, {...initReq, method: "GET"})
  }
}