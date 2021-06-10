/* eslint-disable */
// @ts-nocheck
/*
 * This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
 */

import * as fm from "../../fetch.pb";
export type Application = {
  name?: string;
};

export type ListApplicationsReq = {};

export type ListApplicationsRes = {
  applications?: Application[];
};

export class Applications {
  static ListApplications(
    req: ListApplicationsReq,
    initReq?: fm.InitReq
  ): Promise<ListApplicationsRes> {
    return fm.fetchReq<ListApplicationsReq, ListApplicationsRes>(
      `/api/v1/applications?${fm.renderURLSearchParams(req, [])}`,
      { ...initReq, method: "GET" }
    );
  }
}
