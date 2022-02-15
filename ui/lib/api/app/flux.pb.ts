/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as Gitops_serverV1Source from "./source.pb"
export type Kustomization = {
  namespace?: string
  name?: string
  path?: string
  sourceRef?: Gitops_serverV1Source.SourceRef
  interval?: Gitops_serverV1Source.Interval
  conditions?: Gitops_serverV1Source.Condition[]
  lastAppliedRevision?: string
  lastAttemptedRevision?: string
  lastHandledReconciledAt?: string
}

export type AddKustomizationReq = {
  namespace?: string
  appName?: string
  name?: string
  path?: string
  sourceRef?: Gitops_serverV1Source.SourceRef
  interval?: Gitops_serverV1Source.Interval
}

export type AddKustomizationRes = {
  success?: boolean
  kustomization?: Kustomization
}

export type ListKustomizationsReq = {
  namespace?: string
  appName?: string
}

export type ListKustomizationsRes = {
  kustomizations?: Kustomization[]
}

export type RemoveKustomizationReq = {
  namespace?: string
  appName?: string
  kustomizationName?: string
}

export type RemoveKustomizationRes = {
  success?: boolean
}