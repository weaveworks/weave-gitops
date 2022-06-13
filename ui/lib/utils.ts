import _ from "lodash";
import { toast } from "react-toastify";
import { computeReady } from "../components/KubeStatusIndicator";
import { Condition, HelmRelease, Kustomization } from "./api/core/types.pb";
import { PageRoute } from "./types";

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

export function pageTitleWithAppName(title: string, appName?: string) {
  return `${title}${appName ? ` for ${appName}` : ""}`;
}

interface Statusable {
  conditions: Condition[];
  suspended: boolean;
}

export function statusSortHelper({ suspended, conditions }: Statusable) {
  if (suspended) return 2;
  if (computeReady(conditions)) return 3;
  else return 1;
}

export function automationLastUpdated(a: Kustomization | HelmRelease): string {
  return _.get(_.find(a?.conditions, { type: "Ready" }), "timestamp");
}

const kindPrefix = "Kind";

export function addKind(kind: string): string {
  if (!kind.startsWith(kindPrefix)) {
    return `${kindPrefix}${kind}`;
  }
  return kind;
}

export function removeKind(kind: string): string {
  if (kind.startsWith(kindPrefix)) {
    return kind.slice(kindPrefix.length);
  }
  return kind;
}

export function makeImageString(images: string[]) {
  let imageString = "";
  if (!images[0]) return "-";
  else imageString += images[0];
  if (images[1]) {
    for (let i = 1; i < images.length; i++) imageString += `\n${images[i]}`;
  }
  return imageString;
}

export function formatMetadataKey(key: string) {
  return key;
}
