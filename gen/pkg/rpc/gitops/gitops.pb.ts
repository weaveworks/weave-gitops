/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "../../../fetch.pb"

export enum DeploymentType {
  kustomize = "kustomize",
  helm = "helm",
}

export type LoginReq = {
  state?: string
}

export type LoginRes = {
  redirectUrl?: string
}

export type Application = {
  name?: string
  type?: DeploymentType
}

export type AddApplicationReq = {
  owner?: string
  name?: string
  url?: string
  path?: string
  branch?: string
  deploymentType?: DeploymentType
  privateKey?: string
  dryRun?: boolean
  private?: boolean
  namespace?: string
  dir?: string
}

export type AddApplicationRes = {
  application?: Application
}

export class GitOps {
  static Login(req: LoginReq, initReq?: fm.InitReq): Promise<LoginRes> {
    return fm.fetchReq<LoginReq, LoginRes>(`/gitops.GitOps/Login`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static AddApplication(req: AddApplicationReq, initReq?: fm.InitReq): Promise<AddApplicationRes> {
    return fm.fetchReq<AddApplicationReq, AddApplicationRes>(`/gitops.GitOps/AddApplication`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
}