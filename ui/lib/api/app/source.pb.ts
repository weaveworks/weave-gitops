/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

export enum SourceRefKind {
  GitRepository = "GitRepository",
  Bucket = "Bucket",
  HelmRepository = "HelmRepository",
}

export enum SourceType {
  GitRepository = "GitRepository",
  Bucket = "Bucket",
  HelmRepository = "HelmRepository",
  HelmChart = "HelmChart",
}

export enum BucketProvider {
  Generic = "Generic",
  AWS = "AWS",
  GCP = "GCP",
}

export type Interval = {
  hours?: string
  minutes?: string
  seconds?: string
}

export type SourceRef = {
  kind?: SourceRefKind
  name?: string
}

export type Artifact = {
  checksum?: string
  lastupdateat?: number
  path?: string
  revision?: string
  url?: string
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

export type Source = {
  name?: string
  namespace?: string
  url?: string
  reference?: GitRepositoryRef
  type?: SourceType
  provider?: string
  bucketname?: string
  region?: string
  gitimplementation?: string
  timeout?: string
  secretRefName?: string
  conditions?: Condition[]
  artifact?: Artifact
  interval?: Interval
}

export type GitRepository = {
  namespace?: string
  name?: string
  url?: string
  reference?: GitRepositoryRef
  secretRef?: string
  interval?: Interval
  conditions?: Condition[]
}

export type AddGitRepositoryReq = {
  namespace?: string
  appName?: string
  name?: string
  url?: string
  reference?: GitRepositoryRef
  secretRef?: string
  interval?: Interval
}

export type AddGitRepositoryRes = {
  success?: boolean
  gitRepository?: GitRepository
}

export type ListGitRepositoryReq = {
  namespace?: string
  appName?: string
}

export type ListGitRepositoryRes = {
  gitRepositories?: GitRepository[]
}

export type HelmRepository = {
  namespace?: string
  name?: string
  url?: string
  interval?: Interval
  conditions?: Condition[]
}

export type AddHelmRepositoryReq = {
  namespace?: string
  appName?: string
  name?: string
  url?: string
  interval?: Interval
}

export type AddHelmRepositoryRes = {
  success?: boolean
  helmRepository?: HelmRepository
}

export type ListHelmRepositoryReq = {
  namespace?: string
  appName?: string
}

export type ListHelmRepositoryRes = {
  helmRepositories?: HelmRepository[]
}

export type HelmChart = {
  namespace?: string
  name?: string
  sourceRef?: SourceRef
  chart?: string
  version?: string
  interval?: Interval
  conditions?: Condition[]
}

export type AddHelmChartReq = {
  appName?: string
  namespace?: string
  helmChart?: HelmChart
}

export type AddHelmChartRes = {
  success?: boolean
  helmChart?: HelmChart
}

export type ListHelmChartReq = {
  namespace?: string
  appName?: string
}

export type ListHelmChartRes = {
  helmCharts?: HelmChart[]
}

export type Bucket = {
  namespace?: string
  name?: string
  endpoint?: string
  insecure?: boolean
  interval?: Interval
  provider?: BucketProvider
  region?: string
  secretRefName?: string
  timeout?: number
  conditions?: Condition[]
}

export type AddBucketReq = {
  appName?: string
  namespace?: string
  bucket?: Bucket
}

export type AddBucketRes = {
  success?: boolean
  bucket?: Bucket
}

export type ListBucketReq = {
  namespace?: string
  appName?: string
}

export type ListBucketRes = {
  buckets?: Bucket[]
}

export type HelmRelease = {
  releaseName?: string
  namespace?: string
  name?: string
  interval?: Interval
  helmChart?: HelmChart
  conditions?: Condition[]
}

export type AddHelmReleaseReq = {
  appName?: string
  namespace?: string
  helmRelease?: HelmRelease
}

export type AddHelmReleaseRes = {
  success?: boolean
  helmRelease?: HelmRelease
}

export type ListHelmReleaseReq = {
  namespace?: string
  appName?: string
}

export type ListHelmReleaseRes = {
  helmReleases?: HelmRelease[]
}