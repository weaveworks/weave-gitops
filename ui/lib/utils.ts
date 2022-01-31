import qs from "query-string";
import { toast } from "react-toastify";
import { PageRoute, V2Routes } from "./types";

export const formatURL = (page: string, query: any = {}) => {
  return `${page}?${qs.stringify(query)}`;
};

export const addKustomizationURL = (appName: string) =>
  `${V2Routes.AddKustomization}?${qs.stringify({ appName })}`;

export const getNavValue = (
  currentPage: any
): PageRoute | V2Routes | boolean => {
  switch (currentPage) {
    case "applications":
    case "application_list":
    case "application":
    case "application_detail":
      return V2Routes.ApplicationList;

    default:
      // The "Tabs" component of material-ui wants a bool
      return false;
  }
};

export function notifySuccess(message: string) {
  toast["success"](message);
}

export function notifyError(message: string) {
  toast["error"](`Error: ${message}`);
}

// Must be one of the valid URLs that we have already
// configured on the Gitlab backend for our Oauth app.
export function gitlabOAuthRedirectURI() {
  return `${window.location.origin}${PageRoute.GitlabOAuthCallback}`;
}

export function poller(cb, interval) {
  if (process.env.NODE_ENV === "test") {
    // Stay synchronous in tests
    return cb();
  }

  return setInterval(cb, interval);
}

export function isHTTP(uri) {
  return uri.includes("http") || uri.includes("https");
}

export function convertGitURLToGitProvider(uri: string) {
  if (isHTTP(uri)) {
    return uri;
  }

  const matches = uri.match(/git@(.*)[/|:](.*)\/(.*)/);
  if (!matches) {
    throw new Error(`could not parse url "${uri}"`);
  }
  const [, provider, org, repo] = matches;

  return `https://${provider}/${org}/${repo}`;
}
