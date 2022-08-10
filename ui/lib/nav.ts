import _ from "lodash";
import qs from "query-string";
import { FluxObjectKind } from "./api/core/types.pb";
import { NoNamespace, PageRoute, V2Routes } from "./types";

// getParentNavValue returns the parent for a child page.
// This keeps the nav element highlighted if we are on a child page.
// Example: /sources and /git_repo will both show the "Sources" nav as selected.
export const getParentNavValue = (
  path: string
): PageRoute | V2Routes | boolean => {
  const [, currentPage] = _.split(path, "/");
  switch (`/${currentPage}`) {
    case V2Routes.Automations:
    case V2Routes.Kustomization:
    case V2Routes.HelmRelease:
      return V2Routes.Automations;

    case V2Routes.Sources:
    case V2Routes.GitRepo:
    case V2Routes.HelmChart:
    case V2Routes.HelmRepo:
    case V2Routes.Bucket:
    case V2Routes.OCIRepository:
      return V2Routes.Sources;

    case V2Routes.FluxRuntime:
      return V2Routes.FluxRuntime;

    default:
      // The "Tabs" component of material-ui wants a bool
      return false;
  }
};

const pageTitles = {
  [V2Routes.Automations]: "Applications",
  [V2Routes.Sources]: "Sources",
  [V2Routes.FluxRuntime]: "Flux Runtime",
};

export const getPageLabel = (route: V2Routes): string => {
  return pageTitles[route];
};

export const formatURL = (page: string, query: any = {}) => {
  return `${page}?${qs.stringify(query)}`;
};

export const formatSourceURL = (
  kind: FluxObjectKind,
  name: string,
  namespace: string = NoNamespace,
  clusterName: string
) => {
  return formatURL(objectTypeToRoute(kind), { name, namespace, clusterName });
};

export function objectTypeToRoute(t: FluxObjectKind): V2Routes {
  switch (t) {
    case FluxObjectKind.KindGitRepository:
      return V2Routes.GitRepo;

    case FluxObjectKind.KindBucket:
      return V2Routes.Bucket;

    case FluxObjectKind.KindHelmRepository:
      return V2Routes.HelmRepo;

    case FluxObjectKind.KindHelmChart:
      return V2Routes.HelmChart;

    case FluxObjectKind.KindKustomization:
      return V2Routes.Kustomization;

    case FluxObjectKind.KindHelmRelease:
      return V2Routes.HelmRelease;

    case FluxObjectKind.KindOCIRepository:
      return V2Routes.OCIRepository;

    default:
      break;
  }
}
