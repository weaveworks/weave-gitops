/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "../applications/fetch.pb"
import * as Gitops_serverV1Source from "./source.pb"
export type Kustomization = {
  namespace?: string
  name?: string
  path?: string
  sourceRef?: Gitops_serverV1Source.SourceRef
  interval?: Gitops_serverV1Source.Interval
}

export type AddKustomizationReq = {
  namespace?: string
  appName?: string
  name?: string
  path?: string
  sourceRef?: Gitops_serverV1Source.SourceRef
  interval?: Gitops_serverV1Source.Interval
}

export type AddKustomizationRes = {
  success?: boolean
  kustomization?: Kustomization
}

export type ListKustomizationsReq = {
  namespace?: string
  appName?: string
}

export type ListKustomizationsRes = {
  kustomizations?: Kustomization[]
}

export type RemoveKustomizationReq = {
  namespace?: string
  appName?: string
  kustomizationName?: string
}

export type RemoveKustomizationRes = {
  success?: boolean
}

export class Flux {
  static AddKustomization(req: AddKustomizationReq, initReq?: fm.InitReq): Promise<AddKustomizationRes> {
    return fm.fetchReq<AddKustomizationReq, AddKustomizationRes>(`/v1/namespace/${req["namespace"]}/kustomization`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ListKustomizations(req: ListKustomizationsReq, initReq?: fm.InitReq): Promise<ListKustomizationsRes> {
    return fm.fetchReq<ListKustomizationsReq, ListKustomizationsRes>(`/v1/namespace/${req["namespace"]}/kustomization?${fm.renderURLSearchParams(req, ["namespace"])}`, {...initReq, method: "GET"})
  }
  static RemoveKustomization(req: RemoveKustomizationReq, initReq?: fm.InitReq): Promise<RemoveKustomizationRes> {
    return fm.fetchReq<RemoveKustomizationReq, RemoveKustomizationRes>(`/v1/namespace/${req["namespace"]}/kustomization/${req["kustomizationName"]}`, {...initReq, method: "DELETE"})
  }
  static AddGitRepository(req: Gitops_serverV1Source.AddGitRepositoryReq, initReq?: fm.InitReq): Promise<Gitops_serverV1Source.AddGitRepositoryRes> {
    return fm.fetchReq<Gitops_serverV1Source.AddGitRepositoryReq, Gitops_serverV1Source.AddGitRepositoryRes>(`/v1/namespace/${req["namespace"]}/gitrepository`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ListGitRepositories(req: Gitops_serverV1Source.ListGitRepositoryReq, initReq?: fm.InitReq): Promise<Gitops_serverV1Source.ListGitRepositoryRes> {
    return fm.fetchReq<Gitops_serverV1Source.ListGitRepositoryReq, Gitops_serverV1Source.ListGitRepositoryRes>(`/v1/namespace/${req["namespace"]}/gitrepository?${fm.renderURLSearchParams(req, ["namespace"])}`, {...initReq, method: "GET"})
  }
}