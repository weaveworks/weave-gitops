/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as Wego_serverV1Commits from "./commits.pb"
import * as fm from "./fetch.pb"
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

export class Applications {
  static ListApplications(req: ListApplicationsRequest, initReq?: fm.InitReq): Promise<ListApplicationsResponse> {
    return fm.fetchReq<ListApplicationsRequest, ListApplicationsResponse>(`/v1/applications?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static GetApplication(req: GetApplicationRequest, initReq?: fm.InitReq): Promise<GetApplicationResponse> {
    return fm.fetchReq<GetApplicationRequest, GetApplicationResponse>(`/v1/applications/${req["name"]}?${fm.renderURLSearchParams(req, ["name"])}`, {...initReq, method: "GET"})
  }
  static ListCommits(req: Wego_serverV1Commits.ListCommitsRequest, initReq?: fm.InitReq): Promise<Wego_serverV1Commits.ListCommitsResponse> {
    return fm.fetchReq<Wego_serverV1Commits.ListCommitsRequest, Wego_serverV1Commits.ListCommitsResponse>(`/v1/commits?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
}