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

export type GetAuthenticationProvidersRequest = {
}

export type OauthProvider = {
  name?: string
  authUrl?: string
  redirectUri?: string
}

export type GetAuthenticationProvidersResponse = {
  providers?: OauthProvider[]
}

export type AuthenticateRequest = {
  providerName?: string
  code?: string
}

export type User = {
  email?: string
}

export type AuthenticateResponse = {
  token?: string
}

export type GetUserRequest = {
}

export type GetUserResponse = {
  user?: User
}

export class Applications {
  static ListApplications(req: ListApplicationsRequest, initReq?: fm.InitReq): Promise<ListApplicationsResponse> {
    return fm.fetchReq<ListApplicationsRequest, ListApplicationsResponse>(`/v1/applications?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static GetApplication(req: GetApplicationRequest, initReq?: fm.InitReq): Promise<GetApplicationResponse> {
    return fm.fetchReq<GetApplicationRequest, GetApplicationResponse>(`/v1/applications/${req["name"]}?${fm.renderURLSearchParams(req, ["name"])}`, {...initReq, method: "GET"})
  }
  static GetAuthenticationProviders(req: GetAuthenticationProvidersRequest, initReq?: fm.InitReq): Promise<GetAuthenticationProvidersResponse> {
    return fm.fetchReq<GetAuthenticationProvidersRequest, GetAuthenticationProvidersResponse>(`/v1/auth_providers?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static Authenticate(req: AuthenticateRequest, initReq?: fm.InitReq): Promise<AuthenticateResponse> {
    return fm.fetchReq<AuthenticateRequest, AuthenticateResponse>(`/v1/authenticate/${req["providerName"]}`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static GetUser(req: GetUserRequest, initReq?: fm.InitReq): Promise<GetUserResponse> {
    return fm.fetchReq<GetUserRequest, GetUserResponse>(`/v1/user?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
}