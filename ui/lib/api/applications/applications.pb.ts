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

export enum AutomationKind {
  Kustomize = "Kustomize",
  Helm = "Helm",
}

export enum GitProvider {
  Unknown = "Unknown",
  GitHub = "GitHub",
  GitLab = "GitLab",
}

export type SyncApplicationRequest = {
  name?: string
  namespace?: string
}

export type SyncApplicationResponse = {
  success?: boolean
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
  automationKind?: AutomationKind
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

export type GetGithubDeviceCodeRequest = {
}

export type GetGithubDeviceCodeResponse = {
  userCode?: string
  deviceCode?: string
  validationURI?: string
  interval?: number
}

export type GetGithubAuthStatusRequest = {
  deviceCode?: string
}

export type GetGithubAuthStatusResponse = {
  accessToken?: string
  error?: string
}

export type ParseRepoURLRequest = {
  url?: string
}

export type ParseRepoURLResponse = {
  name?: string
  provider?: GitProvider
  owner?: string
}

export type GetGitlabAuthURLRequest = {
  redirectUri?: string
}

export type GetGitlabAuthURLResponse = {
  url?: string
}

export type AuthorizeGitlabRequest = {
  code?: string
  redirectUri?: string
}

export type AuthorizeGitlabResponse = {
  token?: string
}

export class Applications {
  static ListCommits(req: ListCommitsRequest, initReq?: fm.InitReq): Promise<ListCommitsResponse> {
    return fm.fetchReq<ListCommitsRequest, ListCommitsResponse>(`/v1/applications/${req["name"]}/commits?${fm.renderURLSearchParams(req, ["name"])}`, {...initReq, method: "GET"})
  }
  static GetReconciledObjects(req: GetReconciledObjectsReq, initReq?: fm.InitReq): Promise<GetReconciledObjectsRes> {
    return fm.fetchReq<GetReconciledObjectsReq, GetReconciledObjectsRes>(`/v1/applications/${req["automationName"]}/reconciled_objects`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static GetChildObjects(req: GetChildObjectsReq, initReq?: fm.InitReq): Promise<GetChildObjectsRes> {
    return fm.fetchReq<GetChildObjectsReq, GetChildObjectsRes>(`/v1/applications/child_objects`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static GetGithubDeviceCode(req: GetGithubDeviceCodeRequest, initReq?: fm.InitReq): Promise<GetGithubDeviceCodeResponse> {
    return fm.fetchReq<GetGithubDeviceCodeRequest, GetGithubDeviceCodeResponse>(`/v1/applications/auth_providers/github?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static GetGithubAuthStatus(req: GetGithubAuthStatusRequest, initReq?: fm.InitReq): Promise<GetGithubAuthStatusResponse> {
    return fm.fetchReq<GetGithubAuthStatusRequest, GetGithubAuthStatusResponse>(`/v1/applications/auth_providers/github/status`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static GetGitlabAuthURL(req: GetGitlabAuthURLRequest, initReq?: fm.InitReq): Promise<GetGitlabAuthURLResponse> {
    return fm.fetchReq<GetGitlabAuthURLRequest, GetGitlabAuthURLResponse>(`/v1/applications/auth_providers/gitlab?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static AuthorizeGitlab(req: AuthorizeGitlabRequest, initReq?: fm.InitReq): Promise<AuthorizeGitlabResponse> {
    return fm.fetchReq<AuthorizeGitlabRequest, AuthorizeGitlabResponse>(`/v1/applications/auth_providers/gitlab/authorize`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static SyncApplication(req: SyncApplicationRequest, initReq?: fm.InitReq): Promise<SyncApplicationResponse> {
    return fm.fetchReq<SyncApplicationRequest, SyncApplicationResponse>(`/v1/applications/${req["name"]}/sync`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ParseRepoURL(req: ParseRepoURLRequest, initReq?: fm.InitReq): Promise<ParseRepoURLResponse> {
    return fm.fetchReq<ParseRepoURLRequest, ParseRepoURLResponse>(`/v1/applications/parse_repo_url?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
}