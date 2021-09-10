/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "./fetch.pb"

type Absent<T, K extends keyof T> = { [k in Exclude<keyof T, K>]?: undefined };
type OneOf<T> =
  | { [k in keyof T]?: undefined }
  | (
    keyof T extends infer K ?
      (K extends string & keyof T ? { [k in K]: T[K] } & Absent<T, K>
        : never)
    : never);

export enum GetReconciledObjectsReqAutomationKind {
  Kustomize = "Kustomize",
  Helm = "Helm",
}

export type Condition = {
  type?: string
  status?: string
  reason?: string
  message?: string
  timestamp?: number
}

export type Application = {
  name?: string
  path?: string
  url?: string
  sourceConditions?: Condition[]
  deploymentConditions?: Condition[]
  namespace?: string
  deploymentType?: GroupVersionKind
  reconciledObjectKinds?: GroupVersionKind[]
}

export type AuthenticateRequest = {
  providerName?: string
  accessToken?: string
}

export type AuthenticateResponse = {
  token?: string
}

export type ListApplicationsRequest = {
  namespace?: string
}

export type ListApplicationsResponse = {
  applications?: Application[]
}

export type GetApplicationRequest = {
  name?: string
  namespace?: string
}

export type GetApplicationResponse = {
  application?: Application
}

export type Commit = {
  hash?: string
  date?: string
  author?: string
  message?: string
  url?: string
}


type BaseListCommitsRequest = {
  name?: string
  namespace?: string
  pageSize?: number
}

export type ListCommitsRequest = BaseListCommitsRequest
  & OneOf<{ pageToken: number }>

export type ListCommitsResponse = {
  commits?: Commit[]
  nextPageToken?: number
}

export type GroupVersionKind = {
  group?: string
  kind?: string
  version?: string
}

export type UnstructuredObject = {
  groupVersionKind?: GroupVersionKind
  name?: string
  namespace?: string
  uid?: string
  status?: string
}

export type GetReconciledObjectsReq = {
  automationName?: string
  automationNamespace?: string
  automationKind?: GetReconciledObjectsReqAutomationKind
  kinds?: GroupVersionKind[]
}

export type GetReconciledObjectsRes = {
  objects?: UnstructuredObject[]
}

export type GetChildObjectsReq = {
  groupVersionKind?: GroupVersionKind
  parentUid?: string
}

export type GetChildObjectsRes = {
  objects?: UnstructuredObject[]
}

export class Applications {
  static Authenticate(req: AuthenticateRequest, initReq?: fm.InitReq): Promise<AuthenticateResponse> {
    return fm.fetchReq<AuthenticateRequest, AuthenticateResponse>(`/v1/authenticate/${req["providerName"]}`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ListApplications(req: ListApplicationsRequest, initReq?: fm.InitReq): Promise<ListApplicationsResponse> {
    return fm.fetchReq<ListApplicationsRequest, ListApplicationsResponse>(`/v1/applications?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static GetApplication(req: GetApplicationRequest, initReq?: fm.InitReq): Promise<GetApplicationResponse> {
    return fm.fetchReq<GetApplicationRequest, GetApplicationResponse>(`/v1/applications/${req["name"]}?${fm.renderURLSearchParams(req, ["name"])}`, {...initReq, method: "GET"})
  }
  static ListCommits(req: ListCommitsRequest, initReq?: fm.InitReq): Promise<ListCommitsResponse> {
    return fm.fetchReq<ListCommitsRequest, ListCommitsResponse>(`/v1/applications/${req["name"]}/commits?${fm.renderURLSearchParams(req, ["name"])}`, {...initReq, method: "GET"})
  }
  static GetReconciledObjects(req: GetReconciledObjectsReq, initReq?: fm.InitReq): Promise<GetReconciledObjectsRes> {
    return fm.fetchReq<GetReconciledObjectsReq, GetReconciledObjectsRes>(`/v1/applications/${req["automationName"]}/reconciled_objects`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static GetChildObjects(req: GetChildObjectsReq, initReq?: fm.InitReq): Promise<GetChildObjectsRes> {
    return fm.fetchReq<GetChildObjectsReq, GetChildObjectsRes>(`/v1/applications/child_objects`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
}