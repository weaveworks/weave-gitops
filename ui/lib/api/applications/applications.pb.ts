/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as GoogleLongrunningOperations from "../../google/longrunning/operations.pb"
import * as GoogleProtobufTimestamp from "../../google/protobuf/timestamp.pb"
import * as fm from "./fetch.pb"
export type Application = {
  name?: string
  path?: string
  url?: string
  annotations?: {[key: string]: string}
  createTime?: GoogleProtobufTimestamp.Timestamp
  updateTime?: GoogleProtobufTimestamp.Timestamp
  etag?: string
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

export type DeleteApplicationRequest = {
  applicationName?: string
  namespace?: string
  removeWorkload?: boolean
  etag?: string
}

export class Applications {
  static ListApplications(req: ListApplicationsRequest, initReq?: fm.InitReq): Promise<ListApplicationsResponse> {
    return fm.fetchReq<ListApplicationsRequest, ListApplicationsResponse>(`/v1/applications?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static GetApplication(req: GetApplicationRequest, initReq?: fm.InitReq): Promise<GetApplicationResponse> {
    return fm.fetchReq<GetApplicationRequest, GetApplicationResponse>(`/v1/applications/${req["name"]}?${fm.renderURLSearchParams(req, ["name"])}`, {...initReq, method: "GET"})
  }
  static DeleteApplication(req: DeleteApplicationRequest, initReq?: fm.InitReq): Promise<GoogleLongrunningOperations.Operation> {
    return fm.fetchReq<DeleteApplicationRequest, GoogleLongrunningOperations.Operation>(`/v1/applications/${req["applicationName"]}:remove`, {...initReq, method: "DELETE"})
  }
}