import { UseQueryOptions } from "react-query";
import { ListError } from "./api/core/core.pb";
import { Condition, FluxObjectKind, Interval } from "./api/core/types.pb";

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
  OCIRepository = "/oci",
  Notifications = "/notifications",

  // Use this to allow for certain components to route to a 404 and still compile.
  // We want to keep certain components around for future use.
  NotImplemented = "/not_implemented",
}

export const DefaultCluster = "Default";
export const NoNamespace = "";

export interface Source {
  name?: string;
  namespace?: string;
  kind?: FluxObjectKind;
  conditions?: Condition[];
  interval?: Interval;
  suspended?: boolean;
  clusterName?: string;
  lastUpdatedAt?: string;
}

export interface Syncable {
  name?: string;
  kind?: FluxObjectKind;
  namespace?: string;
  clusterName?: string;
}

export type MultiRequestError = ListError & {
  kind?: FluxObjectKind;
};

// Helper type to work around the weird react-query typedef/api shape
export type ReactQueryOptions<T, E> = Omit<
  UseQueryOptions<T, E>,
  "queryKey" | "queryFn"
>;
