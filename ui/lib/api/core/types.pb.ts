/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

export enum FluxObjectKind {
  KindGitRepository = "KindGitRepository",
  KindBucket = "KindBucket",
  KindHelmRepository = "KindHelmRepository",
  KindHelmChart = "KindHelmChart",
  KindKustomization = "KindKustomization",
  KindHelmRelease = "KindHelmRelease",
  KindCluster = "KindCluster",
  KindOCIRepository = "KindOCIRepository",
}

export enum HelmRepositoryType {
  Default = "Default",
  OCI = "OCI",
}

export type Interval = {
  hours?: string
  minutes?: string
  seconds?: string
}

export type FluxObjectRef = {
  kind?: FluxObjectKind
  name?: string
  namespace?: string
}

export type ObjectRef = {
  kind?: string
  name?: string
  namespace?: string
}

export type Condition = {
  type?: string
  status?: string
  reason?: string
  message?: string
  timestamp?: string
}

export type GitRepositoryRef = {
  branch?: string
  tag?: string
  semver?: string
  commit?: string
}

export type GroupVersionKind = {
  group?: string
  kind?: string
  version?: string
}

export type Kustomization = {
  namespace?: string
  name?: string
  path?: string
  sourceRef?: FluxObjectRef
  interval?: Interval
  conditions?: Condition[]
  lastAppliedRevision?: string
  lastAttemptedRevision?: string
  inventory?: GroupVersionKind[]
  suspended?: boolean
  clusterName?: string
  apiVersion?: string
  tenant?: string
  uid?: string
}

export type HelmChart = {
  namespace?: string
  name?: string
  sourceRef?: FluxObjectRef
  chart?: string
  version?: string
  interval?: Interval
  conditions?: Condition[]
  suspended?: boolean
  lastUpdatedAt?: string
  clusterName?: string
  apiVersion?: string
  tenant?: string
  uid?: string
}

export type HelmRelease = {
  releaseName?: string
  namespace?: string
  name?: string
  interval?: Interval
  helmChart?: HelmChart
  conditions?: Condition[]
  inventory?: GroupVersionKind[]
  suspended?: boolean
  clusterName?: string
  helmChartName?: string
  lastAppliedRevision?: string
  lastAttemptedRevision?: string
  apiVersion?: string
  tenant?: string
  uid?: string
}

export type Object = {
  payload?: string
  clusterName?: string
  tenant?: string
  uid?: string
}

export type Deployment = {
  name?: string
  namespace?: string
  conditions?: Condition[]
  images?: string[]
  suspended?: boolean
  clusterName?: string
  uid?: string
}

export type CrdName = {
  plural?: string
  group?: string
}

export type Crd = {
  name?: CrdName
  version?: string
  kind?: string
  clusterName?: string
  uid?: string
}

export type UnstructuredObject = {
  groupVersionKind?: GroupVersionKind
  name?: string
  namespace?: string
  uid?: string
  status?: string
  conditions?: Condition[]
  suspended?: boolean
  clusterName?: string
  images?: string[]
}

export type Namespace = {
  name?: string
  status?: string
  annotations?: {[key: string]: string}
  labels?: {[key: string]: string}
  clusterName?: string
}

export type Event = {
  type?: string
  reason?: string
  message?: string
  timestamp?: string
  component?: string
  host?: string
  name?: string
  uid?: string
}

export type SuspendReqObj = {
  kind?: FluxObjectKind
  name?: string
  namespace?: string
  clusterName?: string
}