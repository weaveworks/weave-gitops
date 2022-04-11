import {
  AutomationKind,
  Condition,
  Interval,
  SourceRefSourceKind,
} from "./api/core/types.pb";

export enum PageRoute {
  Applications = "/applications",
  ApplicationDetail = "/application_detail",
  ApplicationAdd = "/application_add",
  ApplicationRemove = "/application_remove",
  GitlabOAuthCallback = "/oauth/gitlab",
}

export enum GrpcErrorCodes {
  Unknown = 2,
  Unauthenticated = 16,
  NotFound = 5,
}

export type RequestError = Error & {
  code?: number;
};

export enum V2Routes {
  Automations = "/applications",
  Sources = "/sources",
  FluxRuntime = "/flux_runtime",
  Kustomization = "/kustomization",
  HelmRelease = "/helm_release",
  HelmRepo = "/helm_repo",
  GitRepo = "/git_repo",
  HelmChart = "/helm_chart",
  Bucket = "/bucket",

  // Use this to allow for certain components to route to a 404 and still compile.
  // We want to keep certain components around for future use.
  NotImplemented = "/not_implemented",
}

export const WeGONamespace = "flux-system";
export const DefaultCluster = "Default";

export interface Source {
  name?: string;
  namespace?: string;
  type?: SourceRefSourceKind;
  conditions?: Condition[];
  interval?: Interval;
  suspended?: boolean;
  clusterName?: string;
}

export enum AutomationType {
  Kustomization = "Kustomization",
  HelmRelease = "HelmRelease",
}

export interface Syncable {
  name?: string;
  kind?: AutomationKind;
  namespace?: string;
  clusterName?: string;
}
