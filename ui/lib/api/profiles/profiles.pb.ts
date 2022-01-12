/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as GoogleApiHttpbody from "../../google/api/httpbody.pb"
import * as fm from "../applications/fetch.pb"
export type Maintainer = {
  name?: string
  email?: string
  url?: string
}

export type HelmRepository = {
  name?: string
  namespace?: string
}

export type Profile = {
  name?: string
  home?: string
  sources?: string[]
  description?: string
  keywords?: string[]
  maintainers?: Maintainer[]
  icon?: string
  annotations?: {[key: string]: string}
  kubeVersion?: string
  helmRepository?: HelmRepository
  availableVersions?: string[]
  layer?: string
}

export type GetProfilesRequest = {
}

export type GetProfilesResponse = {
  profiles?: Profile[]
}

export type GetProfileValuesRequest = {
  profileName?: string
  profileVersion?: string
}

export type GetProfileValuesResponse = {
  values?: string
}

export type ProfileValues = {
  name?: string
  version?: string
  values?: string
}

export class Profiles {
  static GetProfiles(req: GetProfilesRequest, initReq?: fm.InitReq): Promise<GetProfilesResponse> {
    return fm.fetchReq<GetProfilesRequest, GetProfilesResponse>(`/v1/profiles?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static GetProfileValues(req: GetProfileValuesRequest, initReq?: fm.InitReq): Promise<GoogleApiHttpbody.HttpBody> {
    return fm.fetchReq<GetProfileValuesRequest, GoogleApiHttpbody.HttpBody>(`/v1/profiles/${req["profileName"]}/${req["profileVersion"]}/values?${fm.renderURLSearchParams(req, ["profileName", "profileVersion"])}`, {...initReq, method: "GET"})
  }
}