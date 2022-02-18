import qs from "query-string";
import { SourceRefSourceKind } from "./api/core/types.pb";
import { PageRoute } from "./types";

export enum V2Routes {
  Automations = "/automations",
  Sources = "/sources",
  FluxRuntime = "/flux_runtime",
  Kustomization = "/kustomization",
  HelmRelease = "/helm_release",
  HelmRepo = "/helm_repo",
  GitRepo = "/git_repo",
  HelmChart = "/helm_chart",
  Bucket = "/bucket",
}

// getParentNavValue returns the parent for a child page.
// This keeps the nav element highlighted if we are on a child page.
// Example: /sources and /git_repo will both show the "Sources" nav as selected.
export const getParentNavValue = (
  currentPage: any
): PageRoute | V2Routes | boolean => {
  switch (currentPage) {
    case V2Routes.Automations:
    case V2Routes.Kustomization:
    case V2Routes.HelmRelease:
      return V2Routes.Automations;

    case V2Routes.Sources:
    case V2Routes.GitRepo:
    case V2Routes.HelmChart:
    case V2Routes.Bucket:
      return V2Routes.Sources;

    case V2Routes.FluxRuntime:
      return V2Routes.FluxRuntime;

    default:
      // The "Tabs" component of material-ui wants a bool
      return false;
  }
};

export const formatURL = (page: string, query: any = {}) => {
  return `${page}?${qs.stringify(query)}`;
};

export function sourceTypeToRoute(t: SourceRefSourceKind): V2Routes {
  switch (t) {
    case SourceRefSourceKind.GitRepository:
      return V2Routes.GitRepo;

    default:
      break;
  }
}
