import _ from "lodash";
import { toast } from "react-toastify";
import { computeReady, ReadyType } from "../components/KubeStatusIndicator";
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
export function gitlabOAuthRedirectURI(): string {
  return `${window.location.origin}${PageRoute.GitlabOAuthCallback}`;
}

export function poller(cb, interval): any {
  if (process.env.NODE_ENV === "test") {
    // Stay synchronous in tests
    return cb();
  }

  return setInterval(cb, interval);
}

// isHTTP checks if something looks link-like enough that we think it
// would be a good idea to auto-link. This is quite strict, and does
// not allow e.g relative links. See also isAllowedLink
export function isHTTP(uri: string): boolean {
  // Regex from Diego Perini's gist: https://gist.github.com/dperini/729294
  // It works better than other regular expressions for validating HTTP and HTTPS URLs.
  const regex = new RegExp(
    /^(?:(?:(?:https?):)?\/\/)(?:\S+(?::\S*)?@)?(?:(?!(?:10|127)(?:\.\d{1,3}){3})(?!(?:169\.254|192\.168)(?:\.\d{1,3}){2})(?!172\.(?:1[6-9]|2\d|3[0-1])(?:\.\d{1,3}){2})(?:[1-9]\d?|1\d\d|2[01]\d|22[0-3])(?:\.(?:1?\d{1,2}|2[0-4]\d|25[0-5])){2}(?:\.(?:[1-9]\d?|1\d\d|2[0-4]\d|25[0-4]))|(?:(?:[a-z0-9\u00a1-\uffff][a-z0-9\u00a1-\uffff_-]{0,62})?[a-z0-9\u00a1-\uffff]\.)+(?:[a-z\u00a1-\uffff]{2,}\.?))(?::\d{2,5})?(?:[/?#]\S*)?$/i
  );

  return regex.test(uri);
}

// isAllowedLink checks if making a "link" clickable will do anybody
// any good. This is quite permissive - it's to stop e.g. oci:// links
// being clickable, because clicking them will make nobody happy. See
// also isAllowedLink
export function isAllowedLink(uri: string): boolean {
  // Regex from https://github.com/cure53/DOMPurify/blob/cce00ac40d33c2aae6422eaa59e6a8aad5c73901/src/regexp.js
  const regex = new RegExp(
    /^(?:(?:(?:f|ht)tps?|mailto|tel|callto|cid|xmpp):|[^a-z]|[a-z+.\-]+(?:[^a-z+.\-:]|$))/i // eslint-disable-line no-useless-escape
  );
  return regex.test(uri);
}

export function convertGitURLToGitProvider(uri: string): string {
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

export function pageTitleWithAppName(title: string, appName?: string): string {
  return `${title}${appName ? ` for ${appName}` : ""}`;
}

interface Statusable {
  conditions: Condition[];
  suspended: boolean;
}

export function statusSortHelper({
  suspended,
  conditions,
}: Statusable): number {
  if (suspended) return 2;
  if (computeReady(conditions) === ReadyType.Reconciling) return 3;
  else if (computeReady(conditions)) return 4;
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

export function makeImageString(images: string[]): string {
  let imageString = "";
  if (!images[0]) return "-";
  else imageString += images[0];
  if (images[1]) {
    for (let i = 1; i < images.length; i++) imageString += `\n${images[i]}`;
  }
  return imageString;
}

export function formatMetadataKey(key: string): string {
  return key
    .replace(/-/g, " ")
    .replace(/\w+/g, (w) => w[0].toUpperCase() + w.slice(1));
}

export const convertImage = (image: string) => {
  const split = image.split("/");

  //remove tags
  const tag = split[split.length - 1];
  if (tag.includes(":"))
    split[split.length - 1] = tag.slice(0, tag.indexOf(":"));

  const prefix = split.shift();
  const noTag = split.join("/");
  let url = "";

  //Github GHCR or Google GCR
  if (prefix === "ghcr.io" || prefix === "gcr.io")
    return `https://${prefix}/${noTag}`;
  //Quay.io
  if (prefix === "quay.io") {
    return `https://quay.io/repository/${noTag}`;
  }
  //complex docker prefix case
  if (prefix === "docker.io") {
    url = "https://hub.docker.com/r/";
    //library alias
    if (split[0] === "library") return url + "_/" + split[1];
    //global
    if (!split[1]) return url + "_/" + split[0];
    //namespaced
    return url + noTag;
  }
  //docker without prefix
  if (prefix === "library") return "https://hub.docker.com/r/_/" + split[0];
  //this one's at risk if we have to add others - global docker images can just be one word apparently
  if (!split[0]) {
    return "https://hub.docker.com/r/_/" + prefix;
  }
  //any other url
  if (prefix.includes(".")) return false;
  //one slash docker images w/o docker.io
  return `https://hub.docker.com/r/${prefix}/${noTag}`;
};

export function calculateZoomRatio(zoomPercent: number): number {
  return (zoomPercent + 20) / 1500;
}

export function calculateNodeOffsetX(
  rootNode: any,
  zoomPercent: number,
  zoomRatio: number
): number {
  if (!rootNode) {
    return 0;
  }

  return zoomPercent * 1.25 + (rootNode.width - rootNode.x) * zoomRatio;
}

export function mapScaleToZoomPercent(scale: number): number {
  return scale * 0.5;
}

export function mapZoomPercentToScale(zoomPercent: number): number {
  return Math.round(zoomPercent * 2);
}
