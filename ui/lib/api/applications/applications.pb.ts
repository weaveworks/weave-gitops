/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "./fetch.pb"
export type Application = {
  name?: string
  path?: string
  url?: string
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
}