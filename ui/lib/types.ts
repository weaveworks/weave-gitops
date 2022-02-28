import { Condition, Interval, SourceRefSourceKind } from "./api/core/types.pb";

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
  Automations = "/automations",
  Sources = "/sources",
  FluxRuntime = "/flux_runtime",
  Kustomization = "/kustomization",
  HelmRelease = "/helm_release",
  HelmRepo = "/helm_repo",
  GitRepo = "/git_repo",
  Bucket = "/bucket",

  // Use this to allow for certain components to route to a 404 and still compile.
  // We want to keep certain components around for future use.
  NotImplemented = "/not_implemented",
}

export const WeGONamespace = "flux-system";

export interface Source {
  name?: string;
  namespace?: string;
  type?: SourceRefSourceKind;
  conditions?: Condition[];
  interval?: Interval;
}

export enum AutomationType {
  Kustomization = "Kustomization",
  HelmRelease = "HelmRelease",
}
