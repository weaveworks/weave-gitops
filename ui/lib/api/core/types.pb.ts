/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

export enum Kind {
  GitRepository = "GitRepository",
  Bucket = "Bucket",
  HelmRepository = "HelmRepository",
  HelmChart = "HelmChart",
  Kustomization = "Kustomization",
  HelmRelease = "HelmRelease",
  Cluster = "Cluster",
  OCIRepository = "OCIRepository",
  Provider = "Provider",
  Alert = "Alert",
  ImageRepository = "ImageRepository",
  ImageUpdateAutomation = "ImageUpdateAutomation",
  ImagePolicy = "ImagePolicy",
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

export type ObjectRef = {
  kind?: string
  name?: string
  namespace?: string
  clusterName?: string
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

export type NamespacedObjectReference = {
  name?: string
  namespace?: string
}

export type Object = {
  payload?: string
  clusterName?: string
  tenant?: string
  uid?: string
  inventory?: GroupVersionKind[]
}

export type Deployment = {
  name?: string
  namespace?: string
  conditions?: Condition[]
  images?: string[]
  suspended?: boolean
  clusterName?: string
  uid?: string
  labels?: {[key: string]: string}
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