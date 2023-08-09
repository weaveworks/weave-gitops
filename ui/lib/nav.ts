import _ from "lodash";
import qs from "query-string";
import { Kind } from "./api/core/types.pb";
import { NoNamespace, PageRoute, V2Routes } from "./types";

// getParentNavValue returns the parent for a child page.
// This keeps the nav element highlighted if we are on a child page.
// Example: /sources and /git_repo will both show the "Sources" nav as selected.
export const getParentNavValue = (
  path: string
): V2Routes | PageRoute | boolean => {
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

    case V2Routes.ImageAutomation:
      return V2Routes.ImageAutomation;

    case V2Routes.ImageAutomationUpdatesDetails:
      return V2Routes.ImageUpdates;

    case V2Routes.ImageAutomationRepositoryDetails:
      return V2Routes.ImageRepositories;

    case V2Routes.ImagePolicyDetails:
      return V2Routes.ImagePolicies;

    case V2Routes.Notifications:
    case V2Routes.Provider:
      return V2Routes.Notifications;

    case V2Routes.UserInfo:
      return V2Routes.UserInfo;

    case V2Routes.Policies:
    case V2Routes.PolicyDetailsPage:
      return V2Routes.Policies;

    case V2Routes.PolicyViolationDetails:
      return location.search.includes("kind=Policy")
        ? V2Routes.Policies
        : V2Routes.Automations;

    default:
      // The "Tabs" component of material-ui wants a bool
      return false;
  }
};

export const getParentNavRouteValue = (
  path: string
): V2Routes | PageRoute | boolean => {
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

    case V2Routes.ImageAutomation:
    case V2Routes.ImageAutomationUpdatesDetails:
    case V2Routes.ImageAutomationRepositoryDetails:
    case V2Routes.ImagePolicyDetails:
      return V2Routes.ImageAutomation;

    case V2Routes.Notifications:
    case V2Routes.Provider:
      return V2Routes.Notifications;

    case V2Routes.UserInfo:
      return V2Routes.UserInfo;

    case V2Routes.Policies:
    case V2Routes.PolicyDetailsPage:
      return V2Routes.Policies;

    case V2Routes.PolicyViolationDetails:
      return location.search.includes("kind=Policy") ||
        location.search.includes("kind=AllPoliciesViolations")
        ? V2Routes.Policies
        : V2Routes.Automations;

    default:
      // The "Tabs" component of material-ui wants a bool
      return false;
  }
};
const pageTitles = {
  [V2Routes.Automations]: "Applications",
  [V2Routes.Sources]: "Sources",
  [V2Routes.FluxRuntime]: "Flux Runtime",
  [V2Routes.Notifications]: "Notifications",
  [V2Routes.ImageAutomation]: "Image Automations",
  [V2Routes.ImagePolicies]: "Image Policies",
  [V2Routes.ImageUpdates]: "Image Updates",
  [V2Routes.ImageRepositories]: "Image Repositories",
  [V2Routes.PolicyViolationDetails]: "Violation Logs",
  [V2Routes.UserInfo]: "User Info",
  [V2Routes.Policies]: "Policies",
};

export const getPageLabel = (route: V2Routes): string => {
  return pageTitles[route];
};

export const formatURL = (page: string, query: any = {}) => {
  return `${page}?${qs.stringify(query)}`;
};

export const formatSourceURL = (
  kind: string,
  name: string,
  namespace: string = NoNamespace,
  clusterName: string
) => {
  return formatURL(objectTypeToRoute(Kind[kind]), {
    name,
    namespace,
    clusterName,
  });
};

export function objectTypeToRoute(t: Kind): V2Routes {
  switch (t) {
    case Kind.GitRepository:
      return V2Routes.GitRepo;

    case Kind.Bucket:
      return V2Routes.Bucket;

    case Kind.HelmRepository:
      return V2Routes.HelmRepo;

    case Kind.HelmChart:
      return V2Routes.HelmChart;

    case Kind.Kustomization:
      return V2Routes.Kustomization;

    case Kind.HelmRelease:
      return V2Routes.HelmRelease;

    case Kind.OCIRepository:
      return V2Routes.OCIRepository;

    case Kind.Provider:
      return V2Routes.Provider;

    default:
      break;
  }

  return "" as V2Routes;
}
