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

export enum SourceType {
  Git = "Git",
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
  deploymentType?: AutomationKind
  reconciledObjectKinds?: GroupVersionKind[]
  kustomization?: Kustomization
  helmRelease?: HelmRelease
  source?: Source
}

export type Kustomization = {
  name?: string
  namespace?: string
  targetNamespace?: string
  path?: string
  conditions?: Condition[]
  interval?: string
  prune?: boolean
  lastAppliedRevision?: string
}

export type HelmRelease = {
  name?: string
  namespace?: string
  targetNamespace?: string
  chart?: HelmChart
  interval?: string
  lastAppliedRevision?: string
  conditions?: Condition[]
}

export type HelmChart = {
  chart?: string
  version?: string
  valuesFiles?: string[]
}

export type Source = {
  name?: string
  url?: string
  type?: SourceType
  namespace?: string
  interval?: string
  reference?: string
  suspend?: boolean
  timeout?: string
  conditions?: Condition[]
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

export type AddApplicationRequest = {
  name?: string
  namespace?: string
  path?: string
  url?: string
  branch?: string
  autoMerge?: boolean
  configRepo?: string
}

export type AddApplicationResponse = {
  success?: boolean
  application?: Application
}

export type RemoveApplicationRequest = {
  name?: string
  namespace?: string
  autoMerge?: boolean
}

export type RemoveApplicationResponse = {
  success?: boolean
  repoUrl?: string
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

export type ValidateProviderTokenRequest = {
  provider?: GitProvider
}

export type ValidateProviderTokenResponse = {
  valid?: boolean
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
  static AddApplication(req: AddApplicationRequest, initReq?: fm.InitReq): Promise<AddApplicationResponse> {
    return fm.fetchReq<AddApplicationRequest, AddApplicationResponse>(`/v1/applications`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static RemoveApplication(req: RemoveApplicationRequest, initReq?: fm.InitReq): Promise<RemoveApplicationResponse> {
    return fm.fetchReq<RemoveApplicationRequest, RemoveApplicationResponse>(`/v1/applications/${req["name"]}`, {...initReq, method: "DELETE", body: JSON.stringify(req)})
  }
  static SyncApplication(req: SyncApplicationRequest, initReq?: fm.InitReq): Promise<SyncApplicationResponse> {
    return fm.fetchReq<SyncApplicationRequest, SyncApplicationResponse>(`/v1/applications/${req["name"]}/sync`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ParseRepoURL(req: ParseRepoURLRequest, initReq?: fm.InitReq): Promise<ParseRepoURLResponse> {
    return fm.fetchReq<ParseRepoURLRequest, ParseRepoURLResponse>(`/v1/applications/parse_repo_url?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static ValidateProviderToken(req: ValidateProviderTokenRequest, initReq?: fm.InitReq): Promise<ValidateProviderTokenResponse> {
    return fm.fetchReq<ValidateProviderTokenRequest, ValidateProviderTokenResponse>(`/v1/applications/validate_token`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
}