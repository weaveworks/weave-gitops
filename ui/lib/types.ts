import { Condition, Interval, SourceType } from "./api/app/source.pb";

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
  ApplicationList = "/application_list",
  NewApp = "/new_app",
  Application = "/application",
  AddKustomization = "/add_kustomization",
  AddHelmRelease = "/add_helm_release",
  AddSource = "/add_source",
  AddGitRepo = "/add_git_repo",
  AddHelmRepo = "/add_helm_repo",
  AddBucket = "/add_bucket",
  Kustomization = "/kustomization",
  HelmRelease = "/helm_release",
  HelmRepo = "/helm_repo",
  Source = "/source",
  AddAutomation = "/add_automation",
  KustomizationList = "/kustomization_list",
  GitRepo = "/git_repo",
  SourcesList = "/sources",
  FluxRuntime = "/flux_runtime",
}

export const WeGONamespace = "flux-system";

export interface Source {
  name?: string;
  namespace?: string;
  type?: SourceType;
  conditions?: Condition[];
  interval?: Interval;
}

export enum AutomationType {
  Kustomization = "Kustomization",
  HelmRelease = "HelmRelease",
}
