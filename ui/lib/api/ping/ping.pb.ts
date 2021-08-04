/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "../applications/fetch.pb"
export type Empty = {
}

export type PingRequest = {
  value?: string
  sleepTimeMs?: number
  errorCodeReturned?: number
}

export type PingResponse = {
  value?: string
  counter?: number
}

export class TestService {
  static PingEmpty(req: Empty, initReq?: fm.InitReq): Promise<PingResponse> {
    return fm.fetchReq<Empty, PingResponse>(`/v1/ping/empty?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static Ping(req: PingRequest, initReq?: fm.InitReq): Promise<PingResponse> {
    return fm.fetchReq<PingRequest, PingResponse>(`/v1/ping`, {...initReq, method: "POST"})
  }
  static PingError(req: PingRequest, initReq?: fm.InitReq): Promise<Empty> {
    return fm.fetchReq<PingRequest, Empty>(`/v1/ping/error`, {...initReq, method: "POST"})
  }
  static PingList(req: PingRequest, entityNotifier?: fm.NotifyStreamEntityArrival<PingResponse>, initReq?: fm.InitReq): Promise<void> {
    return fm.fetchStreamingRequest<PingRequest, PingResponse>(`/v1/ping?${fm.renderURLSearchParams(req, [])}`, entityNotifier, {...initReq, method: "GET"})
  }
}