/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/
export type GitOpsSet = {
  name?: string
  namespace?: string
  inventory?: ResourceRef[]
  conditions?: Condition[]
  generators?: string[]
  clusterName?: string
  type?: string
  labels?: {[key: string]: string}
  annotations?: {[key: string]: string}
  sourceRef?: SourceRef
  suspended?: boolean
}

export type SourceRef = {
  apiVersion?: string
  kind?: string
  name?: string
  namespace?: string
}

export type ListError = {
  namespace?: string
  message?: string
}

export type ResourceRef = {
  id?: string
  version?: string
}

export type Condition = {
  type?: string
  status?: string
  reason?: string
  message?: string
  timestamp?: string
}

export type GitOpsSetGenerator = {
  list?: string
  gitRepository?: string
}